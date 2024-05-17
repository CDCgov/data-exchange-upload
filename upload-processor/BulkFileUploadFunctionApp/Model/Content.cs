using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record Content
    {
        [JsonPropertyName("schema_name")] public string SchemaName { get; init; }
        [JsonPropertyName("schema_version")] public string SchemaVersion { get; init; }
    }
}
