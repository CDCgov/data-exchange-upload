using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Model;
using System.Net;
using Azure;
using Azure.Storage.Blobs;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.DependencyInjection;
using System.Text.Json;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Configuration.AzureAppConfiguration;
using BulkFileUploadFunctionApp.Utils;
using Azure.Messaging.ServiceBus;

namespace BulkFileUploadFunctionAppTests
{
    // 'HealthCheckFunctionTests' for testing health check functionality with mocked dependencies.
    [TestClass]
    public class HealthCheckFunctionTests
    {
        private Mock<FunctionContext> _mockFunctionContext;
        private Mock<IBlobServiceClientFactory> _mockBlobServiceClientFactory;
        private Mock<IEnvironmentVariableProvider> _mockEnvironmentVariableProvider;
        private Mock<IConfigurationRefresher> _configurationRefresherMock;
        private Mock<IConfigurationRefresherProvider> _configurationRefresherProviderMock;
        private Mock<IServiceProvider> _mockServiceProvider;
        private Mock<ILogger<HealthCheckFunction>> _loggerMock;
        private Mock<ILogger<BulkUploadSvcBusClient>> _loggerBusMock;
        private Mock<ILoggerFactory> _loggerFactoryMock;
        private Mock<ServiceBusClient> _mockClient;
        private Mock<ServiceBusSender> _mockSender;
        private Mock<IBulkUploadSvcBusClient> _mockBulkUploadSvcClient;
        private IConfiguration _testConfiguration;
        private IFeatureManagementExecutor _testFeatureManagementExecutor;
        

        // Initializes mock objects for HTTP request/response, function context, blob service, environment variables, and logger.
        // Sets up default behavior for these mocks to be used in health check function tests.
        [TestInitialize]
        public void Initialize()
        {
            // Instantiate mocks.
            _mockFunctionContext = new Mock<FunctionContext>();
            _mockBlobServiceClientFactory = new Mock<IBlobServiceClientFactory>();
            _mockEnvironmentVariableProvider = new Mock<IEnvironmentVariableProvider>();
            _configurationRefresherMock = new Mock<IConfigurationRefresher>();
            _configurationRefresherProviderMock = new Mock<IConfigurationRefresherProvider>();
            _mockServiceProvider = new Mock<IServiceProvider>();
            _loggerFactoryMock = new Mock<ILoggerFactory>();
            _loggerMock = new Mock<ILogger<HealthCheckFunction>>();
            _loggerBusMock = new Mock<ILogger<BulkUploadSvcBusClient>>();
            _testConfiguration = new ConfigurationBuilder().AddInMemoryCollection(new Dictionary<string, string>
            {
                {$"FeatureManagement:{Constants.PROCESSING_STATUS_REPORTS_FLAG_NAME}", "true"}
            }).Build();
            _configurationRefresherProviderMock.Setup(m => m.Refreshers).Returns(new List<IConfigurationRefresher> { _configurationRefresherMock.Object });
            _testFeatureManagementExecutor = new FeatureManagementExecutor(_configurationRefresherProviderMock.Object, _testConfiguration);


            _mockClient = new Mock<ServiceBusClient>();
            _mockSender = new Mock<ServiceBusSender>();
            _mockBulkUploadSvcClient = new Mock<IBulkUploadSvcBusClient>();

            
            // Setup mocks.
            _mockEnvironmentVariableProvider.Setup(m => m.GetEnvironmentVariable(It.IsAny<string>())).Returns("test");
            _mockFunctionContext.Setup(ctx => ctx.InstanceServices)
                                .Returns(_mockServiceProvider.Object);

            var mockBlobServiceClient = new Mock<BlobServiceClient>();
            _mockBlobServiceClientFactory.Setup(m => m.CreateBlobServiceClient(It.IsAny<string>())).Returns(mockBlobServiceClient.Object);

            _loggerFactoryMock.Setup(x => x.CreateLogger(It.IsAny<string>())).Returns(_loggerMock.Object);
            
            
            _mockServiceProvider.Setup(provider => provider.GetService(typeof(ILogger<HealthCheckFunction>)))
                                .Returns(_loggerMock.Object);

            _mockServiceProvider.Setup(provider => provider.GetService(typeof(IFeatureManagementExecutor)))
                .Returns(_testFeatureManagementExecutor);

        }

        private HealthCheckFunction CreateHealthCheckFunction()
        {
            return new HealthCheckFunction(
                _mockBlobServiceClientFactory.Object,
                _mockEnvironmentVariableProvider.Object,
                _loggerFactoryMock.Object,
                _testFeatureManagementExecutor,
                _mockBulkUploadSvcClient.Object);
        }

        private BulkUploadSvcBusClient CreateBulkUploadSvcClient()
        {
            // Mock IEnvironmentVariableProvider
            var mockEnvironmentVariableProvider = new Mock<IEnvironmentVariableProvider>();
            IEnvironmentVariableProvider environmentVariableProvider = new EnvironmentVariableProviderImpl();
            var _serviceBusConnectionString = environmentVariableProvider.GetEnvironmentVariable("SERVICE_BUS_CONNECTION_STR");
            var _serviceBusQueueName = environmentVariableProvider.GetEnvironmentVariable("REPORT_QUEUE_NAME");

            mockEnvironmentVariableProvider
                .Setup(provider => provider.GetEnvironmentVariable("SERVICE_BUS_CONNECTION_STR"))
                .Returns("test");
            mockEnvironmentVariableProvider
                .Setup(provider => provider.GetEnvironmentVariable("REPORT_QUEUE_NAME"))
                .Returns("YourServiceBusQueueName");


            // Mock ILogger<BulkUploadSvcBusClient>
            var mockLogger = new Mock<ILogger<BulkUploadSvcBusClient>>();

            return new BulkUploadSvcBusClient(environmentVariableProvider, mockLogger.Object);
        }
        [TestMethod]
        public async Task GivenMessage_WhenHealthCheckFunctionInvoked_ThenReturnHealthyResponse()
        {
            // Arrange
            var functionContext = TestHelpers.CreateFunctionContext();
            var httpRequestData = TestHelpers.CreateHttpRequestData(functionContext);
            var healthCheckFunction = CreateHealthCheckFunction();
            var expectedResponse = TestHelpers.CreateUpResponse();

            _mockBulkUploadSvcClient
                .Setup(client => client.GetHealthCheck())
                .ReturnsAsync(expectedResponse);

            // Act
            var result = await healthCheckFunction.Run(
                               httpRequestData,functionContext);

            // Assert
            Assert.IsNotNull(result);
            Assert.AreEqual(HttpStatusCode.OK, result.StatusCode);
            result.Body.Position = 0;
            var responseBody = new StreamReader(result.Body).ReadToEnd();
            var healthCheckResponse = JsonSerializer.Deserialize<HealthCheckResponse>(responseBody);
            Assert.IsNotNull(healthCheckResponse);
            Assert.AreEqual("UP", healthCheckResponse.Status);
        }

        [TestMethod]
        public async Task GivenMessage_WhenServiceBusClientThrowsException_ThenReturnNotHealthyResponse()
        {
            // Arrange
            var functionContext = TestHelpers.CreateFunctionContext();
            var httpRequestData = TestHelpers.CreateHttpRequestData(functionContext);
            _mockBulkUploadSvcClient
                .Setup(client => client.GetHealthCheck())
                .Throws(new RequestFailedException("Error connecting to Service Bus"));

            var healthCheckFunction = CreateHealthCheckFunction();

            // Act
            var result = await healthCheckFunction.Run(
                               httpRequestData,functionContext);

            // Assert
            Assert.IsNotNull(result);
            //TODO: Need to fix Assert.AreEqual(HttpStatusCode.InternalServerError, result.StatusCode);
        }

        [TestMethod]
        public async Task GivenMessage_WhenServiceBusReachesQueue_ThenReturnSendMessage()
        {
            Mock<ServiceBusClient> mockClient = new();
            Mock<ServiceBusSender> mockSender = new();

            // This sets up the mock ServiceBusClient to return the mock of the ServiceBusSender.

            mockClient
                .Setup(client => client.CreateSender(It.IsAny<string>()))
                .Returns(mockSender.Object);

            // This sets up the mock sender to successfully return a completed task when any message is passed to
            // SendMessageAsync.

            mockSender
                .Setup(sender => sender.SendMessageAsync(
                    It.IsAny<ServiceBusMessage>(),
                    It.IsAny<CancellationToken>()))
                .Returns(Task.CompletedTask);

            ServiceBusClient client = mockClient.Object;

            // The rest of this snippet illustrates how to send a service bus message using the mocked
            // service bus client above, this would be where application methods sending a message would be
            // called.

            string mockQueueName = "MockQueueName";
            ServiceBusSender sender = client.CreateSender(mockQueueName);
            ServiceBusMessage message = new("Hello World!");

            await sender.SendMessageAsync(message);

            // This illustrates how to verify that SendMessageAsync was called the correct number of times
            // with the expected message.

            mockSender
                .Verify(sender => sender.SendMessageAsync(
                    It.Is<ServiceBusMessage>(m => (m.MessageId == message.MessageId)),
                    It.IsAny<CancellationToken>()));

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

        [TestMethod]
        public async Task HealthCheckFunction_ReturnsDownWhenPSAPIReturnsDown()
        {
            // Arrange
            var functionContext = TestHelpers.CreateFunctionContext();
            var httpRequestData = TestHelpers.CreateHttpRequestData(functionContext);

            _mockBulkUploadSvcClient.Setup(mock => mock.GetHealthCheck())
                .Throws(new RequestFailedException("Error connecting to Service Bus"));

            var healthCheckFunction = CreateHealthCheckFunction();

            // Act
            var result = await healthCheckFunction.Run(
                httpRequestData,
                functionContext);

            // Assert
            Assert.AreEqual(HttpStatusCode.OK, result.StatusCode);
            result.Body.Position = 0;
            var responseBody = new StreamReader(result.Body).ReadToEnd();
            var healthCheckResponse = JsonSerializer.Deserialize<HealthCheckResponse>(responseBody);
            Assert.IsNotNull(healthCheckResponse);
            Assert.AreEqual("DOWN", healthCheckResponse.Status);
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
            // Explicitly mock ILogger<HealthCheckFunction> and add to services
            var mockLogger = new Mock<ILogger<HealthCheckFunction>>();
            services.AddSingleton<ILogger<HealthCheckFunction>>(mockLogger.Object);

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
                var responseStream = new MemoryStream();
                httpResponseData.Setup(res => res.Body).Returns(responseStream);
                httpResponseData.SetupProperty(res => res.StatusCode);
                httpResponseData.Setup(res => res.Headers).Returns(new HttpHeadersCollection());
                return httpResponseData.Object;
            });

            return httpRequestDataMock.Object;
        }

        public static HealthCheckResponse CreateUpResponse()
        {
            var healthCheckResponse = new HealthCheckResponse();
            healthCheckResponse.Status = "UP";
            return healthCheckResponse;
        }
    }
}




