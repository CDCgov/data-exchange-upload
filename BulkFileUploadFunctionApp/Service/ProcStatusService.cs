using Microsoft.Extensions.Logging;
using System.Net.Http;
using Model.Trace;

namespace BulkFileUploadFunctionApp.Utils
{
  internal class ProcStatusService : IProcStatusService
  {
    private readonly string _baseUrl
    private readonly HttpClient _httpClient;
    private readonly ILogger _logger;

    public ProcStatusService(string baseUrl, HttpClient httpClient, ILogger logger)
    {
      _baseUrl = baseUrl;
      _httpClient = httpClient;
      _logger = logger;
    }

    public async Trace GetTraceByUploadId(string uploadId)
    {
      var response = await this.httpClient.GetAsync($"{_baseUrl}/api/trace/traceId/{uploadId}");
      httpResponse.EnsureSuccessfulStatusCode();

      try
      {
        return await response.Content.ReadAsAsync<Trace>();
      }
      catch
      {
        _logger.LogError($"Error deserializing HTTP response {response}")
      }
    }
  }
}
