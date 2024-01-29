// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Model;
using Azure.Identity;
using Azure.Storage.Sas;
using Newtonsoft.Json;
using BulkFileUploadFunctionApp.Utils;
using System.Collections.Concurrent;
using Microsoft.Extensions.Configuration;

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        private readonly IConfiguration _configuration;

        private readonly BlobCopyHelper _blobCopyHelper;

        private readonly string _tusAzureObjectPrefix;

        private readonly string _tusAzureStorageContainer;

        private readonly string _dexAzureStorageAccountName;

        private readonly string _dexAzureStorageAccountKey;

        private readonly string _edavAzureStorageAccountName;

        private readonly string _routingStorageAccountName;

        private readonly string _routingStorageAccountKey;

        private readonly string _edavUploadRootContainerName;

        private readonly string _routingUploadRootContainerName;

        private readonly string _tusHooksFolder;

        private readonly Task<List<DestinationAndEvents>?> _destinationAndEvents;

        private readonly string _targetEdav = "dex_edav";
        private readonly string _targetRouting = "dex_routing";

        private readonly string _destinationAndEventsFileName = "allowed_destination_and_events.json";


        public static string? GetEnvironmentVariable(string name)
        {
            return Environment.GetEnvironmentVariable(name, EnvironmentVariableTarget.Process);
        }

        public BulkFileUploadFunction(ILoggerFactory loggerFactory, IConfiguration configuration)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();

            _configuration = configuration;

            _blobCopyHelper = new(_logger);

            _tusAzureObjectPrefix = GetEnvironmentVariable("TUS_AZURE_OBJECT_PREFIX") ?? "tus-prefix";
            _tusAzureStorageContainer = GetEnvironmentVariable("TUS_AZURE_STORAGE_CONTAINER") ?? "bulkuploads";
            _dexAzureStorageAccountName = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME") ?? "";
            _dexAzureStorageAccountKey = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY") ?? "";
            _edavAzureStorageAccountName = GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME") ?? "";

            _routingStorageAccountName = GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_NAME") ?? "";
            _routingStorageAccountKey = GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_KEY") ?? "";

            _edavUploadRootContainerName = GetEnvironmentVariable("EDAV_UPLOAD_ROOT_CONTAINER_NAME") ?? "upload";
            _routingUploadRootContainerName = GetEnvironmentVariable("ROUTING_UPLOAD_ROOT_CONTAINER_NAME") ?? "routeingress";

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

        /// <summary>
        /// Processeses the given blob created event from the URL provided.
        /// </summary>
        /// <param name="blobCreatedUrl"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        private async Task ProcessBlobCreatedEvent(string? blobCreatedUrl)
        {
            if (blobCreatedUrl == null)
                throw new Exception("Blob url may not be null");

            _logger.LogInformation($"TUS_AZURE_OBJECT_PREFIX={_tusAzureObjectPrefix}, TUS_AZURE_STORAGE_CONTAINER={_tusAzureStorageContainer}, DEX_AZURE_STORAGE_ACCOUNT_NAME={_dexAzureStorageAccountName}");

            try
            {
                _logger.LogInformation($"Processing blob url: {blobCreatedUrl}");

                var sourceBlobUri = new Uri(blobCreatedUrl);
                string tusPayloadFilename = sourceBlobUri.Segments.Last();
                _logger.LogInformation($"tusPayloadFilename is = {tusPayloadFilename}");

                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";
                var sourceContainerName = _tusAzureStorageContainer;
                var tusPayloadPathname = $"/{_tusAzureObjectPrefix}/{tusPayloadFilename}";
                var tusInfoFile = await GetTusFileInfo(connectionString, sourceContainerName, tusPayloadPathname);

                GetRequiredMetaData(tusInfoFile, out string destinationId, out string extEvent);

                var uploadConfig = UploadConfig.Default;
                try
                {
                    // Determine the filename and subfolder creation schemes for this destination/event.
                    var configFilename = $"{destinationId}-{extEvent}.json";
                    var blobReader = new BlobReader(_logger);
                    uploadConfig = await blobReader.GetObjectFromBlobJsonContent<UploadConfig>(connectionString, "upload-configs", configFilename);
                    _logger.LogInformation($"Upload config: FilenameMetadataField={uploadConfig.FilenameMetadataField}, FilenameSuffix={uploadConfig.FilenameSuffix}, FolderStructure={uploadConfig.FolderStructure}");
                }
                catch (Exception e)
                {
                    // use default upload config
                    _logger.LogError($"No upload config found for destination id = {destinationId}, ext event = {extEvent}: exception = ${e.Message}");


                }

                // Determine the destination filename based on the upload config and metadata values provided with the source file.
                GetFilenameFromMetaData(tusInfoFile, uploadConfig.FilenameMetadataField, out string filename);

                var dateTimeNow = DateTime.UtcNow;

                // Determine the folder path and filename suffix from the upload configuration.
                var folderPath = GetFolderPath(uploadConfig, dateTimeNow);
                var filenameSuffix = GetFilenameSuffix(uploadConfig, dateTimeNow);

                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
                var destinationBlobFilename = $"{folderPath}/{fileNameWithoutExtension}{filenameSuffix}{fileExtension}";

                // Container name is "{meta_destination_id}-{extEvent}"
                // There are some restrictions on container names -- underscores not allowed, must be all lowercase
                var destinationContainerName = $"{destinationId.ToLower()}-{extEvent.ToLower()}";

                var tusFileMetadata = tusInfoFile?.MetaData ?? new Dictionary<string, string>();
                tusFileMetadata.Add("tus_tguid", tusPayloadFilename);
                tusFileMetadata.Remove("filename");
                tusFileMetadata.Add("orig_filename", filename);

                // Copy the blob to the DeX storage account specific to the program, partitioned by date
                await CopyBlobFromTusToDexAsync(connectionString, sourceContainerName, tusPayloadPathname, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                await CopyToTargetSystemAsync(destinationId, extEvent, destinationBlobFilename, destinationContainerName, tusFileMetadata);
            }
            catch (Exception e)
            {
                _logger.LogError(e.Message);
            }
        }
        private async Task CopyToTargetSystemAsync(string destinationId, string extEvent, string destinationBlobFilename, string destinationContainerName, Dictionary<string, string> tusFileMetadata)
        {
            var currentDestination = _destinationAndEvents.Result?.Find(d => d.destinationId == destinationId);

            var currentEvent = currentDestination?.extEvents?.Find(e => e.name == extEvent);

            bool isRoutingEnabled = _configuration.GetValue<bool>(".appconfig.featureflag/ROUTING");

            _logger.LogInformation($"Routing Status: {isRoutingEnabled}");

            if (currentEvent != null && currentEvent.copyTargets != null)
            {

                foreach (CopyTarget copyTarget in currentEvent.copyTargets)
                {
                    _logger.LogInformation("Copy Target: " + copyTarget.target);

                    if (copyTarget.target == _targetEdav)
                    {

                        // Now copy the file from DeX to the EDAV storage account, also partitioned by date
                        await CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);

                    }
                    else if (copyTarget.target == _targetRouting)
                    {

                        if (isRoutingEnabled)
                        {
                            // Now copy the file from DeX to the ROUTING storage account, also partitioned by date
                            await CopyBlobFromDexToRoutingAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);
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
                await CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);
            }
        }

        /// <summary>
        /// Returns the metadata from a tus .info file for the pathname provided.
        /// </summary>
        /// <param name="connectionString">Azure storage account connection string</param>
        /// <param name="sourceContainerName">Container where the file to get info for resides</param>
        /// <param name="tusPayloadPathname">Full path of the file to get info on</param>
        /// <returns></returns>
        /// <exception cref="TusInfoFileException"></exception>
        private async Task<TusInfoFile> GetTusFileInfo(string connectionString, string sourceContainerName, string tusPayloadPathname)
        {
            TusInfoFile tusInfoFile;

            try
            {
                string tusInfoPathname = tusPayloadPathname + ".info";

                _logger.LogInformation($"Retrieving tus info file: {tusInfoPathname}");

                var blobReader = new BlobReader(_logger);
                tusInfoFile = await blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(connectionString, sourceContainerName, tusInfoPathname);
            }
            catch (Exception e)
            {
                throw new TusInfoFileException(e.Message);
            }

            _logger.LogInformation($"Info file metadata keys: {string.Join(", ", tusInfoFile.MetaData?.Keys.ToList())}");

            return tusInfoFile;
        }

        /// <summary>
        /// Copies a blob file from DEX to EDAV asynchronously.
        /// </summary>
        /// <param name="sourceContainerName">Source container name</param>
        /// <param name="sourceBlobFilename">Source blob filename</param>
        /// <param name="destinationMetadata">Destination metadata to be associated with the blob file</param>
        /// <returns></returns>
        private async Task CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata)
        {
            try
            {
                BlobServiceClient blobServiceClient = new($"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net");
                BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(sourceContainerName);
                BlobClient dexBlobClient = containerClient.GetBlobClient(sourceBlobFilename);

                var edavBlobServiceClient = new BlobServiceClient(
                    new Uri($"https://{_edavAzureStorageAccountName}.blob.core.windows.net"),
                    new DefaultAzureCredential() // using Service Principal
                );

                // _edavUploadRootContainerName could be set to empty, then no root container in edav

                string destinationContainerName = string.IsNullOrEmpty(_edavUploadRootContainerName) ? sourceContainerName : _edavUploadRootContainerName;
                string destinationBlobFilename = string.IsNullOrEmpty(_edavUploadRootContainerName) ? sourceBlobFilename : $"{sourceContainerName}/{sourceBlobFilename}";

                var edavContainerClient = edavBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                await edavContainerClient.CreateIfNotExistsAsync();

                BlobClient edavDestBlobClient = edavContainerClient.GetBlobClient(destinationBlobFilename);

                using var dexBlobStream = await dexBlobClient.OpenReadAsync();
                {
                    await edavDestBlobClient.UploadAsync(dexBlobStream, null, destinationMetadata);
                    dexBlobStream.Close();
                }
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to Edav");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        /// <summary>
        /// Copies a blob file from DEX to ROUTING asynchronously.
        /// </summary>
        /// <param name="sourceContainerName">Source container name</param>
        /// <param name="sourceBlobFilename">Source blob filename</param>
        /// <param name="destinationMetadata">Destination metadata to be associated with the blob file</param>
        /// <returns></returns>
        private async Task CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata)
        {
            try
            {

                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_routingStorageAccountName};AccountKey={_routingStorageAccountKey};EndpointSuffix=core.windows.net";

                BlobServiceClient blobServiceClient = new($"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net");
                BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(sourceContainerName);
                BlobClient dexBlobClient = containerClient.GetBlobClient(sourceBlobFilename);

                var routingBlobServiceClient = new BlobServiceClient(connectionString);

                // _routingUploadRootContainerName could be set to empty, then no root container in routing

                string destinationContainerName = string.IsNullOrEmpty(_routingUploadRootContainerName) ? sourceContainerName : _routingUploadRootContainerName;
                string destinationBlobFilename = string.IsNullOrEmpty(_routingUploadRootContainerName) ? sourceBlobFilename : $"{sourceContainerName}/{sourceBlobFilename}";

                var routingContainerClient = routingBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                await routingContainerClient.CreateIfNotExistsAsync();

                BlobClient routingDestBlobClient = routingContainerClient.GetBlobClient(destinationBlobFilename);

                using var dexBlobStream = await dexBlobClient.OpenReadAsync();
                {
                    await routingDestBlobClient.UploadAsync(dexBlobStream, null, destinationMetadata);
                    dexBlobStream.Close();
                }
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to ROUTING");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        /// <summary>
        /// Copies a blob from the tus upload folder to the DEX storage account
        /// </summary>
        /// <param name="connectionString">Connection string for both the source and destination storage account</param>
        /// <param name="sourceContainerName">Source container name for the file to copy</param>
        /// <param name="sourceBlobName">Source blob filename to copy</param>
        /// <param name="destinationContainerName">Destination container name for the copied file</param>
        /// <param name="destinationBlobName">Destination blob filename</param>
        /// <param name="destinationMetadata">Metadata to be associated with the destination blob file</param>
        /// <returns></returns>
        private async Task CopyBlobFromTusToDexAsync(string connectionString, string sourceContainerName, string sourceBlobName, string destinationContainerName,
            string destinationBlobName, IDictionary<string, string> destinationMetadata)
        {
            try
            {
                _logger.LogInformation($"Creating destination container client, container name: {destinationContainerName}");

                var sourceContainerClient = new BlobContainerClient(connectionString, sourceContainerName);
                var destinationContainerClient = new BlobContainerClient(connectionString, destinationContainerName);

                // Create the destination container if not exists
                await destinationContainerClient.CreateIfNotExistsAsync();

                _logger.LogInformation("Creating source blob client");

                // Create a BlobClient representing the source blob to copy.
                BlobClient sourceBlob = sourceContainerClient.GetBlobClient(sourceBlobName);

                // Get a BlobClient representing the destination blob with a unique name.
                BlobClient destBlob = destinationContainerClient.GetBlobClient(destinationBlobName);

                await _blobCopyHelper.CopyBlobAsync(sourceBlob, destBlob, destinationMetadata);
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError("Failed to copy from TUS to Dex");
                ExceptionUtils.LogErrorDetails(ex, _logger);
            }
        }

        /// <summary>
        /// Checks that all the required metadata fields are present for a given tus file.
        /// </summary>
        /// <param name="tusInfoFile">Contains all the tus file metadata</param>
        /// <param name="destinationId">Destination ID from the tus info file metadata</param>
        /// <param name="extEvent">External event from the tus info file metadata</param>
        /// <exception cref="TusInfoFileException"></exception>
        /// <exception cref="UploadConfigException"></exception>
        private void GetRequiredMetaData(TusInfoFile tusInfoFile, out string destinationId, out string extEvent)
        {
            if (tusInfoFile.MetaData == null)
                throw new TusInfoFileException("tus info file required metadata is missing");

            var metaDestinationId = tusInfoFile.MetaData!.GetValueOrDefault("meta_destination_id", null);
            if (metaDestinationId == null)
                throw new TusInfoFileException("meta_destination_id is a required metadata field and is missing from the tus info file");
            destinationId = metaDestinationId;

            var metaExtEvent = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_event", null);
            if (metaExtEvent == null)
                throw new TusInfoFileException("meta_ext_event is a required metadata field and is missing from the tus info file");
            extEvent = metaExtEvent;
        }

        /// <summary>
        /// Provides the filename from the metadata using the upload config to tell us what field to look for.
        /// </summary>
        /// <param name="tusInfoFile">Contains all the tus file metadata</param>
        /// <param name="metadataFilenameFieldName">Metadata filename field name to use</param>
        /// <param name="filename">Outputted filename to use</param>
        /// <exception cref="UploadConfigException"></exception>
        /// <exception cref="TusInfoFileException"></exception>
        private void GetFilenameFromMetaData(TusInfoFile tusInfoFile, string? metadataFilenameFieldName, out string filename)
        {
            filename = "";
            if (metadataFilenameFieldName != null)
            {
                var prefFilenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault(metadataFilenameFieldName, null);
                if (prefFilenameFromMetaData != null && prefFilenameFromMetaData.Length > 0)
                    filename = prefFilenameFromMetaData;
                else
                    throw new UploadConfigException($"The metadata field value for filename ({metadataFilenameFieldName}) provided is either empty or missing");
            }

            if (filename == "")
            {
                var filenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("filename", null);
                var extfilenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_filename", null);

                // this is needed for DEX HL7 and is a required field in dex_hl7_metadata_definition.json
                var originalFileName = tusInfoFile.MetaData!.GetValueOrDefault("original_filename", null);

                if (filenameFromMetaData != null)
                    filename = filenameFromMetaData;
                else if (extfilenameFromMetaData != null)
                    filename = extfilenameFromMetaData;
                else if (originalFileName != null)
                    filename = originalFileName;
                else
                    throw new TusInfoFileException("filename, meta_ext_filename, or original_filename is a required metadata field and is missing from the tus info file");
            }
        }

        /// <summary>
        /// Determines the folder path from the upload configuration.
        /// </summary>
        /// <param name="uploadConfig"></param>
        /// <param name="dateTimeNow"></param>
        /// <returns></returns>
        private string GetFolderPath(UploadConfig uploadConfig, DateTime dateTimeNow)
        {
            string folderPath;
            switch (uploadConfig.FolderStructure)
            {
                case "root":
                    // Don't partition uploads into any subfolders - all uploads will reside in the root folder
                    folderPath = "";
                    break;
                case "path":
                    folderPath = uploadConfig.FixedFolderPath ?? "";
                    break;
                case "date_YYYY_MM_DD":
                    // Partitioning is part of the filename where slashes will create subfolders.
                    // Path inside of that is year / month / day / filename
                    folderPath = $"{dateTimeNow.Year}/{dateTimeNow.Month.ToString().PadLeft(2, '0')}/{dateTimeNow.Day.ToString().PadLeft(2, '0')}";
                    break;
                default:
                    _logger.LogWarning("No upload folder structure scheme provided or one provided is unrecognized, using root");
                    folderPath = "";
                    break;
            }

            return folderPath;
        }

        /// <summary>
        /// Determines the filename suffix from the upload configuration.
        /// </summary>
        /// <param name="uploadConfig"></param>
        /// <param name="dateTimeNow"></param>
        /// <returns></returns>
        private string GetFilenameSuffix(UploadConfig uploadConfig, DateTime dateTimeNow)
        {
            string filenameSuffix;
            switch (uploadConfig.FilenameSuffix)
            {
                case "none":
                    filenameSuffix = ""; // no suffix
                    break;
                case "clock_ticks":
                    filenameSuffix = $"_{dateTimeNow.Ticks}";
                    break;
                default:
                    _logger.LogWarning("No filename suffix scheme provided or one provided is unrecognized, using none");
                    filenameSuffix = ""; // no suffix
                    break;
            }

            return filenameSuffix;
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
