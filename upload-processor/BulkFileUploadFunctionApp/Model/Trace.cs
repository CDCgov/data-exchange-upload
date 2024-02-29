using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record Trace
    {
        [JsonPropertyName("trace_id")] public string? TraceId { get; set; }
        [JsonPropertyName("span_id")] public string? SpanId { get; set; }
        [JsonPropertyName("upload_id")] public string? UploadId { get; set; }
        [JsonPropertyName("destination_id")] public string? DestinationId { get; set; }
        [JsonPropertyName("event_type")] public string? EventType { get; set; }
    }
}
