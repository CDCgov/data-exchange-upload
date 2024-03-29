using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        Task<CopyPrereqs> GetCopyPrereqs(string blobCreatedUrl);

        Task CopyAll(CopyPrereqs copyPrereqs);

        //Task CopyFromDexToEdav(CopyPrereqs copyPrereqs);
    
       // Task CopyFromDexToRouting(CopyPrereqs copyPrereqs);

        Task PublishRetryEvent(BlobCopyStage copyStage, CopyPrereqs copyPrereqs);

        Task<string> CopyFromTusToDex(CopyPrereqs copyPrereqs);
    }
}