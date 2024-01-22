using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model.ProcStatus
{
  public class Span
  {
    public static readonly Span Default = new Span();

    [JsonPropertyName("stage_name")]
    public string? stageName { get; set; }
    public string? timestamp { get; set; }
    public string? status { get; set; }
    [JsonPropertyName("ellapsed_millis")]
    public int? ellapsedMillis { get; set; }
  }
}
