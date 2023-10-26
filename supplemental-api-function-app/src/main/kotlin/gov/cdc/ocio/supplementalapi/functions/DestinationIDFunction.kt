package gov.cdc.ocio.supplementalapi.functions

import com.azure.storage.blob.BlobClientBuilder
import com.fasterxml.jackson.databind.ObjectMapper
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.model.Destination
import java.io.ByteArrayOutputStream
import java.io.IOException
import java.util.*
import java.util.stream.Collectors

class DestinationIdFunction {
    fun run(request: HttpRequestMessage<Optional<String>>, context: ExecutionContext): HttpResponseMessage {
        val logger = context.logger

        val blobClient = BlobClientBuilder()
            .endpoint(System.getenv("DexStorageEndpoint"))
            .connectionString(System.getenv("DexStorageConnectionString"))
            .containerName(System.getenv("TusHooksContainerName"))
            .blobName(System.getenv("DestinationsFileName"))
            .buildClient()

        var destinations: Array<Destination>

        try {
            ByteArrayOutputStream().use { outputStream ->
                blobClient.downloadStream(outputStream)
                val mapper = ObjectMapper()
                destinations = mapper.readValue(
                    outputStream.toByteArray(),
                    Array<Destination>::class.java
                )
            }
        } catch (e: IOException) {
            logger.severe(e.message)
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build()
        }

        val destStrs = Arrays.stream(destinations).map { destination: Destination -> destination.destinationId }
            .collect(Collectors.toList())

        return request.createResponseBuilder(HttpStatus.OK).header("Content-Type", "application/json").body(destStrs)
            .build()
    }
}