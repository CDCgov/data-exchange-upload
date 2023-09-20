package test

import com.azure.cosmos.CosmosClient
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
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

        // Create mock objects
        val request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        val context = mock(ExecutionContext::class.java)
        val cosmosClientManager = mock(CosmosClientManager::class.java) // Mock your CosmosClientManager

        // Create a HealthCheckFunction instance
        val healthCheckFunction = HealthCheckFunction()

        // Mock the behavior of your CosmosClientManager to return a mock CosmosClient
        val cosmosClient = mock(CosmosClient::class.java)
        `when`(CosmosClientManager.getCosmosClient()).thenReturn(cosmosClient)


        // Create an instance of HealthCheckFunction
       // val healthCheckFunction = HealthCheckFunction()

        // Call healthCheckFunction.run() with the mock objects
        val response = healthCheckFunction.run(request, context)

        // Assert the response as needed for the success scenario
        assert(response.status == HttpStatus.OK)

    }

    @Test
    fun testRunFailure() {
        // Create mock objects
        val request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        val context = mock(ExecutionContext::class.java)
        val cosmosClientManager = mock(CosmosClientManager::class.java) // Mock your CosmosClientManager

        // Create a HealthCheckFunction instance
        val healthCheckFunction = HealthCheckFunction()

        // Mock the behavior of your CosmosClientManager to throw an exception, simulating an errorCo
        `when`(CosmosClientManager.getCosmosClient()).thenThrow(RuntimeException("CosmosDB error"))

        // Mock other necessary objects and methods as needed for your specific scenario

        // Call the healthCheck function
        val response = healthCheckFunction.run(request, context)

        // Assertions based on the expected behavior of the function
        // Example: Verify that the response status code is HttpStatus.INTERNAL_SERVER_ERROR
        assert(response.status == HttpStatus.INTERNAL_SERVER_ERROR)
    }


}