
using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using BulkFileUploadFunctionApp;
using System.Threading.Tasks;
using System.Net;
using Azure;
using Azure.Storage.Blobs;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionAppTests
{   
   // 'HealthCheckFunctionTests' for testing health check functionality with mocked dependencies.
    [TestClass]
    public class HealthCheckFunctionTests
    {
        private Mock<IHttpRequestDataWrapper> _mockHttpRequestWrapper;
        private Mock<IHttpResponseDataWrapper> _mockResponseWrapper;
        private Mock<FunctionContext> _mockFunctionContext;
        private Mock<IBlobServiceClientFactory> _mockBlobServiceClientFactory;
        private Mock<IEnvironmentVariableProvider> _mockEnvironmentVariableProvider;
        private Mock<IServiceProvider> _mockServiceProvider;
        private Mock<IFunctionLogger> _mockLogger;


        // Initializes mock objects for HTTP request/response, function context, blob service, environment variables, and logger.
        // Sets up default behavior for these mocks to be used in health check function tests.
        [TestInitialize]
        public void Initialize()
        {
            _mockHttpRequestWrapper = new Mock<IHttpRequestDataWrapper>();
            _mockResponseWrapper = new Mock<IHttpResponseDataWrapper>();
            _mockFunctionContext = new Mock<FunctionContext>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            _mockEnvironmentVariableProvider = new Mock<IEnvironmentVariableProvider>();
            _mockLogger = new Mock<IFunctionLogger>();

            _mockHttpRequestWrapper.Setup(m => m.CreateResponse()).Returns(_mockResponseWrapper.Object);

            _mockEnvironmentVariableProvider.Setup(m => m.GetEnvironmentVariable(It.IsAny<string>())).Returns("test");
            
            var mockBlobServiceClient = new Mock<BlobServiceClient>();
            _mockBlobServiceClientFactory.Setup(m => m.CreateBlobServiceClient(It.IsAny<string>())).Returns(mockBlobServiceClient.Object);

           // Configures mock service provider for logging services and sets up the function context to use this provider.
        _mockServiceProvider = new Mock<IServiceProvider>();

        _mockServiceProvider.Setup(provider => provider.GetService(typeof(ILogger)))
                            .Returns(_mockLogger.Object);

        _mockFunctionContext.Setup(ctx => ctx.InstanceServices)
                            .Returns(_mockServiceProvider.Object);
        }

        [TestMethod]
        public async Task HealthCheckFunction_ReturnsHealthyResponse()
        {
            // Arrange
            // setting up a mock response wrapper to simulate the behavior of the actual response object used in the service.
            _mockResponseWrapper.Setup(m => m.WriteStringAsync(It.IsAny<string>())).Returns(Task.CompletedTask);
            _mockResponseWrapper.SetupProperty(m => m.StatusCode, HttpStatusCode.OK);

            // Act
            // Executes the HealthCheckFunction with mocked dependencies to test its behavior.
            var result = await HealthCheckFunction.Run(
                _mockHttpRequestWrapper.Object, // HttpRequestData is not directly used in the function
                _mockFunctionContext.Object,
                _mockBlobServiceClientFactory.Object,
                _mockEnvironmentVariableProvider.Object,
                _mockLogger.Object);

        
            // Asserts that the HealthCheckFunction returns HttpStatusCode.OK, writes "Healthy!" once to the response, 
            // and sets the response status code to HttpStatusCode.OK.
            Assert.AreEqual(HttpStatusCode.OK, result);
            _mockResponseWrapper.Verify(m => m.WriteStringAsync("Healthy!"), Times.Once());
            Assert.AreEqual(HttpStatusCode.OK, _mockResponseWrapper.Object.StatusCode);
        }

        [TestMethod]
        public async Task HealthCheckFunction_ReturnsNotHealthyResponseOnException()
        {
            // Arrange
            // Configures the response wrapper to complete write tasks, set initial status code to InternalServerError,
            // and the blob service client factory to throw an exception on blob service client creation.
            _mockResponseWrapper.Setup(m => m.WriteStringAsync(It.IsAny<string>())).Returns(Task.CompletedTask);
            _mockResponseWrapper.SetupProperty(m => m.StatusCode, HttpStatusCode.InternalServerError);
            _mockBlobServiceClientFactory.Setup(m => m.CreateBlobServiceClient(It.IsAny<string>()))
                .Throws(new RequestFailedException("Error"));

            // Act
            // Executes HealthCheckFunction with mocked dependencies to test its response to predefined conditions.
            var result = await HealthCheckFunction.Run(
                _mockHttpRequestWrapper.Object, // HttpRequestData is not directly used in the function
                _mockFunctionContext.Object,
                _mockBlobServiceClientFactory.Object,
                _mockEnvironmentVariableProvider.Object,
                _mockLogger.Object);

            // Assert
            // Verifies that the function returns InternalServerError, writes "Not Healthy!" once, and sets the response status to InternalServerError.
            Assert.AreEqual(HttpStatusCode.InternalServerError, result);
            _mockResponseWrapper.Verify(m => m.WriteStringAsync("Not Healthy!"), Times.Once());
            Assert.AreEqual(HttpStatusCode.InternalServerError, _mockResponseWrapper.Object.StatusCode);
        }

    }
}




