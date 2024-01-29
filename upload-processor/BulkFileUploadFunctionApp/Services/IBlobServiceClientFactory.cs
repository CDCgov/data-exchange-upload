using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Azure.Storage.Blobs;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBlobServiceClientFactory
    {
        BlobServiceClient CreateBlobServiceClient(string connectionString);
    }
}