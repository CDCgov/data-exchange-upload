using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BulkFileUploadFunctionAppTests
{
    [TestClass]
    public class BulkFileUploadFunctionTests
    {
        private Mock<ILogger<UploadProcessingService>>? _loggerMock;
        private Mock<ILoggerFactory>? _loggerFactoryMock;
        private Mock<ILoggerFactory>? _loggerFactoryBUFMock;
        private BulkFileUploadFunction? _function;
        private Mock<IUploadProcessingService>? _mockUploadProcessingService;

        [TestInitialize]
        public void Initialize()
        {

            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<UploadProcessingService>>();
            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);

            _loggerFactoryBUFMock = new Mock<ILoggerFactory>();
            _loggerFactoryBUFMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            
            _mockUploadProcessingService = new Mock<IUploadProcessingService>();


            // Initialize your function with mocked dependencies
            _function = new BulkFileUploadFunction(
                _loggerFactoryBUFMock.Object,
                _mockUploadProcessingService.Object
                );

        }    

        [TestMethod]
        public async Task GivenValidUri_WhenRunIsCalled_ThenLogEventsCopyAllVerified()
        {
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";

            var blobReaderMock = new Mock<IBlobReader>();
            var blobEvent = new StorageBlobCreatedEvent
            {
                Data = new StorageBlobCreatedEventData { Url = testBlobUrl }
            };
            string[] events = new string[] { JsonSerializer.Serialize(new[]{blobEvent}) };

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

            CopyConfig copyConfig = new CopyConfig
            {
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                TargetEnums = new List<CopyTargetsEnum>()
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

            _mockUploadProcessingService
                .Setup(x => x.GetCopyPrereqs(testBlobUrl))
                .Returns(Task.FromResult(copyPrereqs));

             await _function.Run(events);
            _loggerMock.Verify(x => x.Log(
                It.IsAny<LogLevel>(),
                It.IsAny<EventId>(),
                It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
                It.IsAny<Exception>(),
                It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Once);

            _mockUploadProcessingService.Verify(x => x.CopyAll(copyPrereqs), Times.Once);

        }

        [TestMethod]
        public async Task GivenFail_WhenRunIsCalledWithInvalidCopyPrereqs_ThenCopyAllIsNotCalled()
        {
            string testBlobUrl = "https://example.com/blob/1MB-test-file.txt";

            var blobReaderMock = new Mock<IBlobReader>();
            var blobEvent = new StorageBlobCreatedEvent
            {
                Data = new StorageBlobCreatedEventData { Url = testBlobUrl }
            };
            string[] events = new string[] { JsonSerializer.Serialize(new[] { blobEvent }) };

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

            CopyConfig copyConfig = new CopyConfig
            {
                FilenameSuffix = ".txt",
                FolderStructure = "/blob",
                TargetEnums = new List<CopyTargetsEnum>()
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

            var copyPrereqs = new CopyPrereqs();

            _mockUploadProcessingService
                .Setup(x => x.GetCopyPrereqs(""))
                .Returns(Task.FromResult(copyPrereqs));

            await _function.Run(events);
            _loggerMock.Verify(x => x.Log(
                It.IsAny<LogLevel>(),
                It.IsAny<EventId>(),
                It.Is<It.IsAnyType>((v, t) => v.ToString().Contains("Received events count:")),
                It.IsAny<Exception>(),
                It.IsAny<Func<It.IsAnyType, Exception, string>>()), Times.Once);

            _mockUploadProcessingService.Verify(x => x.CopyAll(copyPrereqs), Times.Never);
        }
    }
}