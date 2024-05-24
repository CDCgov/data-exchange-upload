using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionAppTest.utils;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class UploadProcessingServiceTests
    {
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private Mock<IBlobReader> _mockBlobReader;
        private BlobReaderFactory _blobReaderFactory;
        private Mock<BlobReaderFactory>? _mockBlobReaderFactory;
        private string sourceContainerName;
        private UploadProcessingService _function;
        private Mock<IBulkUploadSvcBusClient> _mockBulkUploadSvcClient;
        private Mock<IUploadEventHubService> _mockUploadEventHubService;
        private Mock<IEnvironmentVariableProvider> _mockEnvironmentVariableProvider;

        [TestInitialize]
        public void Initialize()
        {
            sourceContainerName = "bulkuploads";
            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);

            _mockFeatureManagementExecutor = new Mock<IFeatureManagementExecutor>();
            _blobReaderFactory = new BlobReaderFactory();
            _mockBlobReader = new Mock<IBlobReader>();
            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            _mockBlobReaderFactory
                .Setup(x => x.CreateInstance(It.IsAny<ILogger>()))
                .Returns(_mockBlobReader.Object);

            _mockUploadEventHubService = new Mock<IUploadEventHubService>();
            _mockBulkUploadSvcClient = new Mock<IBulkUploadSvcBusClient>();

            _storageBlobCreatedEvent = new StorageBlobCreatedEvent
            {
                Id = "12323",
                Topic = "routineImmunization",
                Subject = "IZGW",
                EventType = "DD2",
                EventTime = System.DateTime.Now,
                Data = new StorageBlobCreatedEventData { Url = "https://example.com/blob/10MB-test-file" }
            };

            _mockBlobReaderFactory.Setup(x => x.CreateInstance(It.IsAny<ILogger>())).Returns(_mockBlobReader.Object);

            _mockUploadProcessingService = new Mock<IUploadProcessingService>();


            _mockUploadProcessingService.CallBase = true;

            _function = new UploadProcessingService(
                _loggerFactoryMock.Object, 
                _mockBulkUploadSvcClient.Object, 
                _mockFeatureManagementExecutor.Object, 
                _mockUploadEventHubService.Object, 
                _mockBlobReaderFactory.Object);



        }

    }

}
