using Azure.Identity;
using Azure.Storage.Blobs;
using Azure.Storage.Blobs.Models;
using Azure.Storage.Blobs.Specialized;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using Moq;
using System.Diagnostics;
using System.IO;
using System.Reflection.Metadata;
using System.Text;
using System.Text.Json;
using Trace = BulkFileUploadFunctionApp.Model.Trace;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class UploadProcessingServiceTests
    {
        [TestInitialize]
        public void Initialize()
        {
            //TODO: Add any initialization code here
        }


    }

}