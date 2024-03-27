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
        private readonly IBlobClientFactory _blobClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;
        private readonly ILogger _logger;

        // Constructor
        public HealthCheckResultUtil(IBlobClientFactory blobServiceClientFactory,
                                    IEnvironmentVariableProvider environmentVariableProvider,
                                    ILogger logger)
        {
            _blobClientFactory = blobServiceClientFactory;
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
            Dictionary<string, string> _blobFileInfo = new Dictionary<string, string>();

            if (storage == "EDAV Blob Container")
            {
                _blobFileInfo.Add("connectionstring", connectionString);
                _blobFileInfo.Add("containername", "dextesting-testevent1");
                _blobFileInfo.Add("filename", "testevent1.json");
                _blobFileInfo.Add("destination", "EDAV Blob Container");
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
                blobServiceClient = await _blobClientFactory.CreateBlobServiceClientAsync(connectionString);
            }
            else
            {
                containerName = "dextesting-testevent1";
                storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME");
                storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY");
                connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                blobServiceClient = await _blobClientFactory.CreateBlobServiceClientAsync(connectionString);
            }

            _logger.LogInformation($"Checking health for destination: {storage}");

            checkResult = await CheckBlobStorageHealthAsync(storage, containerName, blobServiceClient);

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
    }
}