package processingstatus

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"net/url"
)

type PsSender struct {
	EndpointHealth    string
	EndpointHealthURI *url.URL
	ServiceBusClient  *azservicebus.Client
	ServiceBusQueue   string
} // .PsSender

func New(appConfig appconfig.AppConfig) (*PsSender, error) {

	endpointURI, err := url.ParseRequestURI(appConfig.ProcessingStatusHealthURI)
	if err != nil {
		return nil, err
	} // .if

	sbClient, err := GetServiceBusClient(appConfig)
	if err != nil {
		return nil, err
	} // .if

	return &PsSender{
		EndpointHealth:    appConfig.ProcessingStatusHealthURI,
		EndpointHealthURI: endpointURI,
		ServiceBusClient:  sbClient,
		ServiceBusQueue:   appConfig.ProcessingStatusServiceBusQueue,
	}, nil // .return

} // .New

func GetServiceBusClient(appConfig appconfig.AppConfig) (*azservicebus.Client, error) {

	namespace := appConfig.ProcessingStatusServiceBusNamespace

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	} // .if

	client, err := azservicebus.NewClient(namespace, cred, nil)
	if err != nil {
		return nil, err
	} // .if

	//all good
	return client, nil
} // .GetServiceBusClient
