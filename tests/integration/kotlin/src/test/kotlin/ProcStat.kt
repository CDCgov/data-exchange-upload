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
import org.testng.ITestContext
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

    // Naz updated below code
    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun procStatUploadFileHappyPath(context: ITestContext) {
        val configFile = context.currentXmlTest.getParameter("CONFIG_FILE")
        System.setProperty("CONFIG_FILE", configFile ?: "defaultConfig.properties")
        val propertiesFilePath= "properties/$configFile"
        val metadata = Metadata.generateRequiredMetadataForFile(testFile,propertiesFilePath)

        uploadId = uploadClient.uploadFile(testFile, metadata)
                ?: throw TestNGException("Error uploading file ${testFile.name}")
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.

        traceResponse = procStatReqSpec.get("/api/trace/uploadId/$uploadId")
                .then()
                .statusCode(200)
        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId")
                .then()
                .statusCode(200)
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

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveMetadataVerifyReportWhenFileUploaded() {
        reportResponse.
        body("upload_id", equalTo(uploadId))
                .body("reports.stage_name",
                hasItem("dex-metadata-verify")).body("reports.content.schema_name",
                hasItem("dex-metadata-verify"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveNullIssuesArrayWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).first()

        assertEquals("dex-metadata-verify", metadataVerifyReport.stageName)
        assertNull(metadataVerifyReport.issues)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveUploadStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val uploadStatus = jsonPath.getList<String>("spans.status").last()
        assertEquals("complete", uploadStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveUploadStatusSpanWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "dex-upload")
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveUploadStatusReportWhenFileUploaded() {
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem("dex-upload")).body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val uploadReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-upload" }

        assertNotNull(uploadReport)
    }

    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldHaveValidDestinationAndSourceURLWhenFileUploaded(context: ITestContext) {
        val configFile = context.currentXmlTest.getParameter("CONFIG_FILE")
        System.setProperty("CONFIG_FILE", configFile ?: "defaultConfig.properties")
        // Parse the expected URLs from the parameters
        val expectedSourceUrls = context.currentXmlTest.getParameter("EXPECTED_SOURCE_URLS").split(",")
        val expectedDestinationUrls = context.currentXmlTest.getParameter("EXPECTED_DESTINATION_URLS").split(",")

        val jsonPath = reportResponse.extract().jsonPath()
        val reports = jsonPath.getList("reports", Report::class.java)
        val dexFileCopyReports = reports.filterNotNull().filter { it.stageName == "dex-file-copy" }

        assert(dexFileCopyReports.isNotEmpty()) { "No 'dex-file-copy' reports found" }

        val sourceUrls = dexFileCopyReports.mapNotNull { it.content.fileSourceBlobUrl }
        val destinationUrls = dexFileCopyReports.mapNotNull { it.content.fileDestinationBlobUrl }

        // Validate source URLs
        sourceUrls.forEach { sourceUrl ->
            assert(expectedSourceUrls.any { sourceUrl.contains(it) }) { "Source URL $sourceUrl does not match any expected URLs." }
        }

        // Validate destination URLs
        destinationUrls.forEach { destinationUrl ->
            assert(expectedDestinationUrls.any { destinationUrl.contains(it) }) { "Destination URL $destinationUrl does not match any expected URLs." }
        }
    }






}
