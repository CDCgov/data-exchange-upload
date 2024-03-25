package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
) // .import

// HealthResp, app health response
type HealthResp struct { // TODO: line up with DEX other products and apps

	RootResp // Embedding rootResp, TODO: maybe this is not needed

	Status string `json:"status"` // general app health

	Services []models.HealthServiceResp `json:"services"`
} // .HealthResp

// health responds to /health endpoint with the health of the app including dependency services
func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {

	status := models.STATUS_UP

	var servicesResponses []models.HealthServiceResp

	if hd.cliFlags.RunMode == cliflags.RUN_MODE_LOCAL_TO_AZURE || hd.cliFlags.RunMode == cliflags.RUN_MODE_AZURE {

		ch := make(chan models.HealthServiceResp, 3)

		go func() { ch <- storeaz.CheckTusAzBlobClient(hd.TusAzBlobClient) }()
		go func() { ch <- storeaz.CheckRouterAzBlobClient(hd.RouterAzBlobClient) }()
		go func() { ch <- storeaz.CheckEdavAzBlobClient(hd.EdavAzBlobClient) }()

		servicesResponses = append(servicesResponses, <-ch, <-ch, <-ch)

		for _, sr := range servicesResponses {
			if sr.Status == models.STATUS_DOWN {
				status = models.STATUS_DEGRADED
			} // .if
		} // .if
	} // .if

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
