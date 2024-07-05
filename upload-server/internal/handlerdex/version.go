package handlerdex

import (
	"encoding/json"
	"net/http"
) // .import

// TODO: replace with git SHA in github actions at commit time
const GitRepo = "%%GIT_REPO%%"
const GitVersion = "%%GIT_VERSION%%"
const GitShortVersion = "%%GIT_SHORT_VERSION%%"
const LatestReleaseTag = "%%LATEST_RELEASE_TAG%%"

// VersionResp can be used if needed to populate file metadata
type VersionResp struct {
	GitRepo          string `json:"git_repo"`
	GitVersion       string `json:"git_version"`
	GitShortVersion  string `json:"git_short_version"`
	LatestReleaseTag string `json:"latest_release_tag"`
} // .VersionResp

// version provide git repo and version from where this app was built
func (hd *HandlerDex) version(w http.ResponseWriter, r *http.Request) {

	jsonResp, err := json.Marshal(VersionResp{
		GitRepo:         GitRepo,
		GitVersion:      GitVersion,
		GitShortVersion: GitShortVersion,
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for version response"
		hd.logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
} // .version
