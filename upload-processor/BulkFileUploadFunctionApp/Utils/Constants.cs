using System;

namespace BulkFileUploadFunctionApp.Utils
{
    public static class Constants
    {
        public static readonly string PROC_STAT_REPORT_STAGE_NAME = "dex-file-copy";
        public static readonly string PROC_STAT_REPORT_METADATA_STAGE_NAME = "dex-file-metadata";
        public static readonly string PROC_STAT_FEATURE_FLAG_NAME = "PROCESSING_STATUS";
        public static readonly string ROUTING_FEATURE_FLAG_NAME = "ROUTING";
        public static readonly string PROC_STAT_SERVICE_NAME = "Processing Status API";
    }
}