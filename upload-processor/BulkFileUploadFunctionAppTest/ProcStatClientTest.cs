using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Model;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionAppTest.utils;
using Microsoft.Extensions.Logging;
using Newtonsoft.Json;
using System.Net;

namespace BulkFileUploadFunctionAppTest
{

    [TestClass]
    public class ProcStatClientTest
    {
        private Mock<ILogger<ProcStatClient>> _mockLogger;

        [TestInitialize]
        public void Initialize()
        {
            _mockLogger = new Mock<ILogger<ProcStatClient>>();
        }

        [TestMethod]
        public async Task GivenSuccessfulResponse_WhenGetHealthCheck_ThenReturnsUp()
        {
            // Arrange.
            var responseBody = new HealthCheckResponse
            {
                Status = "UP"
            };
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(JsonConvert.SerializeObject(responseBody))
            };
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetHealthCheck();

            Assert.AreEqual("UP", response.Status);
        }

        [TestMethod]
        public async Task GivenNullResponseContent_WhenGetHealthCheck_ThenReturnsDown()
        {
            // Arrange.
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK);
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetHealthCheck();

            Assert.AreEqual("DOWN", response.Status);
        }

        [TestMethod]
        public async Task GivenException_WhenGetHealthCheck_ThenReturnsDown()
        {
            // Arrange.
            var httpClient = new HttpClient(new MockedHttpMessageHandler(() => throw new HttpRequestException()))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetHealthCheck();

            Assert.AreEqual("DOWN", response.Status);
        }

        [TestMethod]
        public async Task GivenSuccessfulResponse_WhenCreateReport_ThenReturnsTrue()
        {
            // Arrange.
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK);
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);
            var testReport = new CopyReport("test source Url", "test dest Url", "success");

            // Act.
            var response = await client.CreateReport("test upload ID", "test dest ID", "test event type", "test stage name", testReport);

            Assert.IsTrue(response);
        }

        [TestMethod]
        public async Task GivenFailErrorCode_WhenCreateReport_ThenReturnsFalse()
        {
            // Arrange.
            var apiResponse = new HttpResponseMessage(HttpStatusCode.InternalServerError);
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);
            var testReport = new CopyReport("test source Url", "test dest Url", "success");

            // Act.
            var response = await client.CreateReport("test upload ID", "test dest ID", "test event type", "test stage name", testReport);

            Assert.IsFalse(response);
        }

        [TestMethod]
        public async Task GivenSuccessfulResponse_WhenGetTraceByUploadId_ThenReturnsTrace()
        {
            // Arrange.
            var responseBody = new Trace
            {
                TraceId = "1234"
            };
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(JsonConvert.SerializeObject(responseBody))
            };
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetTraceByUploadId("5678");

            Assert.AreEqual("1234", response?.TraceId);
        }

        [TestMethod]
        public async Task GivenNullResponseContent_WhenGetTraceByUploadId_ThenReturnsNull()
        {
            // Arrange.
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK);
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetTraceByUploadId("5678");

            Assert.IsNull(response);
        }

        [TestMethod]
        public async Task GivenException_WhenGetTraceByUploadId_ThenReturnsNull()
        {
            // Arrange.
            var httpClient = new HttpClient(new MockedHttpMessageHandler(() => throw new HttpRequestException()))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.GetTraceByUploadId("5678");

            Assert.IsNull(response);
        }

        [TestMethod]
        public async Task GivenSuccessfulResponse_WhenStartSpanForTrace_ThenReturnsSpan()
        {
            // Arrange.
            var responseBody = new Span
            {
                SpanId = "1234"
            };
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(JsonConvert.SerializeObject(responseBody))
            };
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.StartSpanForTrace("abcd", "defg", "test stage");

            Assert.AreEqual("1234", response?.SpanId);
        }

        [TestMethod]
        public async Task GivenNullResponseContent_WhenStartSpanForTrace_ThenReturnsNull()
        {
            // Arrange.
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK);
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.StartSpanForTrace("abcd", "defg", "test stage");

            Assert.IsNull(response);
        }

        [TestMethod]
        public async Task GivenException_WhenStartSpanForTrace_ThenReturnsNull()
        {
            // Arrange.
            var httpClient = new HttpClient(new MockedHttpMessageHandler(() => throw new HttpRequestException()))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.StartSpanForTrace("abcd", "defg", "test stage");

            Assert.IsNull(response);
        }

        [TestMethod]
        public async Task GivenSuccessfulResponse_WhenStopSpanForTrace_ThenReturnsId()
        {
            // Arrange.
            var responseBody = "1234";
            var apiResponse = new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(responseBody)
            };
            var httpClient = new HttpClient(new MockedHttpMessageHandler(apiResponse))
            {
                BaseAddress = new Uri("http://localhost")
            };
            var client = new ProcStatClient(httpClient, _mockLogger.Object);

            // Act.
            var response = await client.StopSpanForTrace("abcd", "defg");

            Assert.AreEqual("1234", response);
        }
    }
}
