using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataField
    {
        [JsonPropertyName("field_name")] public string? FieldName { get; init; }
        [JsonPropertyName("compat_field_name")] public string? CompatFieldName { get; init; }
        [JsonPropertyName("default_value")] public string? DefaultValue { get; init; }
        [JsonPropertyName("allowed_values")] public List<string>? AllowedValues { get; init; }
        [JsonPropertyName("required")] public bool? Required {  get; init; }
        [JsonPropertyName("description")] public string? Description { get; init; }
    }
}
