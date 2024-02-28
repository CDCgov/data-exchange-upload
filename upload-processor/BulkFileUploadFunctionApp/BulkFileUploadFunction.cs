// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Model;
using Azure.Identity;
using Newtonsoft.Json;
using BulkFileUploadFunctionApp.Utils;
using System.Collections.Concurrent;
using Microsoft.Extensions.Configuration;
using BulkFileUploadFunctionApp.Services;
using System.Runtime.CompilerServices;


namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        
        private readonly ILogger _logger;

        private readonly IConfiguration _configuration;

        private readonly IUploadProcessingService _uploadProcessingService;
 
        public static string? GetEnvironmentVariable(string name)
        {
            return Environment.GetEnvironmentVariable(name, EnvironmentVariableTarget.Process);
        }

        public BulkFileUploadFunction(
        ILoggerFactory loggerFactory,
        IConfiguration configuration,
        IUploadProcessingService uploadProcessingService)
        {

            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();
            _configuration = configuration;
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

                // TODO: PS API fail report if any of these conditions are true.
                if (blobCreatedEvents == null)
                    throw new Exception("Unexpected data content of event; unable to establish a StorageBlobCreatedEvent array");

                if (blobCreatedEvents.Count() == 0)
                    throw new Exception("Unexpected data content of event; there should be at least one element in the array");

                foreach(StorageBlobCreatedEvent blobCreatedEvent in blobCreatedEvents)
                {
                    if (blobCreatedEvent.Data?.Url == null)
                        throw new Exception("Unexpected data content of event; no blob create event url found.");
                    try
                    {
                        await ProcessBlobCreatedEvent(blobCreatedEvent.Data.Url);
                    }
                    catch (Exception ex)
                    {
                        ExceptionUtils.LogErrorDetails(ex, _logger);
                    }
                }
            } // .foreach 

        } // .Task Run        

        private async Task ProcessBlobCreatedEvent(string? blobCreatedUrl)
        {
            if (blobCreatedUrl == null)
                throw new Exception("Blob url may not be null");

            await _uploadProcessingService.ProcessBlob(blobCreatedUrl);
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
