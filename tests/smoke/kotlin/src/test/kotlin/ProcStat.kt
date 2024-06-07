import tus.UploadClient
import io.restassured.RestAssured.*
import dex.DexUploadClient
import model.Report
import org.hamcrest.Matchers.*
import org.testng.Assert.assertNotNull
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import util.*
import util.DataProvider
import kotlin.test.assertNull

@Listeners(UploadIdTestListener::class)
@Test()
class ProcStat {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val procStatReqSpec = given().apply {
        baseUri(EnvConfig.PROC_STAT_URL)
    }
    private lateinit var authToken: String
    private lateinit var testContext: ITestContext

    private lateinit var uploadClient: UploadClient

    @BeforeTest(groups = [Constants.Groups.PROC_STAT])
    fun beforeProcStat(
        context: ITestContext,
    ) {
        authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(groups = [Constants.Groups.PROC_STAT], dataProvider = "validManifestAllProvider", dataProviderClass = DataProvider::class)
    fun shouldHaveReportsForSuccessfulFileUpload(manifest: HashMap<String, String>) {
        val config = ConfigLoader.loadUploadConfig(dexBlobClient, manifest)
        val uid = uploadClient.uploadFile(testFile, manifest) ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(500)

        val reportResponse = procStatReqSpec.get("/api/report/uploadId/$uid")
            .then()
            .statusCode(200)

        // Metadata Verify
        reportResponse.
            body("upload_id", equalTo(uid))
                .body("reports.stage_name",
                    hasItem("dex-metadata-verify")).body("reports.content.schema_name",
                    hasItem("dex-metadata-verify"))

        var jsonPath = reportResponse.extract().jsonPath()
        val metadataVerifyReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-metadata-verify" }
        assertNotNull(metadataVerifyReport)
        assertNull(metadataVerifyReport?.issues)

        // Upload
        reportResponse
            .body("upload_id", equalTo(uid))
                .body("reports.stage_name", hasItem(Constants.UPLOAD_STATUS_REPORT_STAGE_NAME))
                    .body("reports.content.schema_name", hasItem("upload"))

        jsonPath = reportResponse.extract().jsonPath()
        val reports = jsonPath.getList("reports", Report::class.java)
        val uploadReport = reports
            .find { it.stageName == Constants.UPLOAD_STATUS_REPORT_STAGE_NAME }
        assertNotNull(uploadReport)

        val expectedDestinations = config.copyConfig.targets
        val dexFileCopyReports = reports.filterNotNull().filter { it.stageName == "dex-file-copy" }
        assert(dexFileCopyReports.isNotEmpty()) { "No 'dex-file-copy' reports found" }

        val destinations = dexFileCopyReports.mapNotNull { it.content.destination }

        // Post Processing
        expectedDestinations.forEach { dest ->
            assert(destinations.any { dest.trim().contains(it.trim()) }) { "Destination $dest was not found in copy reports" }
        }
    }
}
