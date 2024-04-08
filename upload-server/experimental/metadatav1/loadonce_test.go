package metadatav1

import (
	"path/filepath"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"

	"github.com/joho/godotenv"
) // .import

// TestLoadOnce calls to load metadata v1
func TestLoadOnce(t *testing.T) {

	localEnvPath, err := filepath.Abs("../../configs/local/local.env")
	if err != nil {
		t.Errorf("got %q, wanted %q", err, "local env path no error")
	} // .err

	err = godotenv.Load(localEnvPath)
	if err != nil {
		t.Errorf("got %q, wanted %q", err, "config no error")
	} // .err

	appConfig := appconfig.AppConfig{

		AllowedDestAndEventsPath: "../../configs/allowed_destination_and_events.json",
		DefinitionsPath:          "../../configs/file-hooks/metadata-verify/",
		UploadConfigPath:         "../../configs/upload-configs/",
		HydrateV1ConfigPath:      "../../configs/upload-configs/v2/",
	} // .appConfig

	metaV1, err := LoadOnce(appConfig)
	if err != nil {
		t.Fatalf("got %q, wanted %q", err, "metadata load no error")
	} // .err

	if len((*metaV1).AllowedDestAndEvents) == 0 {
		t.Errorf("got %q, wanted %q", 0, "allowed_destination_and_events loaded")
	} // .err

	if len((*metaV1).Definitions) == 0 {
		t.Errorf("got %q, wanted %q", 0, "definitions loaded")
	} // .err

	if len((*metaV1).UploadConfigs) == 0 {
		t.Errorf("got %q, wanted %q", 0, "upload configs loaded")
	} // .err

} // .TestLoadOnce
