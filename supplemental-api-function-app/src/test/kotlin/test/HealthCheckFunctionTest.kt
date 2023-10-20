package test

import com.azure.cosmos.*
import com.azure.cosmos.util.CosmosPagedIterable
import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpStatus

import gov.cdc.ocio.supplementalapi.functions.HealthCheckFunction
import io.mockk.every
import io.mockk.mockk

import org.mockito.Mockito.mock

import org.testng.Assert.*

import org.testng.annotations.BeforeMethod
import org.testng.annotations.Test

import java.util.*

import gov.cdc.ocio.supplementalapi.model.Item
import com.azure.cosmos.CosmosClient


class HealthCheckFunctionTest {

    private lateinit var request: HttpRequestMessage<Optional<String>>
    private lateinit var context: ExecutionContext

    @BeforeMethod
    fun setUp() {
        // Initialize any mock objects or dependencies needed for testing
        request = mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        context = mock(ExecutionContext::class.java)
    }

    // @Test
    // fun testOkStatusBack() {   
    //     val mockCosmosClient = mockk<CosmosClient>()
    //     val mockCosmosDb = mockk<CosmosDatabase>()
    //     val mockCosmosContainer = mockk<CosmosContainer>()

    //     val items = mockk<CosmosPagedIterable<Item>>()

    //     every { mockCosmosClient.getDatabase(any()) } returns mockCosmosDb
    //     every { mockCosmosDb.getContainer(any()) } returns mockCosmosContainer
    //     every { mockCosmosContainer.queryItems(any<String>(), any<CosmosQueryRequestOptions>(), Item::class.java) } returns items

    //     // Create a HealthCheckFunction instance
    //     val healthCheckFunction = HealthCheckFunction()

    //     // call HealthCheckFunction
    //     val response = healthCheckFunction.run(request, context, mockCosmosClient)

    //     assert(response == HttpStatus.OK)
    // }

    //  @Test
    // fun testFailureStatusBack() {   
    //     val mockCosmosClient = mockk<CosmosClient>()

    //     every { mockCosmosClient.getDatabase(any()) } throws (Exception("CosmosDB error"))

    //     // Create a HealthCheckFunction instance
    //     val healthCheckFunction = HealthCheckFunction()

    //     // call HealthCheckFunction
    //     val response = healthCheckFunction.run(request, context, mockCosmosClient)

    //     assert(response == HttpStatus.INTERNAL_SERVER_ERROR)
    // }
}