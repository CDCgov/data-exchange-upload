package gov.cdc.ocio.cosmossync.functions

import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*
import gov.cdc.ocio.cosmossync.functions.CosmosSyncFunction
import gov.cdc.ocio.cosmossync.functions.HealthCheckFunction
import java.util.Optional

class FunctionKotlinWrappers {
    @FunctionName("HealthCheck")
    fun healthCheck(
        @HttpTrigger(
            name = "req",
            methods = [HttpMethod.GET],
            route = "health",
            authLevel = AuthorizationLevel.ANONYMOUS
        ) request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext
    ): HttpResponseMessage {
        return HealthCheckFunction().run(request, context)
    }

    @FunctionName("CosmosQueueProcessor")
    fun run(
        @QueueTrigger(
            name = "msg",
            queueName = "%CosmosSinkQueueName%",
            connection = "StorageConnectionString"
        ) message: String,
        context: ExecutionContext
    ) {
        CosmosSyncFunction().run(context, message)
    }
}







