using System.Net;
using Azure;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Model;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;
using System.Text.Json;


namespace BulkFileUploadFunctionApp
{

    public class HealthCheckFunction
    {
        const string TestContainerName = "tusd-file-hooks";
        const string TestBlobName = "allowed_destination_and_events.json";
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;
        private readonly IStorageContentReader _storageContentReader;

        // Constructor
        public HealthCheckFunction(IBlobServiceClientFactory blobServiceClientFactory,
                               IEnvironmentVariableProvider environmentVariableProvider,
                               IStorageContentReader storageContentReader)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;
            _storageContentReader = storageContentReader;
        }

        [Function("HealthCheckFunction")]
        public async Task<HttpResponseData> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] HttpRequestData req,
            FunctionContext context)
        {
            var startTime = DateTime.UtcNow; // Record start time of the health check.
            var _logger = context.GetLogger<HealthCheckFunction>();
            _logger.LogInformation("Starting HealthCheckFunction");

            // Initialize the health check response
            var healthCheckResponse = new HealthCheckResponse
            {
                Status = "UP", // Default status.
                TotalChecksDuration = "",  // To be calculated.
                DependencyHealthChecks = new List<HealthCheckResult>() // List of checks to be performed.
            };

            //creating a response for a request and setting its status code to 200 (OK).
            var response = req.CreateResponse();
            response.StatusCode = HttpStatusCode.OK;

            try
            {
                //retrieve the values of these environment variables
                var storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME");
                var storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY");

                _logger.LogInformation("Container name-->" + TestContainerName);
                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";

                //establishing a connection to Azure Blob Storage and accessing a particular container by using the connection string.
                BlobServiceClient blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                
                // Retrieve content from a specific blob.
                string destinations = _storageContentReader.GetContent(blobServiceClient, TestContainerName, TestBlobName);

                _logger.LogInformation("destinations-->" + destinations);
                 
                // Deserialize JSON content to a list of destination and events.
                List<DestinationAndEvents> destinationAndEvents = JsonSerializer.Deserialize<List<DestinationAndEvents>>(destinations)!;
                
                // Concatenate destination names for health checks.
                List<string> listOfDestinations = GetConcatenatedDestinationEventNames(destinationAndEvents);
                 
                 // Perform health checks for each destination.
                foreach (var destination in listOfDestinations)
                {
                    HealthCheckResult checkResult = await CheckBlobStorageHealthAsync(destination, connectionString, blobServiceClient);
                    healthCheckResponse.DependencyHealthChecks.Add(checkResult);
                }

                var endTime = DateTime.UtcNow;
                var duration = endTime - startTime;
                healthCheckResponse.TotalChecksDuration = ToIso8601Duration(duration);

                _logger.LogInformation("TotalChecksDuration-->" + healthCheckResponse.TotalChecksDuration);

                // Determine overall status based on individual checks
                if (healthCheckResponse.DependencyHealthChecks.Any(d => d.Status == "DOWN"))
                {
                    healthCheckResponse.Status = "DOWN"; // Set overall status to DOWN if any check fails.
                }
                response.StatusCode = HttpStatusCode.OK;
                response.Headers.Add("Content-Type", "application/json");
                await response.WriteStringAsync(JsonSerializer.Serialize(healthCheckResponse));

                return response;
            }
            catch (RequestFailedException ex)
            {
                // Log error, respond with "Not Healthy!", and set response status to Internal Server Error (500)
                _logger.LogError(ex, "Error occurred while checking Blob storage container health.");
                response.StatusCode = HttpStatusCode.InternalServerError;
                return response;
            }
        }

        private async Task<HealthCheckResult> CheckBlobStorageHealthAsync(string destination, string connectionString, BlobServiceClient blobServiceClient)
        {
            try
            {
                // Check connectivity by getting blob container reference.
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(destination);

                // Verify the existence of the container as a lightweight operation to check health.
                await container.ExistsAsync();
                
                // If successful, return a healthy result for this destination.
                return new HealthCheckResult(destination, "UP", "Healthy");

            }
            catch (RequestFailedException ex)
            {    
                // In case of any exceptions, consider the destination down and return the error message.
                return new HealthCheckResult(destination, "DOWN", ex.Message);
            }
        }

        public List<string> GetConcatenatedDestinationEventNames(List<DestinationAndEvents> destinations)
        {
            // Create a list to hold concatenated destination and event names.
            var concatenatedList = new List<string>();
            if (destinations == null) return concatenatedList;
            
            // Iterate through each destination and its events, concatenating their names.
            foreach (var destination in destinations)
            {
                if (destination?.extEvents == null) continue;
                foreach (var extEvent in destination.extEvents)
                {
                    if (extEvent?.name == null) continue;
                    string concatenated = $"{destination.destinationId}-{extEvent.name}"; // Concatenate destination ID and event name.
                    concatenatedList.Add(concatenated); // Add the concatenated name to the list.
                }

            }

            return concatenatedList;
        }

        public static string ToIso8601Duration(TimeSpan timeSpan)
        {   
            // Convert TimeSpan to ISO format.
            return "P" +
                   (timeSpan.Days > 0 ? $"{timeSpan.Days}D" : "") +
                   "T" +
                   (timeSpan.Hours > 0 ? $"{timeSpan.Hours}H" : "") +
                   (timeSpan.Minutes > 0 ? $"{timeSpan.Minutes}M" : "") +
                   (timeSpan.Seconds > 0 || timeSpan.Milliseconds > 0 ? $"{timeSpan.Seconds}.{timeSpan.Milliseconds:D3}S" : "");
        }
    }

}
