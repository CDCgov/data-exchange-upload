// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using BulkFileUploadFunctionApp.Model;
using Newtonsoft.Json;
using BulkFileUploadFunctionApp.Utils;
using System.Collections.Concurrent;
using BulkFileUploadFunctionApp.Services;

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        private readonly IUploadProcessingService _uploadProcessingService;

        public BulkFileUploadFunction(ILoggerFactory loggerFactory, IUploadProcessingService uploadProcessingService)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();

            _uploadProcessingService = uploadProcessingService;
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

                if (blobCreatedEvents.Count() == 0)
                    throw new Exception("Unexpected data content of event; there should be at least one element in the array");

                foreach(StorageBlobCreatedEvent blobCreatedEvent in blobCreatedEvents)
                {
                    if (blobCreatedEvent.Data?.Url == null)
                    {
                        _logger.LogInformation($"Received blob created event with null URL: {blobCreatedEvent}");
                    } else {

                        await ProcessBlobCreatedEvent(blobCreatedEvent.Data.Url);
                    }                    
                }
            } // .foreach 

        } // .Task Run 

        private async Task ProcessBlobCreatedEvent(string blobCreatedUrl)
        {
            CopyPreqs copyPreqs = new CopyPreqs();
            copyPreqs.SourceBlobUrl = blobCreatedUrl;

            try
            {
                copyPreqs = await _uploadProcessingService.GetCopyPreqs(blobCreatedUrl);
                _logger.LogInformation($"Copy preqs: {JsonConvert.SerializeObject(copyPreqs)}");
                
 
                await _uploadProcessingService.CopyAll(copyPreqs);
            }
            catch(Exception ex)
            {
                // publish Retry event
                await _uploadProcessingService.PublishRetryEvent(BlobCopyStage.CopyToDex,
                                                                 copyPreqs);
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
