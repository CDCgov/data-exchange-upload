package handlerdex 

import (
	"net/http"
) // .import

const gitAppVersion = "%%DEX_APP_GIT_VERSION_SHA%%"

func (hd *HandlerDex) version(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte(gitAppVersion)) // TODO: replace with git SHA in github actions at commit time

} // .version
