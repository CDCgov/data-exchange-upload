namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        public Task ProcessBlob(string? blobCreatedUrl);

        public Task CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        public Task CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    }
}