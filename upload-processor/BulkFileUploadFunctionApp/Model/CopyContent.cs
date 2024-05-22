using System.Text.Json.Serialization;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public record CopyContent : Content
    {
        public static readonly String DEFAULT_SCHEMA_VERSION = "0.0.1";

        [JsonPropertyName("result")] public string Result { get; set; }
        [JsonPropertyName("destination")] public string Destination { get; set; }
        [JsonPropertyName("error_description")] public string? ErrorDescription { get; set; }

        public CopyContent(string result, string destination, string? errorDesc, string? schemaVersion)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.Result = result;
            Destination = destination;
            this.ErrorDescription = errorDesc;

            if (schemaVersion == null)
            {
                this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
            }
        }

        public CopyContent(string result, string destination)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.Result = result;
            Destination = destination;
            this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
        }
        public CopyContent(string result, string destination, string? errorDesc)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_STAGE_NAME;
            this.Result = result;
            Destination = destination;
            this.ErrorDescription = errorDesc;
            this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
        }

    }
}