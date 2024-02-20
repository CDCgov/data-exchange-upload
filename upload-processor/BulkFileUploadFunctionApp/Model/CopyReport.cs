using System;
using BulkFileUploadFunctionApp.Utils;
using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record CopyReport : Report
    {
        [JsonProperty("file_source_blob_url")] public string FileSourceBlobUrl { get; set; }
        [JsonProperty("file_destination_blob_url")] public string FileDestinationBlobUrl { get; set; }
        [JsonProperty("result")] public string Result { get; set; }
        [JsonProperty("error_description")] public string? ErrorDescription { get; set; }

        public CopyReport(string sourceUrl, string destUrl, string result, string? errorDesc, string? schemaVersion)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.FileSourceBlobUrl = sourceUrl;
            this.FileDestinationBlobUrl = destUrl;
            this.Result = result;
            this.ErrorDescription = errorDesc;

            if (schemaVersion == null)
            {
                this.SchemaVersion = Report.DEFAULT_SCHEMA_VERSION;
            }
        }

        // TODO: add success fail enum and verification.
        public CopyReport(string sourceUrl, string destUrl, string result)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.FileSourceBlobUrl = sourceUrl;
            this.FileDestinationBlobUrl = destUrl;
            this.Result = result;
            this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
        }
        public CopyReport(string sourceUrl, string destUrl, string result, string? errorDesc)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.FileSourceBlobUrl = sourceUrl;
            this.FileDestinationBlobUrl = destUrl;
            this.Result = result;
            this.ErrorDescription = errorDesc;
            this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
        }

    }
}
