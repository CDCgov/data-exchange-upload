using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataField
    {
        [JsonProperty("field_name")] public string? FieldName { get; set; }
        [JsonProperty("compat_field_name")] public string? CompatFieldName { get; set; }
        [JsonProperty("default_value")] public string? DefaultValue { get; set; }
    }
}
