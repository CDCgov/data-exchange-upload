using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record Report
    {
        public static readonly String DEFAULT_SCHEMA_VERSION = "0.0.1";

        [JsonPropertyName("schema_name")] public string? SchemaName { get; init; }
        [JsonPropertyName("schema_version")] public string? SchemaVersion { get; init; }
    }
}
