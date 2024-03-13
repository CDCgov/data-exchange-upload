using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Azure;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    internal class BlobCopyHelper
    {
        private readonly ILogger _logger;

        public BlobCopyHelper(ILogger logger)
        {
            _logger = logger;
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
    }
}
