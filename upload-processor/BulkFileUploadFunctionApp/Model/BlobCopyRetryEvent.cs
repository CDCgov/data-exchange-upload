using System.Text.Json.Serialization;

using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public class BlobCopyRetryEvent
    {
        [JsonPropertyName("copyRetryStage")]
        public BlobCopyStage CopyRetryStage { get; set; }

        [JsonPropertyName("retryAttempt")]
        public int RetryAttempt { get; set; }

        [JsonPropertyName("copyPrereqs")]
        public CopyPrereqs? CopyPrereqs { get; set; }
    }
}