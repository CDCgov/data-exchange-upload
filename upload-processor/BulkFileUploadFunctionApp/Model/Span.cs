using Newtonsoft.Json;
using System;

namespace BulkFileUploadFunctionApp.Model
{
    public record Span
    {
        [JsonProperty("trace_id")] public string? TraceId { get; set; }
        [JsonProperty("span_id")] public string? SpanId {  get; set; }
    }
}
