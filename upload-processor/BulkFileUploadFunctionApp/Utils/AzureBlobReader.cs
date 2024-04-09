using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BulkFileUploadFunctionApp.Utils
{
    public class AzureBlobReader
    {
        private BlobServiceClient _svcClient { get; init; }

        public AzureBlobReader(BlobServiceClient svcClient)
        {
            _svcClient = svcClient;
        }

        public async Task<T?> Read<T>(string containerName, string blobName)
        {
            T? result;

            var sourceContainerClient = _svcClient.GetBlobContainerClient(containerName);
            BlobClient sourceBlob = sourceContainerClient.GetBlobClient(blobName);

            // Ensure that the source blob exists
            if (!await sourceBlob.ExistsAsync())
            {
                throw new Exception("File is missing");
            }

            using (var stream = await sourceBlob.OpenReadAsync())
            {
                result = await JsonSerializer.DeserializeAsync<T>(stream);
            }

            return result;
        }
 
    }
}
