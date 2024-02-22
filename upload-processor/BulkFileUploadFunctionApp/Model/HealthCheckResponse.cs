namespace BulkFileUploadFunctionApp.Model
{
    public class HealthCheckResponse
    {
        public string Status { get; set; }
        public string TotalChecksDuration { get; set; } // Duration of all health checks combined.
        public List<HealthCheckResult> DependencyHealthChecks { get; set; } = new();  // Results of individual dependency health checks.

        public HealthCheckResult ToHealthCheckResult(string serviceName)
        {
            HealthCheckResult result = new HealthCheckResult(serviceName, Status);

            // Aggregate health issues from dependent health checks.
            string healthIssues = string.Join(",", DependencyHealthChecks.Select(check => check.HealthIssues));
            result.HealthIssues = healthIssues;

            return result;
        }
    }
}
