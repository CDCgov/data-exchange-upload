using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataItem
    {
        [JsonPropertyName("field")] public string Field { get; set; }
        [JsonPropertyName("value")] public string Value { get; set; }
    }
}
