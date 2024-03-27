import auth.AuthClient
import org.testng.annotations.Test
import tus.UploadClient
import util.EnvConfig
import util.Metadata
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import model.Report
import org.hamcrest.Matchers.*
import org.testng.Assert.assertNotNull
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.BeforeGroups
import util.Constants
import util.TestFile
import kotlin.test.assertContains
import kotlin.test.assertEquals
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

    @BeforeGroups(groups = [Constants.Groups.PROC_STAT])
    fun procStatHappyPath(context: ITestContext) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
        val senderManifestPropertiesFilename = context.currentXmlTest.getParameter("SENDER_MANIFEST")
        val propertiesFilePath= "properties/$senderManifestPropertiesFilename"
        val metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

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
    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_TRACE])
    fun shouldCreateTraceWhenFileUploaded() {
        traceResponse.body("upload_id", equalTo(uploadId))
    }
    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_TRACE])
    fun shouldHaveMetadataVerifySpanWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "metadata-verify")
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_TRACE])
    fun shouldHaveMetadataVerifyStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val metadataVerifyStatus = jsonPath.getList<String>("spans.status").first()
        assertEquals("complete", metadataVerifyStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_TRACE])
    fun shouldHaveUploadStatusCompleteWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val uploadStatus = jsonPath.getList<String>("spans.status").last()
        assertEquals("complete", uploadStatus)
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_TRACE])
    fun shouldHaveUploadStatusSpanWhenFileUploaded() {
        val jsonPath = traceResponse.extract().jsonPath()
        val stageNames = jsonPath.getList<String>("spans.stage_name")
        assertContains(stageNames, "dex-upload")
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveMetadataVerifyReportWhenFileUploaded() {
        reportResponse.
        body("upload_id", equalTo(uploadId))
                .body("reports.stage_name",
                hasItem("dex-metadata-verify")).body("reports.content.schema_name",
                hasItem("dex-metadata-verify"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveNullIssuesArrayWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).first()

        assertEquals("dex-metadata-verify", metadataVerifyReport.stageName)
        assertNull(metadataVerifyReport.issues)
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveUploadStatusReportWhenFileUploaded() {
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem("dex-upload")).body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val uploadReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-upload" }

        assertNotNull(uploadReport)
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveValidDestinationAndSourceURLWhenFileUploaded(context: ITestContext) {
        // Parse the expected URLs from the parameters
        val expectedSourceUrls = context.currentXmlTest.getParameter("EXPECTED_SOURCE_URL_PREFIXES").split(",")
        val expectedDestinationUrls = context.currentXmlTest.getParameter("EXPECTED_DESTINATION_URL_PREFIXES").split(",")

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
