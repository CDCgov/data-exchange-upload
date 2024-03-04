using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        public Task<CopyPrereqs> GetCopyPrereqs(string blobCreatedUrl);

        public Task CopyAll(CopyPrereqs copyPrereqs);

        public Task CopyFromDexToEdav(CopyPrereqs copyPrereqs);
    
        public Task CopyFromDexToRouting(CopyPrereqs copyPrereqs);

        // public Task CopyFromDexToEdav(string uploadId, string destinationId, string eventType, string dexBlobUrl, string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        // public Task CopyFromDexToRouting(string uploadId, string destinationId, string eventType, string dexBlobUrl, string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);

        public Task PublishRetryEvent(BlobCopyStage copyStage, CopyPrereqs copyPrereqs);
    }
}