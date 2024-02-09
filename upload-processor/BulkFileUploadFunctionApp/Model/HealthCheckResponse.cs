
namespace BulkFileUploadFunctionApp.Model
{
    public class HealthCheckResponse
    {
    public string Status { get; set; }
    public string TotalChecksDuration { get; set; } // Duration of all health checks combined.
    public List<HealthCheckResult> DependencyHealthChecks { get; set; } = new List<HealthCheckResult>(); // Results of individual dependency health checks.
}
    }
