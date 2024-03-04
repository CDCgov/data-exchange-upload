package metadatav1

import (
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

// this gets the metadata already loaded, in emergency case it will reload but this should not happen
func Get() (*MetadataV1, error) {

	if metaV1Instance == nil { // this should be false in almost all cases
		// this should not happen but just in case the call to LoadOnce in main got wiped
		// emergency: metadata is missing, need to reload metadata here
		// now need to load app config here
		appConfig, err := appconfig.ParseConfig()
		if err != nil {
			slog.Error("error parsing app config", "error", err)
			return nil, err
		} // .if

		metaV1Instance, err = LoadOnce(appConfig)
		if err != nil {
			logger := pkgLogger(appConfig)
			logger.Error("error loading metadata v1", "error", err)
			return nil, err
		} // .err

	} // .metaV1Instance

	return metaV1Instance, nil
} // .Get
