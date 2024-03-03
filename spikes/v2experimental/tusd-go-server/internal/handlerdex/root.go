package handlerdex 

import (
	"encoding/json"
	"net/http"
	"time"
) // .import


func (hd *HandlerDex) root(w http.ResponseWriter, r *http.Request) {

	currentTime := time.Now()
    resp := map[string]interface{}{

		// TODO: add from config, Service, App, etc...

        "server_time": currentTime.Format(time.RFC3339),
    } // .resp\

    jsonResp, err := json.Marshal(resp)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    } // .if 

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(jsonResp)
} // .root
