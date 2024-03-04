package handlerdex

import (
	"encoding/json"
	"net/http"
) // .import

func (hd *HandlerDex) metadata(w http.ResponseWriter, r *http.Request) {

	resp := map[string]interface{}{"metadata_versions": hd.appConfig.MetadataVersions}

	json.NewEncoder(w).Encode(resp)
} // .health
