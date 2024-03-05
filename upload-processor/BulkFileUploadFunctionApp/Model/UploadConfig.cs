using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record UploadConfig
    {
        [JsonProperty("filename_suffix")] public string? FilenameSuffix { get; init; }
        [JsonProperty("folder_structure")] public string? FolderStructure { get; init; }
        [JsonProperty("fixed_folder_path")] public string? FixedFolderPath { get; init; }
        [JsonProperty("metadata_config")] public MetadataConfig? MetadataConfig { get; init; }

        public static readonly UploadConfig Default = new UploadConfig()
        {
            FilenameSuffix = "clock_ticks",
            FolderStructure = "date_YYYY_MM_DD",
            FixedFolderPath = null
        };

        public UploadConfig() { }

        // If you want to initialize properties in the constructor, you can add parameters to the constructor
        public UploadConfig(string filenameSuffix, string folderStructure, string fixedFolderPath)
        {
            FilenameSuffix = filenameSuffix;
            FolderStructure = folderStructure;
            FixedFolderPath = fixedFolderPath;
        }
    }
    
}
