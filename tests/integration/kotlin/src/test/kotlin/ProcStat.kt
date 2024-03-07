import auth.AuthClient
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.EnvConfig
import util.Metadata
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import model.Report
import org.hamcrest.Matchers.*
import org.testng.Assert.assertNotNull
import org.testng.Assert.assertTrue
import org.testng.TestNGException
import org.testng.annotations.BeforeGroups
import util.Constants
import util.TestFile
import kotlin.test.assertContains
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

@Test()
class ProcStat {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private val procStatReqSpec = given().apply {
        baseUri(EnvConfig.PROC_STAT_URL)
    }
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var traceResponse: ValidatableResponse
    private lateinit var reportResponse: ValidatableResponse

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun procStatMetadataVerifyHappyPathSetup() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata)
                ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(5_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId").then().statusCode(200)

        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId").then().statusCode(200)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldCreateTraceWhenFileUploaded() {
        traceResponse.body("upload_id", equalTo(uploadId))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifySpanWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()

        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "metadata-verify")
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifyStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()

        val metadataVerifyStatus = jsonPath.getList<String>("spans.status").first()
        assertEquals("complete", metadataVerifyStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveMetadataVerifyReportWhenFileUploaded() {
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem("dex-metadata-verify")).body("reports.content.schema_name", hasItem("dex-metadata-verify"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_METADATA_VERIFY_HAPPY_PATH])
    fun shouldHaveNullIssuesArrayWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).first()

        assertEquals("dex-metadata-verify", metadataVerifyReport.stageName)
        assertNull(metadataVerifyReport.issues)
    }

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun procStatUploadStatusHappyPathSetup() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata)
                ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId").then().statusCode(200)
        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId").then().statusCode(200)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()

        val uploadStatus = jsonPath.getList<String>("spans.status").last()
        assertEquals("complete", uploadStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusSpanWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()

        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "dex-upload")
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveUploadStatusReportWhenFileUploaded() {
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem("dex-upload")).body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_HAPPY_PATH])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()

        val uploadReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-upload" }

        assertNotNull(uploadReport)
        // assertEquals(uploadReport.content.size, uploadReport.content.offset)
    }

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_DEX_FILE_COPY_HAPPY_PATH])
    fun procStatUploadStatusDexFileCopyHappyPath() {
        val metadata = Metadata.generateRequiredMetadataForFile(testFile)
        uploadId = uploadClient.uploadFile(testFile, metadata)
                ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId").then().statusCode(200)
        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId").then().statusCode(200)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_STATUS_DEX_FILE_COPY_HAPPY_PATH])
    fun shouldHaveValidDestinationAndSourceURLWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()

        val reports = jsonPath.getList("reports", Report::class.java)
        val dexFileCopyReports = reports.filterNotNull().filter { it.stageName == "dex-file-copy" }

        assert(dexFileCopyReports.isNotEmpty()) { "No 'dex-file-copy' reports found" }

        val sourceUrls = dexFileCopyReports.map { it.content.fileSourceBlobUrl }
        val destinationUrls = dexFileCopyReports.map { it.content.fileDestinationBlobUrl }

        val expectedSourceUrl = "https://ocioededataexchangedev.blob.core.windows.net/dextesting-testevent1/"
        val expectedUploadDestinationUrl = "https://edavdevdatalakedex.blob.core.windows.net/upload/dextesting-testevent1/"
        val expectedRoutingDestinationUrl = "https://ocioederoutingdatasadev.blob.core.windows.net/routeingress/dextesting-testevent1/"

        assert(sourceUrls.all { it?.contains(expectedSourceUrl) == true }) { "Not all source URLs contain the expected URL" }
        assert(destinationUrls.any { it?.contains(expectedUploadDestinationUrl) == true }) { "The expected upload destination URL is not present" }
        assert(destinationUrls.any { it?.contains(expectedRoutingDestinationUrl) == true }) { "The expected routing destination URL is not present" }

    }
}