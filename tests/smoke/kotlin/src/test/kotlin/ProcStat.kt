import auth.AuthClient
import org.testng.annotations.Test
import tus.UploadClient
import io.restassured.RestAssured.*
import io.restassured.response.ValidatableResponse
import model.Report
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
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private val procStatReqSpec = given().apply {
        baseUri(EnvConfig.PROC_STAT_URL)
    }
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var reportResponse: ValidatableResponse

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.PROC_STAT])
    fun beforeTest(
        context: ITestContext,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("dextesting-testevent1") USE_CASE: String
    ) {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val propertiesFilePath= "properties/$USE_CASE/$SENDER_MANIFEST"
        val metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        uploadId = uploadClient.uploadFile(testFile, metadata)
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
        reportResponse.body("upload_id", equalTo(uploadId)).body("reports.stage_name", hasItem("dex-upload")).body("reports.content.schema_name", hasItem("upload"))
    }

    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveEqualOffsetAndSizeWhenFileUploaded() {
        val jsonPath = reportResponse.extract().jsonPath()
        val uploadReport = jsonPath.getList("reports", Report::class.java).find { it.stageName == "dex-upload" }

        assertNotNull(uploadReport)
    }

    @Parameters("EXPECTED_SOURCE_URL_PREFIXES", "EXPECTED_DESTINATION_URL_PREFIXES")
    @Test(groups = [Constants.Groups.PROC_STAT, Constants.Groups.PROC_STAT_REPORT])
    fun shouldHaveValidDestinationAndSourceURLWhenFileUploaded(
        @Optional("https://ocioededataexchangedev.blob.core.windows.net/dextesting-testevent1") EXPECTED_SOURCE_URL_PREFIXES: String,
        @Optional("https://ocioederoutingdatasadev.blob.core.windows.net/routeingress/dextesting-testevent1,https://edavdevdatalakedex.blob.core.windows.net/upload/dextesting-testevent1") EXPECTED_DESTINATION_URL_PREFIXES: String
    ) {
        // Parse the expected URLs from the parameters
        val expectedSourceUrls = EXPECTED_SOURCE_URL_PREFIXES.split(",")
        val expectedDestinationUrls = EXPECTED_DESTINATION_URL_PREFIXES.split(",")

        val jsonPath = reportResponse.extract().jsonPath()
        val reports = jsonPath.getList("reports", Report::class.java)
        val dexFileCopyReports = reports.filterNotNull().filter { it.stageName == "dex-file-copy" }

        assert(dexFileCopyReports.isNotEmpty()) { "No 'dex-file-copy' reports found" }

        val sourceUrls = dexFileCopyReports.mapNotNull { it.content.fileSourceBlobUrl }
        val destinationUrls = dexFileCopyReports.mapNotNull { it.content.fileDestinationBlobUrl }

        // Validate source URLs
        sourceUrls.forEach { sourceUrl ->
            assert(expectedSourceUrls.any { sourceUrl.trim().contains(it.trim()) }) { "Source URL $sourceUrl does not match any expected URLs." }
        }

        // Validate destination URLs
        destinationUrls.forEach { destinationUrl ->
            assert(expectedDestinationUrls.any { destinationUrl.trim().contains(it.trim()) }) { "Destination URL $destinationUrl does not match any expected URLs." }
        }
    }
}
