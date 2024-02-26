using System.Text.Json.Serialization;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public class BlobCopyRetryEvent
    {
        [JsonPropertyName("copyRetryStage")]
        public BlobCopyStage copyRetryStage { get; set; }

        [JsonPropertyName("retryAttempt")]
        public int retryAttempt { get; set; }

        [JsonPropertyName("sourceBlobUri")]
        public string? sourceBlobUri { get; set; }
        
        [JsonPropertyName("dexContainerName")]
        public string? dexContainerName { get; set; }

        [JsonPropertyName("dexBlobFilename")]
        public string? dexBlobFilename { get; set; }

        [JsonPropertyName("fileMetadata")]
        public Dictionary<string, string>? fileMetadata { get; set; }

        public BlobCopyRetryEvent()
        {
            fileMetadata = new Dictionary<string, string>();
        }
    }
}