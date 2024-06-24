import dex.DexUploadClient
import org.testng.Assert
import org.testng.annotations.BeforeTest
import org.testng.annotations.Test
import util.Constants
import util.EnvConfig

@Test
class Health {
    private val dexUploadClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private lateinit var authToken: String

    @BeforeTest(groups = [Constants.Groups.HEALTH_CHECK])
    fun beforeTest() {
        authToken = dexUploadClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @Test(groups = [Constants.Groups.HEALTH_CHECK])
    fun shouldGetHealthCheck() {
        val expectedDependentServices = arrayOf("Azure Service Bus", "Tus storage", "Redis Locker", "Azure deliver target dex", "Azure deliver target edav", "Azure deliver target routing")
        val healthCheck = dexUploadClient.getHealth(authToken)

        Assert.assertNotNull(healthCheck)
        Assert.assertEquals(healthCheck.status, "UP")
        Assert.assertEquals(healthCheck.services.size, expectedDependentServices.size, "Unexpected number of dependent services: ${healthCheck.services}")

        healthCheck.services.forEach {
            Assert.assertTrue(expectedDependentServices.contains(it.service))
            Assert.assertEquals(it.status, "UP")
        }
    }
}