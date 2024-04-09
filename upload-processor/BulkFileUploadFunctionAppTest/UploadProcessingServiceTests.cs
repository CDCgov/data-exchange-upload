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
using System.Text.Json;
using Trace = BulkFileUploadFunctionApp.Model.Trace;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class UploadProcessingServiceTests
    {
        private UploadProcessingService? _uploadProcessingService;
        private Mock<IProcStatClient>? _mockProcStatClient;
        private Mock<IConfigurationManager>? _mockConfigManager;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private Mock<IUploadEventHubService>? _mockUploadEventHubService;
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private StorageBlobCreatedEvent? _storageBlobCreatedEvent;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<AzureBlobReader>? _mockBlobReader;
        private Mock<BlobClient>? _mockBlobClient;
        private Mock<IBlobServiceClientFactory>? _mockBlobServiceClientFactory;
        private Mock<BlobServiceClient>? _mockBlobServiceClient;
        private Mock<BlobServiceClient>? _mockEdavBlobServiceClient;
        private Mock<Uri>? _mockUri;

        [TestInitialize]
        public void Initialize()
        {
            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            _mockBlobClient = new Mock<BlobClient>();
            _mockBlobServiceClient = new Mock<BlobServiceClient>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            _mockBlobServiceClientFactory
                .Setup(x=> x.CreateInstance(It.IsAny<string>(), It.IsAny<string>()))
                .Returns(_mockBlobServiceClient.Object);
            _mockUri = new Mock<Uri>("https://example.com/blob/1MB-test-file.txt"); //new Mock<Uri>();

            _mockEdavBlobServiceClient = new Mock<BlobServiceClient>();
            _mockBlobServiceClientFactory
                .Setup(x => x.CreateInstance(It.IsAny<string>(), _mockUri.Object, It.IsAny<DefaultAzureCredential>()))
                .Returns(_mockEdavBlobServiceClient.Object);

            _mockConfigManager = new Mock<IConfigurationManager>();
            _mockProcStatClient = new Mock<IProcStatClient>();
            _mockFeatureManagementExecutor = new Mock<IFeatureManagementExecutor>();
            _mockBlobReader = new Mock<AzureBlobReader>();
            _mockUploadEventHubService = new Mock<IUploadEventHubService>();

      


            _storageBlobCreatedEvent = new StorageBlobCreatedEvent
            {
                Id = "12323",
                Topic = "routineImmunization",
                Subject = "IZGW",
                EventType = "DD2",
                EventTime = System.DateTime.Now,
                Data = new StorageBlobCreatedEventData { Url = "https://example.com/blob/10MB-test-file" }
            };



            _mockUploadProcessingService = new Mock<IUploadProcessingService>();

            _mockUploadProcessingService.CallBase = true;

            _uploadProcessingService = new UploadProcessingService(
                _loggerFactoryMock.Object, 
                _mockConfigManager.Object, 
                _mockProcStatClient.Object, _mockFeatureManagementExecutor.Object,
                _mockUploadEventHubService.Object, _mockBlobServiceClientFactory.Object);

        }
    }
}