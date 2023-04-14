package gov.cdc.ocio.cosmossync;

import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*;
import gov.cdc.ocio.cosmossync.functions.CosmosSyncFunction;

public class FunctionJavaWrappers {

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
