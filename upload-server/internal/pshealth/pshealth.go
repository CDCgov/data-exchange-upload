package pshealth

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"net/url"
)

type PsHealth struct {
	EndpointHealth    string
	EndpointHealthURI *url.URL
} // .PsHealth

func New(appConfig appconfig.AppConfig) (*PsHealth, error) {

	endpointURI, err := url.ParseRequestURI(appConfig.ProcessingStatusHealthURI)
	if err != nil {
		return nil, err
	} // .if

	return &PsHealth{
		EndpointHealth:    appConfig.ProcessingStatusHealthURI,
		EndpointHealthURI: endpointURI,
	}, nil // .return

} // .New
