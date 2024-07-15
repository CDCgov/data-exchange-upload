import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobClient
import dex.DexUploadClient
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.ConfigLoader.Companion.loadUploadConfig
import util.DataProvider
import kotlin.collections.HashMap


@Listeners(UploadIdTestListener::class)
@Test()
class FileCopy {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
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
    private val bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)
    private val edavContainerClient = edavBlobClient.getBlobContainerClient(Constants.EDAV_UPLOAD_CONTAINER_NAME)
    private val routingContainerClient =
        routingBlobClient.getBlobContainerClient(Constants.ROUTING_UPLOAD_CONTAINER_NAME)
    private lateinit var authToken: String
    private lateinit var testContext: ITestContext
    private lateinit var uploadClient: UploadClient

    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeFileCopy() {
        authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.FILE_COPY],
        dataProvider = "validManifestAllProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldUploadFile(manifest: HashMap<String, String>) {
        val uid = uploadClient.uploadFile(testFile, manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(2000)

        // First, check bulk upload and .info file.
        val uploadBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uid")
        val uploadInfoBlob =
            bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uid.info")

        Assert.assertTrue(uploadBlob.exists())
        Assert.assertTrue(uploadInfoBlob.exists())
        Assert.assertEquals(uploadBlob.properties.blobSize, testFile.length())

        // Next, check that the file arrived in destination storage.
        val config = loadUploadConfig(dexBlobClient, manifest)
        val filenameSuffix = Filename.getFilenameSuffix(config.copyConfig, uid)
        val expectedFilename = "${
            Metadata.getFilePrefix(config.copyConfig, manifest)
        }${Metadata.getFilename(manifest)}${filenameSuffix}${testFile.extension}"
        var expectedBlobClient: BlobClient?

        if (config.copyConfig.targets.contains("edav")) {
            expectedBlobClient = edavContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }

        if (config.copyConfig.targets.contains("routing")) {
            expectedBlobClient = routingContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }
    }

    @Test(
        groups = [Constants.Groups.FILE_COPY],
        dataProvider = "validManifestV1Provider",
        dataProviderClass = DataProvider::class
    )
    fun shouldTranslateMetadataGivenV1SenderManifest(manifest: HashMap<String, String>) {
        val useCase = Metadata.getUseCaseFromManifest(manifest)
        val dexContainerClient = dexBlobClient.getBlobContainerClient(useCase)
        val v1Config = loadUploadConfig(dexBlobClient, manifest)
        val v2ConfigFilename = v1Config.compatConfigFilename ?: "$useCase.json"
        val v2Config = loadUploadConfig(dexBlobClient, v2ConfigFilename, "v2")
        val metadataMapping = v2Config.metadataConfig.fields.filter { it.compatFieldName != null }
            .associate { it.compatFieldName to it.fieldName }

        val uid = uploadClient.uploadFile(testFile, manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(1000)

        val filenameSuffix = Filename.getFilenameSuffix(v1Config.copyConfig, uid)
        val expectedFilename =
            "${Metadata.getFilePrefix(v1Config.copyConfig)}${Metadata.getFilename(manifest)}$filenameSuffix${testFile.extension}"
        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)
        val blobMetadata = expectedBlobClient.properties.metadata

        metadataMapping.forEach { (v1Key, v2Key) ->
            Assert.assertTrue(
                blobMetadata.containsKey(v2Key),
                "Mismatch: Blob metadata does not contain expected V2 key: $v2Key which should map from V1 key: $v1Key"
            )
        }

        metadataMapping.forEach { (v1Key, v2Key) ->
            val v1Val = manifest[v1Key] ?: ""
            val v2Val = blobMetadata[v2Key] ?: ""
            Assert.assertEquals(v1Val, v2Val, "Expected V1 value: $v1Val does not match with actual V2 value: $v2Val")
        }
    }
}