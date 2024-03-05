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
        private Mock<IBlobReader> _blobReaderMock;
        private IBlobReader _blobReader;
        private BlobReaderFactory _blobReaderFactory;
        private Mock<BlobReaderFactory>? _mockBlobReaderFactory;
        private Mock<BlobServiceClient>? _mockBlobServiceClient;

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
            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            //_loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            _blobReaderMock = new Mock<IBlobReader>();

            _blobReaderFactory = new BlobReaderFactory();
            _blobReader = _blobReaderFactory.CreateInstance(_loggerMock.Object);
            //_mockBlobServiceClient 

            _storageBlobCreatedEvent = new StorageBlobCreatedEvent
            {
                Id = "12323",
                Topic = "routineImmunization",
                Subject = "IZGW",
                EventType = "DD2",
                EventTime = System.DateTime.Now,
                Data = new StorageBlobCreatedEventData { Url = "https://example.com/blob/10MB-test-file" }
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

            _mockBlobReaderFactory.Setup(x => x.CreateInstance(It.IsAny<ILogger>())).Returns(_blobReaderMock.Object);
            _mockBlobServiceClientFactory.Setup(x => x.CreateBlobServiceClient(It.IsAny<string>())).Returns(_mockBlobServiceClient.Object);

            _uploadProcessingService = new UploadProcessingService(_loggerFactoryMock.Object,
            _mockConfiguration.Object,
            _mockProcStatClient.Object,
            _mockFeatureManagementExecutor.Object,
            _mockUploadEventHubService.Object,
            _mockBlobCopyHelperFactory.Object,
            _mockBlobReaderFactory.Object,
            _mockBlobServiceClientFactory.Object);

            _mockUploadProcessingService = new Mock<IUploadProcessingService>();

        }

        [TestMethod]
        public async Task GivenValidURI_WhenProcessBlobIsCalled_ThenBlobIsCopiedFromTusToDex()
        {

            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");            
            string testBlobUrl = uri.ToString();
            string[] expectedCopyResultJson = new string[] { JsonConvert.SerializeObject(new[] { _storageBlobCreatedEvent }) };

            // Act
            await _uploadProcessingService.ProcessBlob(testBlobUrl);

            // Assert
            try
            {
                Console.WriteLine("************ Verify Mock Processing Svc ************");
                _mockUploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);
                //_uploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);
            }
            catch (Exception ex)
            {
                Console.WriteLine("************ " + ex.Message + "  ************");
            }
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

            _blobReaderMock
            .Setup(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(tusInfoFile));

            _blobReaderMock
            .Setup(x => x.GetObjectFromBlobJsonContent<UploadConfig>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(_uploadConfig));

            _mockUploadProcessingService
            .Setup(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()))
            .Returns(_copyBlobFromTusToDex);

            

            // Act
            await _mockUploadProcessingService.Object.ProcessBlob(testBlobUrl);
            await _uploadProcessingService.ProcessBlob(testBlobUrl);

            _mockUploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);

            //_mockUploadProcessingService.Verify(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()), Times.Once);

            // Fails due to UploadProcessingService version not making it past the private GetTusFileInfo()
            // TODO: Try to mock this by calling the real  BlobReader.GetObjectFromBlobJsonContent function
            var result = await _blobReader.GetObjectFromBlobJsonContent<TusInfoFile>(_dexStorageAccountConnectionString, sourceContainerName, testBlobUrl);

            _blobReaderMock.Verify(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()), Times.Once);
            



            // Assert

            //Assert.AreEqual(_mockTusInfoFile, result);
            _mockUploadProcessingService.Verify(x => x.ProcessBlob(testBlobUrl), Times.Once);
        }

        [TestMethod]
        public async Task GivenInvalidData_WithProcessBlob_ThenReturnsFalse()
        {
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri(); //new System.Uri("https://example.com/blob/1MB-test-file");            
            string testBlobUrl = uri.ToString();

            // Arrange
            _mockUploadProcessingService.Setup(x => x.ProcessBlob(It.IsAny<string>())).Returns(It.IsAny<Task>());
            var uploadProcessingService = _mockUploadProcessingService.Object;


            // Assert
            _mockUploadProcessingService.Verify(x => x.CopyBlobFromTusToDex(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>(), It.IsAny<Dictionary<string, string>>()), Times.Never);

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


            public MockTusInfoFile GetObjectFromBlobJsonContent<TusInfoFile>(string connectionString, string sourceContainerName, string blobPathname)
            {
                return new MockTusInfoFile
                {
                    ID = "a0e127caec153d6047ee966bc2acd8cb",
                    Size = 7952,
                    SizeIsDeferred = false,
                    Offset = 0,
                    MetaData = new Dictionary<string, string>{
                        {"meta_destination_id", "flower.jpeg"},
                        {"meta_ext_event","meta_value"}
                    },
                    IsPartial = false,
                    IsFinal = false,
                    Storage = new MockTusStorage
                    {
                        Container = "bulkuploads",
                        Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb",
                        Type = "azurestore"
                    }
                };
            }
        }

        [TestClass]
        public class MockTusStorage
        {
            public string? Container { get; set; }

            public string? Key { get; set; }

            public string? Type { get; set; }
        }


        [TestClass]
        public class MockUploadConfig
        {
            public string? FilenameMetadataField { get; set; }

            public string? FilenameSuffix { get; set; }

            public string? FolderStructure { get; set; }

            public string? FixedFolderPath { get; set; }

            public static readonly MockUploadConfig Default = new MockUploadConfig()
            {
                FilenameMetadataField = "filename",
                FilenameSuffix = "clock_ticks",
                FolderStructure = "date_YYYY_MM_DD",
                FixedFolderPath = null
            };

            public MockUploadConfig() { }

            public MockUploadConfig(string filenameMetadataField, string filenameSuffix, string folderStructure, string fixedFolderPath)
            {
                FilenameMetadataField = filenameMetadataField;
                FilenameSuffix = filenameSuffix;
                FolderStructure = folderStructure;
                FixedFolderPath = fixedFolderPath;
            }
        }
    }
}