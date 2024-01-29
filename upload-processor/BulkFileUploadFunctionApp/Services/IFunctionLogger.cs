using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IFunctionLogger
    {
    void LogInformation(string message);
    void LogError(string message);
    void LogError(Exception ex, string message);
    }
}