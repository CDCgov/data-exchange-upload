using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Azure;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobCopyHelper
    {
        public Task CopyBlobLeaseAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null);

        public Task CopyBlobStreamAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null);
    }

}