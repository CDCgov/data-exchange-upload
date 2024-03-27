using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Collections.Generic;
using Microsoft.Extensions.Azure;
using Azure.Identity;
using Microsoft.Identity.Client.Platforms.Features.DesktopOs.Kerberos;
using BulkFileUploadFunctionApp.Utils;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure;

namespace BulkFileUploadFunctionApp.Services
{
    public class BlobManagementService: IBlobManagementService
    {
        private readonly ILogger _logger;
        private readonly IBlobClientFactory _blobServiceClient;

        public BlobManagementService(ILoggerFactory loggerFactory, IBlobClientFactory blobServiceClient)
        {
            _logger = loggerFactory.CreateLogger<BlobManagementService>();
            _blobServiceClient = blobServiceClient;           
        }

        // For HealthCheck
        public async Task<BlobServiceClient> GetBlobServiceClientAsync(Dictionary<string, string> _blobFileInfo)
        {
            var connectionString = _blobFileInfo["connectionstring"];
            return await _blobServiceClient.CreateBlobServiceClientAsync(connectionString);
        }
        
        // For HealthCheck
        public async Task<BlobServiceClient> GetBlobServiceClientAsync(Uri serviceUri, DefaultAzureCredential credential)
        {
            if (serviceUri == null)
            {
                throw new ArgumentNullException(nameof(serviceUri));
            }
            if (credential == null)
            {
                throw new ArgumentNullException(nameof(credential));
            }
            return await _blobServiceClient.CreateBlobServiceClientAsync(serviceUri, credential);
        }

        // For PreReqs inner methods
        public async Task<BlobClient> GetBlobClientAsync(Dictionary<string, string> _blobFileInfo)
        {
            var connectionString = _blobFileInfo["connectionstring"];
            var containername = _blobFileInfo["containername"];
            var filename = _blobFileInfo["filename"];
            return await _blobServiceClient.CreateBlobClientAsync(connectionString, containername, filename);
        }
        // For CopyTo.. functions
        public async Task<BlobClient> GetBlobClientAsync(BlobCopyStage stage, Dictionary<string, string> _blobFileInfo)
        {
            var connectionString = _blobFileInfo["connectionstring"];
            switch (stage)
            {
                case BlobCopyStage.CopyToDex:
                    return  await _blobServiceClient.CreateBlobClientAsync(connectionString, "dexContainer", "dexBlob");
                case BlobCopyStage.CopyToEdav:
                    return await _blobServiceClient.CreateBlobClientAsync(connectionString, "edavContainer", "edavBlob");
                case BlobCopyStage.CopyToRouting:
                    return await _blobServiceClient.CreateBlobClientAsync(connectionString, "routingContainer", "routingBlob");
                default:
                    throw new ArgumentException($"Invalid stage: {stage}", nameof(stage));
            }
        }

        public async Task<T?> GetObjectFromBlobJsonContent<T>(Dictionary<string, string> _blobFileInfo)
        {
            T? result;

            BlobClient sourceBlob = await _blobServiceClient.CreateBlobClientAsync(_blobFileInfo["connectionstring"], _blobFileInfo["containername"], _blobFileInfo["filename"]);
            _logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

            // Ensure that the source blob exists
            if (!await sourceBlob.ExistsAsync())
            {
                throw new Exception("File is missing");
            }

            _logger.LogInformation("File exists, getting lease on file");

            using (var stream = await sourceBlob.OpenReadAsync())
            {
                result = await JsonSerializer.DeserializeAsync<T>(stream);
            }

            return result;
        }

        public async Task CopyBlobLeaseAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null)
        {
            // Lease the source blob for the copy operation 
            // to prevent another client from modifying it.
            BlobLeaseClient lease = sourceBlob.GetBlobLeaseClient();

            try
            {
                // Get the source blob's properties and display the lease state.
                BlobProperties sourceProperties = await sourceBlob.GetPropertiesAsync();

                _logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

                // Ensure that the source blob exists.
                if (await sourceBlob.ExistsAsync())
                {
                    _logger.LogInformation("File exists, getting lease on file");
                    // Specifying -1 for the lease interval creates an infinite lease.
                    await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

                    _logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");
                    _logger.LogInformation("Starting blob copy");

                    // Start the copy operation.
                    var sourceUriToUse = sourceSasBlobUri != null ? sourceSasBlobUri : sourceBlob.Uri;
                    await destinationBlob.StartCopyFromUriAsync(sourceUriToUse, destinationMetadata);

                    _logger.LogInformation("Finished blob copy");

                    // Get the destination blob's properties and display the copy status.
                    BlobProperties destProperties = await destinationBlob.GetPropertiesAsync();

                    _logger.LogInformation($"Copy status: {destProperties.CopyStatus}");
                    _logger.LogInformation($"Copy progress: {destProperties.CopyProgress}");
                    _logger.LogInformation($"Completion time: {destProperties.CopyCompletedOn}");
                    _logger.LogInformation($"Total bytes: {destProperties.ContentLength}");

                    // Update the source blob's properties.
                    sourceProperties = await sourceBlob.GetPropertiesAsync();
                }
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError(ex.Message);
            }
            finally
            {
                BlobProperties sourceProperties = await sourceBlob.GetPropertiesAsync();
                _logger.LogInformation($"Post-copy Lease state: {sourceProperties.LeaseState}");

                if (sourceProperties.LeaseState == LeaseState.Leased)
                {
                    // Release the lease on the source blob
                    await lease.ReleaseAsync();
                }
            }
        }

        public async Task CopyBlobStreamAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null)
        {
            using var sourceBlobStream = await sourceBlob.OpenReadAsync();
            {
                await destinationBlob.UploadAsync(sourceBlobStream, null, destinationMetadata);
            }
        }

        public async Task<BlobContainerClient> GetBlobContainerClientAsync(BlobServiceClient svc, string containerName)
        {
            return await _blobServiceClient.GetBlobContainerClientAsync(svc, containerName);
        }
    }

} 