using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record Report
    {
        [JsonPropertyName("upload_id")] public string UploadId { get; init; }
        [JsonPropertyName("stage_name")] public string StageName { get; init; }
        [JsonPropertyName("data_stream_id")] public string DataStreamId { get; init; }
        [JsonPropertyName("data_stream_route")] public string DataStreamRoute { get; init; }
        [JsonPropertyName("content_type")] public string ContentType { get; init; }
        [JsonPropertyName("disposition_type")] public string DispositionType { get; init; }
        [JsonPropertyName("content")] public CopyContent Content { get; init; }
    }
}
