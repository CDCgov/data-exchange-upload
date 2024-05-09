import dex.DexUploadClient
import io.tus.java.client.ProtocolException
import okio.IOException
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.BeforeTest
import org.testng.annotations.Listeners
import org.testng.annotations.Optional
import org.testng.annotations.Parameters
import org.testng.annotations.Test
import tus.UploadClient
import util.*

@Listeners(UploadIdTestListener::class)
@Test()
class Info {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val dexUploadClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient
    private lateinit var authToken: String
    private lateinit var uploadId: String
    private lateinit var senderManifest: HashMap<String, String>

    @Parameters("SENDER_MANIFEST", "USE_CASE")
    @BeforeTest(groups = [Constants.Groups.FILE_INFO])
    fun beforeTest(
        context: ITestContext,
        @Optional SENDER_MANIFEST: String?,
        @Optional("dextesting-testevent1") USE_CASE: String
    ) {
        authToken = dexUploadClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val senderManifestDataFile = if (SENDER_MANIFEST.isNullOrEmpty()) "$USE_CASE.properties" else SENDER_MANIFEST
        val propertiesFilePath = "properties/$USE_CASE/$senderManifestDataFile"
        senderManifest = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)

        uploadId = uploadClient.uploadFile(testFile, senderManifest) ?: throw TestNGException("Error uploading file ${testFile.name}")
        context.setAttribute("uploadId", uploadId)
    }

    @Test(groups = [Constants.Groups.FILE_INFO])
    fun shouldGetFileInfo() {
        val fileInfo = dexUploadClient.getFileInfo(uploadId, authToken)

        Assert.assertNotNull(fileInfo)
        Assert.assertNotNull(fileInfo.manifest)
        Assert.assertEquals(fileInfo.fileInfo.sizeBytes, testFile.length())

        fileInfo.manifest.forEach {
            if (senderManifest.containsKey(it.key)) {
                Assert.assertEquals(it.value, senderManifest[it.key])
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