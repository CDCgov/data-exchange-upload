using BulkFileUploadFunctionApp.Model;
using Microsoft.Extensions.Logging;
using Azure.Messaging.ServiceBus;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using BulkFileUploadFunctionApp.Utils;
using System.Text.Json;
using Azure.Messaging.ServiceBus.Administration;
using Microsoft.VisualStudio.TestPlatform.CommunicationUtilities;
using System.Text.Unicode;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBulkUploadSvcBusClient
    {
        Task<HealthCheckResponse?> GetHealthCheck();
        Task<bool> CreateReport<TReport>(string uploadId, string destinationId, string eventType, string stageName, TReport payload);
    }

    public class BulkUploadSvcBusClient : IBulkUploadSvcBusClient
    {
        private readonly ILogger<BulkUploadSvcBusClient> _logger;
        private readonly ServiceBusClient _svcBusClient;
        private readonly string _serviceBusConnectionString;
        private readonly string _serviceBusQueueName;
        private readonly ServiceBusSender _svcBusSender;
        private readonly ServiceBusReceiver _svcBusReceiver;

        public BulkUploadSvcBusClient(string serviceBusConnectionString, string serviceBusQueueName, ILogger<BulkUploadSvcBusClient> logger)
        {
                _serviceBusConnectionString = serviceBusConnectionString;
                _serviceBusQueueName = serviceBusQueueName;
                _logger = logger;
                _svcBusClient = new ServiceBusClient(_serviceBusConnectionString);
                _svcBusSender = _svcBusClient.CreateSender(_serviceBusQueueName);
                _svcBusReceiver = _svcBusClient.CreateReceiver(_serviceBusQueueName);

        }

        public async Task<HealthCheckResponse?> GetHealthCheck()
        {
            try
            {
                try
                {
                   var responseBody = await DoesQueueExistAsync(_serviceBusQueueName);

                    // If the creation of the ServiceBusSender is successful, then the Service Bus is healthy
                    return new HealthCheckResponse()
                    {
                        Status = "UP"
                    };
                }
                catch (Exception ex)
                {
                    // If an exception is thrown, then the Service Bus is not healthy
                    _logger.LogError("Error when checking the health of the Service Bus.");
                    ExceptionUtils.LogErrorDetails(ex, _logger);
                    return new HealthCheckResponse()
                    {
                        Status = "DOWN"
                    };
                }

                
            }
            catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return new HealthCheckResponse()
                {
                    Status = "DOWN"
                };
            }
        }

        public async Task<bool> DoesQueueExistAsync(string queueName)
        {
            var serviceBusAdministrationClient = new ServiceBusAdministrationClient(_serviceBusConnectionString);
            return await serviceBusAdministrationClient.QueueExistsAsync(queueName);
        }

        public Task<bool> CreateReport<TReport>(string uploadId, string destinationId, string eventType, string stageName, TReport payload)
        {
            try
            {
                // build the report json
                var content = Encoding.UTF8.GetBytes(JsonSerializer.Serialize(payload));
                //var response = await _httpClient.PostAsync($"/api/report/json/uploadId/{uploadId}?destinationId={destinationId}&eventType={eventType}&stageName={stageName}", content);

                // add it to a BusMessage object 
                var svcBusMessage = new ServiceBusMessage(uploadId) { Subject=stageName, ContentType="application/json", Body = new BinaryData(content) };

                // send the message
                _svcBusSender.SendMessageAsync(svcBusMessage);

            }
            catch (Exception ex)
            {
                _logger.LogError("Error when calling PS API.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return Task.FromResult(false);
            }

           return Task.FromResult(true);
        }
    }
}
