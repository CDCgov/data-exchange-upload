package cli

import (
	"encoding/json"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/version"
	"net/http"
)

type VersionHandler struct{}

func (vh *VersionHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	resp := &version.Response{
		Repo:                 version.GitRepo,
		LatestReleaseVersion: version.LatestReleaseVersion,
		GitShortSha:          version.GitShortSha,
	}

	rw.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	enc.Encode(resp)
}
