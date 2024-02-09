
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
        Service = service;
        Status = status;
        HealthIssues = healthIssues;
    }
    }
}