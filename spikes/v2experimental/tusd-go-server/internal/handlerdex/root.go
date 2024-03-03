package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"
) // .import

type RootResp struct {
	System     string `json:"system"`
	DexProduct string `json:"dex_product"`
	DexApp     string `json:"dex_app"`
	ServerTime string `json:"server_time"`
} // .rootResp

func (hd *HandlerDex) root(w http.ResponseWriter, r *http.Request) {

	jsonResp, err := json.Marshal(RootResp{
		System:     hd.appConfig.System,
		DexProduct: hd.appConfig.DexProduct,
		DexApp:     hd.appConfig.DexApp,
		ServerTime: time.Now().Format(time.RFC3339),
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for root response"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
} // .root
