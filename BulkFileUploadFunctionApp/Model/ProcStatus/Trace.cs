using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model.ProcStatus
{
  public class Trace
  {
    public static readonly Trace Default = new Trace();

    [JsonPropertyName("trace_id")]
    public string? traceId { get; set; }

    [JsonPropertyName("span_id")]
    public string? spanId { get; set; }

    [JsonPropertyName("upload_id")]
    public string? uploadId { get; set; }

    [JsonPropertyName("destination_id")]
    public string? destinationId { get; set; }

    [JsonPropertyName("event_type")]
    public string? eventType { get; set; }

    public Span[] spans { get; set; }
  }
}
