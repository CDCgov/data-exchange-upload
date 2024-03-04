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

        public Task PublishRetryEvent(BlobCopyStage copyStage, CopyPrereqs copyPrereqs);
    }
}