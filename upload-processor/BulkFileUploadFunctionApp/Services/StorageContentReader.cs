using Azure.Storage.Blobs;

namespace BulkFileUploadFunctionApp.Services
{
    public class StorageContentReader : IStorageContentReader
    {
        public string GetContent(BlobServiceClient blobServiceClient, string containerName, string blobName)
        { 
            // Retrieve a reference to the container and blob within Azure Blob Storage.
            BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(containerName);
            BlobClient blobClient = containerClient.GetBlobClient(blobName);

            // Download the blob's contents as a string
            var downloadInfo = blobClient.DownloadContent();

            // Convert the downloaded blob content to a string and return it.
            string jsonContent = downloadInfo.Value.Content.ToString();

            return jsonContent;
        }
    }
}