using Azure;
using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using Azure.Identity;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public class UploadProcessingService : IUploadProcessingService
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
        private readonly IUploadEventHubService _uploadEventHubService;
        private readonly IFeatureManagementExecutor _featureManagementExecutor;
        private readonly IProcStatClient _procStatClient;
        private readonly string _stageName = "dex-file-copy";
        private readonly string _dexStorageAccountConnectionString;


        public UploadProcessingService(ILoggerFactory loggerFactory, IConfiguration configuration, IProcStatClient procStatClient, IFeatureManagementExecutor featureManagementExecutor, IUploadEventHubService uploadEventHubService)
        {
            _logger = loggerFactory.CreateLogger<UploadProcessingService>();
            _configuration = configuration;
            _blobCopyHelper = new(_logger);

            _featureManagementExecutor = featureManagementExecutor;
            _procStatClient = procStatClient;

            _tusAzureObjectPrefix = Environment.GetEnvironmentVariable("TUS_AZURE_OBJECT_PREFIX", EnvironmentVariableTarget.Process) ?? "tus-prefix";
            _tusAzureStorageContainer = Environment.GetEnvironmentVariable("TUS_AZURE_STORAGE_CONTAINER", EnvironmentVariableTarget.Process) ?? "bulkuploads";
            _dexAzureStorageAccountName = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";
            _edavAzureStorageAccountName = Environment.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";

            _routingStorageAccountName = Environment.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _routingStorageAccountKey = Environment.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";

            _edavUploadRootContainerName = Environment.GetEnvironmentVariable("EDAV_UPLOAD_ROOT_CONTAINER_NAME", EnvironmentVariableTarget.Process) ?? "upload";
            _routingUploadRootContainerName = Environment.GetEnvironmentVariable("ROUTING_UPLOAD_ROOT_CONTAINER_NAME", EnvironmentVariableTarget.Process) ?? "routeingress";

            _tusHooksFolder = Environment.GetEnvironmentVariable("TUSD_HOOKS_FOLDER", EnvironmentVariableTarget.Process) ?? "tusd-file-hooks";

            _destinationAndEvents = GetAllDestinationAndEvents();

            _uploadEventHubService = uploadEventHubService;
            _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";
        }

        /// <summary>
        /// Processeses the given blob created event from the URL provided.
        /// </summary>
        /// <param name="blobCreatedUrl"></param>
        /// <returns></returns>
        /// <exception cref="Exception"></exception>
        public async Task ProcessBlob(string blobCreatedUrl)
        {
            _logger.LogInformation($"TUS_AZURE_OBJECT_PREFIX={_tusAzureObjectPrefix}, TUS_AZURE_STORAGE_CONTAINER={_tusAzureStorageContainer}, DEX_AZURE_STORAGE_ACCOUNT_NAME={_dexAzureStorageAccountName}");

            Trace? trace = null;
            Span? copySpan = null;

            string? destinationContainerName = null;
            string? destinationBlobFilename = null;
            Dictionary<string, string> tusFileMetadata = null;
            string? destinationId = null;
            string? eventType = null;

            TusInfoFile? tusInfoFile = null;

            try
            {
                _logger.LogInformation($"Processing blob url: {blobCreatedUrl}");

                var sourceBlobUri = new Uri(blobCreatedUrl);
                string tusInfoFilename = $"{sourceBlobUri.Segments.Last()}.info";
                _logger.LogInformation($"tusPayloadFilename is {tusInfoFilename}");

                // START SPAN
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    trace = await _procStatClient.GetTraceByUploadId(tusInfoFilename.Replace(".info", ""));
                    copySpan = await _procStatClient.StartSpanForTrace(trace.TraceId, trace.SpanId, _stageName);
                });

                var tusPayloadPathname = $"/{_tusAzureObjectPrefix}/{tusInfoFilename}";
                
                tusInfoFile = await GetTusFileInfo(tusPayloadPathname);

                if (tusInfoFile.ID == null)
                {
                    throw new Exception("Malformed tus info file. No ID provided.");               
                }

                GetRequiredMetaData(tusInfoFile, out destinationId, out eventType);

                // Get V2 upload config file.
                UploadConfig uploadConfig = await GetUploadConfig(MetadataVersion.V2, destinationId, eventType);

                HydrateMetadata(tusInfoFile, uploadConfig, trace.TraceId, trace.SpanId);
                string? filename = tusInfoFile.MetaData.GetValueOrDefault("received_filename", null);

                var dateTimeNow = DateTime.UtcNow;

                // Determine the folder path and filename suffix from the upload configuration.
                var folderPath = GetFolderPath(uploadConfig, dateTimeNow);
                var filenameSuffix = GetFilenameSuffix(uploadConfig, dateTimeNow);

                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
                
                destinationBlobFilename = $"{folderPath}/{fileNameWithoutExtension}{filenameSuffix}{fileExtension}";

                // Container name is "{meta_destination_id}-{extEvent}"
                // There are some restrictions on container names -- underscores not allowed, must be all lowercase
                destinationContainerName = $"{destinationId.ToLower()}-{eventType.ToLower()}";

                // Copy the blob to the DeX storage account specific to the program, partitioned by date
                string dexBlobUrl = await CopyBlobFromTusToDex(tusPayloadPathname, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                await CopyBlobFromDexToTarget(dexBlobUrl, destinationId, eventType, destinationContainerName, destinationBlobFilename, tusFileMetadata);                
            }
            catch (Exception ex)
            {
                _logger.LogInformation($"Errors during blob processing: {blobCreatedUrl}");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                await PublishRetryEvent(BlobCopyStage.CopyToDex, blobCreatedUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                // CREATE FAILURE REPORT
                SendFailureReport(tusInfoFile.ID, destinationId, eventType, blobCreatedUrl, destinationContainerName, $"Failed to copy from Tus to DEX. {ex.Message}");
            }
            finally
            {
                // STOP SPAN
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    if (trace == null)
                    {
                        _logger.LogError("Trace was null when expecting a value.");
                    }

                    if (copySpan == null)
                    {
                        _logger.LogError("Span was null when expecting a value.");
                    }

                    if (trace?.TraceId != null)
                    {
                        if (copySpan?.SpanId != null)
                        {
                            await _procStatClient.StopSpanForTrace(trace.TraceId, copySpan.SpanId);
                        } else
                        {
                            _logger.LogError($"Span ID was null when expecting a value. {copySpan}");
                        }
                    } else
                    {
                        _logger.LogError($"Trace ID was null when expecting a value. {trace}");

                    }
                });                
            }
        }
        private async Task<UploadConfig> GetUploadConfig(MetadataVersion version, string destinationId, string eventType)
        {
            var uploadConfig = UploadConfig.Default;

            try
            {
                // Determine the filename and subfolder creation schemes for this destination/event.
                var configFilename = $"{version}/{destinationId}-{eventType}.json";
                var blobReader = new BlobReader(_logger);
                uploadConfig = await blobReader.GetObjectFromBlobJsonContent<UploadConfig>(_dexStorageAccountConnectionString, "upload-configs", configFilename);
            } catch (Exception e)
            {
                _logger.LogError($"No upload config found for destination id = {destinationId}, ext event = {eventType}.  Using default config. Exception = ${e.Message}");
            }

            return uploadConfig;
        }

        /// <summary>
        /// Copies a blob from the tus upload folder to the DEX storage account
        /// </summary>
        /// <param name="sourceBlobName">Source blob filename to copy</param>
        /// <param name="destinationContainerName">Destination container name for the copied file</param>
        /// <param name="destinationBlobName">Destination blob filename</param>
        /// <param name="destinationMetadata">Metadata to be associated with the destination blob file</param>
        /// <returns></returns>
        private async Task<string> CopyBlobFromTusToDex(string sourceBlobName, string destinationContainerName,
            string destinationBlobName, IDictionary<string, string> destinationMetadata)
        {
            try
            {
                _logger.LogInformation($"Creating destination container client, container name: {destinationContainerName}");

                var sourceContainerClient = new BlobContainerClient(_dexStorageAccountConnectionString, _tusAzureStorageContainer);
                var destinationContainerClient = new BlobContainerClient(_dexStorageAccountConnectionString, destinationContainerName);

                // Create the destination container if not exists
                await destinationContainerClient.CreateIfNotExistsAsync();

                _logger.LogInformation("Creating source blob client");

                // Create a BlobClient representing the source blob to copy.
                BlobClient sourceBlob = sourceContainerClient.GetBlobClient(sourceBlobName);

                // Get a BlobClient representing the destination blob with a unique name.
                BlobClient destBlob = destinationContainerClient.GetBlobClient(destinationBlobName);

                await _blobCopyHelper.CopyBlobAsync(sourceBlob, destBlob, destinationMetadata);

                return destBlob.Uri.ToString();
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError("Failed to copy from TUS to Dex");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                throw ex;
            }
        }

        private async Task CopyBlobFromDexToTarget(string sourceBlobUrl, string destinationId, string eventType, string destinationContainerName, string destinationBlobFilename, Dictionary<string, string> tusFileMetadata)
        {
            var uploadId = tusFileMetadata["tus_tguid"];

            CopyTarget[] targets = GetCopyTargets(destinationId, eventType);

            foreach (CopyTarget copyTarget in targets)
            {
                _logger.LogInformation("Copy Target: " + copyTarget.target);

                if (copyTarget.target == _targetEdav)
                {
                    // Now copy the file from DeX to the EDAV storage account, also partitioned by date
                    try 
                    {
                        var destPath = await CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);

                        // Send copy success report
                        SendSuccessReport(uploadId, destinationId, eventType, sourceBlobUrl, destPath);
                    }
                    catch(Exception ex)
                    {
                        await PublishRetryEvent(BlobCopyStage.CopyToEdav, sourceBlobUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                        // Send copy failure report
                        SendFailureReport(uploadId, destinationId, eventType, sourceBlobUrl, destinationContainerName, $"Failed to copy from Dex to EDAV. {ex.Message}");
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
                            var destPath = await CopyBlobFromDexToRoutingAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);

                            // Send copy success report
                            SendSuccessReport(uploadId, destinationId, eventType, sourceBlobUrl, destPath);
                        }
                        catch(Exception ex)
                        {
                            await PublishRetryEvent(BlobCopyStage.CopyToRouting, sourceBlobUrl, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                            // Send copy failure report
                            SendFailureReport(uploadId, destinationId, eventType, sourceBlobUrl, destinationContainerName, $"Failed to copy from Dex to ROUTING. {ex.Message}");
                        }
                    }
                    else
                    {
                        _logger.LogInformation($"Routing is disabled. Bypassing routing for blob");
                    }
                }
            }        
        }

        /// <summary>
        /// Copies a blob file from DEX to EDAV asynchronously.
        /// </summary>
        /// <param name="sourceContainerName">Source container name</param>
        /// <param name="sourceBlobFilename">Source blob filename</param>
        /// <param name="destinationMetadata">Destination metadata to be associated with the blob file</param>
        /// <returns></returns>
        public async Task<string> CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata)
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

                return edavDestBlobClient.Uri.ToString();
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to Edav");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                throw ex;
            }
        }

        /// <summary>
        /// Copies a blob file from DEX to ROUTING asynchronously.
        /// </summary>
        /// <param name="sourceContainerName">Source container name</param>
        /// <param name="sourceBlobFilename">Source blob filename</param>
        /// <param name="destinationMetadata">Destination metadata to be associated with the blob file</param>
        /// <returns></returns>
        public async Task<string> CopyBlobFromDexToRoutingAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata)
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

                return routingDestBlobClient.Uri.ToString();
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to ROUTING");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                throw ex;
            }
        }
        
        /// <summary>
        /// Returns the metadata from a tus .info file for the pathname provided.
        /// </summary>
        /// <param name="tusPayloadPathname">Full path of the file to get info on</param>
        /// <returns></returns>
        /// <exception cref="TusInfoFileException"></exception>
        private async Task<TusInfoFile> GetTusFileInfo(string tusInfoFilename)
        {
            TusInfoFile tusInfoFile;

            try
            {
                _logger.LogInformation($"Retrieving tus info file: {tusInfoFilename}");

                var blobReader = new BlobReader(_logger);

                tusInfoFile = await blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, _tusAzureStorageContainer, tusInfoFilename);
            }
            catch (Exception e)
            {
                throw new TusInfoFileException(e.Message);
            }

            _logger.LogInformation($"Info file metadata keys: {string.Join(", ", tusInfoFile.MetaData?.Keys.ToList())}");

            return tusInfoFile;
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
        /// <param name="metadata">Contains all the tus file metadata</param>
        /// <param name="metadataFields">All metadata fields for a particular use case</param>
        /// <param name="filenameFieldName">The filename e</param>
        /// <exception cref="UploadConfigException"></exception>
        /// <exception cref="TusInfoFileException"></exception>
        /*
        private string GetFilenameFromMetaData(Dictionary<string, string?> metadata, List<MetadataField> metadataFields, string filenameFieldName)
        {
            string? filename;
            MetadataField filenameField;

            try
            {
                filenameField = metadataFields.Single(field => field.FieldName == filenameFieldName);
            } catch (InvalidOperationException ex)
            {
                throw new UploadConfigException($"The metadata field value for filename ({filenameFieldName}) provided is either empty or missing");
            }

            filename = metadata.GetValueOrDefault(filenameFieldName, null);

            if (filename == null)
            {
                throw new TusInfoFileException($"No filename provided for field {filenameFieldName} and metadata {metadata}");
            }

            return filename;
        }
        */

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

        private void HydrateMetadata(TusInfoFile tusInfoFile, UploadConfig uploadConfig, string traceId, string spanId)
        {
            // Add use-case specific fields and their values.
            foreach (MetadataField field in uploadConfig.MetadataFields)
            {
                tusInfoFile.MetaData.Add(field.FieldName, tusInfoFile.MetaData.GetValueOrDefault(field.CompatFieldName, null));
            }

            // Add common fields and their values.
            tusInfoFile.MetaData.Add("tus_tguid", tusInfoFile.ID); // TODO: verify this field can be replaced with upload_id only.
            tusInfoFile.MetaData.Add("upload_id", tusInfoFile.ID);
            tusInfoFile.MetaData.Add("trace_id", traceId);
            tusInfoFile.MetaData.Add("span_id", spanId);
            tusInfoFile.MetaData.Remove("filename");
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

        private CopyTarget[] GetCopyTargets(string destinationId, string eventType)
        {
            // Default to copy to edav.
            CopyTarget[] targets = { new(_targetEdav) };

            var currentDestination = _destinationAndEvents.Result?.Find(d => d.destinationId == destinationId);
            var currentEvent = currentDestination?.extEvents?.Find(e => e.name == eventType);

            if (currentEvent != null && currentEvent.copyTargets != null)
            {  
                if(currentEvent.copyTargets.Count == 0)
                {
                    _logger.LogInformation($"No copy targets configured for {destinationId} and {eventType}");
                    _logger.LogInformation("Defaulting to EDAV");
                }
                else
                {
                    targets = currentEvent.copyTargets.ToArray();
                }
            }

            return targets;
        }

        private void SendSuccessReport(string uploadId, string destinationId, string eventType, string sourceBlobUrl, string destPath)
        {
            _featureManagementExecutor.ExecuteIfEnabled(Constants.PROC_STAT_FEATURE_FLAG_NAME, () =>
            {
                var successReport = new CopyReport(sourceUrl: sourceBlobUrl, destUrl: destPath, result: "success");
                _procStatClient.CreateReport(uploadId, destinationId, eventType, Constants.PROC_STAT_REPORT_STAGE_NAME, successReport);
            });
        }
        private void SendFailureReport(string uploadId, string destinationId, string eventType, string sourceBlobUrl, string destinationContainerName, string error)
        {
            _featureManagementExecutor.ExecuteIfEnabled(Constants.PROC_STAT_FEATURE_FLAG_NAME, () =>
            {
                CopyReport failReport = new CopyReport(sourceUrl: sourceBlobUrl, destUrl: destinationContainerName, result: "failure", errorDesc: error);
                _procStatClient.CreateReport(uploadId, destinationId, eventType, Constants.PROC_STAT_REPORT_STAGE_NAME, failReport);
            });
        }
    }
}