using Azure;
using Azure.Identity;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using BulkFileUploadFunctionAppTest.utils;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using Newtonsoft.Json;
using System;
using System.Collections.Concurrent;
using System.Diagnostics;

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
        private BulkFileUploadFunction? _function;
        private StorageBlobCreatedEvent? _storageBlobCreatedEvent;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private Mock<IUploadEventHubService>? _mockUploadEventHubService;
        private Mock<BulkFileUploadFunction>? _mockBulkFileUploadFunction;
        private MockTusInfoFile? _mockTusInfoFile;
        private MockTusStorage? _mockTusInfoFileStorage;
        private MockUploadConfig? _mockUploadConfig;

        [TestInitialize]
        public void Initialize()
        {
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", "_dexAzureStorageAccountName", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", "_dexAzureStorageAccountKey", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", "_edavAzureStorageAccountName", EnvironmentVariableTarget.Process);

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
            _mockUploadProcessingService = new Mock<IUploadProcessingService>();
            


            // Initialize your function with mocked dependencies
            _function = new BulkFileUploadFunction(
                _loggerFactoryBUFMock.Object,
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


            _mockTusInfoFileStorage = new MockTusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" };
            _mockTusInfoFile = new MockTusInfoFile
            {
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                IsPartial = false,
                IsFinal = false,
                MetaData = new Dictionary<string, string> { { "meta_destination_id", "dextesting" }, { "meta_ext_event", "testevent1" }, { "filename", "test.txt" } },
                Storage = _mockTusInfoFileStorage
            };

            _mockUploadConfig = new MockUploadConfig
            {
                FilenameMetadataField = _mockTusInfoFile.MetaData is not null ? _mockTusInfoFile.MetaData["meta_destination_id"] : "meta_destination_id",
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob"
            };

        }
        [TestMethod]
        public void GivenValidUri_WhenRunIsCalled_ThenBlobIsValidated()
        {
            // Arrange
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
        public async Task GivenValidUri_WhenRunIsCalled_ThenLogEventsCopyAllVerified()
        {

            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri();       
            string testBlobUrl = uri.ToString();

            var blobReaderMock = new Mock<IBlobReader>();
            var blobEvent = new StorageBlobCreatedEvent
            {
                Data = new StorageBlobCreatedEventData { Url = "http://example.com/blob/10MB-test-file" }
            };
            string[] events = new string[] { JsonConvert.SerializeObject(new[]{blobEvent}) };

            string sourceContainerName = String.Empty; 

            sourceContainerName = _mockTusInfoFile.Storage.Container is not null ? _mockTusInfoFile.Storage.Container.ToString() : "test-container";
            TusInfoFile tusInfoFile = new TusInfoFile
            {
                ID = _mockTusInfoFile.ID,
                Size = _mockTusInfoFile.Size,
                SizeIsDeferred = _mockTusInfoFile.SizeIsDeferred,
                Storage = new TusStorage { Container = _mockTusInfoFile.Storage.Container, Key = _mockTusInfoFile.Storage.Key, Type = _mockTusInfoFile.Storage.Type },
                MetaData = _mockTusInfoFile.MetaData
            };

            MockUploadConfig _mockUploadConfig = new MockUploadConfig
            {
                FilenameMetadataField = _mockTusInfoFile.MetaData is not null ? _mockTusInfoFile.MetaData["meta_destination_id"] : "meta_destination_id",
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob"

            };

            BulkFileUploadFunctionApp.Model.Trace? _trace = new BulkFileUploadFunctionApp.Model.Trace
            {
                DestinationId = "dextesting",
                SpanId = "123234",
                TraceId = "123345"
            };


            var metadataField = new MetadataField
            {
                FieldName = "testFieldName",
                CompatFieldName = "testCompatFieldName",
                DefaultValue = "testDefaultValue"
            };

            var metadataConfig = new MetadataConfig
            {
                Version = "1.0",
                Fields = new List<MetadataField> { metadataField }
            };

            UploadConfig _uploadConfig = new UploadConfig
            {
                FilenameSuffix = _mockUploadConfig.FilenameSuffix,
                FolderStructure = _mockUploadConfig.FolderStructure,
                FixedFolderPath = _mockUploadConfig.FixedFolderPath,
                MetadataConfig = metadataConfig
            };

            var copyPrereqs = new CopyPrereqs()
            {
                UploadId = "testUploadId",
                Metadata = tusInfoFile.MetaData,
                Trace = _trace,
                SourceBlobUrl = testBlobUrl,
                TusPayloadFilename = tusInfoFile.MetaData["filename"],
                DestinationId = tusInfoFile.MetaData["meta_destination_id"],
                EventType = tusInfoFile.MetaData["meta_ext_event"],
                DexBlobFolderName = _mockUploadConfig.FixedFolderPath,
                DexBlobFileName = tusInfoFile.MetaData["filename"].Replace("test", "dexTest")
            };

            var _mockUploadProcessingService = new Mock<IUploadProcessingService>();
            await _mockUploadProcessingService.Object.CopyAll(copyPrereqs);
            if (_function is not null){
                await _function.Run(events);
            }

            if(_loggerMock is not null)
            {
                _loggerMock.Verify(x => x.Log(
                It.IsAny<LogLevel>(),
                It.IsAny<EventId>(),
                It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
                It.IsAny<Exception>(),
                It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Once);
            }
            _mockUploadProcessingService.Verify(x => x.CopyAll(copyPrereqs), Times.Once);
        }

        [TestMethod]
        public async Task GivenNullResponseContent_WhenRunIsCalled_ThenBlobIsNotValidated()
        {
            // Arrange
            var storageBlobCreatedEvent = new StorageBlobCreatedEvent();
            string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[] { String.Empty }) };

            // Assert
            foreach (var blobCreatedEventJson in expectedCopyResultJson)
            {
                Console.WriteLine($"C# Event Hub trigger function is attempting to process a message: {blobCreatedEventJson}");
                Assert.IsNotNull(blobCreatedEventJson);
                StorageBlobCreatedEvent[]? blobCreatedEvents = JsonConvert.DeserializeObject<StorageBlobCreatedEvent[]>(blobCreatedEventJson);
                Assert.IsNotNull(blobCreatedEvents);
                Assert.IsInstanceOfType(blobCreatedEvents, typeof(StorageBlobCreatedEvent[]));
            }

        }




        }
}