using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionAppTest.utils
{
    [TestClass]
    internal class MockUploadConfig
    {
        public string? FilenameMetadataField { get; set; }

        public string? FilenameSuffix { get; set; }

        public string? FolderStructure { get; set; }

        public string? FixedFolderPath { get; set; }

        public static readonly MockUploadConfig Default = new MockUploadConfig()
        {
            FilenameMetadataField = "filename",
            FilenameSuffix = "clock_ticks",
            FolderStructure = "date_YYYY_MM_DD",
            FixedFolderPath = null
        };

        public MockUploadConfig() { }

        public MockUploadConfig(string filenameMetadataField, string filenameSuffix, string folderStructure, string fixedFolderPath)
        {
            FilenameMetadataField = filenameMetadataField;
            FilenameSuffix = filenameSuffix;
            FolderStructure = folderStructure;
            FixedFolderPath = fixedFolderPath;
        }
    }
}
