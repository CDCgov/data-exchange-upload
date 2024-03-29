using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;
using System.Text.Json;

namespace BulkFileUploadFunctionApp
{
    public class UploadProcessorRetry
    {
        private readonly ILogger _logger;
        private readonly IUploadProcessingService _uploadProcessingService;
        private readonly IUploadEventHubService _uploadEventHubService;
        private readonly int _maxRetryAttempts;
        private const int MAX_RETRY_ATTEMPTS_DEFAULT = 2;
        
        public UploadProcessorRetry(ILoggerFactory loggerFactory, IUploadProcessingService uploadProcessingService, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadProcessorRetry>();

            _uploadProcessingService = uploadProcessingService;
            _uploadEventHubService = uploadEventHubService;

            _maxRetryAttempts = int.TryParse(Environment.GetEnvironmentVariable("MAX_RETRY_ATTEMPTS", EnvironmentVariableTarget.Process), out int maxAttemptsFromEnv) ? maxAttemptsFromEnv : MAX_RETRY_ATTEMPTS_DEFAULT;
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

                    BlobCopyRetryEvent? blobCopyRetryEvent = JsonSerializer.Deserialize<BlobCopyRetryEvent>(blobCopyRetryEventJson);

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
                if(blobCopyRetryEvent.RetryAttempt <= _maxRetryAttempts) {

                    // exponential backoff
                    await Task.Delay(1000 * blobCopyRetryEvent.RetryAttempt);

                    _logger.LogInformation($"Copy retry attempt: {blobCopyRetryEvent.RetryAttempt} Stage: {blobCopyRetryEvent.CopyRetryStage}");

                    switch (blobCopyRetryEvent.CopyRetryStage)
                    {
                        case BlobCopyStage.CopyToDex:
                            try
                            {
                                CopyPrereqs copyPrereqs = await _uploadProcessingService.GetCopyPrereqs(blobCopyRetryEvent.CopyPrereqs.SourceBlobUrl);

                                await _uploadProcessingService.CopyAll(copyPrereqs);
                            }
                            catch (Exception ex)
                            {
                                await RePublishEvent(blobCopyRetryEvent);
                            }
                            break;

                        //case BlobCopyStage.CopyToEdav:
                        //    try
                        //    {
                        //        await _uploadProcessingService.CopyFromDexToEdav(blobCopyRetryEvent.CopyPrereqs);
                        //    }
                        //    catch (Exception ex)
                        //    {
                        //        await RePublishEvent(blobCopyRetryEvent);
                        //    }                          
                        //    break;

                        //case BlobCopyStage.CopyToRouting:
                        //    try
                        //    {
                        //        await _uploadProcessingService.CopyFromDexToRouting(blobCopyRetryEvent.CopyPrereqs);
                        //    }
                        //    catch (Exception ex)
                        //    {
                        //        await RePublishEvent(blobCopyRetryEvent);
                        //    }                            
                        //    break;
                        
                        default:
                            _logger.LogInformation("Invalid copy retry stage provided");
                            break;
                    }
                } else {

                    await RePublishEvent(blobCopyRetryEvent);
                }
                
            } catch(Exception ex) {

                _logger.LogError($"Failed to process retry event: " + blobCopyRetryEvent);
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        private async Task RePublishEvent(BlobCopyRetryEvent blobCopyRetryEvent) {

            if (blobCopyRetryEvent.RetryAttempt == _maxRetryAttempts) {

                _logger.LogInformation("Reached max retry attempts - sending event to Replay queue");

                blobCopyRetryEvent.RetryAttempt = 1;
                await _uploadEventHubService.PublishReplayEvent(blobCopyRetryEvent);
            } else {        

                // Increment the retry attempt and put the event back on retry loop
                blobCopyRetryEvent.RetryAttempt += 1;
                await _uploadEventHubService.PublishRetryEvent(blobCopyRetryEvent);
            }
        }
    }
}