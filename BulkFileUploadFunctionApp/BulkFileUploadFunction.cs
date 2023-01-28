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

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        private readonly BlobCopyHelper _blobCopyHelper;

        private readonly string _deploymentPlatform;

        private readonly string _tusAzureObjectPrefix;

        private readonly string _tusAzureStorageContainer;

        private readonly string _dexAzureStorageAccountName;

        private readonly string _dexAzureStorageAccountKey;

        private readonly string _edavAzureStroageAccountName;

        public static string? GetEnvironmentVariable(string name)
        {
            return Environment.GetEnvironmentVariable(name, EnvironmentVariableTarget.Process);
        }

        public BulkFileUploadFunction(ILoggerFactory loggerFactory)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();
            _blobCopyHelper = new(_logger);
            
            _deploymentPlatform = GetEnvironmentVariable("DEPLOYMENT_PLATFORM") ?? "dev";
            _tusAzureObjectPrefix = GetEnvironmentVariable("TUS_AZURE_OBJECT_PREFIX") ?? "tus-prefix";
            _tusAzureStorageContainer = GetEnvironmentVariable("TUS_AZURE_STORAGE_CONTAINER") ?? "bulkuploads";
            _dexAzureStorageAccountName = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME") ?? "dataexchangedev";
            _dexAzureStorageAccountKey = GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY") ?? "";
            _edavAzureStroageAccountName = GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME") ?? "edavdevdatalakedex";
        }

        [Function("BulkFileUploadFunction")]
        public async Task Run([EventGridTrigger] StorageBlobCreatedEvent eventGridEvent)
        {
            _logger.LogInformation(eventGridEvent.Data?.ToString());

            _logger.LogInformation($"DEPLOYMENT_PLATFORM={_deploymentPlatform}, TUS_AZURE_OBJECT_PREFIX={_tusAzureObjectPrefix}, TUS_AZURE_STORAGE_CONTAINER={_tusAzureStorageContainer}, DEX_AZURE_STORAGE_ACCOUNT_NAME={_dexAzureStorageAccountName}");

            try
            {
                _logger.LogInformation("url of event is = {0}", eventGridEvent.Data?.Url);

                if (eventGridEvent.Data == null)
                    throw new Exception("event grid event data can not be null");

                var sourceBlobUri = new Uri(eventGridEvent.Data.Url!);
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
            }
            catch (Exception e)
            {
                _logger.LogError(e.Message);
            }
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

                string destinationContainerName = $"{sourceContainerName}-{_deploymentPlatform}";
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
            
            var filenameFromMetaData = tusInfoFile.MetaData!.GetValueOrDefault("filename", null);
            if (filenameFromMetaData == null)
                throw new TusInfoFileException("filename is a required metadata field and is missing from the tus info file");
            filename = filenameFromMetaData;

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
