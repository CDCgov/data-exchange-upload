using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record CopyConfig
    {
        public static string FILENAME_SUFFIX_NONE = "none";
        public static string FILENAME_SUFFIX_UID = "upload_id";
        [JsonPropertyName("filename_suffix")] public string? FilenameSuffix { get; init; }
        [JsonPropertyName("folder_structure")] public string? FolderStructure { get; init; }
        [JsonPropertyName("targets")] public List<string>? Targets { get; init; }
        public List<CopyTargetsEnum>? TargetEnums { get; set; }

    }
}
