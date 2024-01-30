
using System.Net;
using System.Threading.Tasks;
using Azure;
using Azure.Storage.Blobs;
using BulkFileUploadFunctionApp.Services;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp
{   
    

    public  class HealthCheckFunction
    {

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
        public  async Task<HttpStatusCode> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] IHttpRequestDataWrapper requestWrapper,
            FunctionContext context)
        {
            _logger.LogInformation("HealthCheckFunction");

            //creating a response for a request and setting its status code to 200 (OK).
            var responseWrapper = requestWrapper.CreateResponse();
            responseWrapper.StatusCode = HttpStatusCode.OK;

            try
            {   //names of the environment variables.
                string dexAzureStorageAccountname = "DEX_AZURE_STORAGE_ACCOUNT_NAME";
                string dexAzureStorageAccountKey = "DEX_AZURE_STORAGE_ACCOUNT_KEY";
                string cName = "ndlp-influenzavaccination";
                
                //retrieve the values of these environment variables
                var storageAccountName = _environmentVariableProvider.GetEnvironmentVariable(dexAzureStorageAccountname);
                var storageAccountKey = _environmentVariableProvider.GetEnvironmentVariable(dexAzureStorageAccountKey);

                _logger.LogInformation("Container name-->" + cName);
                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                
                //establishing a connection to Azure Blob Storage and accessing a particular container by using the connection string.
                BlobServiceClient blobServiceClient = _blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(cName);
                
                // Write "Healthy!" to the response and return HTTP status 200 (OK) without blocking
                await responseWrapper.WriteStringAsync("Healthy!");
                return HttpStatusCode.OK;
            }
            catch (RequestFailedException ex)
            {
                // Log error, respond with "Not Healthy!", and set response status to Internal Server Error (500)
                _logger.LogError(ex, "Error occurred while checking Blob storage container health.");
                await responseWrapper.WriteStringAsync("Not Healthy!");
                responseWrapper.StatusCode = HttpStatusCode.InternalServerError;
                return HttpStatusCode.InternalServerError;
            }
        }
    }
   
}
