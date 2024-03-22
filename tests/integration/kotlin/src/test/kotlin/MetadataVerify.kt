import auth.AuthClient
import io.restassured.RestAssured
import io.restassured.response.ValidatableResponse
import io.tus.java.client.ProtocolException
import org.testng.Assert
import org.testng.ITestContext
import org.testng.annotations.BeforeClass
import org.testng.annotations.BeforeGroups
import org.testng.annotations.Test
import tus.UploadClient
import util.*
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT


@Test()
class MetadataVerify {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient
    private lateinit var reportResponse: ValidatableResponse
    private val procStatReqSpec = RestAssured.given().apply {
        baseUri(EnvConfig.PROC_STAT_URL)
    }
    var metadata = HashMap<String, String>()
    val utilities = Utilities()

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    // Naz added below code
    @BeforeGroups(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun procStatUploadFileHappyPath(context: ITestContext) {
        val configFile = context.getCurrentXmlTest().getParameter("CONFIG_FILE")
        System.setProperty("CONFIG_FILE", configFile ?: "defaultConfig.properties")
        val propertiesFilePath="properties/"+configFile
        metadata = Metadata.generateRequiredMetadataForFile(testFile,propertiesFilePath)
    }
    // Naz updated below code
    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldUploadFileGivenRequiredMetadata() {
        println(metadata)
        val uploadId = uploadClient.uploadFile(testFile, metadata)
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.
        println("upload id: " +uploadId )

        Assert.assertNotNull(uploadId)
    }

    // Naz added below code
    @Test(groups = [Constants.Groups.PROC_STAT_UPLOAD_FILE_HAPPY_PATH])
    fun shouldUploadWithValidInputMetadata() {
        println(metadata)
        val uploadId = uploadClient.uploadFile(testFile, metadata)
        Thread.sleep(12_000) // Hard delay to wait for PS API to settle.
        println("upload id: " +uploadId )

        Assert.assertNotNull(uploadId)
        reportResponse = procStatReqSpec.get("/api/report/uploadId/$uploadId")
                .then()
                .statusCode(200)
        val reportResponseAsString = reportResponse.extract().response().asString()
        var actualMetadata= utilities.extractMetadataFromResponseBody(reportResponseAsString)

        println("Expected Meta Data: " +metadata)
        println("Actual Meta Data: " +actualMetadata)

        Assert.assertEquals(metadata, actualMetadata)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenDestinationIDNotProvided() {
        val metadata = hashMapOf(
            "filename" to testFile.name,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenEventNotProvided() {
        val metadata = hashMapOf(
            "filename" to testFile.name,
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_source" to "INTEGRATION-TEST"
        ) as HashMap<String, String>

        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(expectedExceptions = [ProtocolException::class])
    fun shouldReturnErrorWhenFilenameNotProvided() {
        val metadata = hashMapOf(
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        )

        uploadClient.uploadFile(testFile, metadata)
    }
}


