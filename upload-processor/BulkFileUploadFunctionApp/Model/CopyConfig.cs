using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record CopyConfig
    {
        [JsonPropertyName("folder_structure")] public string? FolderStructure { get; init; }
        [JsonPropertyName("targets")] public List<string>? Targets { get; init; }
        public List<CopyTargetsEnum>? TargetEnums { get; set; }

    }
}
