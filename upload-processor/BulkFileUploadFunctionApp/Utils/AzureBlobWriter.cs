using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Services;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    // Derived Writer Classes Implementing IBlobWriter and IEnableable
    public class AzureBlobWriter : Enableable
    {
        public BlobServiceClient Src { get; init; }
        public BlobServiceClient Dest { get; init; }
        public string SrcContainerName { get; init; }
        public string DestContainerName { get; init; }
        public BlobClient SrcBlobClient { get; init; }
        public BlobClient DestBlobClient { get; init; }
        public Dictionary<string, string> MetaData { get; init; }
        private ILogger _logger;

        public AzureBlobWriter(BlobServiceClient src, BlobServiceClient dest, string srcBlobName, string srcContainerName, string destBlobName, string destContainerName,
            Dictionary<string, string> metaData, ILoggerFactory loggerFactory)
        {
            Src = src;
            Dest = dest;
            SrcContainerName = srcContainerName;
            DestContainerName = destContainerName;
            SrcBlobClient = Src.GetBlobContainerClient(SrcContainerName).GetBlobClient(srcBlobName);
            DestBlobClient = Dest.GetBlobContainerClient(DestContainerName).GetBlobClient(destBlobName);
            MetaData = metaData;
            _logger = loggerFactory.CreateLogger<AzureBlobWriter>();
        }

        public AzureBlobWriter(BlobServiceClient src, BlobServiceClient dest, string srcBlobName, string srcContainerName, string destBlobName, string destContainerName,
            Dictionary<string,string> metaData, ILoggerFactory loggerFactory, string featureFlagKey, IFeatureManagementExecutor executor)
        {
            Src = src;
            Dest = dest;
            SrcContainerName = srcContainerName;
            DestContainerName = destContainerName;
            SrcBlobClient = Src.GetBlobContainerClient(SrcContainerName).GetBlobClient(srcBlobName);
            DestBlobClient = Dest.GetBlobContainerClient(DestContainerName).GetBlobClient(destBlobName);
            MetaData = metaData;
            FeatureFlagKey = featureFlagKey;
            Executor = executor;
            _logger = loggerFactory.CreateLogger<AzureBlobWriter>(); ;
        }

        public override void DoIfEnabled(Action callback)
        {
            // Check feature flag and execute callback if enabled
            if(Executor == null || FeatureFlagKey == null)
            {
                callback();
            }
            else
            {
                Executor.ExecuteIfEnabled(FeatureFlagKey, callback);
            }
        }

        public override async Task DoIfEnabledAsync(Func<Task> callback)
        {
            // Check feature flag and execute callback if enabled
            if (Executor == null || FeatureFlagKey == null)
            {
                await callback();
            }
            else
            {
                await Executor.ExecuteIfEnabledAsync(FeatureFlagKey, async () => { await callback(); });
            }
        }

        public async Task WriteLeaseAsync()
        {

            // Lease the source blob for the copy operation 
            // to prevent another client from modifying it.
            BlobLeaseClient lease = SrcBlobClient.GetBlobLeaseClient();
            // Get the source blob's properties and display the lease state.
            BlobProperties sourceProperties = await SrcBlobClient.GetPropertiesAsync();

            _logger.LogInformation($"Checking if source blob with uri {SrcBlobClient.Uri} exists");

            // Ensure that the source blob exists.
            if (await SrcBlobClient.ExistsAsync())
            {
                // Create dest container if not exist.
                await Dest.GetBlobContainerClient(DestContainerName).CreateIfNotExistsAsync();

                _logger.LogInformation("File exists, getting lease on file");
                // Specifying -1 for the lease interval creates an infinite lease.
                await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

                _logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");
                _logger.LogInformation("Starting blob copy");

                // Start the copy operation.
                await DestBlobClient.StartCopyFromUriAsync(SrcBlobClient.Uri, MetaData);

                _logger.LogInformation("Finished blob copy");

                // Get the destination blob's properties and display the copy status.
                BlobProperties destProperties = await DestBlobClient.GetPropertiesAsync();

                _logger.LogInformation($"Copy status: {destProperties.CopyStatus}");
                _logger.LogInformation($"Copy progress: {destProperties.CopyProgress}");
                _logger.LogInformation($"Completion time: {destProperties.CopyCompletedOn}");
                _logger.LogInformation($"Total bytes: {destProperties.ContentLength}");

                // Update the source blob's properties.
                sourceProperties = await SrcBlobClient.GetPropertiesAsync();
            }
            
            sourceProperties = await SrcBlobClient.GetPropertiesAsync();
            _logger.LogInformation($"Post-copy Lease state: {sourceProperties.LeaseState}");

            if (sourceProperties.LeaseState == LeaseState.Leased)
            {
                // Release the lease on the source blob
                await lease.ReleaseAsync();
            }
        }

        public async Task WriteStreamAsync()
        {
            // Create dest container if not exist.
            await Dest.GetBlobContainerClient(DestContainerName).CreateIfNotExistsAsync();

            using var sourceBlobStream = await SrcBlobClient.OpenReadAsync();
            {
                await DestBlobClient.UploadAsync(sourceBlobStream, null, MetaData);
            }
        }
    }

}
