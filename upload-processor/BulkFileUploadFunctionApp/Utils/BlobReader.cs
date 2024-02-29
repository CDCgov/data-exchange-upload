using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BulkFileUploadFunctionApp.Utils
{
    internal class BlobReader
    {
        private readonly ILogger _logger;

        public BlobReader(ILogger logger)
        {
            _logger = logger;
        }

        public async Task<T?> GetObjectFromBlobJsonContent<T>(string connectionString, string sourceContainerName, string blobPathname)
        {
            T? result;
            var sourceContainerClient = new BlobContainerClient(connectionString, sourceContainerName);

            BlobClient sourceBlob = sourceContainerClient.GetBlobClient(blobPathname);

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
    }
}
