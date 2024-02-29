using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record DestinationAndEvents
    {
        [JsonPropertyName("destination_id")]
        public string? destinationId { get; init; }
        
        [JsonPropertyName("ext_events")]
        public List<ExtEvent>? extEvents { get; init; }
    }

    public record ExtEvent
    {
        public string? name { get; init; }

        [JsonPropertyName("definition_filename")]
        public string? definitionFilename { get; init; }

        [JsonPropertyName("copy_targets")]
        public List<CopyTarget>? copyTargets { get; init; }
    }
    
    public record CopyTarget(string target);
}