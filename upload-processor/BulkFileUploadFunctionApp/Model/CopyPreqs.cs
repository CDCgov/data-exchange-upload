using System.Text.Json.Serialization;

using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp.Model
{
    public class CopyPreqs
    {
        private string uploadId;
        private string sourceBlobUrl;
        private string dexBlobUrl;
        private string tusPayloadFilename;
        private string destinationId;
        private string eventType;
        private string destinationContainerName;
        private string destinationBlobName;
        private Dictionary<string, string> destinationMetadata;
        private CopyTarget[] targets;
        private Trace trace;

        public string UploadId
        {
            get { return uploadId; }
            set { uploadId = value; }
        }

        public string SourceBlobUrl
        {
            get { return sourceBlobUrl; }
            set { sourceBlobUrl = value; }
        }     

        public string DexBlobUrl
        {
            get { return dexBlobUrl; }
            set { dexBlobUrl = value; }
        }        

        public string TusPayloadFilename
        {
            get { return tusPayloadFilename; }
            set { tusPayloadFilename = value; }
        }

        public CopyTarget[] Targets
        {
            get { return targets; }
            set { targets = value; }
        }

        public string DestinationId
        {
            get { return destinationId; }
            set { destinationId = value; }
        }

        public string EventType
        {
            get { return eventType; }
            set { eventType = value; }
        }

        public string DestinationContainerName
        {
            get { return destinationContainerName; }
            set { destinationContainerName = value; }
        }

        public string DestinationBlobName
        {
            get { return destinationBlobName; }
            set { destinationBlobName = value; }
        }

        public Dictionary<string, string> DestinationMetadata
        {
            get { return destinationMetadata; }
            set { destinationMetadata = value; }
        }

        public Trace Trace
        {
            get { return trace; }
            set { trace = value; }
        }

        public CopyPreqs()
        {
            this.uploadId = null;
            this.sourceBlobUrl = null;
            this.tusPayloadFilename = null;
            this.destinationId = null;
            this.eventType = null;
            this.destinationContainerName = null;
            this.destinationBlobName = null;
            this.destinationMetadata = null;
            this.targets = null;
            this.trace = null;        
        }
        public CopyPreqs(string uploadId, string sourceBlobUrl, string tusPayloadFilename, string destinationId, string eventType, string destinationContainerName, string destinationBlobName, Dictionary<string, string> destinationMetadata, CopyTarget[] targets, Trace trace)
        {
            this.uploadId = uploadId;
            this.sourceBlobUrl = sourceBlobUrl;
            this.tusPayloadFilename = tusPayloadFilename;
            this.destinationId = destinationId;
            this.eventType = eventType;
            this.destinationContainerName = destinationContainerName;
            this.destinationBlobName = destinationBlobName;
            this.destinationMetadata = destinationMetadata;
            this.targets = targets;
            this.trace = trace;
        }
    }
}