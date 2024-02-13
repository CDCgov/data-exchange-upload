using System;
using BulkFileUploadFunctionApp.Model;

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

        public string GetTraceByUploadId(string uploadId)
        {
            throw new NotImplementedException();
        }

        public string StartSpanForTrace(string traceId, string parentSpanId, string stageName)
        {
            throw new NotImplementedException();
        }

        public string StopSpanForTrace(string traceId, string parentSpanId)
        {
            throw new NotImplementedException();
        }
    }
}
