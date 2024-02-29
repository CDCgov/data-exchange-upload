using System.Net;
using System.Text;
using System.Text.Json;
using Newtonsoft.Json;
using Microsoft.Extensions.Logging;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Azure.Storage.Blobs;

using Azure.Messaging.EventHubs;
using Azure.Messaging.EventHubs.Primitives;
using Azure.Messaging.EventHubs.Consumer;
using Azure.Messaging.EventHubs.Processor;

using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp
{
    public class UploadProcessorReplay
    {
        private readonly ILogger _logger;
        private readonly IUploadEventHubService _uploadEventHubService;
        private readonly string _uploadEventHubNamespaceConnectionString;
        private readonly string _replayEventHubName;
        private readonly string _consumerGroup;
        private readonly string _dexAzureStorageAccountName;
        private readonly string _dexAzureStorageAccountKey;
        private readonly string _dexStorageAccountConnectionString;
        private readonly string _replayCheckpointContainer;

        private static readonly CancellationTokenSource cancellationTokenSource = new CancellationTokenSource();
        private static readonly CancellationToken cancellationToken = cancellationTokenSource.Token;
    
        private DateTimeOffset stopReadingAfterTime;


        public UploadProcessorReplay(ILoggerFactory loggerFactory, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadEventHubService>();
            _uploadEventHubService = uploadEventHubService;

            _uploadEventHubNamespaceConnectionString = Environment.GetEnvironmentVariable("AzureEventHubConnectionString", EnvironmentVariableTarget.Process);
            _replayEventHubName = Environment.GetEnvironmentVariable("ReplayEventHubName", EnvironmentVariableTarget.Process);
            _consumerGroup = Environment.GetEnvironmentVariable("AzureEventHubConsumerGroup", EnvironmentVariableTarget.Process);

            _dexAzureStorageAccountName = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";

            _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";
            _replayCheckpointContainer = "replay-checkpoint";
        }

        [Function("UploadProcessorReplay")]
        public async Task<HttpResponseData> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "post", Route = "replay")] HttpRequestData req,
            FunctionContext context)
        {
            try
            {
                ProcessReplayEventHubEventsAsync();

                return req.CreateResponse(HttpStatusCode.OK);
            
            } catch(Exception ex) {

                _logger.LogError($"Failed to replay events");
                ExceptionUtils.LogErrorDetails(ex, _logger);

                return req.CreateResponse(HttpStatusCode.InternalServerError);
            }
        }

        private async Task ProcessReplayEventHubEventsAsync() 
        {
            _logger.LogInformation("Replaying events...");

            var storageClient = new BlobContainerClient(_dexStorageAccountConnectionString, 
                                                        _replayCheckpointContainer);

            var processorOptions = new EventProcessorClientOptions
            {
                MaximumWaitTime = TimeSpan.FromSeconds(5)
            };

            var replayEventProcessorClient = new EventProcessorClient(storageClient,
                                                                      _consumerGroup,
                                                                      _uploadEventHubNamespaceConnectionString,
                                                                      _replayEventHubName,
                                                                      processorOptions);

            replayEventProcessorClient.ProcessEventAsync += ProcessEventHandler;
            replayEventProcessorClient.ProcessErrorAsync += ProcessErrorHandler;

            try {

                stopReadingAfterTime = DateTimeOffset.UtcNow;

                await replayEventProcessorClient.StartProcessingAsync(cancellationToken);
            }
            catch (TaskCanceledException)
            {
                _logger.LogInformation("Replay stopped");
                await replayEventProcessorClient.StopProcessingAsync();
            }            
        }

        async Task ProcessEventHandler(ProcessEventArgs eventArgs)
        {
            try
            {
                // Check if cancellation is requested
                if (cancellationToken.IsCancellationRequested)
                {
                    // If cancellation is requested, stop processing further events
                    return;
                }

                var eventData = eventArgs.Data;

                string eventJsonString = Encoding.UTF8.GetString(eventData.EventBody.ToArray());

                _logger.LogInformation("Replaying event: " + eventJsonString);

                BlobCopyRetryEvent? replayEvent = JsonConvert.DeserializeObject<BlobCopyRetryEvent>(eventJsonString);

                await _uploadEventHubService.PublishRetryEvent(replayEvent);

                // Cancel processing events if enqueued time exceeds the start time
                if (eventArgs.Data.EnqueuedTime >= stopReadingAfterTime)
                {
                    await eventArgs.UpdateCheckpointAsync(eventArgs.CancellationToken);
                    _logger.LogInformation("Replay Cancelled");
                } else {

                    await eventArgs.UpdateCheckpointAsync();
                }
            }
            catch(Exception ex)
            {

                _logger.LogError($"Error during Replay: {ex.Message}");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        Task ProcessErrorHandler(ProcessErrorEventArgs errorArgs)
        {
            // Handle any errors that occur during event processing
            _logger.LogInformation($"Error processing Replay event: {errorArgs.Exception.Message}");

            return Task.CompletedTask;
        }
    }
}