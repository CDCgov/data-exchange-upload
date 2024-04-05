package handlerdex

import (
	"encoding/json"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadatav1"
) // .import

// metadataV1 returns the full metadata v1 that is configured on the server for visibility
func (hd *HandlerDex) metadataV1(w http.ResponseWriter, r *http.Request) {

	metaV1, err := metadatav1.Get()
	if err != nil {
		errMsg := "error metadata v1 not available"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	json.NewEncoder(w).Encode(metaV1)
} // .health
