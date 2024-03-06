package handlertusd

// TODO: hooks pre-create

import (
	"net/http"
	"slices"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	slog "golang.org/x/exp/slog"
)

// currently in v1 the required fields are wired in the pre-create hook check: meta_destination_id and meta_ext_event

// pre-create hook
func checkManifestV1(logger *slog.Logger) func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	return func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

		// TODO: add ps integration to send report on failure

		senderManifest := hook.Upload.MetaData

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

		// -----------------------------------------------------------------------------
		// check meta_destination_id
		// -----------------------------------------------------------------------------
		metaDestinationId, ok := senderManifest["meta_destination_id"]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_destination_id not found in sent manifest",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .ok
		events, ok := configMetaV1.DestIdsEventsNameMap[metaDestinationId]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_destination_id value is not valid and ext_events are not available",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .ok

		// -----------------------------------------------------------------------------
		// check meta_ext_event
		// -----------------------------------------------------------------------------
		metaExtEvent, ok := senderManifest["meta_ext_event"]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_ext_event not found in sent manifest",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .ok

		if !slices.Contains(events, metaExtEvent) {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "meta_ext_event value is not valid and not found in ext_events",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .if

		// -----------------------------------------------------------------------------
		// check schema for the meta_destination_id - meta_ext_event
		// -----------------------------------------------------------------------------
		eventDefFileName, ok := configMetaV1.DestIdEventFileNameMap[metaDestinationId+metaExtEvent]
		if !ok { // really this should not happen if every destination-event has a schema file
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "schema definition file name not found for meta_destination_id and meta_ext_event combination",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .if

		eventSchemas, ok := configMetaV1.Definitions[eventDefFileName]
		if !ok && len(eventSchemas) == 0 { // this should be also ok, because in v1 every destination-event has one schema file and for some reason the schemas are array of 1
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "schema definition not found for meta_destination_id and meta_ext_event combination",
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		} // .if
		schema := eventSchemas[0] // this was checked above
		schemaFields := schema.Fields

		for _, field := range schemaFields {

			// check if required.
			if field.Required == "true" {

				fieldValue, ok := senderManifest[field.FieldName]
				if !ok {
					httpResponse := tusd.HTTPResponse{
						StatusCode: http.StatusBadRequest,
						Body:       "schema definition required field not sent: " + field.FieldName,
					} // .httpResponse
					return httpResponse, tusd.FileInfoChanges{}, nil
				} // .if

				if field.AllowedValues != nil && len(field.AllowedValues) != 0 {

					if !slices.Contains(field.AllowedValues, fieldValue) {
						httpResponse := tusd.HTTPResponse{
							StatusCode: http.StatusBadRequest,
							Body:       "schema definition required field value not valid for field name: " + field.FieldName,
						} // .httpResponse
						return httpResponse, tusd.FileInfoChanges{}, nil
					} // .if
				} // .if

			} // .if

		} // .for

		// -----------------------------------------------------------------------------
		// check filename per upload config is sent
		// -----------------------------------------------------------------------------
		
		updConfigKey := metaDestinationId + "-" + metaExtEvent
		filename := configMetaV1.UploadConfigs[updConfigKey].FileNameMetadataField
		
		_, ok = senderManifest[filename]
		if !ok {
			httpResponse := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "not found the file name per config: " + filename,
			} // .httpResponse
			return httpResponse, tusd.FileInfoChanges{}, nil
		}// .if 
		
		// -----------------------------------------------------------------------------
		// all checks have passed
		// -----------------------------------------------------------------------------
		httpResponse := tusd.HTTPResponse{
			StatusCode: http.StatusOK,
			Body:       "upload pre-create checks have passed ok",
		} // .httpResponse

		return httpResponse, tusd.FileInfoChanges{}, nil
	} // .return
} // .checkManifestV1
