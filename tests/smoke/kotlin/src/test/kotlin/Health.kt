import dex.DexUploadClient
import org.testng.Assert
import org.testng.annotations.Test
import util.Constants
import util.EnvConfig
import util.TestFile
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import model.HealthResponse

@Test
class Health {
    private val dexUploadClient = DexUploadClient(EnvConfig.UPLOAD_URL)

    @Test(groups = [Constants.Groups.HEALTH_CHECK])
    fun shouldGetHealthCheck() {
        val mapper = jacksonObjectMapper()
        val healthcheckFile = TestFile.getResourceFile(EnvConfig.HEALTHCHECK_CASE).readBytes()
        val expectedHealthCheck : HealthResponse = mapper.readValue(healthcheckFile)

        val actualHealthCheck = dexUploadClient.getHealth()
        
        val actualHealthCheckServices = actualHealthCheck.services.sortedBy { it.service }
        val expectedHealthCheckServices = expectedHealthCheck.services.sortedBy { it.service }

        Assert.assertEquals(actualHealthCheck.status, expectedHealthCheck.status, "Actual healthcheck status is not the same as the expected status")
        Assert.assertEquals(actualHealthCheckServices, expectedHealthCheckServices, "Actual healthcheck services do not match expected services")
    }
}
