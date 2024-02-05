using Azure.Storage.Blobs;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobServiceClientFactory
    {
        BlobServiceClient CreateBlobServiceClient(string connectionString);
    }
}