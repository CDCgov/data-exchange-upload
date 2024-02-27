using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataConfig
    {
        [JsonProperty("version")] public string? Version { get; init; }
        [JsonProperty("fields")] public List<MetadataField>? Fields { get; init; }
    }
}
