using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using Azure.Storage.Blobs;
using Azure;
using Microsoft.Extensions.Logging;
using System.Collections.Generic; // Add this line
using System.Threading.Tasks; // Add this line

namespace BulkFileUploadFunctionApp.Utils
{
    public class BlobCopyHelperFactory
    {
        public virtual IBlobCopyHelper CreateInstance(ILogger logger)
        {
            return new BlobCopyHelper(logger);
        }
    }
}