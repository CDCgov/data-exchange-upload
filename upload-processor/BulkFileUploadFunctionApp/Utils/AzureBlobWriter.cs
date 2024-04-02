using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Azure;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using Microsoft.Extensions.Configuration.AzureAppConfiguration;

namespace BulkFileUploadFunctionApp.Utils
{
    // Derived Writer Classes Implementing IBlobWriter and IEnableable
    public class AzureBlobWriter : Enableable, IRetryable
    {
        public BlobServiceClient Src { get; init; }
        public BlobServiceClient Dest { get; init; }
        public string SrcContainerName { get; init; }
        public string DestContainerName { get; init; }
        public BlobClient SrcBlobClient { get; init; }
        public BlobClient DestBlobClient { get; init; }
        public Dictionary<string, string> MetaData { get; init; }
        private BlobCopyStage _copyStage;

        public AzureBlobWriter(BlobServiceClient src, BlobServiceClient dest, string srcContainerName, string destContainerName,
            string blobName, Dictionary<string, string> metaData, BlobCopyStage copyStage)
        {
            Src = src;
            Dest = dest;
            SrcContainerName = srcContainerName;
            DestContainerName = destContainerName;
            SrcBlobClient = Src.GetBlobContainerClient(SrcContainerName).GetBlobClient(blobName);
            DestBlobClient = Dest.GetBlobContainerClient(DestContainerName).GetBlobClient(blobName);
            MetaData = metaData;
            _copyStage = copyStage; 
        }

        public AzureBlobWriter(BlobServiceClient src, BlobServiceClient dest, string srcContainerName, string destContainerName, 
            string blobName, Dictionary<string,string> metaData, BlobCopyStage copyStage, string featureFlagKey, IFeatureManagementExecutor executor)
        {
            Src = src;
            Dest = dest;
            SrcContainerName = srcContainerName;
            DestContainerName = destContainerName;
            SrcBlobClient = Src.GetBlobContainerClient(SrcContainerName).GetBlobClient(blobName);
            DestBlobClient = Dest.GetBlobContainerClient(DestContainerName).GetBlobClient(blobName);
            MetaData = metaData;
            _copyStage = copyStage;
            FeatureFlagKey = featureFlagKey;
            Executor = executor;
        }

        public void DoWithRetry(Action callback)
        {
            try
            {
                callback();
            } catch (Exception ex)
            {
                throw new RetryException(_copyStage, ex.Message);
            }
        }

        public async Task DoWithRetryAsync(Func<Task> callback)
        {
            try
            {
               await callback();
            }
            catch (Exception ex)
            {
                throw new RetryException(_copyStage, ex.Message);
            }
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

        public async Task WriteLeaseAsync(Uri? sourceSasBlobUri = null)
        {

            // Lease the source blob for the copy operation 
            // to prevent another client from modifying it.
            BlobLeaseClient lease = SrcBlobClient.GetBlobLeaseClient();
            // Get the source blob's properties and display the lease state.
            BlobProperties sourceProperties = await SrcBlobClient.GetPropertiesAsync();

            //_logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

            // Ensure that the source blob exists.
            if (await SrcBlobClient.ExistsAsync())
            {
                //_logger.LogInformation("File exists, getting lease on file");
                // Specifying -1 for the lease interval creates an infinite lease.
                await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

                //_logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");
                //_logger.LogInformation("Starting blob copy");

                // Start the copy operation.
                var sourceUriToUse = sourceSasBlobUri != null ? sourceSasBlobUri : SrcBlobClient.Uri;
                await DestBlobClient.StartCopyFromUriAsync(sourceUriToUse, MetaData);

                //_logger.LogInformation("Finished blob copy");

                // Get the destination blob's properties and display the copy status.
                BlobProperties destProperties = await DestBlobClient.GetPropertiesAsync();

                // _logger.LogInformation($"Copy status: {destProperties.CopyStatus}");
                // _logger.LogInformation($"Copy progress: {destProperties.CopyProgress}");
                // _logger.LogInformation($"Completion time: {destProperties.CopyCompletedOn}");
                // _logger.LogInformation($"Total bytes: {destProperties.ContentLength}");

                // Update the source blob's properties.
                sourceProperties = await SrcBlobClient.GetPropertiesAsync();
            }
            
            sourceProperties = await SrcBlobClient.GetPropertiesAsync();
            // _logger.LogInformation($"Post-copy Lease state: {sourceProperties.LeaseState}");

            if (sourceProperties.LeaseState == LeaseState.Leased)
            {
                // Release the lease on the source blob
                await lease.ReleaseAsync();
            }
        }

        public async Task WriteStreamAsync()
        {
            using var sourceBlobStream = await SrcBlobClient.OpenReadAsync();
            {
                await DestBlobClient.UploadAsync(sourceBlobStream, null, MetaData);
            }
        }
    }

}
