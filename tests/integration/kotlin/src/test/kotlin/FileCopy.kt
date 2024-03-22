import auth.AuthClient
import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobContainerClient
import com.azure.storage.blob.models.BlobListDetails
import com.azure.storage.blob.models.ListBlobsOptions
import org.joda.time.DateTime
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.BeforeGroups
import org.testng.annotations.Test
import tus.UploadClient
import util.Azure
import util.Constants
import util.EnvConfig
import util.Metadata
import util.TestFile
import java.time.Duration

@Test()
class FileCopy {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val edavBlobClient = Azure.getBlobServiceClient(EnvConfig.EDAV_STORAGE_ACCOUNT_NAME,
        ClientSecretCredentialBuilder()
            .clientId(EnvConfig.AZURE_CLIENT_ID)
            .clientSecret(EnvConfig.AZURE_CLIENT_SECRET)
            .tenantId(EnvConfig.AZURE_TENANT_ID)
            .build())
    private val routingBlobClient = Azure.getBlobServiceClient(EnvConfig.ROUTING_STORAGE_CONNECTION_STRING)
    private lateinit var bulkUploadsContainerClient: BlobContainerClient
    private lateinit var edavContainerClient: BlobContainerClient
    private lateinit var routingContainerClient: BlobContainerClient
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String

    @BeforeGroups(groups = [Constants.Groups.FILE_COPY])
    fun fileCopySetup(context: ITestContext) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val senderManifestPropertiesFilename = context.currentXmlTest.getParameter("SENDER_MANIFEST")
        val propertiesFilePath= "properties/$senderManifestPropertiesFilename"
        val metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)
        edavContainerClient = edavBlobClient.getBlobContainerClient(Constants.EDAV_UPLOAD_CONTAINER_NAME)
        routingContainerClient = routingBlobClient.getBlobContainerClient(Constants.ROUTING_UPLOAD_CONTAINER_NAME)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(500) // Hard delay to wait for file to copy.

        Assert.assertTrue(bulkUploadsContainerClient.exists())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldUploadFileToTusContainer() {
        val uploadBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId")
        val uploadInfoBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId.info")

        Assert.assertTrue(uploadBlob.exists())
        Assert.assertTrue(uploadInfoBlob.exists())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldHaveSameSizeFileInTusContainer() {
        val uploadBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId")

        Assert.assertEquals(uploadBlob.properties.blobSize, testFile.length())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldCopyToEdavContainer() {
        val options = ListBlobsOptions()
            .setPrefix(Metadata.getFilePrefixByDate(DateTime.now(), "dextesting-testevent1"))
            .setDetails(BlobListDetails().setRetrieveMetadata(true))
        val edavUploadBlob = edavContainerClient.listBlobs(options, Duration.ofMillis(EnvConfig.AZURE_BLOB_SEARCH_DURATION_MILLIS))
            .first { blob -> blob.metadata?.containsValue(uploadId) == true }

        Assert.assertNotNull(edavUploadBlob)
        Assert.assertEquals(edavUploadBlob.properties.contentLength, testFile.length())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldCopyToRoutingContainer() {
        val options = ListBlobsOptions()
            .setPrefix(Metadata.getFilePrefixByDate(DateTime.now(), "dextesting-testevent1"))
            .setDetails(BlobListDetails().setRetrieveMetadata(true))
        val routingUploadBlob = routingContainerClient.listBlobs(options, Duration.ofMillis(EnvConfig.AZURE_BLOB_SEARCH_DURATION_MILLIS))
            .first { blob -> blob.metadata?.containsValue(uploadId) == true }

        Assert.assertNotNull(routingUploadBlob)
        Assert.assertEquals(routingUploadBlob.properties.contentLength, testFile.length())
    }
}