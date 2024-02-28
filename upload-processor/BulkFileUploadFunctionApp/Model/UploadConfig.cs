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

    }
    
}
