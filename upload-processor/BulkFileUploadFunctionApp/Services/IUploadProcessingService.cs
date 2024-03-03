using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        // public Task<bool> ProcessBlob(string blobCreatedUrl);

        public Task<CopyPreqs> GetCopyPreqs(string blobCreatedUrl);

        public Task CopyAll(CopyPreqs copyPreqs);

        public Task CopyFromDexToEdav(string uploadId, string destinationId, string eventType, string dexBlobUrl, string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        public Task CopyFromDexToRouting(string uploadId, string destinationId, string eventType, string dexBlobUrl, string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);

        public Task PublishRetryEvent(BlobCopyStage copyStage, string uploadId, string destinationId, string eventType, string sourceBlobUrl, string dexBlobUrl, string? dexContainerName, string? dexBlobFilename, Dictionary<string, string>? fileMetadata);
    }
}