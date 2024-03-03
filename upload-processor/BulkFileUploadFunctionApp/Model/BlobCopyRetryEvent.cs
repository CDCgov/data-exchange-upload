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

        [JsonPropertyName("uploadId")]
        public string? uploadId { get; set; }

        [JsonPropertyName("sourceBlobUrl")]
        public string? sourceBlobUrl { get; set; }

        [JsonPropertyName("dexBlobUrl")]
        public string? dexBlobUrl { get; set; }

        [JsonPropertyName("destinationId")]
        public string? destinationId { get; set; }

        [JsonPropertyName("eventType")]
        public string? eventType { get; set; }
        
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