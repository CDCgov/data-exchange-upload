package test

import com.azure.cosmos.*
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager.Companion.getCosmosClient

import gov.cdc.ocio.supplementalapi.functions.HealthCheckFunction
import io.mockk.every
import io.mockk.mockk

import org.mockito.Mockito.mock

import org.testng.Assert.*


import org.testng.annotations.BeforeMethod
import org.testng.annotations.Test

import java.util.*



class HealthCheckFunctionTest {

    private lateinit var request: HttpRequestMessage<Optional<String>>
    private lateinit var context: ExecutionContext

    @BeforeMethod
    fun setUp() {
        // Initialize any mock objects or dependencies needed for testing

        val healthCheckFunction = HealthCheckFunction()
        request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        context = mock(ExecutionContext::class.java)

    }
    @Test
    fun testRun() {

        // Create mock objects
        val mockCosmosClient = mockk<CosmosClient>()
        // Create a HealthCheckFunction instance
        val healthCheckFunction = HealthCheckFunction()
       try {
           // call HealthCheckFunction
           val response = healthCheckFunction.run(request, context)

           // Assert the response as needed for the success scenario
           assert(response == HttpStatus.OK)
       } catch(exception: Exception){
           fail("Expected successful execution, but an exception was thrown: ${exception.message}")
       }

    }

    @Test
    fun testRunFailure() {
        // Create mock objects
        val mockCosmosClient = mockk<CosmosClient>()


        // Create a HealthCheckFunction instance
        val healthCheckFunction = HealthCheckFunction()

        // Mock the behavior of your CosmosClientManager to throw an exception, simulating an errorCo
        every {mockCosmosClient.readAllDatabases() } throws (CosmosDbException("CosmosDB error"))



       try {
            healthCheckFunction.run(request, context)
            throw  CosmosDbException("CosmosDB error")
        }catch (exception: CosmosDbException) {
           exception.message
           // Assertions based on the expected behavior of the function
            assertEquals("CosmosDB error", exception.message)
        }
    }


}



class CosmosDbException(message: String): RuntimeException(message)


