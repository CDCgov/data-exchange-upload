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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
) // .import

// OnUploadComplete gets notification on a tusd upload complete and makes the store copies necessary per config
func (sd ServerDex) onUploadComplete(
	uploadConfig metadatav1.UploadConfig,
	hydrateV1Config metadatav1.HydrateV1Config,
	copyTargets []metadatav1.CopyTarget,
	eventUploadComplete tusd.HookEvent) error {

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

		dstBlobName := getDstBlobName(eventUploadComplete, uploadConfig, ingestDt)

		manifest := make(map[string]*string)
		for mdk, mdv := range eventUploadComplete.Upload.MetaData {
			manifest[mdk] = to.Ptr(mdv)
		} // .for
		// hydrate manifest for v2
		sd.hydrateManifestV1(&manifest, hydrateV1Config)

		// add ingest datetime to file blob metadata for other services to use same folders YYYY/MM/DD
		manifest[models.DEX_INGEST_DATE_TIME_KEY_NAME] = to.Ptr(ingestDt.Format(time.RFC3339))

		// ------------------------------------------------------------------
		// copy from tus raw file + manifest(.info) into dex container as one
		// ------------------------------------------------------------------
		err := sd.copyTusDexWRetry(eventUploadComplete, dstBlobName, manifest)
		if err != nil {
			return err
		} // .if
		sd.Metrics.IncUploadToDex()

		// other copies (files to router and/or edav), based on copy targets metadata config copyTargets
		for _, ct := range copyTargets {

			// ------------------------------------------------------------------
			// ct.Target == models.TARGET_DEX_ROUTER
			// ------------------------------------------------------------------
			if ct.Target == models.TARGET_DEX_ROUTER {

				err = sd.copyTusRouterWRetry(eventUploadComplete, dstBlobName, manifest)
				if err != nil {
					return err
				} // .if
				sd.Metrics.IncUploadToRouter()

			} // .if

			// ------------------------------------------------------------------
			// ct.Target == models.TARGET_DEX_EDAV
			// ------------------------------------------------------------------
			if ct.Target == models.TARGET_DEX_EDAV {

				err = sd.copyTusEdavWRetry(eventUploadComplete, dstBlobName, manifest)
				if err != nil {
					return err
				} // .if
				sd.Metrics.IncUploadToEdav()
			} // .if

		} // .for

	} // .RUN_MODE_AZURE

	// all good
	return nil
} // .OnUploadComplete

// copyTusDexWRetry copy file and metadata from tus to dex container
func (sd ServerDex) copyTusDexWRetry(eventUploadComplete tusd.HookEvent, dstBlobName string, manifest map[string]*string) error {

	logger := sd.logger.With(models.EVENT_UPLOAD_ID, eventUploadComplete.Upload.ID)

	copierDex := storeaz.CopierAzTusToDex{

		SrcTusAzBlobClient:    sd.HandlerDex.TusAzBlobClient,
		SrcTusAzContainerName: sd.AppConfig.TusAzStorageConfig.AzContainerName,
		SrcTusAzBlobName:      eventUploadComplete.Upload.ID,
		//
		DstAzContainerName: sd.AppConfig.DexAzStorageContainerName,
		DstAzBlobName:      dstBlobName,
		Manifest:           manifest,
	} // .copierDex

	for i := 0; i <= sd.AppConfig.CopyRetryTimes; i++ {

		err := copierDex.CopyTusSrcToDst()
		if i == sd.AppConfig.CopyRetryTimes && err != nil {
			logger.Error("error copy file tus to dex, retry times out", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
			return err
		} // .if

		if err != nil {
			logger.Error("error copy file tus to dex, should retry times", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)

			// try refresh the client on first retry
			if i == 1 {
				sd.HandlerDex.TusAzBlobClient, err = storeaz.NewTusAzBlobClient(sd.AppConfig)
				if err != nil {
					logger.Error("error copy file tus to dex, error refresh tus blob client", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
					// quit
					return err
				} // .if
			} // .if

			time.Sleep(time.Millisecond * time.Duration(sd.AppConfig.CopyRetryDelay))
		} else {
			logger.Info("file copied tus to dex with manifest", "retryLoopNum", i)
			break
		} // .else
	} // .for

	// all good
	return nil
} // .copyWRetry

// copyTusRouterWRetry copy file and metadata from tus to router
func (sd ServerDex) copyTusRouterWRetry(eventUploadComplete tusd.HookEvent, dstBlobName string, manifest map[string]*string) error {

	logger := sd.logger.With(models.EVENT_UPLOAD_ID, eventUploadComplete.Upload.ID)

	copierSrcToDst := storeaz.CopierAzSrcDst{

		SrcTusAzBlobClient:    sd.HandlerDex.TusAzBlobClient,
		SrcTusAzContainerName: sd.AppConfig.TusAzStorageConfig.AzContainerName,
		SrcTusAzBlobName:      eventUploadComplete.Upload.ID,
		//
		DstAzBlobClient:    sd.HandlerDex.RouterAzBlobClient,
		DstAzContainerName: sd.AppConfig.RouterAzStorageConfig.AzContainerName,
		DstAzBlobName:      dstBlobName,
		Manifest:           manifest,
	} // .copierDex

	for i := 0; i <= sd.AppConfig.CopyRetryTimes; i++ {

		err := copierSrcToDst.CopyAzSrcToDst()
		if i == sd.AppConfig.CopyRetryTimes && err != nil {
			logger.Error("error copy file dex to router, retry times out", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
			return err
		} // .if

		if err != nil {
			logger.Error("error copy file dex to router, should retry times", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)

			// try refresh the router client, the tus should be good from above copy
			if i == 1 {
				sd.HandlerDex.RouterAzBlobClient, err = storeaz.NewRouterAzBlobClient(sd.AppConfig)
				if err != nil {
					logger.Error("error copy file dex to router, error refresh router blob client", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
					// quit
					return err
				} // .if
			} // .if

			time.Sleep(time.Millisecond * time.Duration(sd.AppConfig.CopyRetryDelay))
		} else {
			sd.logger.Info("file copied dex to router", "retryLoopNum", i)
			break
		} // .else
	} // .for
	// all good
	return nil
} // .copyTusRouterWRetry

// copyTusEdavWRetry copy file and metadata from tus to edav
func (sd ServerDex) copyTusEdavWRetry(eventUploadComplete tusd.HookEvent, dstBlobName string, manifest map[string]*string) error {

	logger := sd.logger.With(models.EVENT_UPLOAD_ID, eventUploadComplete.Upload.ID)

	copierSrcToDst := storeaz.CopierAzSrcDst{

		SrcTusAzBlobClient:    sd.HandlerDex.TusAzBlobClient,
		SrcTusAzContainerName: sd.AppConfig.TusAzStorageConfig.AzContainerName,
		SrcTusAzBlobName:      eventUploadComplete.Upload.ID,
		//
		DstAzBlobClient:    sd.HandlerDex.EdavAzBlobClient,
		DstAzContainerName: sd.AppConfig.EdavAzStorageConfig.AzContainerName,
		DstAzBlobName:      dstBlobName,
		Manifest:           manifest,
	} // .copierDex

	for i := 0; i <= sd.AppConfig.CopyRetryTimes; i++ {

		err := copierSrcToDst.CopyAzSrcToDst()
		if i == sd.AppConfig.CopyRetryTimes && err != nil {
			logger.Error("error copy file dex to edav, retry times out", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
			return err
		} // .if

		if err != nil {
			logger.Error("error copy file dex to edav, should retry times", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)

			// try refresh the edav client, the tus should be good from above copy
			if i == 1 {
				sd.HandlerDex.EdavAzBlobClient, err = storeaz.NewEdavAzBlobClient(sd.AppConfig)
				if err != nil {
					logger.Error("error copy file dex to edav, error refresh edav blob client", "error", err, "retryLoopNum", i, "sd.AppConfig.CopyRetryTimes", sd.AppConfig.CopyRetryTimes)
					// quit
					return err
				} // .if
			} // .if

			time.Sleep(time.Millisecond * time.Duration(sd.AppConfig.CopyRetryDelay))
		} else {
			sd.logger.Info("file copied dex to edav", "retryLoopNum", i)
			break
		} // .else
	} // .for
	// all good
	return nil
} // .copyTusEdavWRetry

// getDstBlobName makes blob name from upload config including folder structure and adding time ticks
func getDstBlobName(eventUploadComplete tusd.HookEvent, uploadConfig metadatav1.UploadConfig, ingestDt time.Time) string {

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

	return dstBlobName
} // .getDstBlobName
