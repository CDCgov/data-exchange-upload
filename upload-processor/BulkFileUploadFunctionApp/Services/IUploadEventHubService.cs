using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadEventHubService
    {
        public Task PublishRetryEvent(BlobCopyRetryEvent blobCopyRetryEvent);

        public Task PublishReplayEvent(BlobCopyRetryEvent blobCopyRetryEvent);
    }
}