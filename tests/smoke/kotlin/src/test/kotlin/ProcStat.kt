import tus.UploadClient
import io.restassured.RestAssured.*
import dex.DexUploadClient
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import util.*
import util.DataProvider
import org.testng.Assert.*
import org.slf4j.LoggerFactory
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import model.DataResponse

@Listeners(UploadIdTestListener::class)
@Test()
class ProcStat {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val procStatReqSpec = given().relaxedHTTPSValidation()
        .apply {
            baseUri(EnvConfig.PROC_STAT_URL)
        }
    private lateinit var authToken: String
    private lateinit var testContext: ITestContext
    private lateinit var uploadClient: UploadClient
    private val log = LoggerFactory.getLogger(ProcStat::class.java)

    @BeforeTest(groups = [Constants.Groups.PROC_STAT])
    fun beforeProcStat(context: ITestContext) {
        authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.PROC_STAT],
        dataProvider = "validManifestAllProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldHaveReportsForSuccessfulFileUpload(manifest: HashMap<String, String>, testContext: ITestContext) {
        val config = ConfigLoader.loadUploadConfig(dexBlobClient, manifest)
        val uid = uploadClient.uploadFile(testFile, manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(2000)

        try {

            val response = procStatReqSpec
                .body(
                    """
                {
                 "query": "query GetReports { getReports(uploadId: \"$uid\", reportsSortedBy: \"timestamp\", sortOrder: Ascending) { content contentType data dataProducerId dataStreamId dataStreamRoute dexIngestDateTime id jurisdiction reportId senderId tags timestamp uploadId stageInfo { action endProcessingTime service startProcessingTime status version issues { level message } } } }",
                  "variables": {}
                }
                """.trimIndent()
                )
                .header("Content-Type", "application/json")
                .log().all()
                .post("pstatus/graphql-service/graphql")

            val jsonResponse = response.asString()
            val objectMapper = jacksonObjectMapper()
            val dataResponse: DataResponse = objectMapper.readValue(jsonResponse, DataResponse::class.java)

            val reportList = dataResponse.data.reports

            reportList.forEach { report ->
                val schemaName = report.content.contentSchemaName
                when (schemaName) {

                    "metadata-verify" -> {
                        val uploadId = report.content.metadata?.uploadId
                        val dexIngestDateTime = report.content.metadata?.dexIngestDateTime
                        val stageInfo = report.stageInfo

                        log.info("Metadata Verify - Upload ID: $uploadId, Dex Ingest DateTime: $dexIngestDateTime")

                        assertEquals(uid, uploadId, "Expected upload ID to match the UID, but found: $uploadId")
                        assertEquals("SUCCESS", stageInfo?.status, "Expected status 'SUCCESS' for metadata verify, but found: ${stageInfo?.status}")
                        assertTrue(stageInfo?.issues.isNullOrEmpty(), "Expected no issues in the metadata verify report, but found: ${stageInfo?.issues}")
                    }


                    "metadata-transform" -> {
                        report.content.transforms?.forEach { transform ->
                            val action = transform.action
                            val field = transform.field
                            val value = transform.value
                            log.info("Transform Action: $action, Field: $field, Value: $value")
                        }
                    }

                    "upload-started", "upload-completed" -> {
                        val status = report.content.status
                        assertEquals("SUCCESS", status, "Expected status 'SUCCESS' for $schemaName, but found: $status")
                        log.info("Schema Name: $schemaName, Status: $status")
                    }

                    "upload-status" -> {
                        val filename = report.content.filename
                        val tguid = report.content.tguid
                        val offset = report.content.offset
                        val size = report.content.size
                        log.info("Upload Status - Filename: $filename, TGUID: $tguid, Offset: $offset, Size: $size")
                    }

                    "blob-file-copy" -> {
                        val sourceUrl = report.content.fileSourceBlobUrl
                        val destinationUrl = report.content.fileDestinationBlobUrl
                        val destinationName = report.content.destinationName
                        assertNotNull(sourceUrl, "Blob source URL is missing for $schemaName")
                        assertNotNull(destinationUrl, "Blob destination URL is missing for $schemaName")
                        assertNotNull(destinationName, "Destination name is missing for $schemaName")
                        log.info("Blob File Copy - Source URL: $sourceUrl, Destination URL: $destinationUrl")

                        val expectedDestinations = config.copyConfig.targets

                        val actualDestination = expectedDestinations.any { expectedDest ->
                            destinationName?.trim()?.equals(expectedDest.trim(), ignoreCase = true) ?: false
                        }

                        assertTrue(
                            actualDestination,
                            "None of the expected destinations ('${expectedDestinations.joinToString(", ")}') were found in the response. Found: $destinationName"
                        )
                    }

                    else -> {
                        log.info("Unknown schema name: $schemaName")
                    }
                }
            }

        } catch (e: Exception) {
            log.error("Test failed due to exception: ${e.message}", e)
            throw TestNGException("Test failed due to exception: ${e.message}")
        }
    }
}