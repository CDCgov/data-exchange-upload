// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Model;
using Google.Protobuf.WellKnownTypes;
using System.Security.Principal;
using Azure.Identity;
using Azure.Storage;
using Azure.Storage.Sas;

namespace BulkFileUploadFunctionApp
{
    public class BulkFileUploadFunction
    {
        private readonly ILogger _logger;

        public BulkFileUploadFunction(ILoggerFactory loggerFactory)
        {
            _logger = loggerFactory.CreateLogger<BulkFileUploadFunction>();
        }

        [Function("BulkFileUploadFunction")]
        public async Task Run([EventGridTrigger] StorageBlobCreatedEvent eventGridEvent)
        {
            _logger.LogInformation(eventGridEvent.Data?.ToString());

            try
            {
                _logger.LogInformation("url of event is = {0}", eventGridEvent.Data?.Url);

                if (eventGridEvent.Data == null)
                    throw new Exception("event grid event data can not be null");

                var sourceBlobUri = new Uri(eventGridEvent.Data.Url!);
                string tusPayloadFilename = sourceBlobUri.Segments.Last();
                _logger.LogInformation($"tusPayloadFilename is = {tusPayloadFilename}");

                var connectionString = "DefaultEndpointsProtocol=https;AccountName=dataexchangedev;AccountKey=lVvJbZ5J+SvLvWpUMwybFKnqYs57J4EF+HBvWTUo9GAHsLheFRWHOxXmVmy2Ojy7m/W8qBbgXIoe+AStzh0IdQ==;EndpointSuffix=core.windows.net";
                var sourceContainerName = "bulkuploads";
                var tusPayloadPathname = "/tus-prefix/" + tusPayloadFilename;
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
                await CopyBlobAsync(connectionString, sourceContainerName, tusPayloadPathname, destinationContainerName, destinationBlobFilename, tusFileMetadata);

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
                const string edavStorageAccount = "edavdevdatalakedex";
                // Using Service Principal
                const string appID = "bd468706-837d-49ff-8aa0-90ce81601291"; // OCIO-EDEEDAV-DEV
                const string appSec = "";
                const string tenantID = "9ce70869-60db-44fd-abe8-d2767077fc8f"; // same tenant for all of CDC including DeX and EDAV

                var clientCred = new ClientSecretCredential(tenantID, appID, appSec);

                var edavBlobServiceClient = new BlobServiceClient(
                    new Uri($"https://{edavStorageAccount}.blob.core.windows.net"),
                    clientCred
                );

                string destinationContainerName = $"{sourceContainerName}-dev";
                var edavContainerClient = edavBlobServiceClient.GetBlobContainerClient(destinationContainerName);

                StorageSharedKeyCredential storageSharedKeyCredential = new("dataexchangedev", "lVvJbZ5J+SvLvWpUMwybFKnqYs57J4EF+HBvWTUo9GAHsLheFRWHOxXmVmy2Ojy7m/W8qBbgXIoe+AStzh0IdQ==");
                Uri blobContainerUri = new($"https://dataexchangedev.blob.core.windows.net/{sourceContainerName}");
                BlobContainerClient dexCombinedSourceContainerClient = new(blobContainerUri, storageSharedKeyCredential);

                string destinationBlobFilename = sourceBlobFilename;
                BlobClient dexSourceBlobClient = dexCombinedSourceContainerClient.GetBlobClient(destinationBlobFilename);
                var dexSasUri = GetServiceSasUriForBlob(dexSourceBlobClient);

                BlobClient edavDestBlobClient = edavContainerClient.GetBlobClient(destinationBlobFilename);

                _logger.LogInformation($"Starting dex to edav blob copy with source uri {dexSourceBlobClient.Uri}, destination uri {edavContainerClient.Uri}");

                // Start the copy operation.
                await edavDestBlobClient.StartCopyFromUriAsync(dexSasUri, destinationMetadata);

                _logger.LogInformation("Finished blob copy to edav");
            }
            catch (RequestFailedException ex)
            {
                _logger.LogError(ex.Message);
            }
        }

        private async Task CopyBlobAsync(string connectionString, string sourceContainerName, string sourceBlobName, string destinationContainerName,
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

                _logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

                // Ensure that the source blob exists.
                if (await sourceBlob.ExistsAsync())
                {
                    _logger.LogInformation("File exists, getting lease on file");

                    // Lease the source blob for the copy operation 
                    // to prevent another client from modifying it.
                    BlobLeaseClient lease = sourceBlob.GetBlobLeaseClient();

                    // Specifying -1 for the lease interval creates an infinite lease.
                    await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

                    // Get the source blob's properties and display the lease state.
                    BlobProperties sourceProperties = await sourceBlob.GetPropertiesAsync();
                    _logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");

                    _logger.LogInformation($"Creating destination blob client, blob filename: {destinationBlobName}");

                    // Get a BlobClient representing the destination blob with a unique name.
                    BlobClient destBlob =
                        destinationContainerClient.GetBlobClient(destinationBlobName);

                    _logger.LogInformation("Starting blob copy");

                    // Start the copy operation.
                    await destBlob.StartCopyFromUriAsync(sourceBlob.Uri, destinationMetadata);

                    _logger.LogInformation("Finished blob copy");

                    // Get the destination blob's properties and display the copy status.
                    BlobProperties destProperties = await destBlob.GetPropertiesAsync();

                    _logger.LogInformation($"Copy status: {destProperties.CopyStatus}");
                    _logger.LogInformation($"Copy progress: {destProperties.CopyProgress}");
                    _logger.LogInformation($"Completion time: {destProperties.CopyCompletedOn}");
                    _logger.LogInformation($"Total bytes: {destProperties.ContentLength}");

                    // Update the source blob's properties.
                    sourceProperties = await sourceBlob.GetPropertiesAsync();

                    if (sourceProperties.LeaseState == LeaseState.Leased)
                    {
                        // Release the lease on the source blob
                        await lease.ReleaseAsync();

                        // Update the source blob's properties to check the lease state.
                        sourceProperties = await sourceBlob.GetPropertiesAsync();
                        _logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");
                    }
                }
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
