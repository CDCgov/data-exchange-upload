import com.azure.storage.blob.BlobContainerClient
import dex.DexUploadClient
import model.UploadConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.ConfigLoader.Companion.loadUploadConfig
import util.DataProvider

@Listeners(UploadIdTestListener::class)
@Test(groups = [Constants.Groups.METADATA_VERIFY])
class MetadataVerify {

    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private lateinit var authToken: String
    private lateinit var uploadClient: UploadClient
    private lateinit var metadata: HashMap<String, String>
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private lateinit var dexContainerClient: BlobContainerClient
    private lateinit var testContext: ITestContext
    private lateinit var useCase: String
    private lateinit var uploadConfig: UploadConfig

    @BeforeTest(groups = [Constants.Groups.METADATA_VERIFY])
    fun beforeTest() {
        authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        dataProvider = "validManifestAllProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldUploadFileGivenRequiredMetadata(manifest: Map<String, String>, context: ITestContext) {
        metadata = HashMap(manifest)
        val uploadId = uploadClient.uploadFile(testFile, metadata)
        context.setAttribute("uploadId", uploadId)
        Assert.assertNotNull(uploadId)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        dataProvider = "validManifestAllProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldValidateMetadataWithManifest(manifest: HashMap<String, String>) {

        val useCase = Metadata.getUseCaseFromManifest(manifest)
        val dexContainerClient = dexBlobClient.getBlobContainerClient(useCase)

        val uploadConfig = loadUploadConfig(dexBlobClient, manifest)

        metadata = Metadata.readMetadataFromJsonFile(useCase)
        val uid = uploadClient.uploadFile(testFile, manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(500)

        val filenameSuffix = Filename.getFilenameSuffix(uploadConfig.copyConfig, uid)
        val expectedFilename =
            "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/${Metadata.getFilename(manifest)}$filenameSuffix${testFile.extension}"
        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)
        val blobMetadata = expectedBlobClient.properties.metadata

        metadata.forEach { (key, value) ->
            val actualValue = blobMetadata[key]
            Assert.assertEquals(
                value, actualValue, "Expected key value: $value does not match with actual key value: $actualValue"
            )
        }
    }
}