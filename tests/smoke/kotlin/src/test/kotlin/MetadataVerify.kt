import com.azure.storage.blob.BlobContainerClient
import dex.DexUploadClient
import io.tus.java.client.ProtocolException
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
    private lateinit var testContext: ITestContext

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
        dataProvider = "invalidManifestRequiredFieldsProvider",
        dataProviderClass = DataProvider::class,
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*field .* was missing"
    )
    fun shouldReturnErrorWhenMissingRequiredField(manifest: Map<String, String>) {
        uploadClient.uploadFile(testFile, manifest)
    }
}