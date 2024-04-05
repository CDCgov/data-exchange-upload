using System;

namespace BulkFileUploadFunctionApp.Utils
{
    public static class Constants
    {
        public const string PROC_STAT_REPORT_STAGE_NAME = "dex-file-copy";
        public const string PROC_STAT_REPORT_METADATA_STAGE_NAME = "dex-file-metadata";
        public const string PROC_STAT_FEATURE_FLAG_NAME = "PROCESSING_STATUS_TRACES";
        public const string ROUTING_FEATURE_FLAG_NAME = "ROUTING";
        public const string PROC_STAT_SERVICE_NAME = "Processing Status API";
        public const string METADATA_VERSION_ONE = "1.0";
        public const string METADATA_VERSION_TWO = "2.0";
        public const string USE_CASE_FIELDNAME_V1 = "meta_destination_id";
        public const string USE_CASE_CATEGORY_FIELDNAME_V1 = "meta_ext_event";
        public const string USE_CASE_FIELDNAME_V2 = "data_stream_id";
        public const string USE_CASE_CATEGORY_FIELDNAME_V2 = "data_stream_route";
    }
}