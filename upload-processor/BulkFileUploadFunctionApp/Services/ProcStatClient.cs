using System.Text;
using System.Text.Json;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Logging;

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

        public async Task<HealthCheckResponse?> GetHealthCheck()
        {
            try
            {
                var response = await _httpClient.GetAsync("/api/health");
                response.EnsureSuccessStatusCode();

                string responseBody = await response.Content.ReadAsStringAsync();
                if (string.IsNullOrEmpty(responseBody))
                {
                    _logger.LogError("Call to PS API returned empty body.");
                    return new HealthCheckResponse()
                    {
                        Status = "DOWN"
                    };
                }
                return JsonSerializer.Deserialize<HealthCheckResponse>(responseBody);
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
        public async Task<bool> CreateReport<TReport>(string uploadId, string destinationId, string eventType, string stageName, TReport payload)
        {
            try
            {
                var content = new StringContent(JsonSerializer.Serialize(payload), Encoding.UTF8, "application/json");
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

                return JsonSerializer.Deserialize<Trace>(responseBody);
            } catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return null;
            }
        }

        public async Task<Span?> StartSpanForTrace(string traceId, string parentSpanId, string stageName)
        {
            try
            {
                var response = await _httpClient.PutAsync($"/api/trace/startSpan/{traceId}/{parentSpanId}?stageName={stageName}", null);
                response.EnsureSuccessStatusCode();

                var responseBody = await response.Content.ReadAsStringAsync();
                if (string.IsNullOrEmpty(responseBody))
                {
                    _logger.LogError("Call to PS API returned empty body.");
                    return null;
                }
                return JsonSerializer.Deserialize<Span>(responseBody);
            }
            catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return null;
            }
        }

        public async Task<string?> StopSpanForTrace(string traceId, string childSpanId)
        {
            try
            {
                var response = await _httpClient.PutAsync($"/api/trace/stopSpan/{traceId}/{childSpanId}", null);
                response.EnsureSuccessStatusCode();

                var responseBody = await response.Content.ReadAsStringAsync();
                if (string.IsNullOrEmpty(responseBody))
                {
                    _logger.LogError("Call to PS API returned empty body.");
                    return null;
                }
                return responseBody;
            }
            catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return null;
            }
        }
    }
}
