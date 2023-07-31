
using System.Net;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Azure.Functions.Worker.Http;
using Microsoft.Extensions.Logging;
using Azure.Messaging.EventHubs;
using Azure.Messaging.EventHubs.Producer;
using System.Text;

namespace BulkFileUploadFunctionApp
{
    public static class HealthCheckFunction
   {   
    
    private const string EventHubConnectionString = "AzureEventHubConnectionString";
    private const string EventHubName = "AzureEventHubName";
    [Function("HealthCheckFunction")]
    public static async Task<HttpResponseData> Run(
        [HttpTrigger(AuthorizationLevel.Function, "get", Route = "health")] HttpRequestData req,        
        FunctionContext context)
    {
        var logger = context.GetLogger("HealthCheckFunction");
        logger.LogInformation("Health check request received.");

        String? eventHubConnectionString = Environment.GetEnvironmentVariable(EventHubConnectionString);
        String? eventHubName = Environment.GetEnvironmentVariable(EventHubName);
         if(String.IsNullOrEmpty(eventHubConnectionString)){
            return CreateErrorResponse(req, "Event Hub connection String not found in the Azure portal configuration.");
         }

         if(String.IsNullOrEmpty(eventHubName)){
            return CreateErrorResponse(req, "Event Hub Name not found in the Azure portal configuration.");
         }
                

            // Check the health of the Event Hub connection
            bool isEventHubHealthy = await CheckEventHubHealthAsync();
            if (!isEventHubHealthy)
            {
                return CreateErrorResponse(req, "Event Hub connection is not healthy.");
            }        
            

            // If all checks pass, return a 200 OK response with a success message
            var response = req.CreateResponse(HttpStatusCode.OK);
            response.Headers.Add("Content-Type", "text/plain; charset=utf-8");
            response.WriteString("Health check passed!");

            return response;
      }
    private static async Task<bool> CheckEventHubHealthAsync()
        {
            var producerClient = new EventHubProducerClient(EventHubConnectionString, EventHubName);           

            try
            {
                // Send a test event to the Event Hub to check if it can send events successfully.
                var eventData = new EventData(Encoding.UTF8.GetBytes("Test Event"));
                await producerClient.SendAsync(new List<EventData> {eventData});
                

                return true;
            }
            catch (Exception ex)
            {                   
               
                return false;
            }
            finally
            {
                await producerClient.CloseAsync();
            }
        }

       

        private static HttpResponseData CreateErrorResponse(HttpRequestData request, string errorMessage)
        {
            var response = request.CreateResponse(HttpStatusCode.ServiceUnavailable);
            response.Headers.Add("Content-Type", "text/plain; charset=utf-8");
            response.WriteString(errorMessage);
            return response;
        }
    }
}
