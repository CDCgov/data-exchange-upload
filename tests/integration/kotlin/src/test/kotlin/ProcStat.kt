import auth.AuthClient
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Env
import util.Metadata
import java.io.File
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import model.Report
import org.hamcrest.Matchers.*
import org.testng.TestNGException
import org.testng.annotations.BeforeGroups
import util.Constants
import kotlin.test.assertContains
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

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
    private lateinit var reportResponse: ValidatableResponse

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(Env.SAMS_USERNAME, Env.SAMS_PASSWORD)
        uploadClient = UploadClient(Env.UPLOAD_URL, authToken)
    }

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun procStatMetadataVerifyHappyPathSetup() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(5_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId")
            .then()
            .statusCode(200)

        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId")
            .then()
            .statusCode(200)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldCreateTraceWhenFileUploaded() {
        traceResponse
            .body("upload_id", equalTo(uploadId))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifySpanWhenFileUploaded() {
        val jsonPath = traceResponse
            .extract().jsonPath()

        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "metadata-verify")
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifyStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse
            .extract().jsonPath()

        val metadataVerifyStatus = jsonPath.getList<String>("spans.status").first()
        assertEquals("complete", metadataVerifyStatus)
    }
    
    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifyReportWhenFileUploaded() {
        reportResponse
            .body("upload_id", equalTo(uploadId))
            .body("reports.stage_name", hasItem("dex-metadata-verify"))
            .body("reports.content.schema_name", hasItem("dex-metadata-verify"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveNullIssuesArrayWhenFileUploaded() {
        val jsonPath = reportResponse
            .extract().jsonPath()
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).first()

        assertEquals("dex-metadata-verify", metadataVerifyReport.stageName)
        assertNull(metadataVerifyReport.issues)
    }

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun procStatUploadStatusHappyPathSetup() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId")
            .then()
            .statusCode(200)
        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId")
            .then()
            .statusCode(200)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse
            .extract().jsonPath()

        val uploadStatus = jsonPath.getList<String>("spans.status").last()
        assertEquals("complete", uploadStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusSpanWhenFileUploaded() {
        val jsonPath = traceResponse
            .extract().jsonPath()

        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "dex-upload")
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusReportWhenFileUploaded() {
        reportResponse
            .body("upload_id", equalTo(uploadId))
            .body("reports.stage_name", hasItem("dex-upload"))
            .body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse
            .extract().jsonPath()

        val uploadReport = jsonPath.getList("reports", Report::class.java)
            .find { it.stageName == "dex-upload" }

        assertNotNull(uploadReport)
        assertEquals(uploadReport.content.size, uploadReport.content.offset)
    }
}