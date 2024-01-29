
using System.Net;
using System.Threading.Tasks;
using Azure;
using Azure.Storage.Blobs;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp
{   
    // 'IBlobServiceClientFactory' for creating BlobServiceClient instances.    
    public interface IBlobServiceClientFactory
    {
        BlobServiceClient CreateBlobServiceClient(string connectionString);
    }
  
    // 'IEnvironmentVariableProvider' for accessing environment variables.
    public interface IEnvironmentVariableProvider
    {
        string GetEnvironmentVariable(string name);
    }

    public static class HealthCheckFunction
    {
        [Function("HealthCheckFunction")]

        //'Run' method responds to HTTP GET requests at the 'health' route, performing health checks and logging.
        //Dependencies like blob service client factory, environment variable provider, and logger are injected.
        public static async Task<HttpStatusCode> Run(
            [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] IHttpRequestDataWrapper requestWrapper,
            FunctionContext context,
            IBlobServiceClientFactory blobServiceClientFactory,
            IEnvironmentVariableProvider environmentVariableProvider,
            IFunctionLogger logger)
        {
            logger.LogInformation("HealthCheckFunction");

            //creating a response for a request and setting its status code to 200 (OK).
            var responseWrapper = requestWrapper.CreateResponse();
            responseWrapper.StatusCode = HttpStatusCode.OK;

            try
            {   //names of the environment variables.
                string dEX_AZURE_STORAGE_ACCOUNT_NAME = "DEX_AZURE_STORAGE_ACCOUNT_NAME";
                string dexAzureStorageAccountKey = "DEX_AZURE_STORAGE_ACCOUNT_KEY";
                string cName = "ndlp-influenzavaccination";
                
                //retrieve the values of these environment variables
                var storageAccountName = environmentVariableProvider.GetEnvironmentVariable(dEX_AZURE_STORAGE_ACCOUNT_NAME);
                var storageAccountKey = environmentVariableProvider.GetEnvironmentVariable(dexAzureStorageAccountKey);

                logger.LogInformation("Container name-->" + cName);
                var connectionString = $"DefaultEndpointsProtocol=https;AccountName={storageAccountName};AccountKey={storageAccountKey};EndpointSuffix=core.windows.net";
                
                //establishing a connection to Azure Blob Storage and accessing a particular container by using the connection string.
                BlobServiceClient blobServiceClient = blobServiceClientFactory.CreateBlobServiceClient(connectionString);
                BlobContainerClient container = blobServiceClient.GetBlobContainerClient(cName);
                
                // Write "Healthy!" to the response and return HTTP status 200 (OK) without blocking
                await responseWrapper.WriteStringAsync("Healthy!");
                return HttpStatusCode.OK;
            }
            catch (RequestFailedException ex)
            {
                // Log error, respond with "Not Healthy!", and set response status to Internal Server Error (500)
                logger.LogError(ex, "Error occurred while checking Blob storage container health.");
                await responseWrapper.WriteStringAsync("Not Healthy!");
                responseWrapper.StatusCode = HttpStatusCode.InternalServerError;
                return HttpStatusCode.InternalServerError;
            }
        }
    }
    //Purpose of the interface: Defines a way to handle HTTP requests and prepare responses.
    public interface IHttpRequestDataWrapper
{
    IHttpResponseDataWrapper CreateResponse();   
}

public class HttpRequestDataWrapper : IHttpRequestDataWrapper
{
    private readonly HttpRequestData _request;

    public HttpRequestDataWrapper(HttpRequestData request)
    {
        _request = request;
    }

    public IHttpResponseDataWrapper CreateResponse()
    {
        var response = _request.CreateResponse();
        return new HttpResponseDataWrapper(response);
    }

}
  // Interface and implementation for managing HTTP response content writing and status code.
public interface IHttpResponseDataWrapper
{
    Task WriteStringAsync(string responseContent);
    HttpStatusCode StatusCode { get; set; }
}

public class HttpResponseDataWrapper : IHttpResponseDataWrapper
{
    private readonly HttpResponseData _response;

    public HttpResponseDataWrapper(HttpResponseData response)
    {
        _response = response;
    }

    public async Task WriteStringAsync(string responseContent)
    {
        await _response.WriteStringAsync(responseContent);
    }

    public HttpStatusCode StatusCode
    {
        get => (HttpStatusCode)_response.StatusCode;
        set => _response.StatusCode = (HttpStatusCode)value;
    }

}

  // Interface 'IFunctionLogger' defines logging capabilities for information and errors. 
public interface IFunctionLogger
{
    void LogInformation(string message);
    void LogError(string message);
    void LogError(Exception ex, string message);
}

//'FunctionLogger' class implements this interface, utilizing an 'ILogger' for actual logging operations.
public class FunctionLogger : IFunctionLogger
{
    private readonly ILogger _logger;

    public FunctionLogger(ILogger logger)
    {
        _logger = logger;
    }

    public void LogInformation(string message)
    {
        _logger.LogInformation(message);
    }

    public void LogError(string message)
    {
        _logger.LogError(message);
    }

    public void LogError(Exception ex, string message)
    {
        _logger.LogError(ex, message);
    }

}



}
