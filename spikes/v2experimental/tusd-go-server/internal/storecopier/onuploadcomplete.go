package storecopier

import (
	"strconv"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storelocal"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// OnUploadComplete gets notification on a tusd upload complete and makes the store copies necessary per config
func OnUploadComplete(flags cliflags.Flags, appConfig appconfig.AppConfig, uploadConfig metadatav1.UploadConfig, copyTargets []metadatav1.CopyTarget, eventUploadComplete tusd.HookEvent) error {

	// ------------------------------------------------------------------
	// RUN_MODE_LOCAL
	// ------------------------------------------------------------------
	if flags.RunMode == cliflags.RUN_MODE_LOCAL {

		sl := storelocal.CopierLocal{

			SrcFileName: eventUploadComplete.Upload.ID,
			SrcFolder:   appConfig.LocalFolderUploadsTus,
			// not using upload config or copy targets for local dev,
			// just copy file to another local folder uploadsA and add time ticks
			DstFolder:   appConfig.LocalFolderUploadsA,
			DstFileName: eventUploadComplete.Upload.MetaData["filename"] + "_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		} // .cl

		err := sl.CopyTusSrcToDst()
		if err != nil {
			return err
		} // .err
	} // .RUN_MODE_LOCAL

	// ------------------------------------------------------------------
	// RUN_MODE_AZURE, RUN_MODE_LOCAL_TO_AZURE
	// ------------------------------------------------------------------
	if flags.RunMode == cliflags.RUN_MODE_AZURE || flags.RunMode == cliflags.RUN_MODE_LOCAL_TO_AZURE {

		saz := storeaz.CopierAz{
			EventUploadComplete: eventUploadComplete,
			UploadConfig:        uploadConfig,
			CopyTargets:         copyTargets,
		} // .saz

		err := saz.CopyTusSrcToDst()
		if err != nil {
			return err
		} // .err
	} // .RUN_MODE_AZURE

	// all good
	return nil
} // .OnUploadComplete
