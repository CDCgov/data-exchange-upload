using Microsoft.Extensions.Logging;
using System.Text;
using System.Text.Json;
using Azure.Messaging.EventHubs;
using Azure.Messaging.EventHubs.Producer;

using BulkFileUploadFunctionApp.Model;


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
            _logger.LogInformation("Publishing Retry Event: " + blobCopyRetryEvent);

            blobCopyRetryEvent.retryAttempt = blobCopyRetryEvent.retryAttempt + 1;
            string jsonPayload = JsonSerializer.Serialize(blobCopyRetryEvent);

            // Create an event data batch
            using (EventDataBatch eventBatch = await _retryEventHubProducerClient.CreateBatchAsync())
            {
                // Add the event data to the batch
                eventBatch.TryAdd(new EventData(Encoding.UTF8.GetBytes(jsonPayload)));

                // Publish the batch of events to the event hub
                await _retryEventHubProducerClient.SendAsync(eventBatch);
                _logger.LogInformation("Retry event published successfully");
            }
        }

        public async Task PublishReplayEvent(BlobCopyRetryEvent blobCopyRetryEvent)
        {
            _logger.LogInformation("Publishing Replay Event: " + blobCopyRetryEvent);

            // publish event to replay 
            blobCopyRetryEvent.retryAttempt = 1;
            string jsonPayload = JsonSerializer.Serialize(blobCopyRetryEvent);

            // Create an event data batch
            using (EventDataBatch eventBatch = await _replayEventHubProducerClient.CreateBatchAsync())
            {
                // Add the event data to the batch
                eventBatch.TryAdd(new EventData(Encoding.UTF8.GetBytes(jsonPayload)));

                // Publish the batch of events to the event hub
                await _replayEventHubProducerClient.SendAsync(eventBatch);
                _logger.LogInformation("Replay event published successfully");
            }
        }
    }
}
