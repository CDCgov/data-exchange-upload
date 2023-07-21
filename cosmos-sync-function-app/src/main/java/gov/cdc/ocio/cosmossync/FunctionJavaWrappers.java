package gov.cdc.ocio.cosmossync;

import java.util.*;
import java.util.Optional;
import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*;
import com.microsoft.azure.functions.ExecutionContext;
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

    @FunctionName("HealthCheck")
    public HttpResponseMessage healthCheck(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "health") HttpRequestMessage<Optional<String>> request,
            final ExecutionContext context
    ) {
        // Perform your health check logic here
        boolean isHealthy = performHealthCheck();

        // Return appropriate response based on health status
        if (isHealthy) {
            return request.createResponseBuilder(HttpStatus.OK).body("Function is healthy").build();
        } else {
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).body("Function is not healthy").build();
        }
    }

    private boolean performHealthCheck() {
        boolean isDatabaseConnected = false;
        try {
            isDatabaseConnected = checkDatabaseConnectivity();
        } catch (Exception e) {
        }
        return isDatabaseConnected;
    }

    private boolean checkDatabaseConnectivity() {
        // Replace these values with your actual database connection details
        String dbUrl = "jdbc:mysql://your-database-url:3306/your-database-name";
        String dbUsername = "your-database-username";
        String dbPassword = "your-database-password";
    
        // try (Connection connection = DriverManager.getConnection(dbUrl, dbUsername, dbPassword)) {
        //     // If the connection is successful, return true to indicate database connectivity
            return true;
        // } catch (SQLException e) {
        //     // If there's an exception (e.g., failed connection), log the error and return false
        //     // You can also handle the exception based on your application's requirements
        //     e.printStackTrace();
        //     return false;
        // }
    }
}