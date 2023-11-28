using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public class DestinationAndEvents
    {
        public string? destination_id { get; set; }
        
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
    
    public class CopyTarget
    {
        public string? target { get; set; }
    } 
}