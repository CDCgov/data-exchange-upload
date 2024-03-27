using Azure.Identity;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobManagementService
    {

        Task<BlobClient> GetBlobClientAsync(BlobCopyStage stage, Dictionary<string, string> _blobFileInfo);
        Task<BlobClient> GetBlobClientAsync(Dictionary<string, string> _blobFileInfo);
        Task<BlobServiceClient> GetBlobServiceClientAsync(Dictionary<string, string> _blobFileInfo);
        Task<BlobServiceClient> GetBlobServiceClientAsync(Uri serviceUri, DefaultAzureCredential credential);
        Task<BlobContainerClient> GetBlobContainerClientAsync(BlobServiceClient svc, string containerName);
        Task<T?> GetObjectFromBlobJsonContent<T>(Dictionary<string, string> _blobFileInfo);
        Task CopyBlobLeaseAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null);
        Task CopyBlobStreamAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null);
    }

}
