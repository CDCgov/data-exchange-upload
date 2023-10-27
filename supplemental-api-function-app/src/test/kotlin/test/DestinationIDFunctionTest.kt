package test

import com.azure.storage.blob.BlobClient
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.functions.DestinationIdFunction
import gov.cdc.ocio.supplementalapi.utils.Blob
import io.mockk.every
import io.mockk.mockk
import io.mockk.mockkObject
import org.mockito.Mockito
import org.mockito.Mockito.doAnswer
import org.mockito.kotlin.any
import org.testng.Assert
import org.testng.annotations.BeforeMethod
import org.testng.annotations.Test
import utils.HttpResponseMessageMock.HttpResponseMessageBuilderMock
import java.io.File
import java.util.*
import kotlin.collections.ArrayList


class DestinationIDFunctionTest {
    private lateinit var request: HttpRequestMessage<Optional<String>>
    private lateinit var context: ExecutionContext
    private lateinit var mockBlobClient: BlobClient

    @BeforeMethod
    fun setUp() {
        // Initialize any mock objects or dependencies needed for testing
        request = Mockito.mock(HttpRequestMessage::class.java) as HttpRequestMessage<Optional<String>>
        context = Mockito.mock(ExecutionContext::class.java)
        // Mocking the Azure Blob Client so we don't connect to the actual remote blob.
        mockBlobClient = mockk()

        // Setup method invocation interception when createResponseBuilder is called to avoid null pointer on real method call.
        doAnswer { invocation ->
            val status = invocation.arguments[0] as HttpStatus
            HttpResponseMessageBuilderMock().status(status)
        }.`when`(request).createResponseBuilder(any())
    }

    @Test
    fun testShouldReturnArrayOfDestinationsGivenDestinationEventJson() {
        // Mocking the Blob utility class to avoid invoking the real Azure Blob Client's download function.
        mockkObject(Blob)
        val testBytes = File("./src/test/kotlin/data/destinations_and_events.json").inputStream().readBytes()

        every { Blob.toByteArray(any()) } returns testBytes

        val destinationIDFunction = DestinationIdFunction()
        val response = destinationIDFunction.run(request, context, mockBlobClient)

        Assert.assertEquals(response.status, HttpStatus.OK)
        Assert.assertEquals(response.body, arrayListOf("test destination 1", "test destination 2", "test destination 3"))
    }

    @Test
    fun testShouldReturnEmptyArrayWhenNoDestinationsFound() {
        // Mocking the Blob utility class to avoid invoking the real Azure Blob Client's download function.
        mockkObject(Blob)
        val emptyBytes = File("./src/test/kotlin/data/empty_file.json").inputStream().readBytes()

        every { Blob.toByteArray(any()) } returns emptyBytes

        val destinationIDFunction = DestinationIdFunction()
        val response = destinationIDFunction.run(request, context, mockBlobClient)

        Assert.assertEquals(response.status, HttpStatus.OK)
        Assert.assertTrue((response.body as ArrayList<String>).isEmpty())
    }
}