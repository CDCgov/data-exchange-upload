namespace BulkFileUploadFunctionApp.Model
{
    public class UploadConfig
    {
        public string? FilenameMetadataField { get; set; }

        public string? FilenameSuffix { get; set; }

        public string? FolderStructure { get; set; }

        public string? FixedFolderPath { get; set; }

        public static readonly UploadConfig Default = new UploadConfig()
        {
            FilenameMetadataField = "filename",
            FilenameSuffix = "clock_ticks",
            FolderStructure = "date_YYYY_MM_DD",
            FixedFolderPath = null
        };

        public UploadConfig() { }

        // If you want to initialize properties in the constructor, you can add parameters to the constructor
        public UploadConfig(string filenameMetadataField, string filenameSuffix, string folderStructure, string fixedFolderPath)
        {
            FilenameMetadataField = filenameMetadataField;
            FilenameSuffix = filenameSuffix;
            FolderStructure = folderStructure;
            FixedFolderPath = fixedFolderPath;
        }
    }
    
}
