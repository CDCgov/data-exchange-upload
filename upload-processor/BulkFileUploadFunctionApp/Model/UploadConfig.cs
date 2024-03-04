
using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record UploadConfig
    {
        [JsonPropertyName("filename_suffix")] public string? FilenameSuffix { get; init; }
        [JsonPropertyName("folder_structure")] public string? FolderStructure { get; init; }
        [JsonPropertyName("fixed_folder_path")] public string? FixedFolderPath { get; init; }
        [JsonPropertyName("metadata_config")] public MetadataConfig? MetadataConfig { get; init; }

        public static readonly UploadConfig Default = new UploadConfig()
        {
            FilenameSuffix = "clock_ticks",
            FolderStructure = "date_YYYY_MM_DD",
            FixedFolderPath = null
        };

    }
    
}
