// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}

namespace BulkFileUploadFunctionApp.Model
{
    public class StorageBlobCreatedEventData
    {
        public string? Url { get; set; }
    }

    public class StorageBlobCreatedEvent
    {
        public string? Id { get; set; }

        public string? Topic { get; set; }

        public string? Subject { get; set; }

        public string? EventType { get; set; }

        public DateTime EventTime { get; set; }

        public StorageBlobCreatedEventData? Data { get; set; }
    }
}
