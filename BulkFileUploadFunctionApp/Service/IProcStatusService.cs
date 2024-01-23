using BulkFileUploadFunctionApp.Model.ProcStatus;

namespace BulkFileUploadFunctionApp.Service
{
  public interface IProcStatService
  {
    Task<Trace> GetTraceByUploadId(string uploadID);
    Task StartSpanForTrace(string traceId, string parentSpanId, string stageName);
    Task StopSpanForTrace(string traceId, string parentSpanId, string stageName);
  }
}
