using System;
using System.Text;
using BulkFileUploadFunctionApp.Model;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Services
{
    public class ProcStatClient : IProcStatClient
    {
        private readonly HttpClient _httpClient;
        private readonly ILogger<ProcStatClient> _logger;

        public ProcStatClient(
            HttpClient httpClient, 
            ILogger<ProcStatClient> logger)
        {
            _httpClient = httpClient;
            _logger = logger;
        }

        public async Task<HealthCheckResponse> GetHealthCheck()
        {
            var response = await _httpClient.GetAsync("/api/health");
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            // TODO: Handle empty body.
            return JsonConvert.DeserializeObject<HealthCheckResponse>(responseBody);
        }
        public async Task CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload)
        {
            var content = new StringContent(JsonConvert.SerializeObject(payload), Encoding.UTF8, "application/json");
            var response = await _httpClient.PostAsync($"/api/report/json/uploadId/{uploadId}?destinationId={destinationId}&eventType={eventType}&stageName={stageName}", content);
            response.EnsureSuccessStatusCode();
        }

        public async Task<Trace> GetTraceByUploadId(string uploadId)
        {
            var response = await _httpClient.GetAsync($"/api/trace/uploadId/{uploadId}");
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            // TODO: Handle empty body.
            return JsonConvert.DeserializeObject<Trace>(responseBody);
        }

        public async Task<Span> StartSpanForTrace(string traceId, string parentSpanId, string stageName)
        {
            var response = await _httpClient.PutAsync($"/api/trace/startSpan/{traceId}/{parentSpanId}?stageName={stageName}", null);
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            // TODO: Handle empty body.
            return JsonConvert.DeserializeObject<Span>(responseBody);
        }

        public async Task<string> StopSpanForTrace(string traceId, string childSpanId)
        {
            var response = await _httpClient.PutAsync($"/api/trace/stopSpan/{traceId}/{childSpanId}", null);
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            return responseBody;
        }
    }
}
