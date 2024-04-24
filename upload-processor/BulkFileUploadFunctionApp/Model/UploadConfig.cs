
using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record UploadConfig
    {
        [JsonPropertyName("metadata_config")] public MetadataConfig? MetadataConfig { get; init; }
        [JsonPropertyName("copy_config")] public CopyConfig? CopyConfig { get; init; }

        public static readonly UploadConfig Default = new UploadConfig()
        {
            MetadataConfig = null,
            CopyConfig = new CopyConfig()
            {
                FolderStructure = "date_YYYY_MM_DD",
                TargetEnums = new List<CopyTargetsEnum> { CopyTargetsEnum.edav, CopyTargetsEnum.routing }
            },
        };

        public UploadConfig() { }

        // If you want to initialize properties in the constructor, you can add parameters to the constructor
        public UploadConfig(MetadataConfig metadataConfig, CopyConfig copyConfig)
        {
            MetadataConfig = metadataConfig;
            CopyConfig = copyConfig;
        }
    }
    
}
