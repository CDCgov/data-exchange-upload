package gov.cdc.ocio.cosmossync;

import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*;
import gov.cdc.ocio.cosmossync.functions.CosmosSyncFunction;
import gov.cdc.ocio.cosmossync.functions.HealthCheckFunction;
import java.util.Optional;

public class FunctionJavaWrappers {

    @FunctionName("HealthCheck")
    public HttpResponseMessage healthCheck(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "health",
                    authLevel = AuthorizationLevel.ANONYMOUS) HttpRequestMessage<Optional<String>> request,
            final ExecutionContext context) {
        return new HealthCheckFunction().run(request, context);
    }

    @FunctionName("CosmosQueueProcessor")
    public void run(
            @QueueTrigger(name = "msg",
                    queueName = "%CosmosSinkQueueName%",
                    connection = "StorageConnectionString") String message,
            final ExecutionContext context
    ) {
        new CosmosSyncFunction().run(context, message);
    }
}
