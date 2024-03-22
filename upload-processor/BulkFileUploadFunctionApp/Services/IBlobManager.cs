using Azure.Storage.Blobs;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobManager
    {
        Task CopyBlobAsync(BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null);
        Task<T?> GetObjectFromBlobJsonContent<T>(string connectionString, string sourceContainerName, string blobPathname);
    }
}
