package processingstatus

import (
	"net/url"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
)

type PsSender struct {
	EndpointHealth    string
	EndpointHealthURI *url.URL
} // .PsSender

func New(appConfig appconfig.AppConfig) (*PsSender, error) {

	endpointURI, err := url.ParseRequestURI(appConfig.ProcessingStatusHealthURI)
	if err != nil {
		return nil, err
	} // .if

	return &PsSender{
		EndpointHealth:    appConfig.ProcessingStatusHealthURI,
		EndpointHealthURI: endpointURI,
	}, nil // .return

} // .New

func SendReport(pss PsSender) error { // TODO: probably not return if an error

	// TODO: what happens when we can't send to processing status and error?
	// should it even be returned? probably not so this can be called on a goroutine

	// TODO send to processing status API

	return nil // all good no errors
} // .Send
