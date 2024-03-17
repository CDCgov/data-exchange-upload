package storecopier

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

// FromTusToDstEvent has configuration and event info needed to copy files from tusd to next folder
type FromTusToDstEvent struct {
	AppConfig       appconfig.AppConfig
	AllUploadConfig metadatav1.AllUploadConfigs

	EventUploadComplete tusd.HookEvent
} // .StoreLocal
