using Microsoft.Extensions.Logging;
using System.Text;
using System.Text.Json;
using Azure.Messaging.EventHubs;
using Azure.Messaging.EventHubs.Producer;

using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;


namespace BulkFileUploadFunctionApp.Services
{
    public class UploadEventHubService : IUploadEventHubService
    {
        private readonly ILogger _logger;
        private readonly string _uploadEventHubNamespaceConnectionString;
        private readonly string _retryEventHubName;
        private readonly string _replayEventHubName;
        private readonly EventHubProducerClient _retryEventHubProducerClient;
        private readonly EventHubProducerClient _replayEventHubProducerClient;
        
        public UploadEventHubService(ILoggerFactory loggerFactory)
        {
            _logger = loggerFactory.CreateLogger<UploadEventHubService>();

            _uploadEventHubNamespaceConnectionString = Environment.GetEnvironmentVariable("AzureEventHubConnectionString", EnvironmentVariableTarget.Process);
            _retryEventHubName = Environment.GetEnvironmentVariable("RetryEventHubName", EnvironmentVariableTarget.Process);
            _replayEventHubName = Environment.GetEnvironmentVariable("ReplayEventHubName", EnvironmentVariableTarget.Process);

            _retryEventHubProducerClient = new EventHubProducerClient(_uploadEventHubNamespaceConnectionString, _retryEventHubName);
            _replayEventHubProducerClient = new EventHubProducerClient(_uploadEventHubNamespaceConnectionString, _replayEventHubName);
        }
        public async Task PublishRetryEvent(BlobCopyRetryEvent blobCopyRetryEvent)
        {
            string jsonPayload = JsonSerializer.Serialize(blobCopyRetryEvent);

            try 
            {
                _logger.LogInformation("Publishing Retry Event: " + jsonPayload);

                await PublishEventAsync(_retryEventHubProducerClient, jsonPayload);

                _logger.LogInformation("Retry event published successfully");

            } catch (Exception ex) {

                _logger.LogError("Failed to publish Retry event: " + jsonPayload);
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        public async Task PublishReplayEvent(BlobCopyRetryEvent blobCopyRetryEvent)
        {
            string jsonPayload = JsonSerializer.Serialize(blobCopyRetryEvent);

            try 
            {
                _logger.LogInformation("Publishing Replay Event: " + jsonPayload);

                await PublishEventAsync(_replayEventHubProducerClient, jsonPayload);

                _logger.LogInformation("Replay event published successfully");
                
            } catch (Exception ex) {

                _logger.LogError("Failed to publish Replay event: " + jsonPayload);
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        private async Task PublishEventAsync(EventHubProducerClient producerClient, string jsonPayload)
        {
            // Create an event data batch
            using (EventDataBatch eventBatch = await producerClient.CreateBatchAsync())
            {
                // Add the event data to the batch
                eventBatch.TryAdd(new EventData(Encoding.UTF8.GetBytes(jsonPayload)));

                // Publish the batch of events to the event hub
                await producerClient.SendAsync(eventBatch);
            }
        }
    }
}