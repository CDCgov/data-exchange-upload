using System.Net;
using Azure;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Model;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;
using System.Text.Json;
using Azure.Identity;


namespace BulkFileUploadFunctionApp
{
    public class HealthCheckFunction
    {
        private readonly string[] StorageNames = { "DEX Blob Container", "EDAV Blob Container", "Routing Blob Container" };
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;

        // Constructor
        public HealthCheckFunction(IBlobServiceClientFactory blobServiceClientFactory,
                                    IEnvironmentVariableProvider environmentVariableProvider)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;

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
                string storageAccountName = string.Empty;
                string storageAccountKey = string.Empty;
                string connectionString = string.Empty;
                string edavAzureStorageAccountName = string.Empty;
                string containerName = string.Empty;

                BlobServiceClient blobServiceClient = null;
                HealthCheckResult checkResult = null;

                // Perform health checks for each destination.
                foreach (var storage in StorageNames)
                {
                    if (storage == "EDAV Blob Container")
                    {
                        containerName = "dextesting-testevent1";
                        edavAzureStorageAccountName = _environmentVariableProvider.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME") ?? "";
                        blobServiceClient = new BlobServiceClient(
                         new Uri($"https://{edavAzureStorageAccountName}.blob.core.windows.net"),
                         new DefaultAzureCredential() // using Service Principal
                     );
                    }
                    else if (storage == "Routing Blob Container")
                    {
                        containerName = "test-routing";
                        // Retrieve the values of these environment variables
                        storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_NAME");
                        storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_KEY");
                        connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                        blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                    }
                    else
                    {
                        containerName = "dextesting-testevent1";
                        storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME");
                        storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY");
                        connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                        blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                    }

                    _logger.LogInformation($"Checking health for destination: {storage}");

                    checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient,_logger);

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

        private async Task<HealthCheckResult> CheckBlobStorageHealthAsync(string destination, string containerName,BlobServiceClient blobServiceClient,ILogger logger)
        {
            try
            {
                // Check connectivity by getting blob container reference.
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(containerName);

                // If successful, return a healthy result for this destination.
                logger.LogInformation($"Health check passed for container: {containerName}");
                return new HealthCheckResult(destination, "UP", "Healthy");
            }
            catch (RequestFailedException ex)
            {
                // In case of any exceptions, consider the destination down and return the error message.
                logger.LogError(ex, "Error occurred while checking {containerName} container health.");
                return new HealthCheckResult(destination, "DOWN", "Unhealthy");
            }
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
