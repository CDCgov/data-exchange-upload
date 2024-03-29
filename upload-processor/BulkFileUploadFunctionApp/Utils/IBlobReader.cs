using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobReader
    {
        //public BlobServiceClient? _svcClient { get; set; }
        //public string? _blobName { get; set; }

        Task<T?> Read<T>(string containerName, string blobName);
    }
}