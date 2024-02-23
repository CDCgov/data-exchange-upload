using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using Azure;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using Azure.Identity;
//using System.Text.Json;
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
        private Mock<IProcStatClient> _mockProcStatClient;
        private Mock<BlobCopyHelperFactory> _mockBlobCopyHelperFactory;
        private Mock<BlobReaderFactory> _mockBlobReaderFactory;
        private Mock<ILogger<BulkFileUploadFunction>> _loggerMock;
        private Mock<ILoggerFactory> _loggerFactoryMock;
        private Mock<IConfiguration> _mockConfiguration;
        private BulkFileUploadFunction _function;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private MockTusInfoFile _mockTusInfoFile;
        private MockTusStorage _mockTusStorage;

        [TestInitialize]
        public void Initialize()
        {
            // Mock the ProcStatClient
            _mockProcStatClient = new Mock<IProcStatClient>();

            // Mock the BlobCopyHelperFactory
            _mockBlobCopyHelperFactory = new Mock<BlobCopyHelperFactory>();
            var mockBlobCopyHelper = new Mock<IBlobCopyHelper>();
            _mockBlobCopyHelperFactory.Setup(f => f.CreateInstance(It.IsAny<ILogger>())).Returns(mockBlobCopyHelper.Object);

            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            var mockBlobReader = new Mock<IBlobReader>();
            _mockBlobReaderFactory.Setup(f => f.CreateInstance(It.IsAny<ILogger>())).Returns(mockBlobReader.Object);

            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<BulkFileUploadFunction>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);


            // Mock IConfiguration
            _mockConfiguration = new Mock<IConfiguration>();

            // real function: 
            //  public BulkFileUploadFunction(IProcStatClient procStatClient, 
            //  BlobCopyHelperFactory blobCopyHelperFactory, 
            //  ILoggerFactory loggerFactory, 
            //  IConfiguration configuration)

            // Initialize your function with mocked dependencies
            _function = new BulkFileUploadFunction(
                _mockProcStatClient.Object,
                _mockBlobCopyHelperFactory.Object,
                _mockBlobReaderFactory.Object,
                _loggerFactoryMock.Object,
                _mockConfiguration.Object);          


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
        public async Task TestValidStorageBlobCreatedEvent()
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
 
        // TODO: Mock function await blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, tusInfoPathname)


        [TestMethod]
        public async Task TestGetObjectFromBlobJsonContent()
        {
            // Arrange
            var blobReaderMock = new Mock<IBlobReader>();
            //var tusInfoFile = new MockTusInfoFile(); // Populate this with the expected return value
            var _dexStorageAccountConnectionString = "YourConnectionString";
            var sourceContainerName = "YourContainerName";
            var tusInfoPathname = "YourPathName";

            blobReaderMock
                .Setup(x => x.GetObjectFromBlobJsonContent<MockTusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, tusInfoPathname))
                .ReturnsAsync(_mockTusInfoFile);

            // Act
            var result = await blobReaderMock.Object.GetObjectFromBlobJsonContent<MockTusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, tusInfoPathname);

            // Assert
            Assert.AreEqual(_mockTusInfoFile, result);
            blobReaderMock.Verify(x => x.GetObjectFromBlobJsonContent<MockTusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, tusInfoPathname), Times.Once);
        }
 
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

            // Set up the BlobCopyHelper mock
            string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[]{_storageBlobCreatedEvent}) };
            // Act
            await _function.Run(expectedCopyResultJson);
            // Add more assertions here as needed
            System.Uri uri = new System.Uri("https://example.com/blob/10MB-test-file");
            blobCopyHelperMock.Verify(x => x.CopyBlobAsync(It.IsAny<BlobClient>(), It.IsAny<BlobClient>(), It.IsAny<IDictionary<string, string>>(), uri), Times.Once);

        } 
         
    }

    [TestClass]
    public class MockTusInfoFile
    {

        public string? ID { get; set; }

        public long Size { get; set; }

        public bool SizeIsDeferred { get; set; }

        public long Offset { get; set; }

        public bool IsPartial { get; set; }

        public bool IsFinal { get; set; }

        public Dictionary<string, string>? MetaData { get; set; }

        public MockTusStorage? Storage { get; set; }


        private MockTusInfoFile GetObjectFromBlobJsonContent<TusInfoFile>(string connectionString, string sourceContainerName, string blobPathname)
        {
            return new MockTusInfoFile{
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                MetaData =  new Dictionary<string, string>{
                    {"filename", "flower.jpeg"},
                    {"meta_field","meta_value"}
                },
                IsPartial = false,
                IsFinal = false,
                Storage = new MockTusStorage{
                Container = "bulkuploads",
                Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb",
                Type = "azurestore"
                }
            };
        }
    }

     public class MockTusStorage
    {
        public string? Container { get; set; }

        public string? Key { get; set; }

        public string? Type { get; set; }
    }

}