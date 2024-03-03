package handlerdex 

import (
	"net/http"
) // .import

func (hd *HandlerDex) health(w http.ResponseWriter, r *http.Request) {


	w.Write([]byte("running ok"))

} // .health
