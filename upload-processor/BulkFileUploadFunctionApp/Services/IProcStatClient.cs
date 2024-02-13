using System;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IProcStatClient
    {
        Task<string> GetHealthCheck();
        string GetTraceByUploadId(string uploadId);
        string StartSpanForTrace(string traceId, string parentSpanId, string stageName);
        string StopSpanForTrace(string traceId, string parentSpanId);
        void CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload);
    }
}
