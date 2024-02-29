package tusdhandler

// TODO: hooks pre-create

import (
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/dexmetadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

func loadConfigMetaV1() (dexmetadatav1.ConfigMetaV1, error) {

	configMetaV1, err := dexmetadatav1.Load()
	if err != nil {
		slog.Error("error starting service, error loading metadata v1 config files", "error", err)
		return dexmetadatav1.ConfigMetaV1{}, err
	} // .if

	return configMetaV1, nil 

} // .loadConfigMetaV1

// currently in v1 the required fields are wired in the pre-create hook check
var REQUIRED_METADATA_FIELDS = [2]string{"meta_destination_id", "meta_ext_event"}

// PreUploadCreateCallback func(hook HookEvent) (HTTPResponse, FileInfoChanges, error)
func checkManifestV1(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

	configMetaV1, err := dexmetadatav1.Load()
	if err != nil {
		slog.Error("error metadata v1 config not available", "error", err)

		httpResponse := tusd.HTTPResponse{
			StatusCode: 500,
			Body: "error metadata v1 config not available",
		} // .httpResponse
		return httpResponse, tusd.FileInfoChanges{}, nil 
	} // .if

	// TODO take out and finalize hook check
	slog.Debug("loaded config metadata v1", "configMetaV1", configMetaV1)
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