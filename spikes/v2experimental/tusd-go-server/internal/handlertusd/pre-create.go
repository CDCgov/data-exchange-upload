package handlertusd

// TODO: hooks pre-create

import (
	"net/http"
	"slices"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	slog "golang.org/x/exp/slog"
)

var (
	ErrMetaV1ConfigNA              = "error metadata v1 config not available"
	ErrMetaDestIdNotFound          = "meta_destination_id not found in manifest"
	ErrMetaDestIdNotValid          = "meta_destination_id value is not valid"
	ErrMetaExtEventNotFound        = "meta_ext_event not found in manifest"
	ErrMetaExtEventNotValid        = "meta_ext_event value is not valid"
	ErrSchemaDefFileNameNA         = "schema definition file name not found for the meta_destination_id and meta_ext_event"
	ErrSchemaDefNA                 = "schema definition not found for the meta_destination_id and meta_ext_event"
	ErrSchemaDefFieldNA            = "schema definition required field not sent: "
	ErrSchemaDefFieldValueNotValid = "schema definition required field value not valid for field name: "
	ErrUpdConfFileNameNA           = "file name not found, required per config: "
) // .var

// checkManifestV1 is a TUSD pre-create hook, checks file manifest for fields and values per metadata v1 requirements
// currently in v1 hooks the required fields are wired in the pre-create hook check: meta_destination_id and meta_ext_event
func checkManifestV1(logger *slog.Logger) func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {
	return func(hook tusd.HookEvent) (tusd.HTTPResponse, tusd.FileInfoChanges, error) {

		// TODO: add ps integration to send report on failure

		senderManifest := hook.Upload.MetaData

		tusdErr := tusd.Error{}

		// -----------------------------------------------------------------------------
		// get the metadata v1 object needed for these checks
		// -----------------------------------------------------------------------------
		configMetaV1, err := metadatav1.Get()
		if err != nil {
			logger.Error(ErrMetaV1ConfigNA, "error", err)
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       ErrMetaV1ConfigNA,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .if

		// -----------------------------------------------------------------------------
		// check meta_destination_id
		// -----------------------------------------------------------------------------
		metaDestinationId, ok := senderManifest["meta_destination_id"]
		if !ok {
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrMetaDestIdNotFound,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .ok
		events, ok := configMetaV1.DestIdsEventsNameMap[metaDestinationId]
		if !ok {
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrMetaDestIdNotValid,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .ok

		// -----------------------------------------------------------------------------
		// check meta_ext_event
		// -----------------------------------------------------------------------------
		metaExtEvent, ok := senderManifest["meta_ext_event"]
		if !ok {
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrMetaExtEventNotFound,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .ok

		if !slices.Contains(events, metaExtEvent) {
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrMetaExtEventNotValid,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .if

		// -----------------------------------------------------------------------------
		// check schema for the meta_destination_id - meta_ext_event
		// -----------------------------------------------------------------------------
		eventDefFileName, ok := configMetaV1.DestIdEventFileNameMap[metaDestinationId+metaExtEvent]
		if !ok { // really this should not happen if every destination-event has a schema file
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrSchemaDefFileNameNA,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .if

		eventSchemas, ok := configMetaV1.Definitions[eventDefFileName]
		if !ok && len(eventSchemas) == 0 { // this should be also ok, because in v1 every destination-event has one schema file and for some reason the schemas are array of 1
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrSchemaDefNA,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .if
		schema := eventSchemas[0] // this was checked above
		schemaFields := schema.Fields

		for _, field := range schemaFields {

			// check if required.
			if field.Required == "true" {

				fieldValue, ok := senderManifest[field.FieldName]
				if !ok {
					httpRes := tusd.HTTPResponse{
						StatusCode: http.StatusBadRequest,
						Body:       ErrSchemaDefFieldNA + field.FieldName,
					} // .httpRes

					tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
					return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
				} // .if

				if field.AllowedValues != nil && len(field.AllowedValues) != 0 {

					if !slices.Contains(field.AllowedValues, fieldValue) {
						httpRes := tusd.HTTPResponse{
							StatusCode: http.StatusBadRequest,
							Body:       ErrSchemaDefFieldValueNotValid + field.FieldName,
						} // .httpRes

						tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
						return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
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
			httpRes := tusd.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       ErrUpdConfFileNameNA + filename,
			} // .httpRes
			tusdErr.HTTPResponse = tusdErr.HTTPResponse.MergeWith(httpRes)
			return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, tusdErr
		} // .if

		// -----------------------------------------------------------------------------
		// all checks have passed
		// -----------------------------------------------------------------------------
		// only to be used if sending additional data to the client
		// tusd choses the appropriate status code
		return tusd.HTTPResponse{}, tusd.FileInfoChanges{}, nil

	} // .return
} // .checkManifestV1
