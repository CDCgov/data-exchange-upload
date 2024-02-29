using System;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IProcStatClient
    {
        Task<HealthCheckResponse?> GetHealthCheck();
        Task<Trace?> GetTraceByUploadId(string uploadId);
        Task<Span?> StartSpanForTrace(string traceId, string parentSpanId, string stageName);
        Task<string?> StopSpanForTrace(string traceId, string parentSpanId);
        Task<bool> CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload);
    }
}
