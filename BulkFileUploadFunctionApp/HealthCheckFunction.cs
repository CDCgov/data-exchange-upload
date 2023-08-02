
using System.Net;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;
using Azure.Storage.Blobs;
using Azure;
using Azure.Storage.Blobs.Models;

namespace BulkFileUploadFunctionApp
{
    public static class HealthCheckFunction
   {   
    
     [Function("HealthCheckFunction")]
    public static async Task<HttpResponseData> Run(
        [HttpTrigger(AuthorizationLevel.Anonymous, "get", Route = "health")] HttpRequestData req,        
        FunctionContext context)
    {
        var logger = context.GetLogger("HealthCheckFunction");
       
          string dEX_AZURE_STORAGE_ACCOUNT_NAME = "DEX_AZURE_STORAGE_ACCOUNT_NAME";
          string dexAzureStorageAccountKey = "DEX_AZURE_STORAGE_ACCOUNT_KEY";
          string cName = "ndlp-influenzavaccination";
          var response = req.CreateResponse();
      try
        {

        String? _dexAzureStorageAccountName = Environment.GetEnvironmentVariable(dEX_AZURE_STORAGE_ACCOUNT_NAME);
        String? _dexAzureStorageAccountKey = Environment.GetEnvironmentVariable(dexAzureStorageAccountKey);
        String? containerName = Environment.GetEnvironmentVariable(cName);    

        var connectionString = $"DefaultEndpointsProtocol=https;AccountName={_dexAzureStorageAccountName};AccountKey={_dexAzureStorageAccountKey};EndpointSuffix=core.windows.net";         
                
        BlobServiceClient blobServiceClient = new BlobServiceClient(connectionString);         
        BlobContainerClient container = blobServiceClient.GetBlobContainerClient(containerName);
        response.StatusCode = (HttpStatusCode)200;
        await response.WriteStringAsync("Healthy!");   
    
         } catch (RequestFailedException ex)
        {
            
            logger.LogError(ex, "Error occurred while checking Blob storage container health.");
            response.StatusCode = (HttpStatusCode)500;
            await response.WriteStringAsync("Not Healthy!");
        }

            return response;
      }       

   }

}
