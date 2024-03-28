using BulkFileUploadFunctionApp.Utils;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json.Serialization;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Model
{
    public record MetadataReport : Report
    {
        [JsonPropertyName("metadata")] public Dictionary<string, string>? Metadata { get; init; }

        public MetadataReport(Dictionary<string, string> metadata)
        {
            this.SchemaName = Constants.PROC_STAT_REPORT_METADATA_STAGE_NAME;
            this.SchemaVersion = DEFAULT_SCHEMA_VERSION;
            this.Metadata = metadata;
        }
    }
}
