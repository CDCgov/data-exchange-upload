namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        public Task<(string, string, string, string, Dictionary<string, string>)> CopyBlobToDex(string? blobCreatedUrl);

        public Task CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    
        public Task CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata);
    }
}