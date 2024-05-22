using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record BulkMetadataTransform
    {
        [JsonPropertyName("action")] public string Action { get; set; }
        [JsonPropertyName("items")] public List<MetadataItem> Items { get; set; }
    }
}
