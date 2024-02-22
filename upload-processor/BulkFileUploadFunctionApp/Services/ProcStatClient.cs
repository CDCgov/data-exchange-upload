using System;
using System.Text;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;
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
            try
            {
                var response = await _httpClient.GetAsync("/api/health");
                response.EnsureSuccessStatusCode();

                var responseBody = await response.Content.ReadAsStringAsync();
                if (string.IsNullOrEmpty(responseBody))
                {
                    _logger.LogError("Call to PS API returned empty body.");
                    return new HealthCheckResponse()
                    {
                        Status = "DOWN"
                    };
                }
                return JsonConvert.DeserializeObject<HealthCheckResponse>(responseBody);
            } catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return new HealthCheckResponse()
                {
                    Status = "DOWN"
                };
            }
        }
        public async Task<bool> CreateReport(string uploadId, string destinationId, string eventType, string stageName, CopyReport payload)
        {
            try
            {
                var content = new StringContent(JsonConvert.SerializeObject(payload), Encoding.UTF8, "application/json");
                var response = await _httpClient.PostAsync($"/api/report/json/uploadId/{uploadId}?destinationId={destinationId}&eventType={eventType}&stageName={stageName}", content);
                response.EnsureSuccessStatusCode();
            } catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return false;
            }

            return true;
        }

        public async Task<Trace?> GetTraceByUploadId(string uploadId)
        {
            try
            {
                var response = await _httpClient.GetAsync($"/api/trace/uploadId/{uploadId}");
                response.EnsureSuccessStatusCode();

                var responseBody = await response.Content.ReadAsStringAsync();
                if (string.IsNullOrEmpty(responseBody))
                {
                    _logger.LogError("Call to PS API returned empty body.");
                    return null;
                }

                return JsonConvert.DeserializeObject<Trace>(responseBody);
            } catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return null;
            }
        }

        public async Task<Span?> StartSpanForTrace(string traceId, string parentSpanId, string stageName)
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
