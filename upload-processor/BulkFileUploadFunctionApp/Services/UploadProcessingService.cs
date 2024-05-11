using Azure;
using Azure.Storage.Blobs; 
using Azure.Messaging.ServiceBus;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Configuration;
using Azure.Identity;
using System.Text.Json;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionApp.Model;
using System.Text.Json.Serialization;

namespace BulkFileUploadFunctionApp.Services
{
    public class UploadProcessingService : IUploadProcessingService
    {
        private readonly ILogger _logger;
        private readonly BlobCopyHelper _blobCopyHelper;
        private readonly IBlobReader _blobReader;
        private readonly string _tusAzureObjectPrefix;
        private readonly string _tusAzureStorageContainer;
        private readonly string _dexAzureStorageAccountName;
        private readonly string _dexAzureStorageAccountKey;
        private readonly string _edavAzureStorageAccountName;
        private readonly string _routingStorageAccountName;
        private readonly string _routingStorageAccountKey;
        private readonly string _edavUploadRootContainerName;
        private readonly string _routingUploadRootContainerName;
        private readonly IUploadEventHubService _uploadEventHubService;
        private readonly IFeatureManagementExecutor _featureManagementExecutor;
        private readonly IProcStatClient _procStatClient;
        private readonly string _dexStorageAccountConnectionString;
        private readonly string _routingStorageAccountConnectionString;
        private readonly BlobServiceClient _dexBlobServiceClient;
        private readonly BlobServiceClient _routingBlobServiceClient;
        private readonly BlobContainerClient _tusContainerClient;
        private readonly BlobServiceClient _edavBlobServiceClient;
        private readonly IBulkUploadSvcBusClient _bulkUploadSvcBusClient;
        private readonly ServiceBusSender _serviceBusSender;
        private readonly string _uploadConfigContainer; 
        private readonly string _ServiceBusConnectionString;
        private readonly string _ServiceBusQueueName;
        


        //TODO: Replace ProcStatClient with BulkUploadSvcBusClient
        public UploadProcessingService(ILoggerFactory loggerFactory, IBulkUploadSvcBusClient bulkUploadSvcBusClient,
        IFeatureManagementExecutor featureManagementExecutor, IUploadEventHubService uploadEventHubService, IBlobReaderFactory blobReaderFactory)
        {
            // Get environment variables.
            _tusAzureObjectPrefix = Environment.GetEnvironmentVariable("TUS_AZURE_OBJECT_PREFIX", EnvironmentVariableTarget.Process) ?? "tus-prefix";
            _tusAzureStorageContainer = Environment.GetEnvironmentVariable("TUS_AZURE_STORAGE_CONTAINER", EnvironmentVariableTarget.Process) ?? "bulkuploads";
            _dexAzureStorageAccountName = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";
            _edavAzureStorageAccountName = Environment.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _routingStorageAccountName = Environment.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _routingStorageAccountKey = Environment.GetEnvironmentVariable("ROUTING_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";
            _edavUploadRootContainerName = Environment.GetEnvironmentVariable("EDAV_UPLOAD_ROOT_CONTAINER_NAME", EnvironmentVariableTarget.Process) ?? "upload";
            _routingUploadRootContainerName = Environment.GetEnvironmentVariable("ROUTING_UPLOAD_ROOT_CONTAINER_NAME", EnvironmentVariableTarget.Process) ?? "routeingress";
            _uploadConfigContainer =  Environment.GetEnvironmentVariable("UPLOAD_CONFIGS", EnvironmentVariableTarget.Process) ?? "upload-configs";
            _ServiceBusConnectionString = Environment.GetEnvironmentVariable("SERVICE_BUS_CONNECTION_STR", EnvironmentVariableTarget.Process) ?? "";
            _ServiceBusQueueName = Environment.GetEnvironmentVariable("REPORT_QUEUE_NAME", EnvironmentVariableTarget.Process) ?? "";

            // Instantiate helper services.
            _logger = loggerFactory.CreateLogger<UploadProcessingService>();
            _blobCopyHelper = new(_logger);
            _blobReader = blobReaderFactory.CreateInstance(_logger);
            _featureManagementExecutor = featureManagementExecutor;

            _uploadEventHubService = uploadEventHubService;
            _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";
            _routingStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_routingStorageAccountName};AccountKey={_routingStorageAccountKey};EndpointSuffix=core.windows.net";

            // Instatiate container client connections.
            _dexBlobServiceClient = new BlobServiceClient(_dexStorageAccountConnectionString);
            _routingBlobServiceClient = new BlobServiceClient(_routingStorageAccountConnectionString);
            _tusContainerClient = _dexBlobServiceClient.GetBlobContainerClient(_tusAzureStorageContainer);
            _edavBlobServiceClient = new BlobServiceClient(
                new Uri($"https://{_edavAzureStorageAccountName}.blob.core.windows.net"),
                new DefaultAzureCredential() // using Service Principal
            );

            _bulkUploadSvcBusClient = bulkUploadSvcBusClient;           
        }

        public async Task<CopyPrereqs> GetCopyPrereqs(string blobCreatedUrl)
        {
            string? uploadId = null;
            MetadataVersion version = MetadataVersion.V1;
            string? useCase = null;
            string? useCaseCategory = null;
            string? destinationContainerName = null;
            Trace? trace = null;

            try
            {
                var sourceBlobUri = new Uri(blobCreatedUrl);
                string tusPayloadFilename = $"/{_tusAzureObjectPrefix}/{sourceBlobUri.Segments.Last()}";

                // Get metadata
                TusInfoFile tusInfoFile = await GetTusInfoFile(tusPayloadFilename);
                uploadId = tusInfoFile.ID;
                if (uploadId == null)
                {
                    throw new TusInfoFileException("received info file without an upload ID");
                }


                HydrateMetadata(tusInfoFile);

                // retrieve version from metadata or default to V1
                version = tusInfoFile.GetMetadataVersion();
                useCase = tusInfoFile.GetUseCase();
                useCaseCategory = tusInfoFile.GetUseCaseCategory();
                destinationContainerName = $"{useCase}-{useCaseCategory}";
                string uploadConfigFilename = $"{useCase}-{useCaseCategory}.json";

                var uploadConfig = await GetUploadConfig(uploadConfigFilename, version);
                _logger.LogInformation($"Got upload config for {version}: {JsonSerializer.Serialize(uploadConfig)}");
                // translate V1 metadata 
                if (version == MetadataVersion.V1)
                {
                    var uploadConfigV2 = await GetUploadConfig(uploadConfigFilename, MetadataVersion.V2);
                    _logger.LogInformation($"Translating to {JsonSerializer.Serialize(uploadConfigV2)}");
                    tusInfoFile.MetaData = TranslateMetadata(tusInfoFile.MetaData, uploadConfigV2);
                }
                
                string? filename = tusInfoFile.MetaData!.GetValueOrDefault("received_filename", null);

                // Get dex folder and filename 
                var dateTimeNow = DateTime.UtcNow;

                // Determine the folder path from the upload configuration.
                var folderPath = GetFolderPath(uploadConfig, dateTimeNow);
                var fileNameSuffix = GetFilenameSuffix(uploadConfig, uploadId);
                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
            
                string destinationBlobFilename = $"{folderPath}/{fileNameWithoutExtension}{fileNameSuffix}{fileExtension}";

                // Get copy targets
                List<CopyTargetsEnum> targets = uploadConfig.CopyConfig.TargetEnums;
                
                return new CopyPrereqs(uploadId,
                                    blobCreatedUrl,
                                    tusPayloadFilename, 
                                    useCase, 
                                    useCaseCategory, 
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
                SendFailureReport(uploadId, useCase, useCaseCategory, blobCreatedUrl, destinationContainerName, $"Failed to get copy preqs: {ex.Message}");

                throw ex;
            }
        }
        public async Task CopyAll(CopyPrereqs copyPrereqs)
        {
            Span? copySpan = null;

            try
            {

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
        public async Task<string> CopyFromTusToDex(CopyPrereqs copyPrereqs)
        {
            try
            {
                _logger.LogInformation($"Creating destination container client, container name: {copyPrereqs.DexBlobFolderName}");

                var destinationContainerClient = new BlobContainerClient(_dexStorageAccountConnectionString, copyPrereqs.DexBlobFolderName);

                // Create the destination container if not exists
                await destinationContainerClient.CreateIfNotExistsAsync();

                _logger.LogInformation("Creating source blob client");

                // Create a BlobClient representing the source blob to copy.
                BlobClient sourceBlob = _tusContainerClient.GetBlobClient(copyPrereqs.TusPayloadFilename);

                // Get a BlobClient representing the destination blob with a unique name.
                BlobClient destBlob = destinationContainerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                await _blobCopyHelper.CopyBlobLeaseAsync(sourceBlob, destBlob, copyPrereqs.Metadata);

                return destBlob.Uri.ToString();
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError("Failed to copy blob from TUS to Dex");

                // Send copy failure report

                SendFailureReport(copyPrereqs.UploadId, 
                                    copyPrereqs.UseCase, 
                                    copyPrereqs.UseCaseCategory, 
                                    copyPrereqs.SourceBlobUrl, 
                                    copyPrereqs.DexBlobFolderName, 
                                    $"Failed to copy blob from TUS to DEX. {ex.Message}");

                throw ex;
            }
        }

        private async Task CopyFromDexToTarget(CopyPrereqs copyPrereqs)
        {
            foreach (CopyTargetsEnum copyTarget in copyPrereqs.Targets)
            {
                _logger.LogInformation("Copy Target: " + copyTarget);

                if (copyTarget == CopyTargetsEnum.edav)
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
                else if (copyTarget == CopyTargetsEnum.routing)
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
            string destinationContainerName = _edavUploadRootContainerName ?? copyPrereqs.DexBlobFolderName;
            string destinationFilename = $"{copyPrereqs.DexBlobFolderName}/{copyPrereqs.DexBlobFileName}" ?? copyPrereqs.DexBlobFileName;

            try
            {
                BlobContainerClient sourceContainerClient = _dexBlobServiceClient.GetBlobContainerClient(copyPrereqs.DexBlobFolderName);
                BlobClient sourceBlobClient = sourceContainerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                BlobContainerClient destContainerClient = _edavBlobServiceClient.GetBlobContainerClient(destinationContainerName);
                await destContainerClient.CreateIfNotExistsAsync();

                BlobClient destBlobClient = destContainerClient.GetBlobClient(destinationFilename);

                await _blobCopyHelper.CopyBlobStreamAsync(sourceBlobClient, destBlobClient, copyPrereqs.Metadata);

                // Send copy success report
                SendSuccessReport(copyPrereqs.UploadId, 
                                    copyPrereqs.UseCase, 
                                    copyPrereqs.UseCaseCategory, 
                                    copyPrereqs.DexBlobUrl, 
                                    destBlobClient.Uri.ToString());
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to Edav");
                ExceptionUtils.LogErrorDetails(ex, _logger);               

                // Send copy failure report
                SendFailureReport(copyPrereqs.UploadId, 
                                    copyPrereqs.UseCase, 
                                    copyPrereqs.UseCaseCategory, 
                                    copyPrereqs.DexBlobUrl, 
                                    destinationContainerName, 
                                    $"Failed to copy blob from DEX to EDAV. {ex.Message}");

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
            string destinationContainerName = _routingUploadRootContainerName ?? copyPrereqs.DexBlobFolderName;
            string destinationFilename = $"{copyPrereqs.DexBlobFolderName}/{copyPrereqs.DexBlobFileName}" ?? copyPrereqs.DexBlobFileName;

            try
            {
                BlobContainerClient sourceContainerClient = _dexBlobServiceClient.GetBlobContainerClient(copyPrereqs.DexBlobFolderName);
                BlobClient sourceBlobClient = sourceContainerClient.GetBlobClient(copyPrereqs.DexBlobFileName);

                BlobContainerClient destContainerClient = _routingBlobServiceClient.GetBlobContainerClient(destinationContainerName);
                await destContainerClient.CreateIfNotExistsAsync();

                BlobClient destBlobClient = destContainerClient.GetBlobClient(destinationFilename);

                await _blobCopyHelper.CopyBlobStreamAsync(sourceBlobClient, destBlobClient, copyPrereqs.Metadata);

                // Send copy success report
                SendSuccessReport(copyPrereqs.UploadId, 
                                    copyPrereqs.UseCase, 
                                    copyPrereqs.UseCaseCategory, 
                                    copyPrereqs.DexBlobUrl, 
                                    destBlobClient.Uri.ToString());
            }
            catch (Exception ex)
            {
                _logger.LogError("Failed to copy from Dex to ROUTING");
                ExceptionUtils.LogErrorDetails(ex, _logger);

                // Send copy failure report
                SendFailureReport(copyPrereqs.UploadId, 
                                    copyPrereqs.UseCase, 
                                    copyPrereqs.UseCaseCategory, 
                                    copyPrereqs.DexBlobUrl, 
                                    destinationContainerName, 
                                    $"Failed to copy blob from DEX to ROUTING. {ex.Message}");

                throw ex;
            }
        }

        private string GetFilenameSuffix(UploadConfig uploadConfig, string uploadId)
        {
            string suffix = string.Empty; // Default to no suffix.

            if (uploadConfig.CopyConfig?.FilenameSuffix == CopyConfig.FILENAME_SUFFIX_UID)
            {
                suffix = $"_{uploadId}";
            }

            return suffix;
        }

        private async Task<TusInfoFile> GetTusInfoFile(string tusPayloadFilename)
        {
            // GET FILE METADATA
            string tusInfoFilename = $"{tusPayloadFilename}.info";             
            _logger.LogInformation($"Retrieving tus info file: {tusInfoFilename}");

            TusInfoFile tusInfoFile = await _blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, _tusAzureStorageContainer, tusInfoFilename);

            if (tusInfoFile.ID == null)
                throw new Exception("Malformed tus info file. No ID provided.");
            
            if (tusInfoFile.MetaData == null)
                throw new TusInfoFileException("tus info file required metadata is missing");

            return tusInfoFile;
        }

        private async Task<UploadConfig> GetUploadConfig(string filename, MetadataVersion versionNum)
        {
            var uploadConfig = UploadConfig.Default;
            var configFilename = $"{versionNum.ToString().ToLower()}/{filename}";

            try
            {
                // Determine the filename and subfolder creation schemes for this destination/event.
                uploadConfig = await _blobReader.GetObjectFromBlobJsonContent<UploadConfig>(_dexStorageAccountConnectionString, _uploadConfigContainer, configFilename);
            } catch (Exception e)
            {
                _logger.LogError($"No upload config found for {configFilename}.  Using default config. Exception = {e.Message}");
            }

            if (uploadConfig == null)
            {
                throw new UploadConfigException($"Unable to parse JSON for upload config {configFilename}");
            }

            // Convert copy target strings to enums.
            List<CopyTargetsEnum> targetEnums = uploadConfig.CopyConfig.Targets.ConvertAll(targetStr =>
            {
                Enum.TryParse(targetStr, out CopyTargetsEnum targetEnum);
                return targetEnum;
            }).ToList();
            uploadConfig.CopyConfig.TargetEnums = targetEnums;

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
            switch (uploadConfig.CopyConfig.FolderStructure)
            {
                case "root":
                    // Don't partition uploads into any subfolders - all uploads will reside in the root folder
                    folderPath = "";
                    break;
                case "path":
                    folderPath = uploadConfig.CopyConfig.FolderStructure ?? "";
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

        private Dictionary<string, string> TranslateMetadata(Dictionary<string, string> fromMetadata, UploadConfig toConfig)
        {
            Dictionary<string, string> toMetadata = new Dictionary<string, string>(fromMetadata);

            if (toConfig.MetadataConfig == null || toConfig.MetadataConfig.Fields == null || toConfig.MetadataConfig.Version == null)
            {
                throw new ArgumentNullException("UploadConfig Metadata fields cannot be null.");
            }

            // Add use-case specific fields and their values.
            foreach (MetadataField field in toConfig.MetadataConfig.Fields)
            {
                if (field.FieldName == null)
                {
                    _logger.LogError("Cannot parse field with null field name.");
                    continue;
                }

                // Skip if field already provided.
                if (toMetadata.ContainsKey(field.FieldName))
                {
                    continue;
                }

                if (field.DefaultValue != null)
                {
                    toMetadata[field.FieldName] = field.DefaultValue;
                    continue;
                }

                if (field.CompatFieldName != null)
                {
                    toMetadata[field.FieldName] = toMetadata.GetValueOrDefault(field.CompatFieldName, "");
                    continue;
                }

                toMetadata.Add(field.FieldName, "");
            }
            toMetadata["version"] = toConfig.MetadataConfig.Version;

            return toMetadata;
        }

        private void HydrateMetadata(TusInfoFile tusInfoFile)
        {
            // Add common fields and their values.
            tusInfoFile.MetaData["tus_tguid"] = tusInfoFile.ID; // TODO: verify this field can be replaced with upload_id only.
            tusInfoFile.MetaData["upload_id"] = tusInfoFile.ID;
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
        private void SendSuccessReport(string uploadId, string destinationId, string eventType, string sourceBlobUrl, string destPath)
        {
            // TODO: refactor to use service bus instead
            _featureManagementExecutor.ExecuteIfEnabled(Constants.PROCESSING_STATUS_REPORTS_FLAG_NAME, () =>
            {
                var successReport = new CopyReport(sourceUrl: sourceBlobUrl, destUrl: destPath, result: "success");
                _bulkUploadSvcBusClient.PublishReport(uploadId, destinationId, eventType, Constants.PROC_STAT_REPORT_STAGE_NAME, successReport);
            });
        }

        private void SendFailureReport(string uploadId, string destinationId, string eventType, string sourceBlobUrl, string destinationContainerName, string error)
        {
            // TODO: refactor to use service bus instead
            _featureManagementExecutor.ExecuteIfEnabled(Constants.PROCESSING_STATUS_REPORTS_FLAG_NAME, () =>
            {
                CopyReport failReport = new CopyReport(sourceUrl: sourceBlobUrl, destUrl: destinationContainerName, result: "failure", errorDesc: error);
                _bulkUploadSvcBusClient.PublishReport(uploadId, destinationId, eventType, Constants.PROC_STAT_REPORT_STAGE_NAME, failReport);
            });
        }
    }
}