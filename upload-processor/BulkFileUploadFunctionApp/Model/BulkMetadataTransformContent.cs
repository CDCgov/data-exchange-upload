using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record BulkMetadataTransformContent : Content
    {
        [JsonPropertyName("transforms")] public BulkMetadataTransform Transforms {  get; init; }
    }
}
