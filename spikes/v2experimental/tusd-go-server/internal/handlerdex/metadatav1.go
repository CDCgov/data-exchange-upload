package handlerdex

import (
	"encoding/json"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
) // .import

func (hd *HandlerDex) metaV1(w http.ResponseWriter, r *http.Request) {

	metaV1, err := metadatav1.Get()
	if err != nil {
		errMsg := "error metadata v1 not available"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	json.NewEncoder(w).Encode(metaV1)
} // .health
