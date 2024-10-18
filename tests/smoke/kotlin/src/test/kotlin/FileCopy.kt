import dex.DexUploadClient
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.ConfigLoader.Companion.loadUploadConfig
import util.DataProvider
import java.net.URLDecoder
import java.nio.charset.StandardCharsets
import java.time.ZonedDateTime
import java.util.TimeZone
import kotlin.collections.HashMap

@Listeners(UploadIdTestListener::class)
@Test()
class FileCopy {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val dexUploadClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private val environment = EnvConfig.ENVIRONMENT
    private lateinit var authToken: String
    private lateinit var testContext: ITestContext
    private lateinit var uploadClient: UploadClient

    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeFileCopy() {
        authToken = dexUploadClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.FILE_COPY],
        dataProvider = "validManifestAllProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldUploadFile(case: TestCase) {
        val uid = uploadClient.uploadFile(testFile, case.manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(5000)
        val uploadInfo = dexUploadClient.getFileInfo(uid, authToken)

        // Check File Info
        val expectedBytes = "10240"
        Assert.assertEquals(uploadInfo.fileInfo.sizeBytes.toString(), expectedBytes)

        // Check Upload Status
        Assert.assertEquals(uploadInfo.uploadStatus.status, "Complete", "File upload status is not 'Complete'")

        // Check Deliveries
        Assert.assertEquals(uploadInfo.deliveries?.size, case.deliveryTargets?.size, "Expected ${case.deliveryTargets?.size ?: 0 } deliveries")
        Assert.assertTrue(uploadInfo.deliveries?.all { it.status == "SUCCESS" }?:false, "Not all deliveries are 'SUCCESS' - Deliveries: ${uploadInfo.deliveries}")

        val expectedDeliveryNames = case.deliveryTargets?.map{ it.name }?.sorted()
        val actualDeliveryNames = uploadInfo.deliveries?.map{ it.name }?.sorted()
        Assert.assertEquals(actualDeliveryNames, expectedDeliveryNames, "Actual delivery targets do not match expected targets")

    
        val currentDateTime = ZonedDateTime.now(TimeZone.getTimeZone("GMT").toZoneId())
        
        uploadInfo.deliveries?.forEach { delivery ->
            Assert.assertEquals(delivery.status, "SUCCESS") // remove the assertion above?
            val actualLocation = URLDecoder.decode(delivery.location, StandardCharsets.UTF_8.toString())
            val pattern = case.deliveryTargets?.find{ it.name == delivery.name}?.pathTemplate?.get(environment)

            val expectedLocation = pattern
                ?.replace("{dataStream}", case.manifest["data_stream_id"].toString())
                ?.replace("{route}", case.manifest["data_stream_route"].toString())
                ?.replace("{year}", currentDateTime.year.toString() )
                ?.replace("{month}", String.format("%02d", currentDateTime.monthValue) )
                ?.replace("{day}", String.format("%02d", currentDateTime.dayOfMonth) )
                ?.replace("{hour}", String.format("%02d", currentDateTime.hour) )
                ?.replace("{filename}", case.manifest["received_filename"].toString())
                ?.replace("{uploadId}",uid)
            Assert.assertTrue(actualLocation.endsWith(expectedLocation.toString()), "Actual location ($actualLocation) does not end with the expected path: $expectedLocation")
            Assert.assertEquals(delivery.issues, null)
        }
    }

    @Test(
        groups = [Constants.Groups.FILE_COPY],
        dataProvider = "validManifestV1Provider",
        dataProviderClass = DataProvider::class
    )
    fun shouldTranslateMetadataGivenV1SenderManifest(case: TestCase) {
        val uid = uploadClient.uploadFile(testFile, case.manifest)
            ?: throw TestNGException("Error uploading file ${testFile.name}")
        testContext.setAttribute("uploadId", uid)
        Thread.sleep(3000)

        val uploadInfo = dexUploadClient.getFileInfo(uid, authToken)
        //Assert.assertTrue(uploadInfo.deliveries?.all { it.status == "SUCCESS" }?:false, "Not all deliveries are 'SUCCESS' - Deliveries: ${uploadInfo.deliveries}")
        case.manifest.forEach{(manifestKey, manifestValue) -> 
            Assert.assertEquals(uploadInfo.manifest[manifestKey], manifestValue.toString(), "Actual manifest value does not equal the expected manifest value")
       }
        Assert.assertNotNull(uploadInfo.manifest["dex_ingest_datetime"])
        Assert.assertEquals(uploadInfo.manifest["upload_id"], uid, "Upload ID on the manifest is not the expected upload ID")
    }
}
