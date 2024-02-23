using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobReader
    {
        Task<T> GetObjectFromBlobJsonContent<T>(string connectionString, string sourceContainerName, string blobPathname);
    }
}