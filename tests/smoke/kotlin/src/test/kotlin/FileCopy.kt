import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobClient
import com.azure.storage.blob.BlobContainerClient
import dex.DexUploadClient
import model.UploadConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import org.testng.annotations.Optional
import tus.UploadClient
import util.*
import util.ConfigLoader.Companion.loadUploadConfig
import kotlin.collections.HashMap


@Listeners(UploadIdTestListener::class)
@Test()
class FileCopy {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val edavBlobClient = Azure.getBlobServiceClient(
        EnvConfig.EDAV_STORAGE_ACCOUNT_NAME,
        ClientSecretCredentialBuilder()
            .clientId(EnvConfig.AZURE_CLIENT_ID)
            .clientSecret(EnvConfig.AZURE_CLIENT_SECRET)
            .tenantId(EnvConfig.AZURE_TENANT_ID)
            .build()
    )
    private val routingBlobClient = Azure.getBlobServiceClient(EnvConfig.ROUTING_STORAGE_CONNECTION_STRING)
    private lateinit var bulkUploadsContainerClient: BlobContainerClient
    private lateinit var dexContainerClient: BlobContainerClient
    private lateinit var edavContainerClient: BlobContainerClient
    private lateinit var routingContainerClient: BlobContainerClient
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var useCase: String
    private lateinit var uploadConfigV1: UploadConfig
    private lateinit var metadata: HashMap<String, String>

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeTest(
        context: ITestContext,
        @Optional SENDER_MANIFEST: String?,
        @Optional("dextesting-testevent1") USE_CASE: String,
    ) {
        useCase = USE_CASE

        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val senderManifestDataFile = if (SENDER_MANIFEST.isNullOrEmpty()) "$USE_CASE.properties" else SENDER_MANIFEST
        val propertiesFilePath = "properties/V1/$USE_CASE/$senderManifestDataFile"
        metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)

        uploadConfigV1 = loadUploadConfig(dexBlobClient, "$USE_CASE.json", "v1")

        dexContainerClient = dexBlobClient.getBlobContainerClient(USE_CASE)
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
        val uploadInfoBlob =
            bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId.info")

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
        val filenameSuffix = Filename.getFilenameSuffix(uploadConfigV1.copyConfig, uploadId)
        val expectedFilename = "${
            Metadata.getFilePrefixByDate(
                DateTime(DateTimeZone.UTC),
                useCase
            )
        }/${testFile.nameWithoutExtension}${filenameSuffix}${testFile.extension}"
        var expectedBlobClient: BlobClient?

        if (uploadConfigV1.copyConfig.targets.contains("edav")) {
            expectedBlobClient = edavContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }

        if (uploadConfigV1.copyConfig.targets.contains("routing")) {
            expectedBlobClient = routingContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldTranslateMetadataGivenV1SenderManifest() {
        val v2ConfigFilename = uploadConfigV1.compatConfigFilename ?: "$useCase.json"
        val uploadConfigV2 = loadUploadConfig(dexBlobClient, v2ConfigFilename, "v2")

        val metadataMapping = uploadConfigV2.metadataConfig.fields.filter { it.compatFieldName != null }
            .associate { it.compatFieldName to it.fieldName }

        val filenameSuffix = Filename.getFilenameSuffix(uploadConfigV1.copyConfig, uploadId)

        val expectedFilename =
            "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/${testFile.nameWithoutExtension}$filenameSuffix.${testFile.extension}"

        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)

        val blobMetadata = expectedBlobClient.properties.metadata

        metadataMapping.forEach { (v1Key, v2Key) ->
            Assert.assertTrue(blobMetadata.containsKey(v2Key), "Mismatch: Blob metadata does not contain expected V2 key: $v2Key which should map from V1 key: $v1Key")
        }

        metadataMapping.forEach{ (v1Key, v2Key) ->
            val v1Val = metadata[v1Key] ?: ""
            val v2Val = blobMetadata[v2Key] ?: ""
            Assert.assertEquals(v1Val, v2Val, "Expected V1 value: $v1Val does not match with actual V2 value: $v2Val")
        }
    }
}