using System;
using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record CopyReport
    {
        [JsonProperty("schema_name")] public string? SchemaName { get; set; }
        [JsonProperty("schema_version")] public string? SchemaVersion { get; set; }
        [JsonProperty("file_source_blob_url")] public string? FileSourceBlobUrl { get; set; }
        [JsonProperty("file_destination_blob_url")] public string? FileDestinationBlobUrl { get; set; }
        [JsonProperty("result")] public string? Result { get; set; }
        [JsonProperty("error_description")] public string? ErrorDescription { get; set; }
    }
}
