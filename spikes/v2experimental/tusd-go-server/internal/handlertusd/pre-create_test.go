package handlertusd

import (
	"net/http"
	"path/filepath"
	"slices"
	"testing"
	// "time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"

	"github.com/joho/godotenv"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// BenchmarkCheckManifestV1 benchmark for checking sender manifest
func BenchmarkCheckManifestV1(t *testing.B) {

	// time.Sleep(time.Duration(time.Duration.Seconds(1)))

	localEnvPath, err := filepath.Abs("../../configs/local/local.env")
	if err != nil {
		t.Errorf("got %q, wanted %q", err, "local env path no error")
	} // .err

	err = godotenv.Load(localEnvPath)
	if err != nil {
		t.Errorf("got %q, wanted %q", err, "config no error")
	} // .err

	appConfig := appconfig.AppConfig{

		System:        "DEX",
		DexProduct:    "Upload API",
		DexApp:        "tusd-go-server",
		LoggerDebugOn: false,

		AllowedDestAndEventsPath: "../../configs/allowed_destination_and_events.json",
		DefinitionsPath:          "../../configs/file-hooks/metadata-verify/",
		UploadConfigPath:         "../../configs/upload-configs/",
	} // .appConfig

	metaV1, err := metadatav1.LoadOnce(appConfig)
	if err != nil {
		t.Errorf("got %q, wanted %q", err, "metadata load no error")
	} // .err

	senderManifest := map[string]string{
		"meta_destination_id": "dextesting",
		"meta_ext_event":      "testevent1",

		"filename": "file.name",
	} // .senderManifest

	metaDestinationId, ok := senderManifest["meta_destination_id"]
	if !ok {
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrMetaDestIdNotFound,
		} // .httpRes

	} // .ok
	events, ok := metaV1.DestIdsEventsNameMap[metaDestinationId]
	if !ok {
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrMetaDestIdNotValid,
		} // .httpRes

	} // .ok

	// -----------------------------------------------------------------------------
	// check meta_ext_event
	// -----------------------------------------------------------------------------
	metaExtEvent, ok := senderManifest["meta_ext_event"]
	if !ok {
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrMetaExtEventNotFound,
		} // .httpRes
	} // .ok

	if !slices.Contains(events, metaExtEvent) {
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrMetaExtEventNotValid,
		} // .httpRes

	} // .if

	// -----------------------------------------------------------------------------
	// check schema for the meta_destination_id - meta_ext_event
	// -----------------------------------------------------------------------------
	eventDefFileName, ok := metaV1.DestIdEventFileNameMap[metaDestinationId+metaExtEvent]
	if !ok { // really this should not happen if every destination-event has a schema file
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrSchemaDefFileNameNA,
		} // .httpRes

	} // .if

	eventSchemas, ok := metaV1.Definitions[eventDefFileName]
	if !ok && len(eventSchemas) == 0 { // this should be also ok, because in v1 every destination-event has one schema file and for some reason the schemas are array of 1
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrSchemaDefNA,
		} // .httpRes
	} // .if
	schema := eventSchemas[0] // this was checked above
	schemaFields := schema.Fields

	for _, field := range schemaFields {

		// check if required.
		if field.Required == "true" {

			fieldValue, ok := senderManifest[field.FieldName]
			if !ok {
				_ = tusd.HTTPResponse{
					StatusCode: http.StatusBadRequest,
					Body:       ErrSchemaDefFieldNA + field.FieldName,
				} // .httpRes
			} // .if

			if field.AllowedValues != nil && len(field.AllowedValues) != 0 {

				if !slices.Contains(field.AllowedValues, fieldValue) {
					_ = tusd.HTTPResponse{
						StatusCode: http.StatusBadRequest,
						Body:       ErrSchemaDefFieldValueNotValid + field.FieldName,
					} // .httpRes

				} // .if
			} // .if

		} // .if

	} // .for

	// -----------------------------------------------------------------------------
	// check filename per upload config is sent
	// -----------------------------------------------------------------------------
	updConfigKey := metaDestinationId + "-" + metaExtEvent
	filename := metaV1.UploadConfigs[updConfigKey].FileNameMetadataField

	_, ok = senderManifest[filename]
	if !ok {
		_ = tusd.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       ErrUpdConfFileNameNA + filename,
		} // .httpRes
	} // .if

} // .BenchmarkCheckManifestV1
