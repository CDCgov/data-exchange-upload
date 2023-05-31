// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Model;
using Azure.Identity;
using Azure.Storage;
using Azure.Storage.Sas;
using Azure.Messaging.EventHubs.Producer;
using Azure.Messaging.EventHubs;
using Newtonsoft.Json;
using BulkFileUploadFunctionApp.Utils;

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        private readonly BlobCopyHelper _blobCopyHelper;

        private readonly string _tusAzureObjectPrefix;

        private readonly string _tusAzureStorageContainer;

        private readonly string _dexAzureStorageAccountName;

        private readonly string _dexAzureStorageAccountKey;

        private readonly string _edavAzureStroageAccountName;

        private readonly string _metadataEventHubEndPoint;
        private readonly string _metadataEventHubHubName;
        private readonly string _metadataEventHubSharedAccessKeyName;
        private readonly string _metadataEventHubSharedAccessKey;

        public static string? GetEnvironmentVariable(string name)
        {
            return Environment.GetEnvironmentVariable(name, EnvironmentVariableTarget.Process);
        }

        public BulkFileUploadFunction(ILoggerFactory loggerFactory)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();
            _blobCopyHelper = new(_logger);

            _tusAzureObjectPrefix = GetEnvironmentVariable("TUS_AZURE_OBJECT_PREFIX") ?? "tus-prefix";
            _tusAzureStorageContainer = GetEnvironmentVariable("TUS_AZURE_STORAGE_CONTAINER") ?? "bulkuploads";
            _dexAzureStorageAccountName = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME") ?? "";
            _dexAzureStorageAccountKey = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY") ?? "";
            _edavAzureStroageAccountName = GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME") ?? "";

            _metadataEventHubEndPoint = GetEnvironmentVariable("DEX_AZURE_EVENTHUB_ENDPOINT_NAME") ?? "";
            _metadataEventHubHubName = GetEnvironmentVariable("DEX_AZURE_EVENTHUB_HUB_NAME") ?? "";
            _metadataEventHubSharedAccessKeyName = GetEnvironmentVariable("DEX_AZURE_EVENTHUB_SHARED_ACCESS_KEY_NAME") ?? "";
            _metadataEventHubSharedAccessKey = GetEnvironmentVariable("DEX_AZURE_EVENTHUB_SHARED_ACCESS_KEY") ?? "";
        }

        [Function("BulkFileUploadFunction")]
        public async Task Run([EventHubTrigger("%AzureEventHubName%", Connection = "AzureEventHubConnectionString", ConsumerGroup = "%AzureEventHubConsumerGroup%")] string[] eventHubTriggerEvent)
        {
            if (eventHubTriggerEvent.Count() < 1)
                throw new Exception("EventHubTrigger triggered with no data");

            string blobCreatedEventJson = eventHubTriggerEvent[0];
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
        }

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

                GetRequiredMetaData(tusInfoFile, out string filename, out string destinationId, out string extEvent);

                // Partitioning is part of the filename where slashes will create subfolders.
                // Container name is "{meta_destination_id}-{extEvent}"
                // Path inside of that is year / month / day / filename
                var dateTimeNow = DateTime.UtcNow;

                var fileNameWithoutExtension = Path.GetFileNameWithoutExtension(filename);
                var fileExtension = Path.GetExtension(filename);
                var destinationBlobFilename = $"{dateTimeNow.Year}/{dateTimeNow.Month.ToString().PadLeft(2, '0')}/{dateTimeNow.Day.ToString().PadLeft(2, '0')}/{fileNameWithoutExtension}_{dateTimeNow.Ticks}{fileExtension}";

                // There are some restrictions on container names -- underscores not allowed, must be all lowercase
                var destinationContainerName = $"{destinationId.ToLower()}-{extEvent.ToLower()}";

                var tusFileMetadata = tusInfoFile?.MetaData ?? new Dictionary<string, string>();
                tusFileMetadata.Add("tus_tguid", tusPayloadFilename);
                tusFileMetadata.Remove("filename");
                tusFileMetadata.Add("orig_filename", filename);

                // Copy the blob to the DeX storage account specific to the program, partitioned by date
                await CopyBlobFromTusToDexAsync(connectionString, sourceContainerName, tusPayloadPathname, destinationContainerName, destinationBlobFilename, tusFileMetadata);

                // Now copy the file from DeX to the EDAV storage account, also partitioned by date
                await CopyBlobFromDexToEdavAsync(destinationContainerName, destinationBlobFilename, tusFileMetadata);

                // Finally, send metadata to eventhub for other consumers
                var metadataRelaySucceeded = await RelayMetaData(tusFileMetadata);
                if (!metadataRelaySucceeded)
                {
                    _logger.LogError($"metadata relay failed for: {tusPayloadPathname}");
                }
            }
            catch (Exception e)
            {
                _logger.LogError(e.Message);
            }
        }

        /// <summary>
        /// Sends Uploaded File metadata to event hub for downstream consumers to proccess
        /// </summary>
        /// <param name="metaData"></param>
        /// <returns></returns>
        private async Task<bool> RelayMetaData(Dictionary<string, string> metaData)
        {
            var relaySucceeded = false;
            EventHubProducerClient? producerClient = null;
            try
            {
                var connectionString = $"Endpoint=sb://{_metadataEventHubEndPoint}.servicebus.windows.net/;SharedAccessKeyName={_metadataEventHubSharedAccessKeyName};SharedAccessKey={_metadataEventHubSharedAccessKey};EntityPath={_metadataEventHubHubName}";

                producerClient = new EventHubProducerClient(connectionString);
                var metaDataEventBody = JsonConvert.SerializeObject(metaData);

                // Create a batch of events 
                using EventDataBatch eventBatch = await producerClient.CreateBatchAsync();

                var eventData = new EventData(metaDataEventBody);
                if (!eventBatch.TryAdd(eventData))
                {
                    // if it is too large for the batch
                    throw new Exception("Metadata Event is too large for the batch and cannot be sent.");
                }

                // Use the producer client to send the batch of events to the event hub
                await producerClient.SendAsync(eventBatch);
                relaySucceeded = true;
                _logger.LogInformation("A batch of 1 metadata events has been published.");
            }
            catch (Exception e)
            {
                _logger.LogError($"Exception caught sending the event batch: {e.Message}");
            }
            finally
            {
                if (producerClient != null)
                {
                    await producerClient.DisposeAsync();
                }
            }
            _logger.LogInformation($"metadata relay result: {relaySucceeded}");

            return relaySucceeded;
        }

        private async Task<TusInfoFile> GetTusFileInfo(string connectionString, string sourceContainerName, string tusPayloadPathname)
        {
            TusInfoFile tusInfoFile;

            string tusInfoPathname = tusPayloadPathname + ".info";

            _logger.LogInformation($"Retrieving tus info file: {tusInfoPathname}");

            var sourceContainerClient = new BlobContainerClient(connectionString, sourceContainerName);

            BlobClient sourceBlob = sourceContainerClient.GetBlobClient(tusInfoPathname);

            _logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

            // Ensure that the source blob exists
            if (!await sourceBlob.ExistsAsync())
            {
                throw new TusInfoFileException("File is missing");
            }

            _logger.LogInformation("File exists, getting lease on file");

            BlobLeaseClient lease = sourceBlob.GetBlobLeaseClient();

            // Specifying -1 for the lease interval creates an infinite lease
            await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

            BlobDownloadResult download = await sourceBlob.DownloadContentAsync();
            tusInfoFile = download.Content.ToObjectFromJson<TusInfoFile>();

            BlobProperties sourceProperties = await sourceBlob.GetPropertiesAsync();

            if (sourceProperties.LeaseState == LeaseState.Leased)
            {
                // Release the lease on the source blob
                await lease.ReleaseAsync();
            }

            _logger.LogInformation($"Info file metadata keys: {tusInfoFile.MetaData?.Keys}");

            return tusInfoFile;
        }

        private Uri? GetServiceSasUriForBlob(BlobClient blobClient, string? storedPolicyName = null)
        {
            // Check whether this BlobClient object has been authorized with Shared Key.
            if (blobClient.CanGenerateSasUri)
            {
                // Create a SAS token that's valid for one hour.
                BlobSasBuilder sasBuilder = new()
                {
                    BlobContainerName = blobClient.GetParentBlobContainerClient().Name,
                    BlobName = blobClient.Name,
                    Resource = "b"
                };

                if (storedPolicyName == null)
                {
                    sasBuilder.ExpiresOn = DateTimeOffset.UtcNow.AddHours(1);
                    sasBuilder.SetPermissions(BlobSasPermissions.Read |
                        BlobSasPermissions.Write);
                }
                else
                {
                    sasBuilder.Identifier = storedPolicyName;
                }

                Uri sasUri = blobClient.GenerateSasUri(sasBuilder);
                _logger.LogInformation($"SAS URI for blob is: {sasUri}");

                return sasUri;
            }
            else
            {
                _logger.LogError("BlobClient must be authorized with Shared Key credentials to create a service SAS.");
                return null;
            }
        }

        private async Task CopyBlobFromDexToEdavAsync(string sourceContainerName, string sourceBlobFilename, IDictionary<string, string> destinationMetadata)
        {
            try
            {
                var edavBlobServiceClient = new BlobServiceClient(
                    new Uri($"https://{_edavAzureStroageAccountName}.blob.core.windows.net"),
                    new DefaultAzureCredential() // using Service Principal
                );

                string destinationContainerName = sourceContainerName;
                var edavContainerClient = edavBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                // Create the destination container if not exists
                await edavContainerClient.CreateIfNotExistsAsync();

                StorageSharedKeyCredential storageSharedKeyCredential = new(_dexAzureStorageAccountName, _dexAzureStorageAccountKey);
                Uri blobContainerUri = new($"https://{_dexAzureStorageAccountName}.blob.core.windows.net/{sourceContainerName}");
                BlobContainerClient dexCombinedSourceContainerClient = new(blobContainerUri, storageSharedKeyCredential);

                string destinationBlobFilename = sourceBlobFilename;
                BlobClient dexSourceBlobClient = dexCombinedSourceContainerClient.GetBlobClient(destinationBlobFilename);
                var dexSasUri = GetServiceSasUriForBlob(dexSourceBlobClient);

                BlobClient edavDestBlobClient = edavContainerClient.GetBlobClient(destinationBlobFilename);

                await _blobCopyHelper.CopyBlobAsync(dexSourceBlobClient, edavDestBlobClient, destinationMetadata, dexSasUri);
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError(ex.Message);
            }
        }

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
                _logger.LogError(ex.Message);
            }
        }

        private static void GetRequiredMetaData(TusInfoFile tusInfoFile, out string filename, out string destinationId, out string extEvent)
        {
            if (tusInfoFile.MetaData == null)
                throw new TusInfoFileException("tus info file required metadata is missing");

            // 27-04-2023: Matt Krystof
            // Below is a temporary hotfix to allow NDLP files to proceed.  IZGW is sending meta_ext_filename, but not filename in the metadata,
            // which is failing here.  Temporary solution is to allow either field, but long-term fix will be to require 'filename' metadata field
            // at time of upload.
            var filenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("filename", null);
            var extfilenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_filename", null);

            // fall-back for using the provided uuid instead of file name
            // this is needed for DEX HL7 and is a required field in dex_hl7_metadata_definition.json
            var extUUIDFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_objectkey", null);

            //if (filenameFromMetaData == null)
            //    throw new TusInfoFileException("filename is a required metadata field and is missing from the tus info file");
            //filename = filenameFromMetaData;
            if (filenameFromMetaData != null)
                filename = filenameFromMetaData;
            else if (extfilenameFromMetaData != null)
                filename = extfilenameFromMetaData;
            else if (extUUIDFromMetaData != null)
                filename = extUUIDFromMetaData;
            else
                throw new TusInfoFileException("filename, meta_ext_filename, or meta_ext_objectkey is a required metadata field and is missing from the tus info file");
            // End of hotfix

            var metaDestinationId = tusInfoFile.MetaData!.GetValueOrDefault("meta_destination_id", null);
            if (metaDestinationId == null)
                throw new TusInfoFileException("meta_destination_id is a required metadata field and is missing from the tus info file");
            destinationId = metaDestinationId;

            var metaExtEvent = tusInfoFile.MetaData!.GetValueOrDefault("meta_ext_event", null);
            if (metaExtEvent == null)
                throw new TusInfoFileException("meta_ext_event is a required metadata field and is missing from the tus info file");
            extEvent = metaExtEvent;
        }
    }
}
