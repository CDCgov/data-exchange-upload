import dex.DexUploadClient
import io.tus.java.client.ProtocolException
import org.testng.ITestContext
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT
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

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        dataProvider = "invalidManifestInvalidValueProvider",
        dataProviderClass = DataProvider::class,
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*had disallowed value.*"
    )
    fun shouldReturnErrorWhenManifestValueInvalid(manifest: Map<String, String>) {
        uploadClient.uploadFile(testFile, manifest)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*invalid character found.*"
    )
    fun shouldReturnErrorWhenFilenameContainsInvalidChars() {
        val manifest = hashMapOf(
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "filename" to "test/-file"
        )
        uploadClient.uploadFile(testFile, manifest)
    }
}