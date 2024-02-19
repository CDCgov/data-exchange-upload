// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Specialized;
using Azure.Identity;
using Azure.Storage.Sas;
using Newtonsoft.Json;
using System.Collections.Concurrent;
using Microsoft.Extensions.Configuration;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        private readonly IConfiguration _configuration;
        private readonly IUploadProcessingService _uploadProcessingService;
        private readonly IUploadEventHubService _uploadEventHubService;
        private readonly string _dexAzureStorageAccountName;
        private readonly string _dexAzureStorageAccountKey;
        private readonly string _tusHooksFolder;
        private readonly Task<List<DestinationAndEvents>?> _destinationAndEvents;
        private readonly string _targetEdav = "dex_edav";
        private readonly string _targetRouting = "dex_routing";
        private readonly string _destinationAndEventsFileName = "allowed_destination_and_events.json";

        public static string? GetEnvironmentVariable(string name)
        {
            return Environment.GetEnvironmentVariable(name, EnvironmentVariableTarget.Process);
        }

        public BulkFileUploadFunction(ILoggerFactory loggerFactory, IConfiguration configuration, IUploadProcessingService uploadProcessingService, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();

            _configuration = configuration;

            _uploadProcessingService = uploadProcessingService;
            _uploadEventHubService = uploadEventHubService;

            _dexAzureStorageAccountName = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME") ?? "";
            _dexAzureStorageAccountKey = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY") ?? "";

            _tusHooksFolder = GetEnvironmentVariable("TUSD_HOOKS_FOLDER") ?? "tusd-file-hooks";

            _destinationAndEvents = GetAllDestinationAndEvents();
        }

        /// <summary>
        /// Entrypoint for processing blob created events.  Note this should only be fired when a tus upload completes.
        /// </summary>
        /// <param name="eventHubTriggerEvent"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        [Function("BulkFileUploadFunction")]
        public async Task Run([EventHubTrigger("%AzureEventHubName%", Connection = "AzureEventHubConnectionString", ConsumerGroup = "%AzureEventHubConsumerGroup%")] string[] eventHubTriggerEvents)
        {
            _logger.LogInformation($"Received events count: {eventHubTriggerEvents.Count()}");

            foreach (var blobCreatedEventJson in eventHubTriggerEvents)
            {
                _logger.LogInformation($"Received event: {blobCreatedEventJson}");

                StorageBlobCreatedEvent[]? blobCreatedEvents = JsonConvert.DeserializeObject<StorageBlobCreatedEvent[]>(blobCreatedEventJson);

                if (blobCreatedEvents == null)
                    throw new Exception("Unexpected data content of event; unable to establish a StorageBlobCreatedEvent array");

                if (blobCreatedEvents.Count() < 1)
                    throw new Exception("Unexpected data content of event; there should be at least one element in the array");

                StorageBlobCreatedEvent blobCreatedEvent = blobCreatedEvents[0];
                if (blobCreatedEvent == null)
                    throw new Exception("Unexpected data content of event; there should be at least one element in the array");

                await ProcessBlobCreatedEvent(blobCreatedEvent?.Data?.Url);

            } // .foreach 

        } // .Task Run        

        private async Task ProcessBlobCreatedEvent(string? blobCreatedUrl)
        {
            if (blobCreatedUrl == null)
                throw new Exception("Blob url may not be null");

            string destinationId = null;
            string extEvent = null;
            string destinationContainerName = null;
            string destinationBlobFilename = null;
            Dictionary<string, string> tusFileMetadata = null;

            try
            {
                var result = await _uploadProcessingService.CopyBlobToDex(blobCreatedUrl);

                destinationId = result.Item1;
                extEvent = result.Item2;
                destinationContainerName = result.Item3;
                destinationBlobFilename = result.Item4;
                tusFileMetadata = result.Item5;

            } catch (Exception ex) {

                PublishRetryEvent(BlobCopyStage.CopyToDex, blobCreatedUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);
                return;
            }

            var currentDestination = _destinationAndEvents.Result?.Find(d => d.destinationId == destinationId);
            if(currentDestination == null) {
                _logger.LogError($"No matching configuration found for destination: {destinationId}" );
                return;
            } 

            var currentEvent = currentDestination?.extEvents?.Find(e => e.name == extEvent);
            if(currentEvent == null) {
                _logger.LogError($"No matching event:{extEvent} found for destination:{destinationId}");
                return;
            } 

            if (currentEvent != null && currentEvent.copyTargets != null && currentEvent.copyTargets.Any())
            {
                foreach (CopyTarget copyTarget in currentEvent.copyTargets)
                {
                    _logger.LogInformation("Copy Target: " + copyTarget.target);

                    if (copyTarget.target == _targetEdav)
                    {
                        // Now copy the file from DeX to the EDAV storage account, also partitioned by date
                        try 
                        {
                            await _uploadProcessingService.CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);
                        }
                        catch(Exception e) 
                        {
                            PublishRetryEvent(BlobCopyStage.CopyToEdav, blobCreatedUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);
                        }                        
                    }
                    else if (copyTarget.target == _targetRouting)
                    {
                        bool isRoutingEnabled = _configuration.GetValue<bool>("FeatureManagement:ROUTING");
                        _logger.LogInformation($"Routing Status: {isRoutingEnabled}");

                        if (isRoutingEnabled)
                        {
                            // Now copy the file from DeX to the ROUTING storage account, also partitioned by date
                            try
                            {
                                await _uploadProcessingService.CopyBlobFromDexToRoutingAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);
                            }
                            catch(Exception e)
                            {
                                PublishRetryEvent(BlobCopyStage.CopyToRouting, blobCreatedUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);
                            }                
                        }
                        else
                        {
                            _logger.LogInformation($"Routing is disabled. Bypassing routing for blob");
                        }
                    }
                }
            }
            else
            {
                _logger.LogInformation("No copy target found. Defaulting to EDAV");

                // Now copy the file from DeX to the EDAV storage account, also partitioned by date
                try
                {
                    await _uploadProcessingService.CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);
                }
                catch(Exception e)
                {                    
                    PublishRetryEvent(BlobCopyStage.CopyToEdav, blobCreatedUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);
                }
            }
        }

        private async Task PublishRetryEvent(BlobCopyStage copyStage, string sourceBlobUri, string dexContainerName, string dexBlobFilename, Dictionary<string, string> fileMetadata)
        {            
            try 
            {
                BlobCopyRetryEvent blobCopyRetryEvent = new BlobCopyRetryEvent();
                blobCopyRetryEvent.copyRetryStage = copyStage;
                blobCopyRetryEvent.retryAttempt = 1;
                blobCopyRetryEvent.sourceBlobUri = sourceBlobUri;
                blobCopyRetryEvent.dexContainerName = dexContainerName;
                blobCopyRetryEvent.dexBlobFilename = dexBlobFilename;
                blobCopyRetryEvent.fileMetadata = fileMetadata;

                await _uploadEventHubService.PublishRetryEvent(blobCopyRetryEvent);
            }
            catch (Exception ex)
            {
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }
        private async Task<List<DestinationAndEvents>?> GetAllDestinationAndEvents()
        {
            var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";

            try
            {
                var blobReader = new BlobReader(_logger);
                var destinationAndEvents = await blobReader.GetObjectFromBlobJsonContent<List<DestinationAndEvents>>(connectionString, _tusHooksFolder, _destinationAndEventsFileName);

                return destinationAndEvents;
            }
            catch (Exception e)
            {
                _logger.LogError("Failed to fetch Destinations and Events");
                ExceptionUtils.LogErrorDetails(e, _logger);

                return new List<DestinationAndEvents>();
            }
        }
    }

    public class JsonLogger : ILogger
    {
        private readonly string _name;
        private readonly Func<JsonLoggerConfiguration> _getCurrentConfig;

        public JsonLogger(string name, Func<JsonLoggerConfiguration> getCurrentConfig)
        {
            _name = name;
            _getCurrentConfig = getCurrentConfig;
        }

        public IDisposable BeginScope<TState>(TState state) => default;

        public bool IsEnabled(LogLevel logLevel)
        {
            var config = _getCurrentConfig();
            return logLevel >= config.LogLevel;
        }

        public void Log<TState>(LogLevel logLevel, EventId eventId, TState state, Exception exception, Func<TState, Exception, string> formatter)
        {
            if (!IsEnabled(logLevel))
            {
                return;
            }

            var config = _getCurrentConfig();
            var timestamp = DateTime.Now.ToString(config.TimestampFormat);
            var logEntry = new
            {
                Timestamp = timestamp,
                LogLevel = logLevel.ToString(),
                Name = _name,
                EventId = eventId.Id,
                Message = formatter(state, exception),
                // Add other desired fields and conventions here
            };

            var json = System.Text.Json.JsonSerializer.Serialize(logEntry, config.JsonSerializerOptions);
            Console.WriteLine(json);
        }
    }

    public class JsonLoggerConfiguration
    {
        public LogLevel LogLevel { get; set; } = LogLevel.Warning;
        public string TimestampFormat { get; set; } = "yyyy-MM-dd HH:mm:ss.fff";
        public System.Text.Json.JsonSerializerOptions JsonSerializerOptions { get; set; } = new System.Text.Json.JsonSerializerOptions
        {
            WriteIndented = true
        };
        // Add other configuration properties here
    }

    public class JsonLoggerProvider : ILoggerProvider
    {
        private readonly JsonLoggerConfiguration _config;
        private readonly ConcurrentDictionary<string, JsonLogger> _loggers = new ConcurrentDictionary<string, JsonLogger>();

        public JsonLoggerProvider(JsonLoggerConfiguration config)
        {
            _config = config;
        }

        public ILogger CreateLogger(string categoryName)
        {
            return _loggers.GetOrAdd(categoryName, name => new JsonLogger(name, () => _config));
        }

        public void Dispose() => _loggers.Clear();
    }
}
