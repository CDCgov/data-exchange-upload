package handlerdex

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/dexmetadatav1"
) // .import

func (hd *HandlerDex) metaV1(w http.ResponseWriter, r *http.Request) {

	configMetaV1, err := dexmetadatav1.Load()
	if err != nil {
		errMessage := "error metadata v1 config not available"
		slog.Error(errMessage, "error", err)
		json.NewEncoder(w).Encode(struct{ error string }{error: errMessage})
		return
	} // .err

	json.NewEncoder(w).Encode(configMetaV1)
} // .health
