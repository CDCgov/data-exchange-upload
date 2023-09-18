package test

import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.functions.HealthCheckFunction
import org.mockito.Mock
import org.mockito.Mockito.mock
import org.mockito.MockitoAnnotations
import org.testng.annotations.AfterMethod
import org.testng.annotations.BeforeMethod
import org.testng.annotations.Test

import org.testng.Assert.*
import java.util.*

class HealthCheckFunctionTest {
    private lateinit var healthCheckFunction: HealthCheckFunction
    private lateinit var request: HttpRequestMessage<Optional<String>>
    private lateinit var context: ExecutionContext

    @BeforeMethod
    fun setUp() {
        //MockitoAnnotations.openMocks(this)
        val endpoint = System.getenv("CosmosDBEndpoint") ?: throw IllegalStateException("CosmosDBEndpoint environment variable is not set.")
        healthCheckFunction = HealthCheckFunction(endpoint)
        request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        context = mock(ExecutionContext::class.java)

    }
    @Test
    fun testRun() {

        // Create an instance of HealthCheckFunction
        //val healthCheckFunction = HealthCheckFunction()

        // Call healthCheckFunction.run() with the mock objects
        val response = healthCheckFunction.run(request, context)

        // Assert the response as needed for the success scenario
        assert(response.status == HttpStatus.OK)

    }


}