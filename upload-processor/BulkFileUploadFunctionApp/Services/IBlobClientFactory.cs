using Azure.Storage.Blobs;
using Azure.Identity;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobClientFactory
    {
        Task<BlobClient> CreateBlobClientAsync(string connectionString, string sourceContainerName, string blobName);
        Task<BlobClient> CreateBlobClientAsync(Uri serviceUri, DefaultAzureCredential credential);
        Task<BlobServiceClient> CreateBlobServiceClientAsync(Uri serviceUri, DefaultAzureCredential credential);
        Task<BlobServiceClient> CreateBlobServiceClientAsync(string connectionString);
        Task<BlobContainerClient> GetBlobContainerClientAsync(BlobServiceClient svc, string containerName);
    }

    public class BlobClientFactory : IBlobClientFactory
    {
        private readonly ILogger<BlobClientFactory> _logger;
        public BlobClientFactory(ILogger<BlobClientFactory> logger)
        {
            _logger = logger;
        }
        private async Task<BlobClient> BuildBlobServiceClientAsync(string connectionString, string sourceContainerName, string blobName)
        {
            var sourceContainerClient = new BlobContainerClient(connectionString, sourceContainerName);
            BlobClient sourceBlob = sourceContainerClient.GetBlobClient(blobName);
            return await Task.FromResult(sourceBlob);
        }
        public async Task<BlobClient> CreateBlobClientAsync(string connectionString, string sourceContainerName, string blobName)
        {
            return await BuildBlobServiceClientAsync(connectionString, sourceContainerName, blobName);
        }

        public async Task<BlobClient> CreateBlobClientAsync(Uri serviceUri, DefaultAzureCredential credential)
        {
            if (serviceUri == null)
            {
                _logger.LogError("Call to BlobClientFactory returned empty body.");
                throw new ArgumentNullException(nameof(serviceUri));
            }
            if (credential == null)
            {
                _logger.LogError("Call to BlobClientFactory returned empty body.");
                throw new ArgumentNullException(nameof(credential));
            }
            return await Task.FromResult(new BlobClient(serviceUri, credential));
        }
        public async Task<BlobServiceClient> CreateBlobServiceClientAsync(Uri serviceUri, DefaultAzureCredential credential)
        {
            if (serviceUri == null)
            {
                throw new ArgumentNullException(nameof(serviceUri));
            }
            if (credential == null)
            {
                throw new ArgumentNullException(nameof(credential));
            }
            return await Task.FromResult(new BlobServiceClient(serviceUri, credential));
        }
        public async Task<BlobServiceClient> CreateBlobServiceClientAsync(string connectionString)
        {
            if (string.IsNullOrEmpty(connectionString))
            {
                // Handle the case where connectionString is null or empty.
                // You might want to throw an exception or handle it in another way.
                throw new ArgumentException("Connection string cannot be null or empty.");
            }

            return await Task.FromResult(new BlobServiceClient(connectionString));
        }

        public async Task<BlobContainerClient> GetBlobContainerClientAsync(BlobServiceClient svc, string containerName)
        {
            BlobContainerClient blobContainer = svc.GetBlobContainerClient(containerName);
            await blobContainer.CreateIfNotExistsAsync();
            return blobContainer;
        }
    }
}