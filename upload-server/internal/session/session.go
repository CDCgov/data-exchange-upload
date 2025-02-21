package session

import "github.com/gorilla/sessions"

var store sessions.Store

func Init() {
	store = sessions.NewCookieStore([]byte("temp"))
}

func Store() sessions.Store {
	return store
}
