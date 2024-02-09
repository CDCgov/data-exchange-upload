using Azure.Storage.Blobs;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IStorageContentReader
    {
        string GetContent(BlobServiceClient blobServiceClient, string containerName, string blobName);
    }
}