using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;

using System;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp
{
    public class UploadProcessorRetry
    {
        private readonly ILogger _logger;
        private readonly IUploadProcessingService _uploadProcessingService;

        private readonly IUploadEventHubService _uploadEventHubService;

        private const int MAX_RETRY_ATTEMPTS = 2;
        
        public UploadProcessorRetry(ILoggerFactory loggerFactory, IUploadProcessingService uploadProcessingService, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadProcessorRetry>();

            _uploadProcessingService = uploadProcessingService;
            _uploadEventHubService = uploadEventHubService;
        }
        
        [Function("UploadProcessorRetry")]
        public async Task Run([EventHubTrigger("%RetryEventHubName%", Connection = "AzureEventHubConnectionString", ConsumerGroup = "%AzureEventHubConsumerGroup%")] string[] blobCopyRetryEvents)
        {      
            try
            {
                _logger.LogInformation($"Received blob copy retry events: {blobCopyRetryEvents.Count()}");

                foreach (var blobCopyRetryEventJson in blobCopyRetryEvents)
                {
                    _logger.LogInformation($"Received blob copy retry event: {blobCopyRetryEventJson}");

                    BlobCopyRetryEvent? blobCopyRetryEvent = JsonConvert.DeserializeObject<BlobCopyRetryEvent>(blobCopyRetryEventJson);

                    if (blobCopyRetryEvent != null) {

                        await ProcessCopyRetryEvent(blobCopyRetryEvent); 
                    }                
                }
            } catch (Exception ex) {

                _logger.LogError($"Failed to process retry events");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        private async Task ProcessCopyRetryEvent(BlobCopyRetryEvent blobCopyRetryEvent)
        {
            try
            {
                if(blobCopyRetryEvent.copyRetryStage != null) {

                    if(blobCopyRetryEvent.retryAttempt <= MAX_RETRY_ATTEMPTS) {

                        await Task.Delay(1000 * blobCopyRetryEvent.retryAttempt);
                        _logger.LogInformation($"Copy retry attempt: {blobCopyRetryEvent.retryAttempt} Stage: {blobCopyRetryEvent.copyRetryStage}");

                        switch (blobCopyRetryEvent.copyRetryStage)
                        {
                            case BlobCopyStage.CopyToDex:
                                try
                                {
                                    await _uploadProcessingService.ProcessBlob(blobCopyRetryEvent.sourceBlobUri);
                                }
                                catch (Exception ex)
                                {
                                    await RePublishEvent(blobCopyRetryEvent);
                                }
                                break;

                            case BlobCopyStage.CopyToEdav:
                                try
                                {
                                    await _uploadProcessingService.CopyBlobFromDexToEdavAsync(blobCopyRetryEvent.dexContainerName, blobCopyRetryEvent.dexBlobFilename, blobCopyRetryEvent.fileMetadata);
                                }
                                catch (Exception ex)
                                {
                                    await RePublishEvent(blobCopyRetryEvent);
                                }                          
                                break;

                            case BlobCopyStage.CopyToRouting:
                                try
                                {
                                    await _uploadProcessingService.CopyBlobFromDexToRoutingAsync(blobCopyRetryEvent.dexContainerName, blobCopyRetryEvent.dexBlobFilename, blobCopyRetryEvent.fileMetadata);
                                }
                                catch (Exception ex)
                                {
                                    await RePublishEvent(blobCopyRetryEvent);
                                }                            
                                break;
                            
                            default:
                                _logger.LogInformation("Invalid copy retry stage provided");
                                break;
                        }
                    } else {

                        await RePublishEvent(blobCopyRetryEvent);
                    }
                }
            } catch(Exception ex) {

                _logger.LogError($"Failed to process retry event: " + blobCopyRetryEvent);
                ExceptionUtils.LogErrorDetails(ex, _logger);                
            }
        }

        private async Task RePublishEvent(BlobCopyRetryEvent blobCopyRetryEvent) {

            if (blobCopyRetryEvent.retryAttempt == MAX_RETRY_ATTEMPTS) {
                blobCopyRetryEvent.retryAttempt = 1;
                await _uploadEventHubService.PublishReplayEvent(blobCopyRetryEvent);
            } else {
                blobCopyRetryEvent.retryAttempt += 1;
                await _uploadEventHubService.PublishRetryEvent(blobCopyRetryEvent);
            }
        }
    }
}