package handlerdex 

import (
	"encoding/json"
	"net/http"
) // .import

// TODO: replace with git SHA in github actions at commit time
const GitRepo ="%%git remote get-url remote-name%%"
const GitVersion = "%%GIT_VERSION%%"
const GitShortVersion = "%%GIT_SHORT_VERSION%%"

type VersionResp struct {
    GitRepo string `json:"git_repo"`
    GitVersion string `json:"git_version"`
	GitShortVersion string `json:"git_short_version"`
} // .VersionResp

func (hd *HandlerDex) version(w http.ResponseWriter, r *http.Request) {

	jsonResp, err := json.Marshal(VersionResp{
        GitRepo: GitRepo,
        GitVersion: GitVersion,
		GitShortVersion: GitShortVersion,
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
} // .version
