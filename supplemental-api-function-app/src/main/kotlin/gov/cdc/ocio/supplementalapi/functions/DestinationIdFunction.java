package gov.cdc.ocio.supplementalapi.functions;

import com.azure.identity.DefaultAzureCredentialBuilder;
import com.azure.storage.blob.BlobClient;
import com.azure.storage.blob.BlobClientBuilder;
import com.azure.storage.blob.specialized.BlobInputStream;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.microsoft.azure.functions.ExecutionContext;
import com.microsoft.azure.functions.HttpRequestMessage;
import com.microsoft.azure.functions.HttpResponseMessage;
import com.microsoft.azure.functions.HttpStatus;

import java.io.IOException;
import java.util.Map;
import java.util.Optional;

public class DestinationIdFunction {
    public HttpResponseMessage run(HttpRequestMessage<Optional<String>> request, final ExecutionContext context) {
        BlobClient blobClient = new BlobClientBuilder()
                .endpoint("https://ocioededataexchangedev.blob.core.windows.net") // TODO: Make env vars for url and container name.
                .credential(new DefaultAzureCredentialBuilder().build())
                .containerName("tusd-file-hooks")
                .blobName("allowed_destination_and_events.json")
                .buildClient();

        Map<String, Object> jsonMap = null;

        try (BlobInputStream blobInputStream = blobClient.openInputStream()) {
            blobInputStream.read();
            ObjectMapper mapper = new ObjectMapper();
            jsonMap = mapper.readValue(blobInputStream, Map.class);
        } catch (IOException e) {
            e.printStackTrace();
        }

        if (jsonMap == null) {
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
        return request.createResponseBuilder(HttpStatus.OK).header("Content-Type", "application/json").body(jsonMap).build();
    }
}
