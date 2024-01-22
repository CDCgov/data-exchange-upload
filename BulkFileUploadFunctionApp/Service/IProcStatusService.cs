using Model.ProcStatus.Trace;

namespace BulkFileUploadFunctionApp.Service
{
  public interface IProcStatService
  {
    string baseUrl;
    HttpClient httpClient;
    ILogger logger;

    Trace GetTraceByUploadId(string uploadID);
  }
}
