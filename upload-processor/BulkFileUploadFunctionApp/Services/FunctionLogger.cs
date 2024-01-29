using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp
{
    public class FunctionLogger : IFunctionLogger
{
    private readonly ILogger _logger;

    public FunctionLogger(ILogger logger)
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