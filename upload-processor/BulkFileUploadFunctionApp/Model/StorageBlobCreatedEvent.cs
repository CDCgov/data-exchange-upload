// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}

namespace BulkFileUploadFunctionApp.Model
{
    public class StorageBlobCreatedEventData
    {
        [JsonPropertyName("url")]
        public string? Url { get; set; }
    }

    public class StorageBlobCreatedEvent
    {
        [JsonPropertyName("id")]
        public string? Id { get; set; }
        [JsonPropertyName("topic")]
        public string? Topic { get; set; }
        [JsonPropertyName("subject")]
        public string? Subject { get; set; }
        [JsonPropertyName("eventType")]
        public string? EventType { get; set; }
        [JsonPropertyName("eventTime")]
        public DateTime EventTime { get; set; }
        [JsonPropertyName("data")]
        public StorageBlobCreatedEventData? Data { get; set; }
    }
}
