using System;
using System.Security.Cryptography.X509Certificates;
using BulkFileUploadFunctionApp.Model;
using Newtonsoft.Json;

namespace BulkFileUploadFunctionApp.Services
{
    public class ProcStatClient : IProcStatClient
    {
        private readonly HttpClient _httpClient;
        private readonly string _remoteServiceBaseUrl;

        public ProcStatClient(HttpClient httpClient)
        {
            _httpClient = httpClient;
        }

        public async Task<string> GetHealthCheck()
        {
            var response = await _httpClient.GetAsync("/api/health");
            response.EnsureSuccessStatusCode();

            var responseBody = await response.Content.ReadAsStringAsync();
            return responseBody;
        }
        public void CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload)
        {
            throw new NotImplementedException();
        }

        public async Task<Trace> GetTraceByUploadId(string uploadId)
        {
            var response = await _httpClient.GetAsync($"/api/trace/uploadId/{uploadId}");
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
