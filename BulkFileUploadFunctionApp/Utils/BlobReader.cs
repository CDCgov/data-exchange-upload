using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    internal class BlobReader
    {
        private readonly ILogger _logger;

        public BlobReader(ILogger logger)
        {
            _logger = logger;
        }

        public async Task<T> GetObjectFromBlobJsonContent<T>(string connectionString, string sourceContainerName, string blobPathname)
        {
            var sourceContainerClient = new BlobContainerClient(connectionString, sourceContainerName);

            BlobClient sourceBlob = sourceContainerClient.GetBlobClient(blobPathname);

            _logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

            // Ensure that the source blob exists
            if (!await sourceBlob.ExistsAsync())
            {
                throw new Exception("File is missing");
            }

            _logger.LogInformation("File exists, getting lease on file");

            BlobLeaseClient lease = sourceBlob.GetBlobLeaseClient();

            // Specifying -1 for the lease interval creates an infinite lease
            await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

            BlobDownloadResult download = await sourceBlob.DownloadContentAsync();
            var objectData = download.Content.ToObjectFromJson<T>();

            BlobProperties sourceProperties = await sourceBlob.GetPropertiesAsync();

            if (sourceProperties.LeaseState == LeaseState.Leased)
            {
                // Release the lease on the source blob
                await lease.ReleaseAsync();
            }

            return objectData;
        }
    }
}
