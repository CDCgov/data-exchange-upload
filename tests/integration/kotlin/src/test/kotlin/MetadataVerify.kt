import auth.AuthClient
import io.tus.java.client.ProtocolException
import org.testng.Assert
import org.testng.ITestContext
import org.testng.annotations.BeforeTest
import org.testng.annotations.Test
import tus.UploadClient
import util.Constants
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT
import util.EnvConfig
import util.Metadata
import util.TestFile


@Test()
class MetadataVerify {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient
    private lateinit var matadataHappyPath: HashMap<String, String>

    @BeforeTest(groups = [Constants.Groups.METADATA_VERIFY])
    fun beforeTest(context: ITestContext) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val senderManifestPropertiesFilename = context.currentXmlTest.getParameter("SENDER_MANIFEST")
        val useCase = context.currentXmlTest.getParameter("USE_CASE")
        val propertiesFilePath= "properties/$useCase/$senderManifestPropertiesFilename"
        matadataHappyPath = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)
    }

    @Test(groups = [Constants.Groups.METADATA_VERIFY])
    fun shouldUploadFileGivenRequiredMetadata() {
        val uploadId = uploadClient.uploadFile(testFile, matadataHappyPath)
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
