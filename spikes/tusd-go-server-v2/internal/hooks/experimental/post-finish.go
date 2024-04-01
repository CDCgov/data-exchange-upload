package experimental

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func LocalPostProcess(event handler.HookEvent) (res hooks.HookResponse, err error) {

	logger := sloger.DefaultLogger
	logger.Info("upload finished", models.EVENT_UPLOAD_ID, event.Upload.ID)

	MetaV1, _ := metadatav1.LoadOnce(appconfig.AppConfig{})
	// --------------------------------------------------------------
	// 	pulling from metadata v1 the upload config and copy targets for this event
	// --------------------------------------------------------------
	// TODO: meta_destination_id and meta_ext_event were checked in pre-check so they would be in metadata
	// TODO: could add ok check just in case ^
	uploadConfigKey := event.Upload.MetaData[models.META_DESTINATION_ID]
	uploadConfigKey += "-"
	uploadConfigKey += event.Upload.MetaData[models.META_EXT_EVENT]
	uploadConfig := MetaV1.UploadConfigs[uploadConfigKey]
	hydrateV1Config := MetaV1.HydrateV1ConfigsMap[uploadConfigKey]

	copyTargets := getCopyTargets(MetaV1.AllowedDestAndEvents, event.Upload.MetaData)

	err = onLocalUploadComplete(uploadConfig, hydrateV1Config, copyTargets, event)
	return res, err
}

func getCopyTargets(allowed []metadatav1.AllowedDestAndEvents, metadata map[string]string) []metadatav1.CopyTarget {
	for _, v := range allowed {
		if v.DestinationId == metadata[models.META_DESTINATION_ID] {

			for _, ev := range v.ExtEvents {
				if ev.Name == metadata[models.META_EXT_EVENT] {
					return ev.CopyTargets
				}
			} // .for
		} // .if
	} // .for
	return nil
}

func (az *AzureUploadCompleteHandler) AzurePostProcess(event handler.HookEvent) (res hooks.HookResponse, err error) {
	logger := sloger.DefaultLogger
	logger.Info("upload finished", models.EVENT_UPLOAD_ID, event.Upload.ID)

	MetaV1, _ := metadatav1.LoadOnce(appconfig.AppConfig{})
	// --------------------------------------------------------------
	// 	pulling from metadata v1 the upload config and copy targets for this event
	// --------------------------------------------------------------
	// TODO: meta_destination_id and meta_ext_event were checked in pre-check so they would be in metadata
	// TODO: could add ok check just in case ^
	uploadConfigKey := event.Upload.MetaData[models.META_DESTINATION_ID]
	uploadConfigKey += "-"
	uploadConfigKey += event.Upload.MetaData[models.META_EXT_EVENT]
	uploadConfig := MetaV1.UploadConfigs[uploadConfigKey]
	hydrateV1Config := MetaV1.HydrateV1ConfigsMap[uploadConfigKey]

	copyTargets := getCopyTargets(MetaV1.AllowedDestAndEvents, event.Upload.MetaData)

	err = az.onAzureUploadComplete(uploadConfig, hydrateV1Config, copyTargets, event)
	return res, err

}
