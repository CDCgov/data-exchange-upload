
using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record HealthCheckResult
    {
        [JsonPropertyName("service")]
        public string Service { get; set; }
        [JsonPropertyName("status")]
        public string Status { get; set; }
        [JsonPropertyName("health_issues")]
        public string HealthIssues { get; set; }

        // Constructor for easily setting properties upon creation
        public HealthCheckResult(string service, string status, string healthIssues = "")
        {
            Service = service ?? throw new ArgumentNullException(nameof(service), "Service cannot be null.");
            Status = status ?? throw new ArgumentNullException(nameof(status), "Status cannot be null.");
            HealthIssues = healthIssues;
        }
    }
}