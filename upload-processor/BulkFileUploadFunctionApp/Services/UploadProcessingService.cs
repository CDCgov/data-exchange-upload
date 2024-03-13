using Azure;
using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using Azure.Identity;
using System.Text.Json;

using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;

namespace BulkFileUploadFunctionApp.Services
{
    public class UploadProcessingService : IUploadProcessingService
    {
        private readonly ILogger _logger;
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
        public async Task<CopyPrereqs> GetCopyPrereqs(string blobCreatedUrl)
        {
            string? uploadId = null;
            string? destinationId = null;
            string? eventType = null;

            string? destinationContainerName = null;

            Trace? trace = null;

            try
            {
                var sourceBlobUri = new Uri(blobCreatedUrl);
                string tusPayloadFilename = $"/{_tusAzureObjectPrefix}/{sourceBlobUri.Segments.Last()}";

                // Get metadata
                TusInfoFile tusInfoFile = await GetTusInfoFile(tusPayloadFilename);
                uploadId = tusInfoFile.ID;

                // Get trace
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    trace = await _procStatClient.GetTraceByUploadId(uploadId);
                });

                // Get Destination and Event type
                // TODO: Refactor to something with more generic language, as destination and event are deprecated terms.
                var metaDestinationId = tusInfoFile.MetaData!.GetValueOrDefault("meta_destination_id", null);
                if (metaDestinationId == null)
                    throw new TusInfoFileException("meta_destination_id is a required metadata field and is missing from the tus info file");
                destinationId = metaDestinationId;

                var metaExtEvent = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_event", null);
                if (metaExtEvent == null)
                    throw new TusInfoFileException("meta_ext_event is a required metadata field and is missing from the tus info file");
                eventType = metaExtEvent;

                // Get upload configs for destination and event type
                UploadConfig uploadConfig = await GetUploadConfig(MetadataVersion.V2, destinationId, eventType);

                // Hydrate metadata
                HydrateMetadata(tusInfoFile, uploadConfig, trace.TraceId, trace.SpanId);
                string? filename = tusInfoFile.MetaData!.GetValueOrDefault("received_filename", null);

                // Get dex folder and filename 
                var dateTimeNow = DateTime.UtcNow;

                // Determine the folder path and filename suffix from the upload configuration.
                var folderPath = GetFolderPath(uploadConfig, dateTimeNow);
                var filenameSuffix = GetFilenameSuffix(uploadConfig, dateTimeNow);

                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
            
                string destinationBlobFilename = $"{folderPath}/{fileNameWithoutExtension}{filenameSuffix}{fileExtension}";

                // Container name is "{meta_destination_id}-{extEvent}"
                // There are some restrictions on container names -- underscores not allowed, must be all lowercase
                destinationContainerName = $"{destinationId.ToLower()}-{eventType.ToLower()}";

                // Get copy targets
                CopyTarget[] targets = GetCopyTargets(destinationId, eventType);
                
                return new CopyPrereqs(uploadId,
                                       blobCreatedUrl,
                                       tusPayloadFilename, 
                                       destinationId, 
                                       eventType, 
                                       destinationContainerName, 
                                       destinationBlobFilename, 
                                       tusInfoFile.MetaData, 
                                       targets,
                                       trace);
            }
            catch(Exception ex)
            {
                _logger.LogError("Failed to copy from TUS to Dex");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                
                // Send copy failure report
                SendFailureReport(uploadId, destinationId, eventType, blobCreatedUrl, destinationContainerName, $"Failed to get copy preqs: {ex.Message}");

                throw ex;
            }
        }
        public async Task CopyAll(CopyPrereqs copyPrereqs)
        {
            Span? copySpan = null;

            try
            {
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    copySpan = await _procStatClient.StartSpanForTrace(copyPrereqs.Trace.TraceId, copyPrereqs.Trace.SpanId, _stageName);
                });

                copyPrereqs.DexBlobUrl = await CopyFromTusToDex(copyPrereqs);

                // copy to targets
                await CopyFromDexToTarget(copyPrereqs);
            }
            catch(Exception ex)
            {
                ExceptionUtils.LogErrorDetails(ex, _logger);
                throw ex;
            }
            finally
            {
                if(copySpan != null) 
                {
                    await _procStatClient.StopSpanForTrace(copyPrereqs.Trace.TraceId, copySpan.SpanId);
                }
            }
        }
               
        /// <summary>
        /// Copies a blob from the tus upload folder to the DEX storage account
        /// </summary>
        /// <param name="copyPreqs">Copy preqs</param>
        /// <returns>dexBlobUrl</returns>
        private async Task<string> CopyFromTusToDex(CopyPrereqs copyPrereqs)
        {
            try
            {
                _logger.LogInformation($"Creating destination container client, container name: {copyPrereqs.DexBlobFolderName}");

                var sourceContainerClient = new BlobContainerClient(_dexStorageAccountConnectionString, _tusAzureStorageContainer);
                var destinationContainerClient = new BlobContainerClient(_dexStorageAccountConnectionString, copyPrereqs.DexBlobFolderName);

                // Create the destination container if not exists
                await destinationContainerClient.CreateIfNotExistsAsync();

                _logger.LogInformation("Creating source blob client");

                // Create a BlobClient representing the source blob to copy.
                BlobClient sourceBlob = sourceContainerClient.GetBlobClient(copyPrereqs.TusPayloadFilename);

                // Get a BlobClient representing the destination blob with a unique name.
                BlobClient destBlob = destinationContainerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                await _blobCopyHelper.CopyBlobAsync(sourceBlob, destBlob, copyPrereqs.Metadata);

                return destBlob.Uri.ToString();
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError("Failed to copy blob from TUS to Dex");

                // Send copy failure report
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    SendFailureReport(copyPrereqs.UploadId, 
                                      copyPrereqs.DestinationId, 
                                      copyPrereqs.EventType, 
                                      copyPrereqs.SourceBlobUrl, 
                                      copyPrereqs.DexBlobFolderName, 
                                      $"Failed to copy blob from TUS to DEX. {ex.Message}");
                });

                throw ex;
            }
        }

        private async Task CopyFromDexToTarget(CopyPrereqs copyPrereqs)
        {
            foreach (CopyTarget copyTarget in copyPrereqs.Targets)
            {
                _logger.LogInformation("Copy Target: " + copyTarget.target);

                if (copyTarget.target == _targetEdav)
                {
                    try
                    {
                        await CopyFromDexToEdav(copyPrereqs);
                    }
                    catch(Exception ex)
                    {
                        // publish retry event
                        await PublishRetryEvent(BlobCopyStage.CopyToEdav,
                                                copyPrereqs);
                    }
                }
                else if (copyTarget.target == _targetRouting)
                {
                    try
                    {
                        await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.ROUTING_FEATURE_FLAG_NAME, async () =>
                        {
                            await CopyFromDexToRouting(copyPrereqs);
                        });

                        _featureManagementExecutor.ExecuteIfDisabled(Constants.ROUTING_FEATURE_FLAG_NAME, () =>
                        {
                            _logger.LogInformation($"Routing is disabled. Bypassing routing for blob");
                        });
                    } catch (Exception ex)
                    {
                        // publish retry event
                        await PublishRetryEvent(BlobCopyStage.CopyToRouting,
                                                copyPrereqs);
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
        /// 
        public async Task CopyFromDexToEdav(CopyPrereqs copyPrereqs)
        {
            string? destinationContainerName = null;

            try
            {
                BlobServiceClient blobServiceClient = new($"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net");
                BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(copyPrereqs.DexBlobFolderName);
                BlobClient dexBlobClient = containerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                var edavBlobServiceClient = new BlobServiceClient(
                    new Uri($"https://{_edavAzureStorageAccountName}.blob.core.windows.net"),
                    new DefaultAzureCredential() // using Service Principal
                );

                // _edavUploadRootContainerName could be set to empty, then no root container in edav

                destinationContainerName = string.IsNullOrEmpty(_edavUploadRootContainerName) ? copyPrereqs.DexBlobFolderName : _edavUploadRootContainerName;
                string destinationBlobFilename = string.IsNullOrEmpty(_edavUploadRootContainerName) ? copyPrereqs.DexBlobFileName : $"{copyPrereqs.DexBlobFolderName}/{copyPrereqs.DexBlobFileName}";

                var edavContainerClient = edavBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                await edavContainerClient.CreateIfNotExistsAsync();

                BlobClient edavDestBlobClient = edavContainerClient.GetBlobClient(destinationBlobFilename);

                using var dexBlobStream = await dexBlobClient.OpenReadAsync();
                {
                    await edavDestBlobClient.UploadAsync(dexBlobStream, null, copyPrereqs.Metadata);
                    dexBlobStream.Close();
                }

                // Send copy success report
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    SendSuccessReport(copyPrereqs.UploadId, 
                                      copyPrereqs.DestinationId, 
                                      copyPrereqs.EventType, 
                                      copyPrereqs.DexBlobUrl, 
                                      edavDestBlobClient.Uri.ToString());
                });
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to Edav");
                ExceptionUtils.LogErrorDetails(ex, _logger);               

                // Send copy failure report
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    SendFailureReport(copyPrereqs.UploadId, 
                                      copyPrereqs.DestinationId, 
                                      copyPrereqs.EventType, 
                                      copyPrereqs.DexBlobUrl, 
                                      destinationContainerName, 
                                      $"Failed to copy blob from DEX to EDAV. {ex.Message}");
                });

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
        public async Task CopyFromDexToRouting(CopyPrereqs copyPrereqs)
        {
            string? destinationContainerName = null;

            try
            {
                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_routingStorageAccountName};AccountKey={_routingStorageAccountKey};EndpointSuffix=core.windows.net";

                BlobServiceClient blobServiceClient = new($"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net");
                BlobContainerClient containerClient = blobServiceClient.GetBlobContainerClient(copyPrereqs.DexBlobFolderName);
                BlobClient dexBlobClient = containerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                var routingBlobServiceClient = new BlobServiceClient(connectionString);

                // _routingUploadRootContainerName could be set to empty, then no root container in routing

                destinationContainerName = string.IsNullOrEmpty(_routingUploadRootContainerName) ? copyPrereqs.DexBlobFolderName : _routingUploadRootContainerName;
                string destinationBlobFilename = string.IsNullOrEmpty(_routingUploadRootContainerName) ? copyPrereqs.DexBlobFileName : $"{copyPrereqs.DexBlobFolderName}/{copyPrereqs.DexBlobFileName}";

                var routingContainerClient = routingBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                await routingContainerClient.CreateIfNotExistsAsync();

                BlobClient routingDestBlobClient = routingContainerClient.GetBlobClient(destinationBlobFilename);

                using var dexBlobStream = await dexBlobClient.OpenReadAsync();
                {
                    await routingDestBlobClient.UploadAsync(dexBlobStream, null, copyPrereqs.Metadata);
                    dexBlobStream.Close();
                }

                // Send copy success report
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    SendSuccessReport(copyPrereqs.UploadId, 
                                      copyPrereqs.DestinationId, 
                                      copyPrereqs.EventType, 
                                      copyPrereqs.DexBlobUrl, 
                                      routingDestBlobClient.Uri.ToString());
                });
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to ROUTING");
                ExceptionUtils.LogErrorDetails(ex, _logger);

                // Send copy failure report
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    SendFailureReport(copyPrereqs.UploadId, 
                                      copyPrereqs.DestinationId, 
                                      copyPrereqs.EventType, 
                                      copyPrereqs.DexBlobUrl, 
                                      destinationContainerName, 
                                      $"Failed to copy blob from DEX to ROUTING. {ex.Message}");
                });

                throw ex;
            }
        }

        private async Task<TusInfoFile> GetTusInfoFile(string tusPayloadFilename)
        {
            // GET FILE METADATA
            string tusInfoFilename = $"{tusPayloadFilename}.info";             
            _logger.LogInformation($"Retrieving tus info file: {tusInfoFilename}");

            var blobReader = new BlobReader(_logger);

            TusInfoFile tusInfoFile = await blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, _tusAzureStorageContainer, tusInfoFilename);

            if (tusInfoFile.ID == null)
                throw new Exception("Malformed tus info file. No ID provided.");
            
            if (tusInfoFile.MetaData == null)
                throw new TusInfoFileException("tus info file required metadata is missing");

            return tusInfoFile;
        }

        private async Task<UploadConfig> GetUploadConfig(MetadataVersion version, string destinationId, string eventType)
        {
            var uploadConfig = UploadConfig.Default;
            var configFilename = $"{version.ToString().ToLower()}/{destinationId}-{eventType}.json";

            try
            {
                // Determine the filename and subfolder creation schemes for this destination/event.
                var blobReader = new BlobReader(_logger);
                uploadConfig = await blobReader.GetObjectFromBlobJsonContent<UploadConfig>(_dexStorageAccountConnectionString, "upload-configs", configFilename);
            } catch (Exception e)
            {
                _logger.LogError($"No upload config found for destination id = {destinationId}, ext event = {eventType}.  Using default config. Exception = ${e.Message}");
            }

            if (uploadConfig == null)
            {
                throw new UploadConfigException($"Unable to parse JSON for upload config {configFilename}");
            }

            return uploadConfig;
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

        private void HydrateMetadata(TusInfoFile tusInfoFile, UploadConfig uploadConfig, string traceId, string spanId)
        {
            if (tusInfoFile.MetaData == null || tusInfoFile.ID == null)
            {
                throw new ArgumentNullException("Metadata cannot be null.");
            }

            if (uploadConfig.MetadataConfig == null || uploadConfig.MetadataConfig.Fields == null)
            {
                throw new ArgumentNullException("Metadata fields cannot be null.");
            }

            // Add use-case specific fields and their values.
            foreach (MetadataField field in uploadConfig.MetadataConfig.Fields)
            {
                if (field.FieldName == null)
                {
                    _logger.LogError("Cannot parse field with null field name.");
                    continue;
                }

                // Skip if field already provided.
                if (tusInfoFile.MetaData.ContainsKey(field.FieldName))
                {
                    continue;
                }

                if (field.DefaultValue != null)
                {
                    tusInfoFile.MetaData[field.FieldName] = field.DefaultValue;
                    continue;
                }

                if (field.CompatFieldName != null)
                {
                    tusInfoFile.MetaData[field.FieldName] = tusInfoFile.MetaData.GetValueOrDefault(field.CompatFieldName, "");
                    continue;
                }

                tusInfoFile.MetaData.Add(field.FieldName, "");
            }

            // Add common fields and their values.
            tusInfoFile.MetaData["version"] = uploadConfig.MetadataConfig.Version;
            tusInfoFile.MetaData["tus_tguid"] = tusInfoFile.ID; // TODO: verify this field can be replaced with upload_id only.
            tusInfoFile.MetaData["upload_id"] = tusInfoFile.ID;
            tusInfoFile.MetaData["trace_id"] = traceId;
            tusInfoFile.MetaData["parent_span_id"] = spanId;
            tusInfoFile.MetaData.Remove("filename"); // Remove filename field to use standard received_filename field.
        }

        private async Task<List<DestinationAndEvents>?>  GetAllDestinationAndEvents()
        {
            var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";

            try
            {
                _logger.LogInformation($"Fetching Destinations and Events from  {_tusHooksFolder}/{_destinationAndEventsFileName}");

                var blobReader = new BlobReader(_logger);
                var destinationAndEvents = await blobReader.GetObjectFromBlobJsonContent<List<DestinationAndEvents>>(connectionString, _tusHooksFolder, _destinationAndEventsFileName);

                _logger.LogInformation($"Destinations And Events: {JsonSerializer.Serialize(destinationAndEvents)}");

                return destinationAndEvents;
            }
            catch (Exception e)
            {
                _logger.LogError("Failed to fetch Destinations and Events");
                ExceptionUtils.LogErrorDetails(e, _logger);

                return new List<DestinationAndEvents>();
            }
        }

        public async Task PublishRetryEvent(BlobCopyStage copyStage, CopyPrereqs copyPrereqs)
        {            
            try 
            {
                BlobCopyRetryEvent blobCopyRetryEvent = new BlobCopyRetryEvent();
                blobCopyRetryEvent.CopyRetryStage = copyStage;
                blobCopyRetryEvent.RetryAttempt = 1;
                blobCopyRetryEvent.CopyPrereqs = copyPrereqs;

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
            CopyTarget[] defaultTargets = { new(_targetEdav) };

            if (_destinationAndEvents.Result == null) {
                _logger.LogError($"Empty Destinations and Events: {_destinationAndEvents.Result}");
                throw new Exception("Empty Destinations and Events");
            }

            var currentDestination = _destinationAndEvents.Result?.Find(d => d.destinationId == destinationId);
            var currentEvent = currentDestination?.extEvents?.Find(e => e.name == eventType);

            if (currentEvent == null) {
                _logger.LogError($"No copy targets configured for {destinationId} and {eventType} - defaulting to EDAV");
                _logger.LogError($"Destinations And Events: {JsonSerializer.Serialize(_destinationAndEvents)}");
                return defaultTargets;
            }

            if (currentEvent.copyTargets == null || currentEvent.copyTargets.Count == 0)
            {
                _logger.LogError($"No copy targets configured for {destinationId} and {eventType} - defaulting to EDAV");
                _logger.LogError($"Destinations And Events: {JsonSerializer.Serialize(_destinationAndEvents)}");
                return defaultTargets;
            }

            return currentEvent.copyTargets.ToArray();
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