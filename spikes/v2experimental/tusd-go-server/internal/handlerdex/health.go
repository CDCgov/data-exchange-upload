package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"
) // .import

type HealthResp struct { // TODO: line up with DEX other products and apps

	RootResp        // Embedding rootResp, TODO: maybe this is not needed
	Health   string `json:"health"` // TODO: line-up

} // .Health

func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {

	jsonResp, err := json.Marshal(HealthResp{
		RootResp: RootResp{
			System:     hd.appConfig.System,
			DexProduct: hd.appConfig.DexProduct,
			DexApp:     hd.appConfig.DexApp,
			ServerTime: time.Now().Format(time.RFC3339),
		},
		Health: "All Good",
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
