using Newtonsoft.Json;
using System;

namespace BulkFileUploadFunctionApp.Model
{
    public record Trace
    {
        [JsonProperty("trace_id")] public string? TraceId { get; set; }
        [JsonProperty("span_id")] public string? SpanId { get; set; }
        [JsonProperty("upload_id")] public string? UploadId { get; set; }
        [JsonProperty("destination_id")] public string? DestinationId { get; set; }
        [JsonProperty("event_type")] public string? EventType { get; set; }
    }
}
