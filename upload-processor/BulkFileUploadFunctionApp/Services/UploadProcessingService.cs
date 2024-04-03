using Azure;
using Azure.Identity;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Services
{
    public class UploadProcessingService : IUploadProcessingService
    {
        private readonly ILogger _logger;
        private readonly ILoggerFactory _loggerFactory;
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly AzureBlobReader _dexBlobReader;
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
        private readonly BlobServiceClient _edavBlobServiceClient;
        private readonly string _uploadConfigContainer;

        public UploadProcessingService(ILoggerFactory loggerFactory, IConfiguration configuration, IProcStatClient procStatClient,
        IFeatureManagementExecutor featureManagementExecutor, IUploadEventHubService uploadEventHubService, IBlobServiceClientFactory blobServiceClientFactory)
        {
            _loggerFactory = loggerFactory;
            _logger = loggerFactory.CreateLogger<UploadProcessingService>();
            _blobServiceClientFactory = blobServiceClientFactory;         
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
            _uploadConfigContainer =  Environment.GetEnvironmentVariable("UPLOAD_CONFIGS", EnvironmentVariableTarget.Process) ?? "upload-configs";
            
            // Instantiate helper services.
            _logger = loggerFactory.CreateLogger<UploadProcessingService>();
            _featureManagementExecutor = featureManagementExecutor;
            _procStatClient = procStatClient;

            _uploadEventHubService = uploadEventHubService;
            _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";
            _routingStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_routingStorageAccountName};AccountKey={_routingStorageAccountKey};EndpointSuffix=core.windows.net";

            // Create or retrieve singleton instances with specified connection strings
            _dexBlobServiceClient = _blobServiceClientFactory.CreateInstance("dex", _dexStorageAccountConnectionString);
            _routingBlobServiceClient = _blobServiceClientFactory.CreateInstance("routing", _routingStorageAccountConnectionString);
            _edavBlobServiceClient = _blobServiceClientFactory.CreateInstance("edav", new Uri($"https://{_edavAzureStorageAccountName}.blob.core.windows.net"),
                new DefaultAzureCredential());
            _dexBlobReader = new AzureBlobReader(_dexBlobServiceClient);
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

                TusInfoFile tusInfoFile = await GetTusInfoFile(tusPayloadFilename);
                uploadId = tusInfoFile.ID;

                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    trace = await _procStatClient.GetTraceByUploadId(uploadId);
                });

                HydrateMetadata(tusInfoFile, trace?.TraceId, trace?.SpanId);

                // retrieve version from metadata or default to V1
                version = tusInfoFile.GetMetadataVersion();
                useCase = tusInfoFile.GetUseCase();
                useCaseCategory = tusInfoFile.GetUseCaseCategory();
                destinationContainerName = $"{useCase}-{useCaseCategory}";
                string uploadConfigFilename = $"{useCase}-{useCaseCategory}.json";

                var uploadConfig = await GetUploadConfig(uploadConfigFilename, version);
                _logger.LogInformation($"Got upload config for {version}.");
                // translate V1 metadata 
                if (version == MetadataVersion.V1)
                {
                    var uploadConfigV2 = await GetUploadConfig(uploadConfigFilename, MetadataVersion.V2);
                    _logger.LogInformation($"Translating to {MetadataVersion.V2}");
                    tusInfoFile.MetaData = TranslateMetadata(tusInfoFile.MetaData, uploadConfigV2);
                }
                
                string? filename = tusInfoFile.MetaData!.GetValueOrDefault("received_filename", null);

                // Get dex folder and filename 
                var dateTimeNow = DateTime.UtcNow;

                // Determine the folder path and filename suffix from the upload configuration.
                var folderPath = GetFolderPath(uploadConfig, dateTimeNow);
                var filenameSuffix = GetFilenameSuffix(uploadConfig, dateTimeNow);

                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
            
                string destinationBlobFilename = $"{folderPath}/{fileNameWithoutExtension}{filenameSuffix}{fileExtension}";

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
                _logger.LogError("Failed to get copy inputs.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                
                // Send copy failure report
                SendFailureReport(uploadId, useCase, useCaseCategory, blobCreatedUrl, destinationContainerName, $"Failed to get copy preqs: {ex.Message}");

                throw new RetryException(BlobCopyStage.CopyToDex, ex.Message);
            }
        }
        
        public async Task CopyAll(CopyPrereqs copyPrereqs)
        {
            Span? copySpan = null;
            _logger.LogInformation($"Creating destination container client, container name: {copyPrereqs.DexBlobFolderName}");

            AzureBlobWriter tusToDexBlobWriter = CreateWriterForStage(BlobCopyStage.CopyToDex, copyPrereqs);
            AzureBlobWriter dexToEdavBlobWriter = CreateWriterForStage(BlobCopyStage.CopyToEdav, copyPrereqs);
            AzureBlobWriter dexToRoutingBlobWriter = CreateWriterForStage(BlobCopyStage.CopyToRouting, copyPrereqs);

            List<AzureBlobWriter> writers = copyPrereqs.Targets.Select(target =>
            {
                switch (target)
                {
                    case CopyTargetsEnum.edav:
                        return dexToEdavBlobWriter;
                    case CopyTargetsEnum.routing:
                        return dexToRoutingBlobWriter;
                    default:
                        return dexToEdavBlobWriter;
                };
            }).ToList();

            try
            {
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    copySpan = await _procStatClient.StartSpanForTrace(copyPrereqs.Trace.TraceId, copyPrereqs.Trace.SpanId, Constants.PROC_STAT_REPORT_STAGE_NAME);
                });

                copyPrereqs.DexBlobUrl = await CopyFromTusToDex(tusToDexBlobWriter);
                await CopyFromDexToTargets(writers, copyPrereqs);
            }
            catch (RetryException ex)
            {
                await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                {
                    AzureBlobWriter? writer = writers.Find(writer => writer.CopyStage == ex.Stage);

                    if (writer != null)
                    {
                        string srcUrl = writer.SrcBlobClient.Uri.ToString();
                        string destUrl = writer.DestBlobClient.Uri.ToString();
                        SendFailureReport(copyPrereqs.UploadId,
                                      copyPrereqs.UseCase,
                                      copyPrereqs.UseCaseCategory,
                                      copyPrereqs.SourceBlobUrl,
                                      copyPrereqs.DexBlobFolderName,
                                      $"Failed to copy blob from {srcUrl} to {destUrl}. {ex.Message}");
                    }
                });

                await PublishRetryEvent(ex.Stage, copyPrereqs);
            }
            finally
            {
                if(copySpan != null) 
                {
                    await _procStatClient.StopSpanForTrace(copyPrereqs.Trace.TraceId, copySpan.SpanId);
                }
            }
        }               
        
        public async Task<string> CopyFromTusToDex(AzureBlobWriter tusToDexBlobWriter)
        {
            try
            {
                await tusToDexBlobWriter.WriteLeaseAsync();

                return tusToDexBlobWriter.DestBlobClient.Uri.ToString();
            } catch (RequestFailedException ex)
            {
                _logger.LogError($"Failed to copy from {tusToDexBlobWriter.SrcBlobClient.Uri} to {tusToDexBlobWriter.DestBlobClient.Uri}.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                throw new RetryException(BlobCopyStage.CopyToDex, ex.Message);
            }
        }
        
        public async Task CopyFromDexToTargets(List<AzureBlobWriter> writers, CopyPrereqs copyPrereqs)
        {
            _logger.LogInformation($"Writting to {writers.Count} target(s).");
            foreach (AzureBlobWriter writer in writers)
            {
                try
                {
                    await writer.DoIfEnabledAsync(async () =>
                    {
                        _logger.LogInformation($"Copying to {writer.DestBlobClient.Uri}");
                        await writer.WriteStreamAsync();
                        _logger.LogInformation("Complete.");

                        // Send copy success report
                        await _featureManagementExecutor.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, async () =>
                        {
                            SendSuccessReport(copyPrereqs.UploadId,
                                                copyPrereqs.UseCase,
                                                copyPrereqs.UseCaseCategory,
                                                copyPrereqs.DexBlobUrl,
                                                writer.DestBlobClient.Uri.ToString());
                        });
                    });
                }
                catch (RequestFailedException ex)
                {
                    _logger.LogError($"Failed to copy from {writer.SrcBlobClient.Uri} to {writer.DestBlobClient.Uri}.");
                    ExceptionUtils.LogErrorDetails(ex, _logger);
                    throw new RetryException(writer.CopyStage, ex.Message);
                }
            }
        }

        public AzureBlobWriter CreateWriterForStage(BlobCopyStage stage, CopyPrereqs copyPrereqs)
        {
            string dexToEdavDestinationContainerName = _edavUploadRootContainerName ?? copyPrereqs.DexBlobFolderName;
            string dexToTargetFilename = $"{copyPrereqs.DexBlobFolderName}/{copyPrereqs.DexBlobFileName}" ?? copyPrereqs.DexBlobFileName;
            string dexToRoutingDestinationContainerName = _routingUploadRootContainerName ?? copyPrereqs.DexBlobFolderName;

            switch (stage)
            {
                case BlobCopyStage.CopyToDex:
                    return new AzureBlobWriter(
                        _dexBlobServiceClient,
                        _dexBlobServiceClient,
                        copyPrereqs.TusPayloadFilename,
                        _tusAzureStorageContainer,
                        copyPrereqs.DexBlobFileName,
                        copyPrereqs.DexBlobFolderName,
                        copyPrereqs.Metadata,
                        BlobCopyStage.CopyToDex,
                        _loggerFactory);
                case BlobCopyStage.CopyToEdav:
                    return new AzureBlobWriter(
                        _dexBlobServiceClient,
                        _edavBlobServiceClient,
                        copyPrereqs.DexBlobFileName,
                        copyPrereqs.DexBlobFolderName,
                        dexToTargetFilename,
                        dexToEdavDestinationContainerName,
                        copyPrereqs.Metadata,
                        BlobCopyStage.CopyToEdav,
                        _loggerFactory);
                case BlobCopyStage.CopyToRouting:
                    return new AzureBlobWriter(
                        _dexBlobServiceClient,
                        _routingBlobServiceClient,
                        copyPrereqs.DexBlobFileName,
                        copyPrereqs.DexBlobFolderName,
                        dexToTargetFilename,
                        dexToRoutingDestinationContainerName,
                        copyPrereqs.Metadata,
                        BlobCopyStage.CopyToRouting,
                        _loggerFactory,
                        Constants.ROUTING_FEATURE_FLAG_NAME,
                        _featureManagementExecutor);
                default:
                    throw new ArgumentException($"Unsupported stage {stage}.");
            }
        }

        private async Task<TusInfoFile> GetTusInfoFile(string tusPayloadFilename)
        {
            string tusInfoFilename = $"{tusPayloadFilename}.info";             
            _logger.LogInformation($"Retrieving tus info file: {tusInfoFilename}");
            
            var tusInfoFile = await _dexBlobReader.Read<TusInfoFile>(_tusAzureStorageContainer, tusInfoFilename);

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
                uploadConfig = await _dexBlobReader.Read<UploadConfig>(_uploadConfigContainer, configFilename); 

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

        /// <summary>
        /// Determines the filename suffix from the upload configuration.
        /// </summary>
        /// <param name="uploadConfig"></param>
        /// <param name="dateTimeNow"></param>
        /// <returns></returns>
        private string GetFilenameSuffix(UploadConfig uploadConfig, DateTime dateTimeNow)
        {
            string filenameSuffix;
            switch (uploadConfig.CopyConfig.FilenameSuffix)
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

        private void HydrateMetadata(TusInfoFile tusInfoFile, string? traceId, string? spanId)
        {
            // Add common fields and their values.
            tusInfoFile.MetaData["tus_tguid"] = tusInfoFile.ID; // TODO: verify this field can be replaced with upload_id only.
            tusInfoFile.MetaData["upload_id"] = tusInfoFile.ID;
            tusInfoFile.MetaData["trace_id"] = traceId;
            tusInfoFile.MetaData["parent_span_id"] = spanId;
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