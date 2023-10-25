
using System.Net;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp;
using Castle.Core.Logging;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Moq;
namespace BulkFileUploadFunctionAppTest;

[TestClass]
public class HealthCheckFunctionTests
{
    [TestMethod]
    public async Task HealthCheckFunction_ReturnsHealthyResponse()
    {
        //Arrange
        var request = new Mock<HttpRequestData>();
        var response = request.Object.CreateResponse();
        var context = new Mock<FunctionContext>();
        var logger = new Mock<ILogger>();

        context.Setup(c => c.GetLogger("HealthCheckFunction")).Returns((Microsoft.Extensions.Logging.ILogger)logger.Object);

        // Set up your environment variables
        Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", "DEX_AZURE_STORAGE_ACCOUNT_NAME");
        Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", "DEX_AZURE_STORAGE_ACCOUNT_KEY");
        // Mock BlobServiceClient and BlobContainerClient
        var blobServiceClientMock = new Mock<BlobServiceClient>();
        var blobContainerClientMock = new Mock<BlobContainerClient>();
        blobServiceClientMock.Setup(x => x.GetBlobContainerClient(It.IsAny<string>())).Returns(blobContainerClientMock.Object);
      
        var result = await HealthCheckFunction.Run(request.Object, context.Object);

        //Assert
       Assert.AreEqual(HttpStatusCode.OK, result);
       

    }

    [TestMethod]
    public async Task HealthCheckFunction_ReturnsUnHealthyResponse()
    {
        //Arrange
        var request = new Mock<HttpRequestData>();
        var response = request.Object.CreateResponse();
        var context = new Mock<FunctionContext>();
        var logger = new Mock<ILogger>();

        context.Setup(c => c.GetLogger("HealthCheckFunction")).Throws(new Exception("Log error"));

        // Set up your environment variables
        Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME", "DEX_AZURE_STORAGE_ACCOUNT_NAME");
        Environment.SetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY", "DEX_AZURE_STORAGE_ACCOUNT_KEY");
        // Mock BlobServiceClient and BlobContainerClient
        var blobServiceClientMock = new Mock<BlobServiceClient>();
        var blobContainerClientMock = new Mock<BlobContainerClient>();
        blobServiceClientMock.Setup(x => x.GetBlobContainerClient(It.IsAny<string>())).Returns(blobContainerClientMock.Object);
      
        var result = await HealthCheckFunction.Run(request.Object, context.Object);

        //Assert
       Assert.AreEqual(HttpStatusCode.InternalServerError, result);
       

    }
}



