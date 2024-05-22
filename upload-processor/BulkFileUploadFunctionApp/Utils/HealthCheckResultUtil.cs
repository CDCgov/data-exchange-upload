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

namespace BulkFileUploadFunctionApp.Utils
{
    public class HealthCheckResultUtil
    {
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;
        private readonly ILogger _logger;

        // Constructor
        public HealthCheckResultUtil(IBlobServiceClientFactory blobServiceClientFactory,
                                    IEnvironmentVariableProvider environmentVariableProvider,
                                    ILogger logger)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;
            _logger = logger;
        }

        public async Task<HealthCheckResult> GetResult(string storage)
        {
            string storageAccountName = string.Empty;
            string storageAccountKey = string.Empty;
            string connectionString = string.Empty;
            string edavAzureStorageAccountName = string.Empty;
            string containerName = string.Empty;

            BlobServiceClient blobServiceClient = null;
            HealthCheckResult checkResult = null;

            if (storage == "EDAV Blob Container")
            {
                containerName = "dextesting-testevent1";
                edavAzureStorageAccountName = _environmentVariableProvider.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME") ?? "";
                blobServiceClient = new BlobServiceClient(
                 new Uri($"https://{edavAzureStorageAccountName}.blob.core.windows.net"),
                 new DefaultAzureCredential() // using Service Principal
                 checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient);
             );
            } else if (storage == "PS API Service Bus"){
                serviceBusName= "PS API Service Bus";
                string keyVaultUrl = "https://ocio-dev-upload-vault.vault.azure.net";
                string secretName = "ps-service-bus-connection-str";
                string connectionString = await GetServiceBusConnectionString(keyVaultUrl, secretName);    
                bool isServiceBusHealthy = await IsServiceBusHealthy(connectionString);
                checkResult = await CheckServiceBusHealthAsync(storage, serviceBusName, isServiceBusHealthy);

            }
            else if (storage == "Routing Blob Container")
            {
                containerName = "test-routing";
                // Retrieve the values of these environment variables
                storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_NAME");
                storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_KEY");
                connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient);
            }
            else
            {
                containerName = "dextesting-testevent1";
                storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME");
                storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY");
                connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient);
            }

            _logger.LogInformation($"Checking health for destination: {storage}");

           // checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient);

            return checkResult;

        }

        private async Task<HealthCheckResult> CheckBlobStorageHealthAsync(string destination, string containerName, BlobServiceClient blobServiceClient)
        {
            try
            {
                // Check connectivity by getting blob container reference.
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(containerName);

                // If successful, return a healthy result for this destination.
                _logger.LogInformation($"Health check passed for container: {containerName}");
                return new HealthCheckResult(destination, "UP", "Healthy");
            }
            catch (RequestFailedException ex)
            {
                // In case of any exceptions, consider the destination down and return the error message.
                _logger.LogError(ex, $"Error occurred while checking {containerName} container health.");
                return new HealthCheckResult(destination, "DOWN", "Unhealthy");
            }
        }

        private async Task<HealthCheckResult> CheckServiceBusHealthAsync(string destination, string serviceBusName, bool isServiceBusHealthy)
        {
           
               if (isServiceBusHealthy)
                 {
                    _logger.LogInformation($"Health check passed for Service Bus: {serviceBusName}");
                     return new HealthCheckResult(destination, "UP", "Healthy");                     
                 }
                else
                 {
                     _logger.LogError(ex, $"Error occurred while checking {serviceBusName} Service Bus health.");
                     return new HealthCheckResult(destination, "DOWN", "Unhealthy");
                 }
          
        }

        private static async Task<bool> IsServiceBusHealthy(string connectionString)
        {
         try
        {
        await using (var client = new ServiceBusClient(connectionString))
        {            
            await sender.SendMessageAsync(new ServiceBusMessage("Health check"));
            return true;
        }
        }
        catch (Exception ex)
        {
        _logger.LogError($"Failed to connect to the service bus: {ex.Message}");
        return false;
        }
       }
       }
}