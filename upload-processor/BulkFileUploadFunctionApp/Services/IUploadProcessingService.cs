using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        public Task ProcessBlob(string? blobCreatedUrl);

        public Task<TusInfoFile> GetTusFileInfo(string tusPayloadPathname);

        public Task<UploadConfig> GetUploadConfig(string destinationId, string eventType);

        public Task<string> CopyBlobFromTusToDex(string sourceBlobName, string destinationContainerName, string destinationBlobName, IDictionary<string, string> destinationMetadata);

        public Task<string> CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        public Task<string> CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    }
}