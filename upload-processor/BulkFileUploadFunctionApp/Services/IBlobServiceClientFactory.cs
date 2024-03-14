using Azure.Storage.Blobs;
using Azure.Identity;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobServiceClientFactory
    {
        BlobServiceClient CreateBlobServiceClient(string connectionString);
        BlobServiceClient CreateBlobServiceClient(Uri serviceUri, DefaultAzureCredential credential);
    }
}