using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataTransformContent : Content
    {
        [JsonPropertyName("action")] public string Action { get; set; }
        [JsonPropertyName("field")] public string Field { get; set; }
        [JsonPropertyName("value")] public string Value { get; set; }
    }
}
