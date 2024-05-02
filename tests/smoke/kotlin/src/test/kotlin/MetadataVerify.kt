import auth.AuthClient
import io.tus.java.client.ProtocolException
import org.testng.Assert
import org.testng.ITestContext
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT

@Listeners(UploadIdTestListener::class)
@Test()
class MetadataVerify {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient
    private lateinit var metadataHappyPath: HashMap<String, String>
    private lateinit var metadataInvalidFilename: HashMap<String, String>
    private lateinit var metadataNoDestId: HashMap<String, String>
    private lateinit var metadataNoEvent: HashMap<String, String>

    @Parameters(
        "USE_CASE",
        "SENDER_MANIFEST",
        "SENDER_MANIFEST_INVALID_FILENAME",
        "SENDER_MANIFEST_NO_DEST_ID",
        "SENDER_MANIFEST_NO_EVENT",
    )
    @BeforeTest(groups = [Constants.Groups.METADATA_VERIFY])
    fun beforeTest(
        @Optional("dextesting-testevent1") USE_CASE: String,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("invalid-filename.properties") SENDER_MANIFEST_INVALID_FILENAME: String,
        @Optional("no-dest-id.properties") SENDER_MANIFEST_NO_DEST_ID: String,
        @Optional("no-event.properties") SENDER_MANIFEST_NO_EVENT: String
    ) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        metadataHappyPath = Metadata.convertPropertiesToMetadataMap("properties/$USE_CASE/$SENDER_MANIFEST")
        metadataInvalidFilename = Metadata.convertPropertiesToMetadataMap("properties/$USE_CASE/$SENDER_MANIFEST_INVALID_FILENAME")
        metadataNoDestId = Metadata.convertPropertiesToMetadataMap("properties/$USE_CASE/$SENDER_MANIFEST_NO_DEST_ID")
        metadataNoEvent = Metadata.convertPropertiesToMetadataMap("properties/$USE_CASE/$SENDER_MANIFEST_NO_EVENT")
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY])
    fun shouldUploadFileGivenRequiredMetadata(context: ITestContext) {
        val uploadId = uploadClient.uploadFile(testFile, metadataHappyPath)
        context.setAttribute("uploadId", uploadId)

        Assert.assertNotNull(uploadId)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*")
    fun shouldReturnErrorWhenDestinationIDNotProvided() {
        uploadClient.uploadFile(testFile, metadataNoDestId)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*"
    )
    fun shouldReturnErrorWhenEventNotProvided() {
        uploadClient.uploadFile(testFile, metadataNoEvent)
    }

    @Test(groups = [
        Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*field filename was missing"
    )
    fun shouldReturnErrorWhenFilenameNotProvided() {
        val metadata = hashMapOf(
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        )

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(groups = [
        Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*invalid character found.*"
    )
    fun shouldReturnErrorWhenFilenameContainsInvalidChars() {
        uploadClient.uploadFile(testFile, metadataInvalidFilename)
    }
}
