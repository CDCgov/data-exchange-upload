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
using System.Runtime.Serialization;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IBulkUploadSvcBusClient
    {
        Task<HealthCheckResponse?> GetHealthCheck();
        Task PublishReport(string uploadId, string destinationId, string eventType, string stageName, Report payload);

    }

    public class BulkUploadSvcBusClient : IBulkUploadSvcBusClient
    {
        private readonly ILogger<BulkUploadSvcBusClient> _logger;
        private readonly ServiceBusClient _svcBusClient;
        private readonly string _serviceBusConnectionString;
        private readonly string _serviceBusQueueName;
        private readonly ServiceBusSender _svcBusSender;


        public BulkUploadSvcBusClient(IEnvironmentVariableProvider environmentVariableProvider, ILogger<BulkUploadSvcBusClient> logger)
        {
            _serviceBusConnectionString = environmentVariableProvider.GetEnvironmentVariable("SERVICE_BUS_CONNECTION_STR");
            _serviceBusQueueName = environmentVariableProvider.GetEnvironmentVariable("REPORT_QUEUE_NAME");
            _logger = logger;
            _svcBusClient = new ServiceBusClient(_serviceBusConnectionString);
            _svcBusSender = _svcBusClient.CreateSender(_serviceBusQueueName);
        }

        public async Task<HealthCheckResponse?> GetHealthCheck()
        {
            try
            {
                var responseBody = await DoesQueueExistAsync();

                // If the queue exists, then the Service Bus is healthy
                if (responseBody)
                {
                    return new HealthCheckResponse()
                    {
                        Status = "UP"
                    };
                }
                else
                {
                    _logger.LogWarning("The queue does not exist in the Service Bus.");
                    return new HealthCheckResponse()
                    {
                        Status = "DOWN"
                    };
                }
            }
            catch (ServiceBusException sbe)
            {
                _logger.LogError($"Error when checking the health of the Service Bus: {sbe.Reason.ToString()}");
                ExceptionUtils.LogErrorDetails(sbe, _logger);
                // Delay before retrying (you can implement exponential backoff here)
                await Task.Delay(1000); // 1 second delay before retrying                
                return new HealthCheckResponse()
                {
                    Status = "DOWN"
                };

            }
            catch (Exception ex)
            {
                // If an exception is thrown, then the Service Bus is not healthy
                _logger.LogError("Error when checking the health of the Service Bus.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return new HealthCheckResponse()
                {
                    Status = "UNKNOWN"
                };
            }
        }


        public async Task<bool> DoesQueueExistAsync()
        {
            var serviceBusAdministrationClient = new ServiceBusAdministrationClient(_serviceBusConnectionString);
            return await serviceBusAdministrationClient.QueueExistsAsync(_serviceBusQueueName);
        }

        public async Task PublishReport(string uploadId, string destinationId, string eventType, string stageName, Report payload)
        {
            const int maxRetryAttempts = 3;
            int currentRetryAttempt = 0;


            while (currentRetryAttempt < maxRetryAttempts)
            {
                try
                {
                    // build the report json
                    var payloadString = JsonSerializer.Serialize(payload);
                    var content = Encoding.UTF8.GetBytes(payloadString);
                    _logger.LogInformation($"Payload for Service Bus Report Message : {payloadString}");

                    // add it to a BusMessage object 
                    var svcBusMessage = new ServiceBusMessage(uploadId) { Subject = stageName, ContentType = "application/json", Body = new BinaryData(content) };

                    // send the message
                    await _svcBusSender.SendMessageAsync(svcBusMessage);

                    // Message sent successfully, break out of the retry loop
                    break;
                }
                catch (SerializationException se)
                {
                    _logger.LogError($"Serialization error: Failed to send success report to service bus: {se.Message}");
                    ExceptionUtils.LogErrorDetails(se, _logger);
                    throw;
                }
                catch (ServiceBusException sbe)
                {
                    // Increment the retry attempt
                    currentRetryAttempt++;

                    if (currentRetryAttempt >= maxRetryAttempts)
                    {
                        _logger.LogError($"After 3 failed attempts, the system failed to send success report to service bus: {sbe.Reason.ToString()}");
                        ExceptionUtils.LogErrorDetails(sbe, _logger);
                        // Max retry attempts reached, throw the exception
                        throw;
                    }
                }
                catch (Exception ex)
                {
                    _logger.LogError("Error when calling Service Bus.");
                    ExceptionUtils.LogErrorDetails(ex, _logger);
                    throw;
                }
            }
        }
    }
}
