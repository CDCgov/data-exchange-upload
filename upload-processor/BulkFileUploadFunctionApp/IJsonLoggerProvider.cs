// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp
{
    public interface IJsonLoggerProvider
    {
        ILogger CreateLogger(string categoryName);
        void Dispose();
    }
}