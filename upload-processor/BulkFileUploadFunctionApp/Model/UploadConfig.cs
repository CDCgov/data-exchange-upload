
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
            FixedFolderPath = null,
            MetadataConfig = null,
        };

        public UploadConfig() { }

        // If you want to initialize properties in the constructor, you can add parameters to the constructor
        public UploadConfig(string filenameSuffix, string folderStructure, string fixedFolderPath, MetadataConfig metadataConfig)
        {
            FilenameSuffix = filenameSuffix;
            FolderStructure = folderStructure;
            FixedFolderPath = fixedFolderPath;
            MetadataConfig = metadataConfig;
        }
    }
    
}
