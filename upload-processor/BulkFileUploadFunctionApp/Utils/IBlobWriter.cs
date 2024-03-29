using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobWriter
    {
        BlobServiceClient? _src { get; set; }
        BlobServiceClient? _dest { get; set; }
        string? _srcContainerName { get; set; }
        string? _destContainerName { get; set; }
        string? _destBlobName { get; set; }
        Task Write(string srcFileName, string destContainerName);
    }
}
