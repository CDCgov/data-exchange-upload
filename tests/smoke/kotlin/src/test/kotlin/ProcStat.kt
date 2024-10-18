import tus.UploadClient
import io.restassured.RestAssured.*
import dex.DexUploadClient
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import org.testng.Assert.*
import org.slf4j.LoggerFactory
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import model.DataResponse
import util.*
import util.DataProvider

@Listeners(UploadIdTestListener::class)
@Test()
class ProcStat {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    //private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
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
    fun shouldHaveReportsForSuccessfulFileUpload(case: TestCase, testContext: ITestContext) {
        //val config = ConfigLoader.loadUploadConfig(dexBlobClient, case.manifest)
        val uid = uploadClient.uploadFile(testFile, case.manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        log.debug("File uploaded successfully with UID: $uid")
        Thread.sleep(2000)

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
        log.debug("Received response: $jsonResponse")

        val objectMapper = jacksonObjectMapper()
        val dataResponse: DataResponse = objectMapper.readValue(jsonResponse, DataResponse::class.java)
        log.debug("Parsed DataResponse: $dataResponse")

        val reportList = dataResponse.data.reports

        reportList.forEach { report ->
            val schemaName = report.content.contentSchemaName
            log.debug("Validating schema: $schemaName")

            when (schemaName) {
                "metadata-verify" -> {
                    val uploadId = report.content.metadata?.uploadId
                    val dexIngestDateTime = report.content.metadata?.dexIngestDateTime
                    val stageInfo = report.stageInfo

                    log.info("Metadata Verify - Upload ID: $uploadId, Dex Ingest DateTime: $dexIngestDateTime")

                    log.debug("Validating uploadId for metadata-verify report. Expected: $uid, Actual: $uploadId")
                    assertEquals(uid, uploadId, "Expected upload ID to match the UID, but found: $uploadId")

                    log.debug("Validating stageInfo status. Expected: SUCCESS, Actual: ${stageInfo?.status}")
                    assertEquals(
                        "SUCCESS",
                        stageInfo?.status,
                        "Expected status 'SUCCESS' for metadata verify, but found: ${stageInfo?.status}"
                    )
                    log.debug("Checking for issues in stageInfo")
                    assertTrue(
                        stageInfo?.issues.isNullOrEmpty(),
                        "Expected no issues in the metadata verify report, but found: ${stageInfo?.issues}"
                    )
                }

                "metadata-transform" -> {
                    val transforms = report.content.transforms
                    log.debug("Validating transforms in metadata-transform report")

                    assertNotNull(
                        transforms,
                        "No transforms found in the metadata-transform report; expected at least one transform."
                    )
                    assertTrue(
                        transforms!!.isNotEmpty(),
                        "The transforms list in the metadata-transform report is empty; expected at least one transform."
                    )

                    transforms.forEach { transform ->
                        val action = transform.action
                        val field = transform.field
                        val value = transform.value
                        log.info("Transform Action: $action, Field: $field, Value: $value")
                    }
                }

                "upload-started", "upload-completed" -> {
                    val status = report.content.status
                    log.debug("Validating upload-started/upload-completed schema: $schemaName with status: $status")

                    assertEquals("SUCCESS", status, "Expected status 'SUCCESS' for $schemaName, but found: $status")

                    log.debug("Processing report for schema: $schemaName")
                    log.debug("Report content: Status: $status, Report ID: ${report.reportId}")
                    log.info("Schema Name: $schemaName, Status: $status")
                }

                "upload-status" -> {
                    val filename = report.content.filename
                    val tguid = report.content.tguid
                    val offset = report.content.offset
                    val size = report.content.size
                    log.debug("Validating upload status with Filename: $filename, Offset: $offset, Size: $size")

                    log.info("Upload Status - Filename: $filename, TGUID: $tguid, Offset: $offset, Size: $size")
                    assertEquals(
                        size,
                        offset,
                        "Upload-status mismatch: expected offset to equal size, but found size: $size and offset: $offset"
                    )
                }

                "blob-file-copy" -> {
                    val sourceUrl = report.content.fileSourceBlobUrl
                    val destinationUrl = report.content.fileDestinationBlobUrl
                    val destinationName = report.content.destinationName
                    log.debug("Validating blob-file-copy with Source URL: $sourceUrl, Destination URL: $destinationUrl")

                    assertNotNull(sourceUrl, "Blob source URL is missing for $schemaName")
                    assertNotNull(destinationUrl, "Blob destination URL is missing for $schemaName")
                    assertNotNull(destinationName, "Destination name is missing for $schemaName")
                    log.info("Blob File Copy - Source URL: $sourceUrl, Destination URL: $destinationUrl")

//                    val expectedDestinations = config.copyConfig.targets

                    val expectedDestinations = case.deliveryTargets!!.map{ it.name }.sorted()

//                    val actualDestination = expectedDestinations?.any { expectedDest ->
//                        destinationName?.trim()?.equals(expectedDest.trim(), ignoreCase = true) ?: false
//                    }
                    assertTrue(expectedDestinations.contains(destinationName), "Actual destination $destinationName is not in expected target destinations ${expectedDestinations.toString()}")
//                    assertTrue(
//                        actualDestination?,
//                        "None of the expected destinations ('${expectedDestinations.joinToString(", ")}') were found in the response. Found: $destinationName"
//                    )
                }

                else -> {
                    log.info("Unknown schema name: $schemaName")
                }
            }
        }
    }
}