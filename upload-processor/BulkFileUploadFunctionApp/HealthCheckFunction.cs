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
using Microsoft.Extensions.Configuration.AzureAppConfiguration;
using Microsoft.Extensions.Configuration;


namespace BulkFileUploadFunctionApp
{
    public class HealthCheckFunction
    {
        private readonly string[] StorageNames = { "DEX Blob Container", "EDAV Blob Container", "Routing Blob Container" };
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;
        private readonly ILogger _logger;
        private readonly IFeatureManagementExecutor _featureManagementExecutor;
        private readonly IProcStatClient _procStatClient;

        // Constructor
        public HealthCheckFunction(IBlobServiceClientFactory blobServiceClientFactory,
                                    IEnvironmentVariableProvider environmentVariableProvider,
                                    ILoggerFactory loggerFactory,
                                    IFeatureManagementExecutor featureManagementExecutor,
                                    IProcStatClient procStatClient)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;
            _logger = loggerFactory.CreateLogger<HealthCheckFunction>();

            _featureManagementExecutor = featureManagementExecutor;
            _procStatClient = procStatClient;
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
            // Perform health checks for each destination.
            try
            {
                foreach (var storage in StorageNames)
                {
                    HealthCheckResultUtil healthCheckResultUtil = new HealthCheckResultUtil(_blobServiceClientFactory,
                                    _environmentVariableProvider,
                                    _logger);

                    HealthCheckResult checkResult = await healthCheckResultUtil.GetResult(storage);
                    healthCheckResponse.DependencyHealthChecks.Add(checkResult);
                }
            }
            catch (RequestFailedException ex)
            {
                // TODO: Append DOWN for a specific storage instead of 500 error.
                // Log error, respond with "Not Healthy!", and set response status to Internal Server Error (500)
                _logger.LogError(ex, "Error occurred while checking Blob storage container health.");
                response.StatusCode = HttpStatusCode.InternalServerError;
                return response;
            }


            // Perform health check for Processing Status.
            try
            {
                await _featureManagementExecutor
                .ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    if(_procStatClient == null)
                    {
                        throw new Exception("ProcStatClient is null");
                    }
                    HealthCheckResponse? procStatHealthCheck = await _procStatClient.GetHealthCheck();
                    if (procStatHealthCheck == null)
                    {
                        throw new Exception("ProcStatHealthCheck is null");
                    }
                    healthCheckResponse.DependencyHealthChecks.Add(procStatHealthCheck.ToHealthCheckResult(Constants.PROC_STAT_SERVICE_NAME));
                });
            } catch (Exception ex)
            {
                _logger.LogError("Error occured while getting PS API health.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                healthCheckResponse.DependencyHealthChecks.Add(new HealthCheckResult(Constants.PROC_STAT_SERVICE_NAME, "DOWN", ex.Message));
            }
            

            var endTime = DateTime.UtcNow;
            var duration = endTime - startTime;
            healthCheckResponse.TotalChecksDuration = duration.ToString("c");

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

    }

}
