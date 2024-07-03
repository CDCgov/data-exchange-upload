package mmsapihealth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type MmsApiHealth struct {
	HealthURL  string
	HTTPClient *http.Client
}

func New(uri string) (*MmsApiHealth, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Check base URI is valid:
	healthUrl, err := url.ParseRequestURI(uri)
	if err != nil {
		return nil, err
	}

	healthUrl.Path, err = url.JoinPath(healthUrl.Path, "health")
	if err != nil {
		return nil, err
	}

	return &MmsApiHealth{
		HealthURL:  healthUrl.String(),
		HTTPClient: client,
	}, nil

} // .New

func (mmsApiHealth MmsApiHealth) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.MMS_API

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mmsApiHealth.HealthURL, nil)
	if err != nil {
		return mmsApiDown(err)
	}

	resp, err := mmsApiHealth.HTTPClient.Do(req)
	if err != nil {
		return mmsApiDown(err)
	}
	defer resp.Body.Close()

	// When unhealthy, the MMS API responds with status = http.StatusInternalServerError
	if resp.StatusCode != http.StatusOK {
		return mmsApiDown(fmt.Errorf("mms api health failed with status code %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)))
	}

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}

func mmsApiDown(err error) models.ServiceHealthResp {
	return models.ServiceHealthResp{
		Service:     models.MMS_API,
		Status:      models.STATUS_DOWN,
		HealthIssue: err.Error(),
	}
}
