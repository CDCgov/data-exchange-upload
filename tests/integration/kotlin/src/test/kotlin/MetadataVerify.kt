import auth.AuthClient
import io.restassured.RestAssured.given
import io.restassured.http.ContentType
import io.restassured.path.json.JsonPath
import io.tus.java.client.ProtocolException
import org.hamcrest.CoreMatchers.equalTo
import org.testng.Assert
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT
import util.EnvConfig
import util.TestFile




@Test()
class MetadataVerify {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test()
    fun shouldUploadFileGivenRequiredMetadata() {
        val metadata = hashMapOf(
                "filename" to testFile.name,
                "meta_destination_id" to TEST_DESTINATION,
                "meta_ext_event" to TEST_EVENT,
                "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        val uploadId = uploadClient.uploadFile(testFile, metadata)
        Assert.assertNotNull(uploadId)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenDestinationIDNotProvided() {
        val metadata = hashMapOf(
                "filename" to testFile.name,
                "meta_ext_event" to TEST_EVENT,
                "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenEventNotProvided() {
        val metadata = hashMapOf(
                "filename" to testFile.name,
                "meta_destination_id" to TEST_DESTINATION,
                "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenFilenameNotProvided() {
        val metadata = hashMapOf(
                "meta_destination_id" to TEST_DESTINATION,
                "meta_ext_event" to TEST_EVENT,
                "meta_ext_source" to "INTEGRATION-TEST"
        )

        uploadClient.uploadFile(testFile, metadata)
    }
}


