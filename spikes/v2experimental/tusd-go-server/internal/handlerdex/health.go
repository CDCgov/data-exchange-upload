package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
) // .import

type HealthResp struct { // TODO: line up with DEX other products and apps

	RootResp // Embedding rootResp, TODO: maybe this is not needed

	Status string `json:"status"`

	ErrInfo string `json:"error_info"`
} // .Health

const STATUS_UP = "UP"
const STATUS_DOWN = "DOWN"

// health responds to /health endpoint with the health of the app
// TODO: line-up with DEX standards
// TODO: check the dependencies such as storages
func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {

	status := STATUS_UP
	errInfo := "nil"

	_, err := storeaz.NewAzBlobClient(hd.appConfig)
	if err != nil {
		status = STATUS_DOWN
		errInfo = err.Error()
	} // .if

	jsonResp, err := json.Marshal(HealthResp{
		RootResp: RootResp{
			System:      hd.appConfig.System,
			DexProduct:  hd.appConfig.DexProduct,
			DexApp:      hd.appConfig.DexApp,
			ServerTime:  time.Now().Format(time.RFC3339),
			Environment: hd.cliFlags.Environment,
		},
		Status:  status,
		ErrInfo: errInfo,
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
