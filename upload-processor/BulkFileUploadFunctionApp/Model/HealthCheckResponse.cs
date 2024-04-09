using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Model
{
    public record HealthCheckResponse
    {
        [JsonPropertyName("status")]
        public string? Status { get; set; }
        [JsonPropertyName("total_checks_duration")]
        public string? TotalChecksDuration { get; set; } // Duration of all health checks combined.
        [JsonPropertyName("dependency_health_checks")]
        public List<HealthCheckResult> DependencyHealthChecks { get; set; } = new();  // Results of individual dependency health checks.

        public HealthCheckResult ToHealthCheckResult(string serviceName)
        {
            if(string.IsNullOrEmpty(Status))
            {
                throw new InvalidOperationException("Health check status is not set.");
            }
            HealthCheckResult result = new HealthCheckResult(serviceName, Status);
            List<string> healthIssuesList = DependencyHealthChecks
                        .Select(check => check.HealthIssues)
                        .Where(issue => !string.IsNullOrEmpty(issue))
                        .ToList();

            // Aggregate health issues from dependent health checks.
            string healthIssues = string.Join(",", healthIssuesList);
            result.HealthIssues = healthIssues;

            return result;
        }
    }
}
