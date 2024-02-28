package server 

// TODO: hooks pre-create

import (
	"log" // TODO: slog 
	tusd "github.com/tus/tusd/v2/pkg/handler"
)


// PreUploadCreateCallback func(hook HookEvent) (HTTPResponse, FileInfoChanges, error)
func checkMeta(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

	log.Print(hook.Upload.MetaData)
	// hook.Upload.MetaData["Filename"] = hook.HTTPRequest.Header.Get("Filename")

	_, ok := hook.Upload.MetaData["filename"]
	if !ok {
		httpResponse := tusd.HTTPResponse{
			StatusCode: 400,
			Body: "filename not found in the provided manifest",
		} // .httpResponse
		return httpResponse, tusd.FileInfoChanges{}, nil 
	} // .ok

	httpResponse := tusd.HTTPResponse{
		StatusCode: 200,
		Body: "all good",
	} // .httpResponse
        
	return httpResponse, tusd.FileInfoChanges{}, nil 
} // .checkMeta