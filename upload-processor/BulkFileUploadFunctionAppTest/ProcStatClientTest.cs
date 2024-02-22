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
        public async Task GivenSuccessfulResponse_WhenGetHealthCheck_ThenReturnsOk()
        {
            // Arrange.
            var request = new HttpRequestMessage(HttpMethod.Get, "/api/health");
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

    }
}
