package gov.cdc.ocio.supplementalapi.functions;

import com.azure.storage.blob.BlobClient;
import com.azure.storage.blob.BlobClientBuilder;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.microsoft.azure.functions.ExecutionContext;
import com.microsoft.azure.functions.HttpRequestMessage;
import com.microsoft.azure.functions.HttpResponseMessage;
import com.microsoft.azure.functions.HttpStatus;
import gov.cdc.ocio.supplementalapi.model.Destination;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.util.*;
import java.util.logging.Logger;
import java.util.stream.Collectors;

public class DestinationIdFunction {
    public HttpResponseMessage run(HttpRequestMessage<Optional<String>> request, final ExecutionContext context) {
        Logger logger = context.getLogger();

        BlobClient blobClient = new BlobClientBuilder()
                .endpoint(System.getenv("DexStorageEndpoint"))
                .connectionString(System.getenv("DexStorageConnectionString"))
                .containerName(System.getenv("TusHooksContainerName"))
                .blobName(System.getenv("DestinationsFileName"))
                .buildClient();

        Destination[] destinations;

        try (ByteArrayOutputStream outputStream = new ByteArrayOutputStream()) {
            blobClient.downloadStream(outputStream);
            ObjectMapper mapper = new ObjectMapper();
            destinations = mapper.readValue(outputStream.toByteArray(), Destination[].class);
        } catch (IOException e) {
            logger.severe(e.getMessage());
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        if (destinations == null) {
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        List<String> destStrs = Arrays.stream(destinations).map(destination -> destination.destinationId).collect(Collectors.toList());

        return request.createResponseBuilder(HttpStatus.OK).header("Content-Type", "application/json").body(destStrs).build();
    }
}
