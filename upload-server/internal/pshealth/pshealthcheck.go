package pshealth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

func (pss PsHealth) Health(ctx context.Context) models.ServiceHealthResp {

	var shr models.ServiceHealthResp
	shr.Service = models.PROCESSING_STATUS_APP

	// call processing status health endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pss.EndpointHealth, nil)
	if err != nil {
		return processingStatusDown(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		return processingStatusDown(err)
	} // .if
	defer res.Body.Close()

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return processingStatusDown(err)
	} // .if

	var resMap map[string]json.RawMessage
	err = json.Unmarshal(resData, &resMap)
	if err != nil {
		return processingStatusDown(err)
	} // .if

	status, ok := resMap["status"]
	if !ok {
		return processingStatusDown(errors.New("processing status response not found"))
	} // .if

	if string(status) != (models.STATUS_UP) {
		return processingStatusDown(fmt.Errorf("processing status: %s", status))
	} // .if

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
} // .HealthCheck

func processingStatusDown(err error) models.ServiceHealthResp {
	return models.ServiceHealthResp{
		Service:     models.PROCESSING_STATUS_APP,
		Status:      models.STATUS_DOWN,
		HealthIssue: err.Error(),
	} // .return
} // .processingStatusDown
