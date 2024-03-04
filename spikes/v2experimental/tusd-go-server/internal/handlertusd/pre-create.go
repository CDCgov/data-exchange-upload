package handlertusd

// TODO: hooks pre-create

import (
	"net/http"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	slog "golang.org/x/exp/slog"
)

// currently in v1 the required fields are wired in the pre-create hook check: meta_destination_id and meta_ext_event

// pre-create hook
func checkManifestV1(logger *slog.Logger) func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	return func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

		// -----------------------------------------------------------------------------
		// get the metadata v1 object needed for these checks
		// -----------------------------------------------------------------------------
		configMetaV1, err := metadatav1.Get()
		if err != nil {
			logger.Error("error metadata v1 config not available", "error", err)
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "error metadata v1 config not available",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .if

		// TODO take out and finalize hook check
		logger.Debug("loaded config metadata v1", "configMetaV1", configMetaV1)
		// hook.Upload.MetaData["Filename"] = hook.HTTPRequest.Header.Get("Filename")

		// -----------------------------------------------------------------------------
		// meta_destination_id
		// -----------------------------------------------------------------------------
		_ /*metaDestinationId*/, ok := hook.Upload.MetaData["meta_destination_id"]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_destination_id not found in the provided manifest",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .ok

		// -----------------------------------------------------------------------------
		// meta_ext_event
		// -----------------------------------------------------------------------------
		_ /*metaExtEvent*/, ok = hook.Upload.MetaData["meta_ext_event"]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_ext_event not found in the provided manifest",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .ok

		httpResponse := tusd.HTTPResponse{
			StatusCode: 200,
			Body:       "all good",
		} // .httpResponse

		return httpResponse, tusd.FileInfoChanges{}, nil
	} // .return
} // .checkManifestV1
