import auth.AuthClient
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Env
import util.Metadata
import java.io.File
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import org.hamcrest.Matchers.*
import org.testng.TestNGException
import org.testng.annotations.BeforeGroups
import util.Groups
import kotlin.test.assertContains
import kotlin.test.assertEquals

@Test()
class ProcStat {
    private val testFile = File(MetadataVerify::class.java.getResource("10KB-test-file").file)
    private val authClient = AuthClient(Env.UPLOAD_URL)
    private val procStatReqSpec = given().apply {
        baseUri(Env.PROC_STAT_URL)
    }
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var traceResponse: ValidatableResponse

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(Env.SAMS_USERNAME, Env.SAMS_PASSWORD)
        uploadClient = UploadClient(Env.UPLOAD_URL, authToken)
    }

    @BeforeGroups(groups = [Groups.PROC_STAT_TRACE_HAPPY_PATH])
    fun procStatHappyPathSetup() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(5_000)
        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId")
            .then()
    }

    @Test(groups = [Groups.PROC_STAT_TRACE_HAPPY_PATH])
    fun shouldCreateTraceWhenFileUploaded() {
        traceResponse
            .statusCode(200)
            .body("upload_id", equalTo(uploadId))
    }

    @Test(groups = [Groups.PROC_STAT_TRACE_HAPPY_PATH])
    fun shouldHaveMetadataVerifySpanWhenFileUploaded() {
        val jsonPath = traceResponse
            .statusCode(200)
            .extract().jsonPath()

        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "metadata-verify")
    }

    @Test(groups = [Groups.PROC_STAT_TRACE_HAPPY_PATH])
    fun shouldHaveMetadataVerifyStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse
            .statusCode(200)
            .extract().jsonPath()

        val metadataVerifyStatus = jsonPath.getList<String>("spans.status").first()
        assertEquals(metadataVerifyStatus, "complete")
    }
}