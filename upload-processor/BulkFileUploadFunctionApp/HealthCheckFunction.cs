using System.Net;
using Azure;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;
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
        private readonly ILogger _logger;

        // Constructor
        public HealthCheckFunction(IBlobServiceClientFactory blobServiceClientFactory,
                                    IEnvironmentVariableProvider environmentVariableProvider,
                                    ILoggerFactory loggerFactory)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;
            _logger = loggerFactory.CreateLogger<HealthCheckFunction>();
        }

        [Function("HealthCheckFunction")]
        public async Task<HttpResponseData> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] HttpRequestData req,
            FunctionContext context)
        {
            var startTime = DateTime.UtcNow; // Record start time of the health check.            
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
                // Perform health checks for each destination.
                foreach (var storage in StorageNames)
                {
                    HealthCheckResultUtil healthCheckResultUtil = new HealthCheckResultUtil(_blobServiceClientFactory,
                                    _environmentVariableProvider,
                                    _logger);

                    HealthCheckResult checkResult = await healthCheckResultUtil.GetResult(storage);
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
