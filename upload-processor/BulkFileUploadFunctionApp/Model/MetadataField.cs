using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataField
    {
        [JsonProperty("field_name")] public string? FieldName { get; init; }
        [JsonProperty("compat_field_name")] public string? CompatFieldName { get; init; }
        [JsonProperty("default_value")] public string? DefaultValue { get; init; }
    }
}
