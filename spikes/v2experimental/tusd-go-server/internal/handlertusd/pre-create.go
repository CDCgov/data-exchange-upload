package handlertusd

// TODO: hooks pre-create

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"log/slog"
)

// currently in v1 the required fields are wired in the pre-create hook check
var REQUIRED_METADATA_FIELDS = [2]string{"meta_destination_id", "meta_ext_event"}

// pre-create hook
func checkManifestV1(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

	// TODO find a way to move from slog to custom logger if feasible, e.g. wrap function

	configMetaV1, err := metadatav1.Get()
	if err != nil {
		slog.Error("error metadata v1 config not available", "error", err)

		httpResponse := tusd.HTTPResponse{
			StatusCode: 500,
			Body:       "error metadata v1 config not available",
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
			Body:       "filename not found in the provided manifest",
		} // .httpResponse
		return httpResponse, tusd.FileInfoChanges{}, nil
	} // .ok

	httpResponse := tusd.HTTPResponse{
		StatusCode: 200,
		Body:       "all good",
	} // .httpResponse

	return httpResponse, tusd.FileInfoChanges{}, nil
} // .checkMeta
