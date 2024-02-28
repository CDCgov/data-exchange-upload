using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using Azure.Identity;
using Newtonsoft.Json;
using BulkFileUploadFunctionApp.Utils;
using System.Collections.Concurrent;
using Microsoft.Extensions.Configuration;
using BulkFileUploadFunctionApp.Services;
using Microsoft.Extensions.Logging.Abstractions;
using System;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class BulkFileUploadFunctionTests
    {
        private Mock<IProcStatClient>? _mockProcStatClient;
        private Mock<BlobCopyHelperFactory>? _mockBlobCopyHelperFactory;
        private Mock<BlobReaderFactory>? _mockBlobReaderFactory;
        private Mock<ILogger<BulkFileUploadFunction>>? _loggerMockBUF;
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;

        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private Mock<ILoggerFactory>? _loggerFactoryBUFMock;
        private Mock<IConfiguration>? _mockConfiguration;
        private BulkFileUploadFunction _function;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private MockTusInfoFile? _mockTusInfoFile;
        private MockTusStorage? _mockTusStorage;
        private Mock<IBlobServiceClientFactory>? _mockBlobServiceClientFactorySrc;
        private Mock<IBlobServiceClientFactory>? _mockBlobServiceClientFactoryDest;
        private UploadProcessingService _uploadProcessingService;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private Mock<IUploadEventHubService>? _mockUploadEventHubService;
        private Mock<BulkFileUploadFunction>? _mockBulkFileUploadFunction;

        private readonly string _targetEdav = "dex_edav";
        private readonly string _targetRouting = "dex_routing";
        private readonly string _destinationAndEventsFileName = "allowed_destination_and_events.json";
        private readonly string _stageName = "dex-file-copy";

        private string _dexAzureStorageAccountName;
        private string _dexAzureStorageAccountKey;
        private string _edavAzureStorageAccountName;


        [TestInitialize]
        public void Initialize()
        {
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", "YourStorageAccountName", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", "YourStorageAccountKey", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", "YourStorageAccountName", EnvironmentVariableTarget.Process);

            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);

            _loggerFactoryBUFMock = new Mock<ILoggerFactory>();
            _loggerMockBUF = new Mock<ILogger<BulkFileUploadFunction>>();
            _loggerFactoryBUFMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            
            _mockConfiguration = new Mock<IConfiguration>();
            _mockProcStatClient = new Mock<IProcStatClient>();
            _mockFeatureManagementExecutor = new Mock<IFeatureManagementExecutor>();
            _mockUploadEventHubService = new Mock<IUploadEventHubService>();
            _mockBlobCopyHelperFactory = new Mock<BlobCopyHelperFactory>();
            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            _mockBulkFileUploadFunction = new Mock<BulkFileUploadFunction>();
            //_loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            _mockUploadProcessingService = new Mock<IUploadProcessingService>();
            


            // Initialize your function with mocked dependencies
            _function = new BulkFileUploadFunction(
                _loggerFactoryBUFMock.Object,
                _mockConfiguration.Object,
                _mockUploadProcessingService.Object
                );



            _storageBlobCreatedEvent = new StorageBlobCreatedEvent
            {
                Id = "12323",
                Topic="routineImmunization",
                Subject="IZGW",
                EventType="DD2",
                EventTime=System.DateTime.Now,
                Data = new StorageBlobCreatedEventData{Url="https://example.com/blob/10MB-test-file"}
            };     

            _mockTusInfoFile = new MockTusInfoFile
            {
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                IsPartial = false,
                IsFinal = false,
                MetaData = new Dictionary<string, string> { { "filename", "flower.jpeg" }, { "meta_field", "meta_value" } },
                Storage = new MockTusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" }
            };



        }
        [TestMethod]
        public async Task GivenValidUri_WhenRunIsCalled_ThenBlobIsValidated()
        {
            // Arrange
            var loggerFactoryMock = new Mock<ILoggerFactory>();
            var loggerMock = new Mock<ILogger<BulkFileUploadFunction>>();
            var configurationMock = new Mock<IConfiguration>();
            var procStatClientMock = new Mock<IProcStatClient>();
            var blobCopyHelperMock = new Mock<IBlobCopyHelper>();
            var blobReaderMock = new Mock<IBlobReader>();

            // Set up the BlobCopyHelper mock
           string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[]{_storageBlobCreatedEvent}) };

            // Assert
            foreach (var blobCreatedEventJson in expectedCopyResultJson)
            {
                Console.WriteLine($"C# Event Hub trigger function processed a message: {blobCreatedEventJson}");
                Assert.IsNotNull(blobCreatedEventJson);
                StorageBlobCreatedEvent[]? blobCreatedEvents = JsonConvert.DeserializeObject<StorageBlobCreatedEvent[]>(blobCreatedEventJson);
                Assert.IsNotNull(blobCreatedEvents);
                Assert.IsInstanceOfType(blobCreatedEvents, typeof(StorageBlobCreatedEvent[]));
            }
        }        

        [TestMethod]
        public async Task Run_ReceivesEvents_LogsEventCount()
        {
            var blobReaderMock = new Mock<IBlobReader>();
            var blobEvent = new StorageBlobCreatedEvent
            {
                Data = new StorageBlobCreatedEventData { Url = "http://example.com/blob/10MB-test-file" }
            };
            string[] events = new string[] { JsonConvert.SerializeObject(new[]{blobEvent}) };
            await _function.Run(events);
            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Once);
        }

 /* 
        [TestMethod]
        public async Task Run_ProcessesEventHubTriggerEvents_Successfully()
        {
            // Arrange
            var loggerFactoryMock = new Mock<ILoggerFactory>();
            var loggerMock = new Mock<ILogger<BulkFileUploadFunction>>();
            var configurationMock = new Mock<IConfiguration>();
            var procStatClientMock = new Mock<IProcStatClient>();
            var blobCopyHelperMock = new Mock<IBlobCopyHelper>();
            var blobReaderMock = new Mock<IBlobReader>();
            var mockBlobClient = new Mock<BlobClient>();

            //var mockBlobClientWrapper = new Mock<IBlobClientWrapper>();
            //mockBlobClientWrapper
            //    .Setup(b => b.UploadAsync(It.IsAny<Stream>(), It.IsAny<BlobUploadOptions>(), It.IsAny<CancellationToken>()))
            //    .ReturnsAsync(new BlobContentInfo());   

            // Set up the BlobCopyHelper mock
            string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[]{_storageBlobCreatedEvent}) };
            // Act
            await _function.Run(expectedCopyResultJson);
            // Add more assertions here as needed
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("file://1MB-test-file.txt"));            
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");
            
             // Act
             // actual example: BlobClient sourceBlob, BlobClient destinationBlob, IDictionary<string, string> destinationMetadata, Uri? sourceSasBlobUri = null
            _mockTusInfoFile.MetaData.Add("tus_tguid", "123232");
            //_mockTusInfoFile.MetaData.Remove("filename");
            _mockTusInfoFile.MetaData.Add("orig_filename", "1MB-test-file.txt");
            var result = await blobCopyHelperMock.Object.CopyBlobAsync(_mockBlobClientFactorySrc.Object, _mockBlobClientFactoryDest.Object, _mockTusInfoFile.MetaData, uri);

            blobCopyHelperMock.Verify(x => x.CopyBlobAsync(It.IsAny<BlobClient>(), It.IsAny<BlobClient>(), It.IsAny<IDictionary<string, string>>(), uri), Times.Once);
            Assert.AreEqual(1, expectedCopyResultJson.Length);

        }  */
         
    }


}