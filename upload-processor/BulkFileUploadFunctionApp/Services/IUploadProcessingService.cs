using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IUploadProcessingService
    {
        Task<CopyPrereqs> GetCopyPrereqs(string blobCreatedUrl);

        Task CopyAll(CopyPrereqs copyPrereqs);

        Task CopyFromDexToTargets(Dictionary<BlobCopyStage, AzureBlobWriter> writers, CopyPrereqs copyPrereqs);

        Task PublishRetryEvent(BlobCopyStage copyStage, CopyPrereqs copyPrereqs);

        Task<string> CopyFromTusToDex(AzureBlobWriter tusToDexBlobWriter);

        AzureBlobWriter CreateWriterForStage(BlobCopyStage stage, CopyPrereqs copyPrereqs);
    }
}