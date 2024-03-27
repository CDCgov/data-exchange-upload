package serverdex

import (
	"strconv"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storelocal"
	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

// OnUploadComplete gets notification on a tusd upload complete and makes the store copies necessary per config
func (sd ServerDex) onUploadComplete(uploadConfig metadatav1.UploadConfig, copyTargets []metadatav1.CopyTarget, eventUploadComplete tusd.HookEvent) error {

	// ------------------------------------------------------------------
	// RUN_MODE_LOCAL
	// ------------------------------------------------------------------
	if sd.CliFlags.RunMode == cliflags.RUN_MODE_LOCAL {

		sl := storelocal.CopierLocal{

			SrcFileName: eventUploadComplete.Upload.ID,
			SrcFolder:   sd.AppConfig.LocalFolderUploadsTus,
			// not using upload config or copy targets for local dev,
			// just copy file to another local folder uploadsA and add time ticks
			DstFolder:   sd.AppConfig.LocalFolderUploadsA,
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
	if sd.CliFlags.RunMode == cliflags.RUN_MODE_AZURE || sd.CliFlags.RunMode == cliflags.RUN_MODE_LOCAL_TO_AZURE {

		// time of ingest
		ingestDt := time.Now().UTC()

		dstBlobName := eventUploadComplete.Upload.MetaData[models.META_DESTINATION_ID]
		dstBlobName += "-"
		dstBlobName += eventUploadComplete.Upload.MetaData[models.META_EXT_EVENT]
		dstBlobName += "/"

		if uploadConfig.FolderStructure == models.DATE_YYYY_MM_DD {
			// Format MM-DD-YYYY
			ingestDtParts := strings.Split(ingestDt.Format("01-02-2006"), "-")
			mm := ingestDtParts[0]
			dd := ingestDtParts[1]
			yyyy := ingestDtParts[2]
			dstBlobName += yyyy + "/" + mm + "/" + dd + "/"
		} // .if

		dstBlobName += eventUploadComplete.Upload.MetaData[models.FILENAME]

		if uploadConfig.FileNameSuffix == models.CLOCK_TICKS {
			dstBlobName += "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
		} // .if

		saz := storeaz.CopierAzTusToDex{
			EventUploadComplete: eventUploadComplete,
			UploadConfig:        uploadConfig,
			CopyTargets:         copyTargets,
			//
			SrcTusAzBlobClient:    sd.HandlerDex.TusAzBlobClient,
			SrcTusAzContainerName: sd.AppConfig.TusAzStorageConfig.AzContainerName,
			SrcTusAzBlobName:      eventUploadComplete.Upload.ID,
			//
			DstAzContainerName: sd.AppConfig.DexAzStorageContainerName,
			DstAzBlobName:      dstBlobName,
			IngestDt:           ingestDt,
		} // .saz

		err := saz.CopyTusSrcToDst()
		if err != nil {
			return err
		} // .err

		// TODO: more copies as routing files, based on copy targets copyTargets

	} // .RUN_MODE_AZURE

	// all good
	return nil
} // .OnUploadComplete
