package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
) // .import

// HealthResp, app health response
type HealthResp struct { // TODO: line up with DEX other products and apps

	RootResp // Embedding rootResp, TODO: maybe this is not needed

	Status string `json:"status"` // general app health

	Services []ServiceHealthResp `json:"services"`
} // .HealthResp

// ServiceHealthResp, health response from an app service dependency
type ServiceHealthResp struct {
	Service     string `json:"service"`
	Status      string `json:"status"`
	HealthIssue string `json:"health_issue"`
} // .ServiceHealthResp

// general app health statuses
const STATUS_UP = "UP"
const STATUS_DEGRADED = "DEGRADED"
const STATUS_DOWN = "DOWN"
const HEALTH_ISSUE_NONE = "None reported"

// health responds to /health endpoint with the health of the app including dependency services
func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {

	status := STATUS_UP

	var servicesResponses []ServiceHealthResp

	_, err := storeaz.NewTusAzBlobClient(hd.appConfig)
	if err != nil {
		servicesResponses = append(servicesResponses, ServiceHealthResp{
			Service:     "AzBlobTusUpload",
			Status:      STATUS_DOWN,
			HealthIssue: err.Error()})
	} else {
		servicesResponses = append(servicesResponses, ServiceHealthResp{
			Service:     "AzBlobTusUpload",
			Status:      STATUS_UP,
			HealthIssue: HEALTH_ISSUE_NONE})
	} // .else

	jsonResp, err := json.Marshal(HealthResp{
		RootResp: RootResp{
			System:     hd.appConfig.System,
			DexProduct: hd.appConfig.DexProduct,
			DexApp:     hd.appConfig.DexApp,
			ServerTime: time.Now().Format(time.RFC3339),
			RunMode:    hd.cliFlags.RunMode,
		},
		Status:   status,
		Services: servicesResponses,
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for health response"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
} // .health
