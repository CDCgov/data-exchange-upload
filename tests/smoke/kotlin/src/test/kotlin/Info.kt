import dex.DexUploadClient
import okio.IOException
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.BeforeMethod
import org.testng.annotations.BeforeTest
import org.testng.annotations.Listeners
import org.testng.annotations.Test
import tus.UploadClient
import util.*

@Listeners(UploadIdTestListener::class)
@Test
class Info {
    private val testFile = TestFile.getResourceFile("10KB-test-file")
    private val dexUploadClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient
    private lateinit var authToken: String
    private lateinit var testContext: ITestContext

    @BeforeTest(groups = [Constants.Groups.FILE_INFO])
    fun beforeTest() {
        authToken = dexUploadClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @BeforeMethod
    fun setupUpload(context: ITestContext) {
        testContext = context
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.FILE_INFO],
        dataProvider = "validManifestProvider",
        dataProviderClass = DataProvider::class
    )
    fun shouldGetFileInfo(case: TestCase) {
        val uid: String = uploadClient.uploadFile(testFile, case.manifest)
            ?: throw TestNGException("Error uploading file given manifest ${case.manifest}")
        testContext.setAttribute("uploadId", uid)

        val fileInfo = dexUploadClient.getFileInfo(uid, authToken)

        Assert.assertNotNull(fileInfo, "File info should not be null")
        Assert.assertEquals(fileInfo.fileInfo.sizeBytes, testFile.length(), "File size should match the uploaded file")

        Assert.assertNotNull(fileInfo.manifest, "Manifest should not be null")
        fileInfo.manifest.forEach {
            if (case.manifest.containsKey(it.key)) {
                Assert.assertEquals(it.value,case.manifest[it.key], "Manifest values should match")
            }
        }

        val uploadStatus = fileInfo.uploadStatus
        Assert.assertNotNull(uploadStatus, "Upload status should not be null")
        uploadStatus.let {
            Assert.assertEquals(it.status, "Complete", "Upload status should be 'Complete'")
            Assert.assertNotNull(it.chunkReceivedAt, "Chunk received timestamp should not be null")
        }

        if (fileInfo.deliveries != null) {
            Assert.assertTrue(fileInfo.deliveries.isNotEmpty(), "Deliveries list should not be empty")
            fileInfo.deliveries.forEach { delivery ->
                Assert.assertEquals(delivery.status, "SUCCESS", "Delivery status should be SUCCESS")

                val expectedFilename = fileInfo.manifest["received_filename"] ?: testFile.name
                Assert.assertTrue(
                    delivery.location.contains(uid) || delivery.location.contains(expectedFilename),
                    "Delivery location should contain either the uploadId or the filename"
                )
                Assert.assertNotNull(delivery.deliveredAt, "Delivery timestamp should not be null")
                Assert.assertTrue(delivery.issues?.isEmpty() ?: true, "There should be no issues in delivery")
            }
        }
    }

    @Test(
        groups = [Constants.Groups.FILE_INFO],
        expectedExceptions = [IOException::class],
        expectedExceptionsMessageRegExp = "Error getting file info.*"
    )
    fun shouldReturnNotFoundGivenInvalidId() {
        dexUploadClient.getFileInfo("blah", authToken)
    }
}
