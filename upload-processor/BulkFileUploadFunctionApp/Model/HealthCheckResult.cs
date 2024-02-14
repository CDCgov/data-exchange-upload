
namespace BulkFileUploadFunctionApp.Model
{
    public class HealthCheckResult
    {
        public string Service { get; set; }
        public string Status { get; set; }
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