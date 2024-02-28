namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        public Task<bool> ProcessBlob(string blobCreatedUrl, bool retryAttempt);

        public Task<string> CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        public Task<string> CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    }
}