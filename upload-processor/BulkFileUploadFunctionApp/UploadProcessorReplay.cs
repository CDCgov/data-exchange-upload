using System.Net;
using System.Text;
using System.Text.Json;
using Microsoft.Extensions.Logging;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Azure.Storage.Blobs;

using Azure.Messaging.EventHubs;
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
        private DateTimeOffset stopProcessingAfterTime;
        private EventProcessorClient? replayEventProcessorClient;


        public UploadProcessorReplay(ILoggerFactory loggerFactory, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadEventHubService>();
            _uploadEventHubService = uploadEventHubService;

            _uploadEventHubNamespaceConnectionString = Environment.GetEnvironmentVariable("AzureEventHubConnectionString", EnvironmentVariableTarget.Process) ?? "";
            _replayEventHubName = Environment.GetEnvironmentVariable("ReplayEventHubName", EnvironmentVariableTarget.Process) ?? "";
            _consumerGroup = Environment.GetEnvironmentVariable("AzureEventHubConsumerGroup", EnvironmentVariableTarget.Process) ?? "";

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
                await ProcessReplayEventHubEventsAsync();

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

            replayEventProcessorClient = new EventProcessorClient(storageClient,
                                                                _consumerGroup,
                                                                _uploadEventHubNamespaceConnectionString,
                                                                _replayEventHubName,
                                                                processorOptions);

            replayEventProcessorClient.ProcessEventAsync += ProcessEventHandler;
            replayEventProcessorClient.ProcessErrorAsync += ProcessErrorHandler;

            stopProcessingAfterTime = DateTimeOffset.UtcNow;

            await replayEventProcessorClient.StartProcessingAsync();
        }

        async Task ProcessEventHandler(ProcessEventArgs eventArgs)
        {
            try
            {
                if(!eventArgs.HasEvent) {
                    _logger.LogInformation("No replay event found. Stopping replay.");
                    await replayEventProcessorClient.StopProcessingAsync();                   
                    return;
                }

                var eventData = eventArgs.Data;
                if(eventData == null) {
                    _logger.LogInformation("No event data found.");
                    return;
                }

                string eventJsonString = Encoding.UTF8.GetString(eventData.EventBody.ToArray());

                BlobCopyRetryEvent? replayEvent = JsonSerializer.Deserialize<BlobCopyRetryEvent>(eventJsonString);

                _logger.LogInformation("Replaying event: " + eventJsonString);

                if(replayEvent == null)
                {
                    _logger.LogInformation("Failed to deserialize replay event.");
                    return;
                }

                await _uploadEventHubService.PublishRetryEvent(replayEvent);

                _logger.LogInformation("Updating replay checkpoint");
                await eventArgs.UpdateCheckpointAsync();

                // Cancel processing events if enqueued time exceeds the start time
                if (eventArgs.Data.EnqueuedTime >= stopProcessingAfterTime)
                {
                    _logger.LogInformation("Stopping replay");
                    await replayEventProcessorClient.StopProcessingAsync();                   
                }
            }
            catch(Exception ex)
            {
                _logger.LogError($"Error during event replay: {ex.Message}");
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