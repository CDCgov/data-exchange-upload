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
    public class AzureBlobWriter
    {
        public BlobServiceClient _src { get; init; }
        public BlobServiceClient _dest { get; init; }
        public string _srcContainerName { get; init; }
        public string _destContainerName { get; init; }
        public BlobClient _srcBlobClient { get; init; }
        public BlobClient _destBlobClient { get; init; }
        public Dictionary<string, string> _metaData { get; init; }
        public string? _featureFlagKey { get; init; }

        public AzureBlobWriter(BlobServiceClient src, BlobServiceClient dest, string srcContainerName, string destContainerName, 
            string blobName, Dictionary<string,string> metaData, string featureFlagKey)
        {
            _src = src;
            _dest = dest;
            _srcContainerName = srcContainerName;
            _destContainerName = destContainerName;
            _srcBlobClient = _src.GetBlobContainerClient(_srcContainerName).GetBlobClient(blobName);
            _destBlobClient = _dest.GetBlobContainerClient(_destContainerName).GetBlobClient(blobName);
            _metaData = metaData;
            _featureFlagKey = featureFlagKey;
        }

        public void DoIfEnabled(string featureFlagKey, Action callback)
        {
            // Check feature flag and execute callback if enabled
            if (featureFlagKey == Constants.PROC_STAT_FEATURE_FLAG_NAME) //  "DexFeatureEnabled" ?
            {
                callback();
            }
        }

        public async Task WriteLeaseAsync(Uri? sourceSasBlobUri = null)
        {

            // Lease the source blob for the copy operation 
            // to prevent another client from modifying it.
            BlobLeaseClient lease = _srcBlobClient.GetBlobLeaseClient();
            // Get the source blob's properties and display the lease state.
            BlobProperties sourceProperties = await _srcBlobClient.GetPropertiesAsync();

            try
            {

                //_logger.LogInformation($"Checking if source blob with uri {sourceBlob.Uri} exists");

                // Ensure that the source blob exists.
                if (await _srcBlobClient.ExistsAsync())
                {
                    //_logger.LogInformation("File exists, getting lease on file");
                    // Specifying -1 for the lease interval creates an infinite lease.
                    await lease.AcquireAsync(TimeSpan.FromSeconds(-1));

                    //_logger.LogInformation($"Lease state: {sourceProperties.LeaseState}");
                    //_logger.LogInformation("Starting blob copy");

                    // Start the copy operation.
                    var sourceUriToUse = sourceSasBlobUri != null ? sourceSasBlobUri : _srcBlobClient.Uri;
                    await _destBlobClient.StartCopyFromUriAsync(sourceUriToUse, _metaData);

                    //_logger.LogInformation("Finished blob copy");

                    // Get the destination blob's properties and display the copy status.
                    BlobProperties destProperties = await _destBlobClient.GetPropertiesAsync();

                    // _logger.LogInformation($"Copy status: {destProperties.CopyStatus}");
                    // _logger.LogInformation($"Copy progress: {destProperties.CopyProgress}");
                    // _logger.LogInformation($"Completion time: {destProperties.CopyCompletedOn}");
                    // _logger.LogInformation($"Total bytes: {destProperties.ContentLength}");

                    // Update the source blob's properties.
                    sourceProperties = await _srcBlobClient.GetPropertiesAsync();
                }
            }
            catch (RequestFailedException ex)
            {
                // _logger.LogError(ex.Message);
            }
            finally
            {
                sourceProperties = await _srcBlobClient.GetPropertiesAsync();
                // _logger.LogInformation($"Post-copy Lease state: {sourceProperties.LeaseState}");

                if (sourceProperties.LeaseState == LeaseState.Leased)
                {
                    // Release the lease on the source blob
                    await lease.ReleaseAsync();
                }
            }
        }

        public async Task WriteStreamAsync()
        {
            using var sourceBlobStream = await _srcBlobClient.OpenReadAsync();
            {
                await _destBlobClient.UploadAsync(sourceBlobStream, null, _metaData);
            }
        }
    }

}
