using BulkFileUploadFunctionApp.Model;
using Microsoft.Extensions.Logging;
using Azure.Messaging.ServiceBus;
using System.Text;
using BulkFileUploadFunctionApp.Utils;
using System.Text.Json;
using Azure.Messaging.ServiceBus.Administration;
using System.Runtime.Serialization;


namespace BulkFileUploadFunctionApp.Services
{
    public interface IBulkUploadSvcBusClient
    {
        Task<HealthCheckResponse?> GetHealthCheck();
        Task PublishReport(Report payload);

    }

    public class BulkUploadSvcBusClient : IBulkUploadSvcBusClient
    {
        private readonly ILogger<BulkUploadSvcBusClient> _logger;
        private readonly ServiceBusClient _svcBusClient;
        private readonly ServiceBusAdministrationClient _serviceBusAdministrationClient;
        private readonly string _serviceBusConnectionString;
        private readonly string _serviceBusQueueName;
        private readonly ServiceBusSender _svcBusSender;


        public BulkUploadSvcBusClient(IEnvironmentVariableProvider environmentVariableProvider, ILogger<BulkUploadSvcBusClient> logger)
        {
            _serviceBusConnectionString = environmentVariableProvider.GetEnvironmentVariable("SERVICE_BUS_CONNECTION_STR");
            _serviceBusQueueName = environmentVariableProvider.GetEnvironmentVariable("REPORT_QUEUE_NAME");
            _logger = logger;
            _svcBusClient = new ServiceBusClient(_serviceBusConnectionString, new ServiceBusClientOptions
            {
                TransportType = ServiceBusTransportType.AmqpWebSockets,
                RetryOptions = new ServiceBusRetryOptions
                {
                    TryTimeout = TimeSpan.FromSeconds(60),
                    MaxRetries = 3,
                    Delay = TimeSpan.FromSeconds(1)
                }
            });

            _svcBusSender = _svcBusClient.CreateSender(_serviceBusQueueName);
            _serviceBusAdministrationClient = new ServiceBusAdministrationClient(_serviceBusConnectionString);
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
            try
            {
                _logger.LogInformation("Checking if queue {0} exists", _serviceBusQueueName);
                

                if (await _serviceBusAdministrationClient.QueueExistsAsync(_serviceBusQueueName).ConfigureAwait(false))
                {
                    return true;
                }
                else
                {
                    return false;
                }
            }
            catch(UnauthorizedAccessException uae) {                 
                _logger.LogError($"Unauthorized access error: {uae.Message}.Queue '{_serviceBusQueueName}' could not be checked / created, "+
                    "likely due to missing the 'Manage' permission. \r\n You must either grant the 'Manage' permission, or set ServiceBusQueueOptions.CheckAndCreateQueues to false");
                ExceptionUtils.LogErrorDetails(uae, _logger);
                return false;
                       }
            catch (ServiceBusException sbe)
            {
                _logger.LogError($"Error when checking if the queue exists in the Service Bus: {sbe.Reason.ToString()}");
                ExceptionUtils.LogErrorDetails(sbe, _logger);
                return false;
            }   
            catch (Exception ex)
            {
                _logger.LogError("Error when checking if the queue exists in the Service Bus.");
                ExceptionUtils.LogErrorDetails(ex, _logger);
                return false;
            }

        }

        public async Task PublishReport(Report payload)
        {
                try
                {
                    // build the report json
                    var payloadString = JsonSerializer.Serialize(payload);
                    var content = Encoding.UTF8.GetBytes(payloadString);
                    _logger.LogInformation($"Payload for Service Bus Report Message : {payloadString}");

                    // add it to a BusMessage object 
                    var svcBusMessage = new ServiceBusMessage(new BinaryData(content));

                    // send the message
                    await _svcBusSender.SendMessageAsync(svcBusMessage);

                }
                catch (SerializationException se)
                {
                    _logger.LogError($"Serialization error: Failed to send report to service bus: {se.Message}");
                    ExceptionUtils.LogErrorDetails(se, _logger);
                }
                catch (ServiceBusException sbe)
                {
                    _logger.LogError($"After 3 failed attempts, the system failed to send report to service bus: {sbe.Reason.ToString()}");
                    ExceptionUtils.LogErrorDetails(sbe, _logger);
                    // Max retry attempts reached, throw the exception
                }
                catch (Exception ex)
                {
                    _logger.LogError("Error when calling Service Bus.");
                    ExceptionUtils.LogErrorDetails(ex, _logger);
                }
        }
    }
}