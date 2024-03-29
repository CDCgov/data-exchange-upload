using Azure.Storage.Blobs;
using Azure.Identity;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Logging;
using System.Collections.Concurrent;

namespace BulkFileUploadFunctionApp.Services
{

    public interface IBlobServiceClientFactory
    {
        BlobServiceClient CreateInstance(string instanceName, string connectionString);
        BlobServiceClient CreateInstance(string instanceName, Uri serviceUri, DefaultAzureCredential credential);
    }

    // Implementation of the IBlobServiceClientFactory with Singleton Pattern
    public class BlobServiceClientFactory : IBlobServiceClientFactory
    {
        private readonly ConcurrentDictionary<string, BlobServiceClient> _instances = new ConcurrentDictionary<string, BlobServiceClient>();
        private readonly ILogger<BlobServiceClientFactory> _logger;

        public BlobServiceClientFactory(ILogger<BlobServiceClientFactory> logger)
        {
            _logger = logger ?? throw new ArgumentNullException(nameof(logger));
        }

        public BlobServiceClient CreateInstance(string instanceName, string connectionString)
        {
            return _instances.GetOrAdd(instanceName, name =>
            {
                _logger.LogInformation($"Creating new BlobServiceClient instance for {name}.");
                return new BlobServiceClient(connectionString);
            });
        }

        public BlobServiceClient CreateInstance(string instanceName, Uri serviceUri, DefaultAzureCredential credential)
        {
            return _instances.GetOrAdd(instanceName, name =>
            {
                _logger.LogInformation($"Creating new BlobServiceClient instance for {name}.");
                return new BlobServiceClient(serviceUri, credential);
            });
        }
    }

}