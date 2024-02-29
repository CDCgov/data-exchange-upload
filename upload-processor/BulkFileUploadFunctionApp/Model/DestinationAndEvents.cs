using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Model
{
    public record DestinationAndEvents
    {
        [JsonProperty("destination_id")]
        public string? destinationId { get; init; }
        
        [JsonProperty("ext_events")]
        public List<ExtEvent>? extEvents { get; init; }
    }

    public record ExtEvent
    {
        public string? name { get; init; }

        [JsonProperty("definition_filename")]
        public string? definitionFilename { get; init; }

        [JsonProperty("copy_targets")]
        public List<CopyTarget>? copyTargets { get; init; }
    }
    
    public record CopyTarget(string target);
}