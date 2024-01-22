using BulkFileUploadFunctionApp.Model.ProcStatus;
using System.Net.Http.Json;

namespace BulkFileUploadFunctionApp.Service
{
  internal class ProcStatusService : IProcStatService
  {
    private readonly HttpClient _httpClient;

    public ProcStatusService(HttpClient httpClient)
    {
      _httpClient = httpClient;
    }

    public async Task<Trace> GetTraceByUploadId(string uploadId)
    {
      return await _httpClient.GetFromJsonAsync<Trace>($"api/trace/traceId/{uploadId}");
    }

    public async Task StartSpanForTrace(string traceId, string parentSpanId, string stageName)
    {
      await _httpClient.GetAsync($"api/trace/addSpan/{traceId}/{parentSpanId}?stageName={stageName}&spanMark=start");
    }

    public async Task StopSpanForTrace(string traceId, string parentSpanId, string stageName)
    {
      await _httpClient.GetAsync($"api/trace/addSpan/{traceId}/{parentSpanId}?stageName={stageName}&spanMark=stop");
    }
  }
}
