import auth.AuthClient
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Env
import util.Metadata
import java.io.File
import io.restassured.RestAssured.*
import org.hamcrest.Matchers.*

@Test()
class ProcStat {
    private val testFile = File(MetadataVerify::class.java.getResource("10KB-test-file").file)
    private val authClient = AuthClient(Env.UPLOAD_URL)
    private val procStatReqSpec = given().apply {
        baseUri(Env.PROC_STAT_URL)
    }
    private lateinit var uploadClient: UploadClient

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(Env.SAMS_USERNAME, Env.SAMS_PASSWORD)
        uploadClient = UploadClient(Env.UPLOAD_URL, authToken)
    }

    @Test()
    fun shouldCreateTraceWhenFileUploaded() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        val uploadId = uploadClient.uploadFile(testFile, metadata)
        Thread.sleep(5_000)

        procStatReqSpec.get("/api/trace/uploadId/$uploadId")
            .then()
            .body("upload_id", equalTo(uploadId))
    }
}