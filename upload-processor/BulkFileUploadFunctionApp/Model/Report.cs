using Newtonsoft.Json;
using System;

namespace BulkFileUploadFunctionApp.Model
{
    public record Report
    {
        public static readonly String DEFAULT_SCHEMA_VERSION = "0.0.1";

        [JsonProperty("schema_name")] public string SchemaName { get; init; }
        [JsonProperty("schema_version")] public string SchemaVersion { get; init; }
    }
}
