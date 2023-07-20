package gov.cdc.ocio.supplementalapi;

import java.util.*;
import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.sql.Connection;
import java.sql.DriverManger;
import java.sql.SQLException;
import gov.cdc.ocio.supplementalapi.functions.StatusForDestinationFunction;
import gov.cdc.ocio.supplementalapi.functions.StatusForTguidFunction;

public class FunctionJavaWrappers {

    @FunctionName("StatusForTguid")
    public HttpResponseMessage statusForTguid(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "status/{tguid}",
                    authLevel = AuthorizationLevel.FUNCTION) HttpRequestMessage<Optional<String>> request,
            @BindingName("tguid") String tguid,
            final ExecutionContext context) {
        return new StatusForTguidFunction().run(request, tguid, context);
    }

    @FunctionName("StatusForDestination")
    public HttpResponseMessage statusForDestination(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "status/destination/{destinationName}",
                    authLevel = AuthorizationLevel.FUNCTION) HttpRequestMessage<Optional<String>> request,
            @BindingName("destinationName") String destinationName,
            final ExecutionContext context) {
        return new StatusForDestinationFunction().run(request, destinationName, context);
    }

    @FunctionName("HealthCheckDev")
    public HttpResponseMessage healthCheckDev(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "health",
                    authLevel = AuthorizationLevel.ANONYMOUS) HttpRequestMessage<Optional<String>> request,
            final ExecutionContext context) {
        // Perform health checks specific to the dev environment
        boolean isHealthy = performHealthChecksDev();

        // Return appropriate response based on health status
        if (isHealthy) {
            return request.createResponseBuilder(HttpStatus.OK).body("API is healthy in dev environment").build();
        } else {
            return request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).body("API is not healthy in dev environment").build();
        }
    }

    private boolean performHealthChecksDev() {
        // Health check variables
        boolean isDatabaseConnected = false;
        boolean isExternalServiceHealthy = false;

        // Perform health checks for database connectivity in the dev environment
        try {
            // Assuming you have a method to check database connectivity in the dev environment
            isDatabaseConnected = checkDatabaseConnectivity();
        } catch (Exception e) {
            // Log or handle the exception if necessary
        }

        // Perform health checks for external service availability in the dev environment
        try {
            // Assuming you have a method to check external service health in the dev environment
            isExternalServiceHealthy = checkExternalServiceHealth();
        } catch (Exception e) {
            // Log or handle the exception if necessary
        }

        // You can add more health checks here as needed for the dev environment

        // Return true if all health checks pass, otherwise false
        return isDatabaseConnected && isExternalServiceHealthy;
    }

    private boolean checkDatabaseConnectivity() {
        // Assume you have a database connection configuration
        String dbUrl = "jdbc:mysql://localhost:3306/dev_database";
        String dbUsername = "dev_user";
        String dbPassword = "dev_password";

        try (Connection connection = DriverManager.getConnection(dbUrl, dbUsername, dbPassword)) {
            // If the connection is successful, return true to indicate database connectivity
            return true;
        } catch (SQLException e) {
            // If there's an exception (e.g., failed connection), log the error and return false
            // You can also handle the exception based on your application's requirements
            e.printStackTrace();
            return false;
        }
    }

    private boolean checkExternalServiceHealth() {
        // Define the URL of the external service you want to check
        String externalServiceUrl = "https://example.com/api/health";

        try {
            // Create a connection to the external service
            URL url = new URL(externalServiceUrl);
            HttpURLConnection connection = (HttpURLConnection) url.openConnection();

            // Set the request method to GET
            connection.setRequestMethod("GET");

            // Set a reasonable timeout for the connection (in milliseconds)
            int timeoutMillis = 5000;
            connection.setConnectTimeout(timeoutMillis);
            connection.setReadTimeout(timeoutMillis);

            // Send the request and check the response code
            int responseCode = connection.getResponseCode();
            connection.disconnect();

            // Consider the service healthy if the response code is in the 2xx range (e.g., 200 OK)
            return responseCode >= 200 && responseCode < 300;
        } catch (Exception e) {
            // If there's an exception (e.g., failed connection or timeout), log the error and return false
            // You can also handle the exception based on your application's requirements
            e.printStackTrace();
            return false;
        }
    }
}
