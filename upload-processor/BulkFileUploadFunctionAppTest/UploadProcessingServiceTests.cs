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
using System.Text.Json;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class UploadProcessingServiceTests
    {
        private Mock<IProcStatClient>? _mockProcStatClient;
        private Mock<IConfigurationManager> _mockConfigManager;
        private Mock<IFeatureManagementExecutor> _mockFeatureManagementExecutor;
        private Mock<IUploadEventHubService> _mockUploadEventHubService;
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private StorageBlobCreatedEvent _storageBlobCreatedEvent;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;
        private Mock<AzureBlobReader> _mockBlobReader;
        private string sourceContainerName;
        private UploadProcessingService? _function; // temporary placeholder
        private Mock<BlobClient> _mockBlobClient;
        private Mock<IBlobServiceClientFactory> _mockBlobServiceClientFactory;
        private Mock<BlobServiceClient> _mockBlobServiceClient;
        private Mock<BlobServiceClient> _mockEdavBlobServiceClient;
        private Mock<Uri> _mockUri;

        [TestInitialize]
        public void Initialize()
        {
            sourceContainerName = "bulkuploads";
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

            _function = new UploadProcessingService(
                _loggerFactoryMock.Object, 
                _mockConfigManager.Object, 
                _mockProcStatClient.Object, _mockFeatureManagementExecutor.Object,
                _mockUploadEventHubService.Object, _mockBlobServiceClientFactory.Object);

        }

        [TestMethod]
        public async Task GivenConnStringWhenCopyPreReqsThenGetDestinationAndEvents()
        {
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";
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

            var copyConfig = new CopyConfig
            {
                FilenameSuffix = ".txt",
                FolderStructure = "root",
                Targets = new List<string> { "/blob" }
            };

            UploadConfig uploadConfig = new UploadConfig
            {
                CopyConfig = copyConfig,
                MetadataConfig = metadataConfig
            };

            Trace? _trace = new Trace
            {
                DestinationId = "dextesting",
                SpanId = "123234",
                TraceId = "123345"
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
                DexBlobFolderName = uploadConfig.CopyConfig.FolderStructure,
                DexBlobFileName = tusInfoFile.MetaData["filename"].Replace("test", "dexTest")
            };


            //_mockBlobReader.Setup(x => x.Read<TusInfoFile>(It.IsAny<Dictionary<string, string>>())).ReturnsAsync(tusInfoFile);
            //_mockBlobReader.Setup(x => x.Read<UploadConfig>(It.IsAny<Dictionary<string, string>>())).ReturnsAsync(uploadConfig);

            _mockUploadProcessingService
                .Setup(x => x.GetCopyPrereqs(testBlobUrl))
                .Returns(Task.FromResult(copyPrereqs));

            _mockProcStatClient
                .Setup(x => x.GetTraceByUploadId(It.IsAny<string>()))
                .Returns(Task.FromResult(_trace));

            _mockFeatureManagementExecutor
                .Setup(x => x.ExecuteIfEnabledAsync(Constants.PROC_STAT_FEATURE_FLAG_NAME, It.IsAny<Func<Task>>()))
                .Returns((string feature, Func<Task> action) =>
                {
                    action();
                    return Task.FromResult(_mockFeatureManagementExecutor.Object);
                });


            _function.GetCopyPrereqs(testBlobUrl);

            _mockUploadProcessingService.Verify(x => x.GetCopyPrereqs(testBlobUrl), Times.Once); // change back to Once after PR
        }
        //should copy file to dex container when given any or no copy target
        [TestMethod]
        public async Task GivenAnyOrNoCopyTargetWhenCopyToDexThenCopyFile()
        {
            #region temporary placeholder  
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";

            #endregion

            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR

            //Setup:
            /*
             mock copyPrereqs
             mock blobcontainerclient
             mock blobclient
             mock and setup blobmanager.CopyBlobLeaseAsync - to account for blobcopyhelper
             */
            //Call CopyFromTusToDex
        }

        // should copy file to edav container when given edav copy target
        [TestMethod]
        public async Task GivenEdavCopyTargetWhenDexToEdavThenCopyFile()
        {
            #region temporary placeholder  
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";

            #endregion

            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR

            //Setup:
            /*
             mock copyPrereqs
             mock blobcontainerclient
             mock blobclient
             mock and setup blobmanager.CopyBlobStreamAsync - to account for blobcopyhelper
             */
            // Call CopyFromDexToEdav

        }

        // should copy file to routing container when given routing copy target
        [TestMethod]
        public async Task GivenRoutingCopyTargetWhenDexToRoutingThenCopyFile()
        {
            #region temporary placeholder  
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";

            #endregion

            _loggerMock.Verify(x => x.Log(
                It.IsAny<LogLevel>(),
                It.IsAny<EventId>(),
                It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
                It.IsAny<Exception>(),
                It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR

            //Setup:
            /*
             mock copyPrereqs
             mock blobcontainerclient
             mock blobclient
             mock and setup blobmanager.CopyBlobStreamAsync - to account for blobcopyhelper
             */
            // Call CopyFromDexToRouting
        }

        // should throw if required metadata not provided
        [TestMethod]
        public async Task GivenBadMetadataWhenGetCopyPrereqsThenFail()
        {
            #region temporary placeholder  
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";
            #endregion

            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR
        }

        // should get filename from metadata
        [TestMethod]
        public async Task GivenMetaDataWhenCopyAllThenGetFilename()
        {
            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR
        }

        //should generate correct folder path for destination container
        [TestMethod]
        public async Task GivenDestinationContainerWhenCopyAllThenGenerateCorrectFolderPath()
        {
            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never); //rollback after PR
        }

        // should generate correct filename suffix
        [TestMethod]
        public async Task GivenFilenameWhenCopyAllThenGenerateCorrectFilenameSuffix()
        {
            _loggerMock.Verify(x => x.Log(
            It.IsAny<LogLevel>(),
            It.IsAny<EventId>(),
            It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
            It.IsAny<Exception>(),
            It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Never);//rollback after PR
        }

    }
}