using System.Net;
using Azure;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Services;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;


namespace BulkFileUploadFunctionApp
{

    public class HealthCheckFunction
    {
        const string TestContainerName = "dextesting-testevent1";
        private readonly IBlobServiceClientFactory _blobServiceClientFactory;
        private readonly IEnvironmentVariableProvider _environmentVariableProvider;
        private readonly IFunctionLogger<HealthCheckFunction> _logger;

        // Constructor
        public HealthCheckFunction(IBlobServiceClientFactory blobServiceClientFactory,
                               IEnvironmentVariableProvider environmentVariableProvider,
                               IFunctionLogger<HealthCheckFunction> logger)
        {
            _blobServiceClientFactory = blobServiceClientFactory;
            _environmentVariableProvider = environmentVariableProvider;
            _logger = logger;
        }

        [Function("HealthCheckFunction")]
        public async Task<HttpResponseData> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] HttpRequestData req,
            FunctionContext context)
        {
            _logger.LogInformation("HealthCheckFunction");

            //creating a response for a request and setting its status code to 200 (OK).
            var response = req.CreateResponse();
            response.StatusCode = HttpStatusCode.OK;

            try
            {
                //retrieve the values of these environment variables
                var storageAccountName = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_NAME");
                var storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable("DEX_AZURE_STORAGE_ACCOUNT_KEY");

                _logger.LogInformation("Container name-->" + TestContainerName);
                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";

                //establishing a connection to Azure Blob Storage and accessing a particular container by using the connection string.
                BlobServiceClient blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(TestContainerName);

                
                response.StatusCode = HttpStatusCode.OK;
                return response;
            }
            catch (RequestFailedException ex)
            {
                // Log error, respond with "Not Healthy!", and set response status to Internal Server Error (500)
                _logger.LogError(ex, "Error occurred while checking Blob storage container health.");
                response.StatusCode = HttpStatusCode.InternalServerError;
                return response;
            }
        }
    }

}
