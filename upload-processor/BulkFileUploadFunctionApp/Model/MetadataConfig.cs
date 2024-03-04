using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataConfig
    {
        [JsonPropertyName("version")] public string? Version { get; init; }
        [JsonPropertyName("fields")] public List<MetadataField>? Fields { get; init; }
    }
}
