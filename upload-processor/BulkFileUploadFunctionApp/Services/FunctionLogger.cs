
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Services
{
    public class FunctionLogger<T> : IFunctionLogger<T>
{
    private readonly ILogger<T> _logger;

    public FunctionLogger(ILogger<T> logger)
    {
        _logger = logger;
    }

    public void LogInformation(string message)
    {
        _logger.LogInformation(message);
    }

    public void LogError(string message)
    {
        _logger.LogError(message);
    }

    public void LogError(Exception ex, string message)
    {
        _logger.LogError(ex, message);
    }

}
}