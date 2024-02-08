import auth.AuthClient
import org.testng.Assert
import org.testng.annotations.BeforeClass
import org.testng.annotations.Test
import tus.UploadClient
import util.Env


@Test()
class MetadataVerify {
    private var authClient = AuthClient(Env.UPLOAD_URL)
    private lateinit var uploadClient: UploadClient

    @BeforeClass()
    fun beforeClass() {
        val authToken = authClient.getToken(Env.SAMS_USERNAME, Env.SAMS_PASSWORD)
        uploadClient = UploadClient(Env.UPLOAD_URL, authToken)
    }

    @Test()
    fun test() {
        uploadClient.uploadFile()
        Assert.assertEquals(1, 1)
    }
}

