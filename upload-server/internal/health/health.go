package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
) // .import

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type Checkable interface {
	Health(context.Context) models.ServiceHealthResp
}

// HealthResp, app health response
type HealthResp struct {
	Status   string                     `json:"status"` // general app health
	Services []models.ServiceHealthResp `json:"services"`
} // .HealthResp

type SystemHealthCheck struct {
	Services []Checkable
}

func (hc *SystemHealthCheck) Register(checks ...any) error {
	var errs error
	for _, c := range checks {
		if cc, ok := c.(Checkable); ok {
			hc.Services = append(hc.Services, cc)
		} else {
			errs = errors.Join(errs, fmt.Errorf("Could not register %+V health check", c))
		}
	}
	return errs
}

// health responds to /health endpoint with the health of the app including dependency services
func (hc *SystemHealthCheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	status := models.STATUS_UP

	var servicesResponses []models.ServiceHealthResp

	for _, check := range hc.Services {
		sr := check.Health(r.Context())
		servicesResponses = append(servicesResponses, sr)
		if sr.Status == models.STATUS_DOWN {
			status = models.STATUS_DEGRADED
		} // .if
	} // .if

	jsonResp, err := json.Marshal(HealthResp{
		Status:   status,
		Services: servicesResponses,
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for health response"
		logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
} // .health

var DefaultSystemHealthCheck = &SystemHealthCheck{}

func Register(c ...any) error {
	return DefaultSystemHealthCheck.Register(c...)
}

func Handler() http.Handler {
	return DefaultSystemHealthCheck
}
