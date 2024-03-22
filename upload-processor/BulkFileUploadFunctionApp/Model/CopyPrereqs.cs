using System.Text.Json.Serialization;

using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public class CopyPrereqs
    {
        public string? UploadId { get; set; }
        public string? SourceBlobUrl { get; set; }
        public string? TusPayloadFilename { get; set; }
        public string? DexBlobUrl { get; set; }
        public string? DestinationId { get; set; }
        public string? EventType { get; set; }
        public string? DexBlobFolderName { get; set; }
        public string? DexBlobFileName { get; set; }
        public Dictionary<string, string>? Metadata { get; set; }
        public CopyTarget[]? Targets { get; set; }
        public Trace? Trace { get; set; }

        public CopyPrereqs() { }

        // create a Default constructor
        public CopyPrereqs(string uploadId, string sourceBlobUrl, string tusPayloadFilename, string destinationId, string eventType, string dexBlobFolderName, string dexBlobFileName, Dictionary<string, string> metadata, CopyTarget[] targets, Trace trace)
        {
            UploadId = uploadId;
            SourceBlobUrl = sourceBlobUrl;
            TusPayloadFilename = tusPayloadFilename;
            DestinationId = destinationId;
            EventType = eventType;
            DexBlobFolderName = dexBlobFolderName;
            DexBlobFileName = dexBlobFileName;
            Metadata = metadata;
            Targets = targets;
            Trace = trace;
        }
    }
}