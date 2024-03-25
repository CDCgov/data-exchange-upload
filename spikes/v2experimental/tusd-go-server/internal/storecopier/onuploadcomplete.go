package storecopier

import (
	"strconv"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storelocal"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// OnUploadComplete gets notification on a tusd upload complete and makes the store copies necessary per config
func OnUploadComplete(flags cliflags.Flags, appConfig appconfig.AppConfig, allUploadConfig metadatav1.AllUploadConfigs, eventUploadComplete tusd.HookEvent) error {

	evInfo := FromTusToDstEvent{
		AppConfig:           appConfig,
		AllUploadConfig:     allUploadConfig,
		EventUploadComplete: eventUploadComplete,
	} // .eventInfo

	// ------------------------------------------------------------------
	// RUN_MODE_LOCAL
	// ------------------------------------------------------------------
	if flags.RunMode == cliflags.RUN_MODE_LOCAL {

		sl := storelocal.CopierLocal{

			SrcFileName: evInfo.EventUploadComplete.Upload.ID,
			SrcFolder:   evInfo.AppConfig.LocalFolderUploadsTus,

			// TODO config destination per respective upload config
			// TODO: adding file ticks, change per config
			// _ = cl.AllUploadConfig // TODO

			DstFolder:   evInfo.AppConfig.LocalFolderUploadsA,
			DstFileName: evInfo.EventUploadComplete.Upload.MetaData["filename"] + "_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		} // .cl

		err := sl.CopyTusSrcToDst()
		if err != nil {
			return err
		} // .err

	} // .RUN_MODE_LOCAL

	// ------------------------------------------------------------------
	// RUN_MODE_AZURE
	// ------------------------------------------------------------------
	if flags.RunMode == cliflags.RUN_MODE_AZURE {

	} // .RUN_MODE_AZURE

	return nil
} // .OnUploadComplete
