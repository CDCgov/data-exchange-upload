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
<<<<<<< HEAD
<<<<<<< HEAD
import java.util.logging.Logger;
=======
>>>>>>> d848de7 (returns array of destinations)
=======
import java.util.logging.Logger;
>>>>>>> 0f69193 (use context logger)
import java.util.stream.Collectors;

public class DestinationIdFunction {
    public HttpResponseMessage run(HttpRequestMessage<Optional<String>> request, final ExecutionContext context) {
<<<<<<< HEAD
<<<<<<< HEAD
        Logger logger = context.getLogger();

=======
>>>>>>> d848de7 (returns array of destinations)
=======
        Logger logger = context.getLogger();

>>>>>>> 0f69193 (use context logger)
        BlobClient blobClient = new BlobClientBuilder()
                .endpoint(System.getenv("DEX_STORAGE_ENDPOINT"))
                .connectionString(System.getenv("DEX_STORAGE_CONNECTION_STRING"))
                .containerName(System.getenv("TUS_HOOKS_CONTAINER_NAME"))
                .blobName(System.getenv("DESTINATIONS_FILE_NAME"))
                .buildClient();

<<<<<<< HEAD
<<<<<<< HEAD
        Destination[] destinations;
=======
        Destination[] destinations = null;
>>>>>>> d848de7 (returns array of destinations)
=======
        Destination[] destinations;
>>>>>>> 0f69193 (use context logger)

        try (ByteArrayOutputStream outputStream = new ByteArrayOutputStream()) {
            blobClient.downloadStream(outputStream);
            ObjectMapper mapper = new ObjectMapper();
            destinations = mapper.readValue(outputStream.toByteArray(), Destination[].class);
        } catch (IOException e) {
<<<<<<< HEAD
<<<<<<< HEAD
            logger.severe(e.getMessage());
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
=======
            e.printStackTrace();
>>>>>>> d848de7 (returns array of destinations)
=======
            logger.severe(e.getMessage());
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
>>>>>>> 0f69193 (use context logger)
        }

        if (destinations == null) {
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }

        List<String> destStrs = Arrays.stream(destinations).map(destination -> destination.destinationId).collect(Collectors.toList());

        return request.createResponseBuilder(HttpStatus.OK).header("Content-Type", "application/json").body(destStrs).build();
    }
}
