namespace BulkFileUploadFunctionApp.Model
{
    internal class UploadConfig
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

    }
    
}
