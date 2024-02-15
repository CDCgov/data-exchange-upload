using System;
using BulkFileUploadFunctionApp.Model;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Services
{
    public class ProcStatClient : IProcStatClient
    {
        private readonly HttpClient _httpClient;
        private readonly string _remoteServiceBaseUrl;
        private readonly ILogger<ProcStatClient> _logger;

        public ProcStatClient(HttpClient httpClient, ILogger<ProcStatClient> logger)
        {
            _httpClient = httpClient;
            _logger = logger;
        }

        public async Task<string> GetHealthCheck()
        {
            var response = await _httpClient.GetAsync("/api/health");
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            return responseBody;
        }
        public async Task CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload)
        {
            var content = new StringContent(payload.ToString());
            var response = await _httpClient.PostAsync($"/api/report/json/uploadId/{uploadId}?destinationId={destinationId}&eventType={eventType}", content);
            response.EnsureSuccessStatusCode();
        }

        public async Task<Trace> GetTraceByUploadId(string uploadId)
        {
            var response = await _httpClient.GetAsync($"/api/trace/uploadId/{uploadId}");
            _logger.LogInformation($"*****{response.Content.ReadAsStringAsync().Result}");

            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            // TODO: Handle empty body.
            return JsonConvert.DeserializeObject<Trace>(responseBody);
        }

        public async Task<string> StartSpanForTrace(string traceId, string parentSpanId, string stageName)
        {
            var response = await _httpClient.PutAsync($"/api/trace/startSpan/{traceId}/{parentSpanId}?stageName={stageName}", null);
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            return responseBody;
        }

        public async Task<string> StopSpanForTrace(string traceId, string parentSpanId)
        {
            var response = await _httpClient.PutAsync($"/api/trace/stopSpan/{traceId}/{parentSpanId}", null);
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            return responseBody;
        }
    }
}
