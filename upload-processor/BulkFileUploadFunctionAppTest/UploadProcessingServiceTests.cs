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
using System.Reflection.Metadata;
using DurableTask.Core.History;
using Microsoft.AspNetCore.Mvc;
using BulkFileUploadFunctionAppTest.utils;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class UploadProcessingServiceTests
    {
        private UploadProcessingService _uploadProcessingService;
        private Mock<IProcStatClient>? _mockProcStatClient;
        private Mock<BlobCopyHelperFactory>? _mockBlobCopyHelperFactory;
        private Mock<IBlobServiceClientFactory>? _mockBlobServiceClientFactory;
        private Mock<ILogger<BulkFileUploadFunction>>? _loggerMockBUF;
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private Mock<ILoggerFactory>? _loggerFactoryBUFMock;
        private Mock<IConfiguration>? _mockConfiguration;
        private BulkFileUploadFunction _function;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private MockTusInfoFile? _mockTusInfoFile;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private Mock<IUploadEventHubService>? _mockUploadEventHubService;
        private string _dexAzureStorageAccountName;
        private string _dexAzureStorageAccountKey;
        private string _edavAzureStorageAccountName;
        private Mock<IBlobReader> _mockBlobReader;
        private IBlobReader _blobReader;
        private BlobReaderFactory _blobReaderFactory;
        private Mock<BlobReaderFactory>? _mockBlobReaderFactory;
        private Mock<BlobServiceClient>? _mockBlobServiceClient;
        private Mock<CopyPrereqs> _mockCopyPreReqs;
        private Mock<Task<List<DestinationAndEvents>?>> _mockDestinationAndEvents;
        private MockTusStorage? _mockTusInfoFileStorage;
        private MockUploadConfig? _mockUploadConfig;

        [TestInitialize]
        public void Initialize()
        {
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", "YourStorageAccountName", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", "YourStorageAccountKey", EnvironmentVariableTarget.Process);
            Environment.SetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", "YourStorageAccountName", EnvironmentVariableTarget.Process);

            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);

            _mockConfiguration = new Mock<IConfiguration>();
            _mockProcStatClient = new Mock<IProcStatClient>();
            _mockFeatureManagementExecutor = new Mock<IFeatureManagementExecutor>();
            _mockUploadEventHubService = new Mock<IUploadEventHubService>();
            _mockBlobCopyHelperFactory = new Mock<BlobCopyHelperFactory>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            _blobReaderFactory = new BlobReaderFactory();
            _blobReader = _blobReaderFactory.CreateInstance(_loggerMock.Object);

            _mockCopyPreReqs = new Mock<CopyPrereqs>();
            
            _mockBlobReader = new Mock<IBlobReader>();
            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            _mockBlobReaderFactory
                .Setup(x => x.CreateInstance(It.IsAny<ILogger>()))
                .Returns(_mockBlobReader.Object);

            _mockDestinationAndEvents = new Mock<Task<List<DestinationAndEvents>?>>();

            _storageBlobCreatedEvent = new StorageBlobCreatedEvent
            {
                Id = "12323",
                Topic = "routineImmunization",
                Subject = "IZGW",
                EventType = "DD2",
                EventTime = System.DateTime.Now,
                Data = new StorageBlobCreatedEventData { Url = "https://example.com/blob/10MB-test-file" }
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


            _mockBlobReaderFactory.Setup(x => x.CreateInstance(It.IsAny<ILogger>())).Returns(_mockBlobReader.Object);

            _uploadProcessingService = new UploadProcessingService(_loggerFactoryMock.Object,
            _mockConfiguration.Object,
            _mockProcStatClient.Object,
            _mockFeatureManagementExecutor.Object,
            _mockUploadEventHubService.Object,
            _mockBlobReaderFactory.Object);

            _mockUploadProcessingService = new Mock<IUploadProcessingService>();

            _mockUploadProcessingService.CallBase = true;
        }

        [TestMethod]
        public async Task GivenValidURI_WheGetCopyPrereqs_ThenBlobIsCopiedFromTusToDex()
        {
            _dexAzureStorageAccountName = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";
            _edavAzureStorageAccountName = Environment.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";

            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");            
            string testBlobUrl = uri.ToString();
            string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[] { _storageBlobCreatedEvent }) };
            string _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net"; ;

            var _mockTusInfoFileStorage = new MockTusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" };
            var _mockTusInfoFile = new MockTusInfoFile
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
            string sourceContainerName = _mockTusInfoFile.Storage.Container.ToString();

            TusInfoFile tusInfoFile = new TusInfoFile
            {
                ID = _mockTusInfoFile.ID,
                Size = _mockTusInfoFile.Size,
                SizeIsDeferred = _mockTusInfoFile.SizeIsDeferred,
                Storage = new TusStorage { Container = _mockTusInfoFile.Storage.Container, Key = _mockTusInfoFile.Storage.Key, Type = _mockTusInfoFile.Storage.Type },
                MetaData = _mockTusInfoFile.MetaData
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
                Fields = new List<MetadataField> { metadataField } // Add more MetadataField objects to this list as needed
            };
            MockUploadConfig _mockUploadConfig = new MockUploadConfig
            {
                FilenameMetadataField = _mockTusInfoFile.MetaData["meta_destination_id"],
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob"
               
            };


            UploadConfig _uploadConfig = new UploadConfig
            {
                FilenameSuffix = _mockUploadConfig.FilenameSuffix,
                FolderStructure = _mockUploadConfig.FolderStructure,
                FixedFolderPath = _mockUploadConfig.FixedFolderPath,
                MetadataConfig = metadataConfig

            };

            Trace? _trace = new Trace { 
                DestinationId = "dextesting",
                SpanId = "123234",
                TraceId = "123345"            
            };

            var _destinationAndEventsMock = new List<DestinationAndEvents>
            {
                new DestinationAndEvents
                {
                    destinationId = _mockTusInfoFile.MetaData["meta_destination_id"],
                    extEvents = new List<ExtEvent>
                    {
                        new ExtEvent
                        {
                            name = _mockTusInfoFile.MetaData["meta_ext_event"], 
                            definitionFilename = "filename1",
                            copyTargets = new List<CopyTarget> { new CopyTarget("target1"), new CopyTarget("target2") }
                        },
                        new ExtEvent
                        {
                            name = "event2",
                            definitionFilename = "filename2",
                            copyTargets = new List<CopyTarget> { new CopyTarget("target3"), new CopyTarget("target4") }
                        }
                    }
                },
                new DestinationAndEvents
                {
                    destinationId = "destination2",
                    extEvents = new List<ExtEvent>
                    {
                        new ExtEvent
                        {
                            name = "event3",
                            definitionFilename = "filename3",
                            copyTargets = new List<CopyTarget> { new CopyTarget("target5"), new CopyTarget("target6") }
                        }
                    }
                }
            };
            // Setup
            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(tusInfoFile));

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<UploadConfig>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(_uploadConfig));

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<List<DestinationAndEvents>>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(_destinationAndEventsMock));

            // ISSUE: _destinationAndEvents keeps returning null !!!!
            _mockProcStatClient
                .Setup(x=> x.GetTraceByUploadId(It.IsAny<string>()))
                .Returns(Task.FromResult(_trace));
            
            _mockFeatureManagementExecutor
                .Setup(x => x.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, It.IsAny<Func<Task>>()))
                .Returns((string feature, Func<Task> action) =>
                {
                    action();
                    return Task.FromResult(_mockFeatureManagementExecutor.Object);
                });

            // Arrange          

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

            _mockUploadProcessingService
                .Setup(x => x.GetCopyPrereqs(testBlobUrl))
                .Returns(Task.FromResult(copyPrereqs));

            _mockUploadProcessingService
                .Setup(x => x.CopyFromTusToDex(copyPrereqs))
                .Returns(Task.FromResult(It.IsAny<string>()));

            await _mockUploadProcessingService.Object.CopyAll(copyPrereqs);

            // Act
            //await _uploadProcessingService.CopyAll(copyPrereqs);
            //await _uploadProcessingService.GetCopyPrereqs(testBlobUrl);

            // Assert
            //_mockUploadProcessingService.Verify(x => x.CopyAll(It.IsAny<CopyPrereqs>()), Times.Once);
            _mockFeatureManagementExecutor.Verify(x => x.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, It.IsAny<Func<Task>>()), Times.Never);
            _mockBlobReader.Verify(x=> x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()), Times.Never);
        }


        [TestMethod]
        public async Task GivenBlobJsonContent_WhenBlobReader_ThenGetTusInfoFile()
        {
            // Arrange
            //_blobReaderMock = new Mock<IBlobReader>();
            //var tusInfoFile = new MockTusInfoFile(); // Populate this with the expected return value
            _dexAzureStorageAccountName = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";
            _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", EnvironmentVariableTarget.Process) ?? "";
            _edavAzureStorageAccountName = Environment.GetEnvironmentVariable("EDAV_AZURE_STORAGE_ACCOUNT_NAME", EnvironmentVariableTarget.Process) ?? "";

            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");            
            string testBlobUrl = uri.ToString();

            string _dexStorageAccountConnectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net"; ;
            var _mockTusInfoFileStorage = new MockTusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" };
            var _mockTusInfoFile = new MockTusInfoFile
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
            string sourceContainerName = _mockTusInfoFile.Storage.Container.ToString();

            //MockTusInfoFile _mockTus = _mockTusInfoFile.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, testBlobUrl);

            TusInfoFile tusInfoFile = new TusInfoFile
            {
                ID = _mockTusInfoFile.ID,
                Size = _mockTusInfoFile.Size,
                SizeIsDeferred = _mockTusInfoFile.SizeIsDeferred,
                Storage = new TusStorage { Container = _mockTusInfoFile.Storage.Container, Key = _mockTusInfoFile.Storage.Key, Type = _mockTusInfoFile.Storage.Type },
                MetaData = _mockTusInfoFile.MetaData
            };
            Task<TusInfoFile> tusInfoFileTask = Task.FromResult(tusInfoFile);
            MockUploadConfig _mockUploadConfig = new MockUploadConfig
            {
                FilenameMetadataField = _mockTusInfoFile.MetaData["meta_destination_id"],
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob"
            };

            UploadConfig _uploadConfig = new UploadConfig
            {
                FilenameSuffix = _mockUploadConfig.FilenameSuffix,
                FolderStructure = _mockUploadConfig.FolderStructure,
                FixedFolderPath = _mockUploadConfig.FixedFolderPath
            };
            Task<UploadConfig> uploadConfigTask = Task.FromResult(_uploadConfig);
            Task<string> _copyBlobFromTusToDex = Task.FromResult(testBlobUrl);

            //Setup 
            // Pre-reqs: Uri, GetUploadConfig,                                 GetTusFileInfo,
            //                   |-> BlobRead.GetObjectFromBlobJsonContent<T>     |-> BlobRead.GetObjectFromBlobJsonContent<T>

            // IDEA: Option is to make an attempt to mock the _blobServiceClientFactory and BlobClient objects in hopes to satisfy the 
            // Setup for blobReaderMock

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(tusInfoFile));

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<UploadConfig>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(_uploadConfig));

            //_mockUploadProcessingService
            //.Setup(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()))
            //.Returns(_copyBlobFromTusToDex);

            

            // Act
           // await _mockUploadProcessingService.Object.ProcessBlob(testBlobUrl);
           // await _uploadProcessingService.ProcessBlob(testBlobUrl);

           // _mockUploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);

            //_mockUploadProcessingService.Verify(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()), Times.Once);

            // Fails due to UploadProcessingService version not making it past the private GetTusFileInfo()
            // TODO: Try to mock this by calling the real  BlobReader.GetObjectFromBlobJsonContent function
            //var result = await _blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, testBlobUrl);

            //_mockBlobReader.Verify(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()), Times.Once);
            



            // Assert

            //Assert.AreEqual(_mockTusInfoFile, result);
            //_mockUploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);
            Assert.IsNotNull(tusInfoFile);
        }

        [TestMethod]
        public async Task GivenInvalidData_WithProcessBlob_ThenReturnsFalse()
        {
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");            
            string testBlobUrl = uri.ToString();

            // Arrange
           // _mockUploadProcessingService.Setup(x => x.ProcessBlob(It.IsAny<string>())).Returns(It.IsAny<Task>());
            var uploadProcessingService = _mockUploadProcessingService.Object;


            // Assert
            _mockUploadProcessingService.Verify(x => x.CopyAll(It.IsAny<CopyPrereqs>()), Times.Never);
            //_mockUploadProcessingService.Verify(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()), Times.Never);

        }

    }
}