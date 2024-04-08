package handlerdex

import (
	"encoding/json"
	"net/http"
) // .import

// metadata returns DEX Upload Api all supported metadata versions
// TODO: currently this is loaded from app config and should be changed to metadata service when available
func (hd *HandlerDex) metadata(w http.ResponseWriter, r *http.Request) {

	resp := map[string]interface{}{"metadata_versions": hd.appConfig.MetadataVersions}

	json.NewEncoder(w).Encode(resp)
} // .health
