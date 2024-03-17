package storecopier

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storecopierlocal"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// OnUploadComplete gets notification on a tusd upload complete and makes the store copies necessary per config
func OnUploadComplete(flags cliflags.Flags, appConfig appconfig.AppConfig, allUploadConfig metadatav1.AllUploadConfigs, eventUploadComplete tusd.HookEvent) error {

	// LOCAL
	if flags.Environment == cliflags.ENV_LOCAL {

		// making only one local copy as concept from tusd raw to a local folder
		lc := storecopierlocal.FromTusToDstCopier{
			AppConfig:           appConfig,
			AllUploadConfig:     allUploadConfig,
			EventUploadComplete: eventUploadComplete,
		} // .lc

		err := lc.CopyTusSrcToDst()
		if err != nil {
			return err
		} // .err

	} // .LOCAL

	// AZURE

	return nil
} // .OnUploadComplete
