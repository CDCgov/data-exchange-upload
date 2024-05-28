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
@Test()
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

    @Test(groups = [Constants.Groups.FILE_INFO], dataProvider = "validManifestAllProvider", dataProviderClass = DataProvider::class)
    fun shouldGetFileInfo(manifest: HashMap<String, String>) {
        val uid: String = uploadClient.uploadFile(testFile, manifest) ?: throw TestNGException("Error uploading file given manifest $manifest")
        testContext.setAttribute("uploadId", uid)

        val fileInfo = dexUploadClient.getFileInfo(uid, authToken)

        Assert.assertNotNull(fileInfo)
        Assert.assertNotNull(fileInfo.manifest)
        Assert.assertEquals(fileInfo.fileInfo.sizeBytes, testFile.length())

        fileInfo.manifest.forEach {
            if (manifest.containsKey(it.key)) {
                Assert.assertEquals(it.value, manifest[it.key])
            }
        }
    }

    @Test(
        groups = [Constants.Groups.FILE_INFO],
        expectedExceptions = [IOException::class],
        expectedExceptionsMessageRegExp = "Error getting file info.*")
    fun shouldReturnNotFoundGivenInvalidId() {
        dexUploadClient.getFileInfo("blah", authToken)
    }
}