import auth.AuthClient
import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobClient
import com.azure.storage.blob.BlobContainerClient
import com.azure.storage.blob.models.BlobItem
import com.azure.storage.blob.models.BlobListDetails
import com.azure.storage.blob.models.ListBlobsOptions
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import model.UploadConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import tus.UploadClient
import util.*
import java.time.Duration

@Listeners(UploadIdTestListener::class)
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
    private lateinit var uploadConfigBlobClient: BlobClient
    private lateinit var uploadConfig: UploadConfig
    private lateinit var edavContainerClient: BlobContainerClient
    private lateinit var routingContainerClient: BlobContainerClient
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var useCase: String

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeTest(
        context: ITestContext,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("dextesting-testevent1") USE_CASE: String
    ) {
        useCase = USE_CASE

        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val propertiesFilePath= "properties/$USE_CASE/$SENDER_MANIFEST"
        val metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)

        uploadConfigBlobClient = dexBlobClient
            .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
            .getBlobClient("v1/${USE_CASE}.json")
        uploadConfig = jacksonObjectMapper().readValue(uploadConfigBlobClient.downloadContent().toString())

        edavContainerClient = edavBlobClient.getBlobContainerClient(Constants.EDAV_UPLOAD_CONTAINER_NAME)
        routingContainerClient = routingBlobClient.getBlobContainerClient(Constants.ROUTING_UPLOAD_CONTAINER_NAME)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        context.setAttribute("uploadId", uploadId)
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
    fun shouldCopyToDestinationContainers() {
        val filenameSuffix = if (uploadConfig.copyConfig.filenameSuffix == "upload_id") "_${uploadId}" else ""
        val expectedFilename = "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC), useCase)}/${testFile.nameWithoutExtension}${filenameSuffix}${testFile.extension}"
        var expectedBlobClient: BlobClient?

        if (uploadConfig.copyConfig.targets.contains("edav")) {
            expectedBlobClient = edavContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }

        if (uploadConfig.copyConfig.targets.contains("routing")) {
            expectedBlobClient = routingContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun verifyMetadataTranslation(){
        val options = ListBlobsOptions()
            .setPrefix(Metadata.getFilePrefixByDate(DateTime.now(),useCase))
            .setDetails(BlobListDetails().setRetrieveMetadata(true))
        val routingUploadBlob =routingContainerClient.listBlobs(options, Duration.ofMillis(EnvConfig.AZURE_BLOB_SEARCH_DURATION_MILLIS))
            .first{blob ->blob.metadata?.containsValue(uploadId) == true }

        Assert.assertNotNull(routingUploadBlob)
        Assert.assertEquals(routingUploadBlob.properties.contentLength, testFile.length())

        val metaDataKeys = uploadClient.getAllMetadataKeys(routingUploadBlob)
        for (key in metaDataKeys) {
            assert(routingUploadBlob.metadata?.containsKey(key) == true) {"Metadata key '$key' not found"}
        }

    }

}