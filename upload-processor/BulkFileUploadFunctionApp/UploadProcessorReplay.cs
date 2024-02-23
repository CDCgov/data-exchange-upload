using System.Net;
using System.Text;
using System.Text.Json;
using Newtonsoft.Json;
using Microsoft.Extensions.Logging;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;

using Azure.Messaging.EventHubs;
using Azure.Messaging.EventHubs.Consumer;

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

        public UploadProcessorReplay(ILoggerFactory loggerFactory, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadEventHubService>();
            _uploadEventHubService = uploadEventHubService;

            _uploadEventHubNamespaceConnectionString = Environment.GetEnvironmentVariable("AzureEventHubConnectionString", EnvironmentVariableTarget.Process);
            _replayEventHubName = Environment.GetEnvironmentVariable("ReplayEventHubName", EnvironmentVariableTarget.Process);
            _consumerGroup = Environment.GetEnvironmentVariable("AzureEventHubConsumerGroup", EnvironmentVariableTarget.Process);
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
            EventHubConsumerClient replayConsumerClient = null;

            try
            {             
                replayConsumerClient = new EventHubConsumerClient(_consumerGroup, _uploadEventHubNamespaceConnectionString, _replayEventHubName);

                _logger.LogInformation("Replaying events...");

                // Get current timestanp when the trigger was invoked
                DateTimeOffset stopReadingAfterTime = DateTimeOffset.UtcNow;

                await foreach (PartitionEvent partitionEvent in replayConsumerClient.ReadEventsAsync())
                {
                    DateTimeOffset enqueueTime = partitionEvent.Data.EnqueuedTime;

                    // Do not process evets submitted after the trigger was invoked to avoid going into a loop in case of retry failures
                    if (enqueueTime > stopReadingAfterTime)
                    {
                        await ProcessEvent(partitionEvent);
                        break; 
                    } 
                    else 
                    {
                        await ProcessEvent(partitionEvent);
                    }
                }
            }
            catch (Exception ex)
            {
                _logger.LogError("Error while replaying events.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
            finally
            {
                if(replayConsumerClient != null) 
                {                   
                    await replayConsumerClient.CloseAsync();
                }                
            }
        }

        private async Task ProcessEvent(PartitionEvent partitionEvent) 
        {
            string eventJsonString = Encoding.UTF8.GetString(partitionEvent.Data.Body.ToArray());

            _logger.LogInformation("Replaying event: " + eventJsonString);

            BlobCopyRetryEvent? replayEvent = JsonConvert.DeserializeObject<BlobCopyRetryEvent>(eventJsonString);

            await _uploadEventHubService.PublishRetryEvent(replayEvent);
        }
    }
}