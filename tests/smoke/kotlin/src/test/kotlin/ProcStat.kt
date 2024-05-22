import org.testng.annotations.Test
import tus.UploadClient
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import dex.DexUploadClient
import model.Report
import model.UploadConfig
import org.hamcrest.Matchers.*
import org.testng.Assert.assertNotNull
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.BeforeTest
import org.testng.annotations.Listeners
import org.testng.annotations.Optional
import org.testng.annotations.Parameters
import util.*
import kotlin.test.assertNull

@Listeners(UploadIdTestListener::class)
@Test()
class ProcStat {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val procStatReqSpec = given().apply {
        baseUri(EnvConfig.PROC_STAT_URL)
    }
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var reportResponse: ValidatableResponse
    private lateinit var uploadConfig: UploadConfig

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.PROC_STAT])
    fun beforeTest(
        context: ITestContext,
        @Optional SENDER_MANIFEST: String?,
        @Optional("dextesting-testevent1") USE_CASE: String
    ) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val senderManifestDataFile = if (SENDER_MANIFEST.isNullOrEmpty()) "$USE_CASE.properties" else SENDER_MANIFEST
        val propertiesFilePath = "properties/$USE_CASE/$senderManifestDataFile"
        val senderManifest = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        uploadConfig = ConfigLoader.loadUploadConfig(dexBlobClient, "$USE_CASE.json", "v1")

        uploadId = uploadClient.uploadFile(testFile, senderManifest)
                ?: throw TestNGException("Error uploading file ${testFile.name}")
        context.setAttribute("uploadId", uploadId)

        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.

        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId")
                .then()
                .statusCode(200)
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
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-metadata-verify" }

        assertNotNull(metadataVerifyReport)
        assertNull(metadataVerifyReport?.issues)
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveUploadStatusReportWhenFileUploaded() {
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem(Constants.UPLOAD_STATUS_REPORT_STAGE_NAME)).body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val uploadReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == Constants.UPLOAD_STATUS_REPORT_STAGE_NAME }

        assertNotNull(uploadReport)
    }

    @Parameters("EXPECTED_SOURCE_URL_PREFIXES", "EXPECTED_DESTINATION_URL_PREFIXES")
    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveValidDestinationAndSourceURLWhenFileUploaded() {
        // Parse the expected URLs from the parameters
        val expectedDestinations = uploadConfig.copyConfig.targets

        val jsonPath = reportResponse.extract().jsonPath()
        val reports = jsonPath.getList("reports", Report::class.java)
        val dexFileCopyReports = reports.filterNotNull().filter { it.stageName == "dex-file-copy" }

        assert(dexFileCopyReports.isNotEmpty()) { "No 'dex-file-copy' reports found" }

        val destinations = dexFileCopyReports.mapNotNull { it.content.destination }

        // Validate source URLs
        destinations.forEach { dest ->
            assert(expectedDestinations.any { dest.trim().contains(it.trim()) }) { "Destination ${dest} does not match any expected destinations" }
        }
    }
}
