using Microsoft.VisualStudio.TestTools.UnitTesting;
using Moq;
using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Services;
using System.Threading.Tasks;
using System.Net;
using Azure;
using Azure.Storage.Blobs;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using System.Text;
using Microsoft.Extensions.DependencyInjection;

namespace BulkFileUploadFunctionAppTests
{
    // 'HealthCheckFunctionTests' for testing health check functionality with mocked dependencies.
    [TestClass]
    public class HealthCheckFunctionTests
    {
        private Mock<FunctionContext> _mockFunctionContext;
        private Mock<IBlobServiceClientFactory> _mockBlobServiceClientFactory;
        private Mock<IEnvironmentVariableProvider> _mockEnvironmentVariableProvider;
        private Mock<IServiceProvider> _mockServiceProvider;
        private Mock<IFunctionLogger<HealthCheckFunction>> _mockLogger;


        // Initializes mock objects for HTTP request/response, function context, blob service, environment variables, and logger.
        // Sets up default behavior for these mocks to be used in health check function tests.
        [TestInitialize]
        public void Initialize()
        {
            _mockFunctionContext = new Mock<FunctionContext>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            _mockEnvironmentVariableProvider = new Mock<IEnvironmentVariableProvider>();
            _mockLogger = new Mock<IFunctionLogger<HealthCheckFunction>>();

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

        private HealthCheckFunction CreateHealthCheckFunction()
        {
            return new HealthCheckFunction(
                _mockBlobServiceClientFactory.Object,
                _mockEnvironmentVariableProvider.Object,
                _mockLogger.Object);
        }

        [TestMethod]
        public async Task HealthCheckFunction_ReturnsHealthyResponse()
        {
            // Arrange
            // setting up a mock response wrapper to simulate the behavior of the actual response object used in the service.
            var functionContext = TestHelpers.CreateFunctionContext();
            var httpRequestData = TestHelpers.CreateHttpRequestData(functionContext);

            var healthCheckFunction = CreateHealthCheckFunction();

            // Act
            // Executes the HealthCheckFunction with mocked dependencies to test its behavior.
            var result = await healthCheckFunction.Run(
                httpRequestData,
                functionContext);

            // Check response is not null, status code is OK, and 'Healthy!' was written once.
            Assert.IsNotNull(result);
            Assert.AreEqual(HttpStatusCode.OK, result.StatusCode);
           
        }

        [TestMethod]
        public async Task HealthCheckFunction_ReturnsNotHealthyResponseOnException()
        {
            // Arrange            
            var functionContext = TestHelpers.CreateFunctionContext();
            var httpRequestData = TestHelpers.CreateHttpRequestData(functionContext);
            _mockBlobServiceClientFactory.Setup(m => m.CreateBlobServiceClient(It.IsAny<string>()))
                .Throws(new RequestFailedException("Error"));


            var healthCheckFunction = CreateHealthCheckFunction();
            // Act
            // Executes HealthCheckFunction with mocked dependencies to test its response to predefined conditions.
            var result = await healthCheckFunction.Run(
                httpRequestData,
                functionContext);

            // Assert
            Assert.IsNotNull(result);
            Assert.AreEqual(HttpStatusCode.InternalServerError, result.StatusCode);

        }

    }

    // Defines a static class that contains helper methods for creating mock instances
    // of FunctionContext and HttpRequestData for use in unit testing Azure Functions.
    public static class TestHelpers
    {
        // Creates and returns a mock FunctionContext with a configured service provider.
        // This allows for testing functions that depend on services registered in the FunctionContext.
        public static FunctionContext CreateFunctionContext()
        {
            var services = new ServiceCollection();
            services.AddLogging(builder => builder.AddConsole());

            var serviceProvider = services.BuildServiceProvider();

            // Creates a mock FunctionContext.
            var functionContext = new Mock<FunctionContext>();
            // Sets up the InstanceServices property to return the built service provider,
            // allowing for dependency injection within the test environment.
            functionContext.Setup(ctx => ctx.InstanceServices).Returns(serviceProvider);
            return functionContext.Object;
        }

        // Creates and returns a mock HttpRequestData object for use in testing HTTP-triggered functions.
        public static HttpRequestData CreateHttpRequestData(FunctionContext functionContext)
        {
            // Creates a MemoryStream to represent the HTTP request body.
            var memoryStream = new MemoryStream();

            // Retrieves an ILoggerFactory from the function context's service provider,
            // allowing for logging within the mock HttpRequestData.
            var loggerFactory = functionContext.InstanceServices.GetService<ILoggerFactory>();
            var logger = loggerFactory.CreateLogger("Test");

            // Creates a mock HttpRequestData object, passing in the mock function context.
            var httpRequestDataMock = new Mock<HttpRequestData>(functionContext);

            // Sets up the Body property to return the previously created MemoryStream,
            // simulating an HTTP request body.
            httpRequestDataMock.Setup(req => req.Body).Returns(memoryStream);

            // Sets up the CreateResponse method to return a mock HttpResponseData object,
            // allowing for verification of response creation in tests.
            httpRequestDataMock.Setup(req => req.CreateResponse()).Returns(() =>
            {
                // Creates a mock HttpResponseData object.
                var httpResponseData = new Mock<HttpResponseData>(functionContext);
                // Sets up properties to simulate a real HTTP response.
                httpResponseData.SetupProperty(res => res.Body);
                httpResponseData.SetupProperty(res => res.StatusCode);
                httpResponseData.Setup(res => res.Headers).Returns(new HttpHeadersCollection());
                return httpResponseData.Object;
            });

            return httpRequestDataMock.Object;
        }
    }
}




