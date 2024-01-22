using BulkFileUploadFunctionApp.Model.ProcStatus;

namespace BulkFileUploadFunctionApp.Service
{
  public interface IProcStatService
  {
    Task<Trace> GetTraceByUploadId(string uploadID);
  }
}
