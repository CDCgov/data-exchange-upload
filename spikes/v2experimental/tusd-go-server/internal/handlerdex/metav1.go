package handlerdex

import (
	"encoding/json"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/dexmetadatav1"
) // .import

func (hd *HandlerDex) metaV1(w http.ResponseWriter, r *http.Request) {

	configMetaV1, err := dexmetadatav1.Load()
	if err != nil {
		errMsg := "error metadata v1 config not available"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	json.NewEncoder(w).Encode(configMetaV1)
} // .health
