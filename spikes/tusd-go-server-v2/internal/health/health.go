package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
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
type HealthResp struct { // TODO: line up with DEX other products and apps
	Status   string                     `json:"status"` // general app health
	Services []models.ServiceHealthResp `json:"services"`
} // .HealthResp

type SystemHealthCheck struct {
	Services []Checkable
}

func (hc *SystemHealthCheck) Register(c Checkable) {
	hc.Services = append(hc.Services, c)
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

func Register(c Checkable) {
	DefaultSystemHealthCheck.Register(c)
}

func Handler() http.Handler {
	return DefaultSystemHealthCheck
}
