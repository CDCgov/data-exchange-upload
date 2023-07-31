
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
        logger.LogInformation("Health check request received.");

        

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
             
         
       // Create a BlobServiceClient object from the connection string
        BlobServiceClient blobServiceClient = new BlobServiceClient(connectionString);

         // Get a reference to the specified container
        BlobContainerClient container = blobServiceClient.GetBlobContainerClient(containerName);

       

        // Check if the container exists
         if (await container.ExistsAsync())
         {
            BlobContainerProperties containerProperties=await container.GetPropertiesAsync();

            var lastModified = containerProperties.LastModified;
            var ETag = containerProperties.ETag;
            // Container exists, return a success message with container name and properties
            response.StatusCode = (HttpStatusCode)200;
            await response.WriteStringAsync($"Container '{container.Name}' exists. LastModified: {lastModified}, Etag: {ETag}");
        }
        else
        {
         // Container doesn't exist, return an error message
            response.StatusCode = (HttpStatusCode)404;
            await response.WriteStringAsync($"Container '{containerName}' does not exist.");
         }
          
         } catch (RequestFailedException ex)
        {
            // Handle any exceptions that might occur during the health check
            logger.LogError(ex, "Error occurred while checking Blob storage container health.");
            response.StatusCode = (HttpStatusCode)500;
            await response.WriteStringAsync("Error occurred while checking Blob storage container health.");
        }

            return response;
      }       

   }

}
