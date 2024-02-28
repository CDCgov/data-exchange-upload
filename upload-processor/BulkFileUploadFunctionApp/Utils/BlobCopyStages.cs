using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Azure;

namespace BulkFileUploadFunctionApp.Utils
{
    public enum BlobCopyStage
    {
        CopyToDex,
        CopyToEdav,
        CopyToRouting
    }
}