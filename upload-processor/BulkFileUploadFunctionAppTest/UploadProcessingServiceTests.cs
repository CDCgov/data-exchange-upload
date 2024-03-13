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
        private Mock<IProcStatClient>? _mockProcStatClient;
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<IFeatureManagementExecutor>? _mockFeatureManagementExecutor;
        private string _dexAzureStorageAccountName;
        private string _dexAzureStorageAccountKey;
        private Mock<IBlobReader> _mockBlobReader;
        private BlobReaderFactory _blobReaderFactory;
        private Mock<BlobReaderFactory>? _mockBlobReaderFactory;
        private UploadConfig _uploadConfig;
        private TusInfoFile tusInfoFile;
        private string sourceContainerName;

        [TestInitialize]
        public void Initialize()
        {
            sourceContainerName = "bulkuploads";
            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);

            _mockProcStatClient = new Mock<IProcStatClient>();
            _mockFeatureManagementExecutor = new Mock<IFeatureManagementExecutor>();
            _blobReaderFactory = new BlobReaderFactory();
            _mockBlobReader = new Mock<IBlobReader>();
            _mockBlobReaderFactory = new Mock<BlobReaderFactory>();
            _mockBlobReaderFactory
                .Setup(x => x.CreateInstance(It.IsAny<ILogger>()))
                .Returns(_mockBlobReader.Object);

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

        }

        [TestMethod]
        public async Task GivenValidURI_WheGetCopyPrereqs_ThenBlobIsCopiedFromTusToDex()
        {
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri();
            string testBlobUrl = uri.ToString();
            string[] expectedCopyResultJson = new string[] { JsonSerializer.Serialize(new[] { _storageBlobCreatedEvent }) };

           
            TusInfoFile tusInfoFile = new TusInfoFile
            {
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                IsPartial = false,
                IsFinal = false,
                MetaData = new Dictionary<string, string> { { "meta_destination_id", "dextesting" }, { "meta_ext_event", "testevent1" }, { "filename", "test.txt" } },
                Storage = new TusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" }
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
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob",
                MetadataConfig = metadataConfig
            };

            Trace? _trace = new Trace
            {
                DestinationId = "dextesting",
                SpanId = "123234",
                TraceId = "123345"
            };

            var _destinationAndEventsMock = new List<DestinationAndEvents>
            {
                new DestinationAndEvents
                {
                    destinationId = tusInfoFile.MetaData["meta_destination_id"],
                    extEvents = new List<ExtEvent>
                    {
                        new ExtEvent
                        {
                            name = tusInfoFile.MetaData["meta_ext_event"], 
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
                DexBlobFolderName = _uploadConfig.FixedFolderPath,
                DexBlobFileName = tusInfoFile.MetaData["filename"].Replace("test", "dexTest")
            };

            _mockUploadProcessingService
                .Setup(x => x.GetCopyPrereqs(testBlobUrl))
                .Returns(Task.FromResult(copyPrereqs));

            _mockUploadProcessingService
                .Setup(x => x.CopyFromTusToDex(copyPrereqs))
                .Returns(Task.FromResult(It.IsAny<string>()));

            await _mockUploadProcessingService.Object.CopyAll(copyPrereqs);


            _mockFeatureManagementExecutor.Verify(x => x.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, It.IsAny<Func<Task>>()), Times.Never);
            _mockBlobReader.Verify(x=> x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()), Times.Never);
        }


        [TestMethod]
        public async Task GivenBlobJsonContent_WhenBlobReader_ThenGetTusInfoFile()
        {
            // Arrange
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri();     
            string testBlobUrl = uri.ToString();

            TusInfoFile tusInfoFile = new TusInfoFile
            {
                ID = "a0e127caec153d6047ee966bc2acd8cb",
                Size = 7952,
                SizeIsDeferred = false,
                Offset = 0,
                IsPartial = false,
                IsFinal = false,
                MetaData = new Dictionary<string, string> { { "meta_destination_id", "dextesting" }, { "meta_ext_event", "testevent1" }, { "filename", "test.txt" } },
                Storage = new TusStorage { Container = "bulkuploads", Key = "tus-prefix/a0e127caec153d6047ee966bc2acd8cb", Type = "azurestore" }
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
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                FixedFolderPath = "/blob",
                MetadataConfig = metadataConfig
            };

            Trace? _trace = new Trace
            {
                DestinationId = "dextesting",
                SpanId = "123234",
                TraceId = "123345"
            };

            Task<UploadConfig> uploadConfigTask = Task.FromResult(_uploadConfig);
            Task<string> _copyBlobFromTusToDex = Task.FromResult(testBlobUrl);

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<TusInfoFile>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(tusInfoFile));

            _mockBlobReader
            .Setup(x => x.GetObjectFromBlobJsonContent<UploadConfig>(It.IsAny<string>(), It.IsAny<string>(), It.IsAny<string>()))
            .Returns(Task.FromResult(_uploadConfig));

            Assert.IsNotNull(tusInfoFile);
        }

        [TestMethod]
        public async Task GivenInvalidData_WithProcessBlob_ThenReturnsFalse()
        {
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri();
            string testBlobUrl = uri.ToString();
            // Arrange
            var uploadProcessingService = _mockUploadProcessingService.Object;
            // Assert
            _mockUploadProcessingService.Verify(x => x.CopyAll(It.IsAny<CopyPrereqs>()), Times.Never);
        }

        [TestMethod]
        public async Task GivenData_WithActualBlobReader_ThenReturnsBlobReader()
        {
            var mockUriWrapper = new Mock<IUriWrapper>();
            mockUriWrapper.Setup(u => u.GetUri()).Returns(new Uri("https://example.com/blob/1MB-test-file.txt"));
            System.Uri uri = mockUriWrapper.Object.GetUri();
            string testBlobUrl = uri.ToString();
            // Arrange
            var uploadProcessingService = _mockUploadProcessingService.Object;
            // Assert
            _mockUploadProcessingService.Verify(x => x.CopyAll(It.IsAny<CopyPrereqs>()), Times.Never);
        }
    }
}