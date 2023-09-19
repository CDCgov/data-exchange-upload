package test

import com.azure.cosmos.CosmosClient
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.functions.HealthCheckFunction
import org.mockito.Mockito.mock
import org.mockito.Mockito.`when`
import org.testng.Assert
import org.testng.annotations.BeforeMethod
import org.testng.annotations.Test

import java.util.*

class HealthCheckFunctionTest {
    private lateinit var request: HttpRequestMessage<Optional<String>>
    private lateinit var context: ExecutionContext

    @BeforeMethod
    fun setUp() {
        // Initialize any mock objects or dependencies needed for testing
        request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        context = mock(ExecutionContext::class.java)
    }
    @Test
    fun testRun() {


        // Create an instance of HealthCheckFunction
        val healthCheckFunction = HealthCheckFunction()

        // Call healthCheckFunction.run() with the mock objects
        val response = healthCheckFunction.run(request, context)

        // Assert the response as needed for the success scenario

        Assert.assertEquals(response , HttpStatus.OK)

    }

    @Test
    fun testRunFailure() {
        // Create an instance of HealthCheckFunction
        val healthCheckFunction = HealthCheckFunction()

        // Call healthCheckFunction.run() with the mock objects
        val response = healthCheckFunction.run(request, context)

        // Assert the response as needed for the failure scenario
        Assert.assertNotEquals(response,HttpStatus.INTERNAL_SERVER_ERROR)
        // You can also assert other properties of the response if needed
    }


}