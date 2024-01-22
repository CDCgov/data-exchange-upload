using BulkFileUploadFunctionApp.Model.ProcStatus;
using Microsoft.Extensions.Logging;
using System.Net.Http.Json;

namespace BulkFileUploadFunctionApp.Service
{
  internal class ProcStatusService : IProcStatService
  {
    private readonly Uri _baseUrl;
    private readonly HttpClient _httpClient;
    private readonly ILogger _logger;

    public ProcStatusService(string baseUrl, HttpClient httpClient, ILogger logger)
    {
      _baseUrl = new Uri(baseUrl);
      _httpClient = httpClient;
      _logger = logger;
    }

    public async Task<Trace> GetTraceByUploadId(string uploadId)
    {
      return await _httpClient.GetFromJsonAsync<Trace>($"api/trace/traceId/{uploadId}");
    }
  }
}
