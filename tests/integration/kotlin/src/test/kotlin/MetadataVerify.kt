import auth.AuthClient
import org.testng.Assert
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Env
import java.io.File


@Test()
class MetadataVerify {
    private val TEST_DESTINATION = "dextesting"
    private val TEST_EVENT = "testevent1"
    private var authClient = AuthClient(Env.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(Env.SAMS_USERNAME, Env.SAMS_PASSWORD)
        uploadClient = UploadClient(Env.UPLOAD_URL, authToken)
    }

    @Test()
    fun shouldUploadFileGivenRequiredMetadata() {
        // TODO: Handle test file not found.
        val file = File(MetadataVerify::class.java.getResource("10KB-test-file").file)
        val metadata = hashMapOf(
            "filename" to file.name,
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        val uploadId = uploadClient.uploadFile(file, metadata)
        Assert.assertNotNull(uploadId)
    }
}

