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

    private val environment: String = listOf("dev", "tst", "stg", "prd").find {
        EnvConfig.UPLOAD_URL.contains("api$it")
    } ?: throw IllegalArgumentException("Unknown environment URL: ${EnvConfig.UPLOAD_URL}")

    @BeforeTest(groups = [Constants.Groups.HEALTH_CHECK])
    fun beforeTest() {
        authToken = dexUploadClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
    }

    @Test(groups = [Constants.Groups.HEALTH_CHECK])
    fun shouldGetHealthCheck() {

        val commonServices = arrayOf(
            "Event Publishing processing-status-cosmos-db-report-sink-topics",
            "Tus storage",
            "Redis Locker",
            "Azure deliver target edav",
            "Azure deliver target routing",
            "Azure deliver target ehdi",
            "Azure deliver target eicr",
            "Azure deliver target ncird"
        )

        val envSpecificServices = arrayOf(
            "Event Publishing ocio-ede-$environment-upload-file-ready-topic",
            "ocio-ede-$environment-upload-file-ready-subscription Event Subscriber"
        )

        val expectedDependentServices = commonServices + envSpecificServices
        val healthCheck = dexUploadClient.getHealth(authToken)

        Assert.assertNotNull(healthCheck)
        Assert.assertEquals(healthCheck.status, "UP")
        Assert.assertEquals(
            healthCheck.services.size,
            expectedDependentServices.size,
            "Unexpected number of dependent services: ${healthCheck.services}"
        )

        val actualServices = healthCheck.services.map { it.service }

        Assert.assertEqualsNoOrder(
            actualServices.toTypedArray(),
            expectedDependentServices,
            buildString {
                append("The actual service is not matched with the expected service.\n")
                val expServices = expectedDependentServices.filter { it !in actualServices }
                val actServices = actualServices.filter { it !in expectedDependentServices }
                if (expServices.isNotEmpty()) {
                    append("Expected service: ${expServices.joinToString(", ")}\n")
                }
                if (actServices.isNotEmpty()) {
                    append("Actual service: ${actServices.joinToString(", ")}\n")
                }
            }
        )
    }
}