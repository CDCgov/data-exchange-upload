using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public class DestinationAndEvents
    {
        [JsonPropertyName("destination_id")]
        public string? destinationId { get; set; }

        [JsonPropertyName("ext_events")]
        public List<ExtEvent>? extEvents { get; set; }

        public static readonly DestinationAndEvents Default = new DestinationAndEvents();
    }

    public class ExtEvent
    {
        public string? name { get; set; }

        [JsonPropertyName("definition_filename")]
        public string? definitionFilename { get; set; }

        [JsonPropertyName("copy_targets")]
        public List<CopyTarget>? copyTargets { get; set; }
    }

    public record CopyTarget(string target);
}