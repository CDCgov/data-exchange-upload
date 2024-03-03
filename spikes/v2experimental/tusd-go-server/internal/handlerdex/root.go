package handlerdex

import (
	"encoding/json"
	"net/http"
	"time"

) // .import

type rootResp struct {
    System string `json:"system"`
    DexProduct string `json:"dex_product"`
    DexApp string `json:"dex_app"` 
    ServerTime string `json:"server_time"`
} // .rootAppConfig


func (hd *HandlerDex) root(w http.ResponseWriter, r *http.Request) {

    jsonResp, err := json.Marshal(rootResp{

        System: hd.config.System,
        DexProduct: hd.config.DexProduct,
        DexApp: hd.config.DexApp,
        
        ServerTime:  time.Now().Format(time.RFC3339),
    }) // .jsonResp
    if err != nil {
        // TODO log error 
        // TODO: don't expose errors
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    } // .if 

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResp)
} // .root
