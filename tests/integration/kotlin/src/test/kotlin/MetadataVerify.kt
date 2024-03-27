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

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.METADATA_VERIFY])
    fun beforeTest(
        context: ITestContext,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("dextesting-testevent1") USE_CASE: String
    ) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val propertiesFilePath= "properties/$USE_CASE/$SENDER_MANIFEST"
        metadataHappyPath = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY])
    fun shouldUploadFileGivenRequiredMetadata(context: ITestContext) {
        val uploadId = uploadClient.uploadFile(testFile, metadataHappyPath)
        context.setAttribute("uploadId", uploadId)

        Assert.assertNotNull(uploadId)
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY], expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenDestinationIDNotProvided() {
        val metadata = hashMapOf(
            "filename" to testFile.name,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY], expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenEventNotProvided() {
        val metadata = hashMapOf(
            "filename" to testFile.name,
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY], expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenFilenameNotProvided() {
        val metadata = hashMapOf(
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        )

        uploadClient.uploadFile(testFile, metadata)
    }
}
