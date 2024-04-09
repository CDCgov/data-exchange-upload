// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
namespace BulkFileUploadFunctionApp
{
    public class JsonLoggerBase
    {

        // figure out a way to implement BeginScope<TState> without violating lint rule CS8633
        public IDisposable? BeginScope<TState>(TState state) => default;
    }
}