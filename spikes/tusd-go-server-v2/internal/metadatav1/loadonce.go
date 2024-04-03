package metadatav1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

var metaV1Instance *MetadataV1

// avoids racing for metadata loads if program changes and LoadOnce is called from other places
var lock = &sync.Mutex{}

// LoadOnce metadata once from files, this should be called only once in main
// however it is ok to call LoadOnce multiple time since it handles concurrency and only (re) loads once
func LoadOnce(appConfig appconfig.AppConfig) (*MetadataV1, error) {

	// only to be called once from main
	// the if metaV1Instance == nil checks would not be needed if is this is only called once
	// but things can change and the method can handle multiple calls

	if metaV1Instance == nil { // check cheap once outside lock, this should usually be false because the metadata should be loaded

		lock.Lock()
		defer lock.Unlock()
		if metaV1Instance == nil { // more expensive check one more time inside lock because access to this point was not blocked

			logger := pkgLogger()

			// ----------------------------------------------------------------------
			// allowed destination and events
			// ----------------------------------------------------------------------
			allowedDestAndEventsPath, err := filepath.Abs(appConfig.AllowedDestAndEventsPath)
			if err != nil {
				logger.Error("error reading file from path", "AllowedDestAndEventsPath", appConfig.AllowedDestAndEventsPath)
				return nil, err
			} // .if

			fileContent, err := os.ReadFile(allowedDestAndEventsPath)
			if err != nil {
				logger.Error("error loading allowed destination and events from file", "error", err)
				return nil, err
			} // .err

			var allowedDestAndEvents []AllowedDestAndEvents
			err = json.Unmarshal(fileContent, &allowedDestAndEvents)
			if err != nil {
				logger.Error("error reading allowed destination and events into typed object", "error", err)
				return nil, err
			} // .err

			// ----------------------------------------------------------------------
			// all definitions and all upload configs
			// ----------------------------------------------------------------------
			allDefinitions := make(AllDefinitions)
			allUploadConfigs := make(AllUploadConfigs)
			destIdsEventsNameMap := make(DestIdsEventsNameMap)
			destIdEventFileNameMap := make(DestIdEventFileNameMap)
			hydrateV1ConfigMap := make(map[string]HydrateV1Config)

			for _, allowed := range allowedDestAndEvents {

				destIdsEventsNameMap[allowed.DestinationId] = []string{}

				extEvents := allowed.ExtEvents
				for _, event := range extEvents {

					destIdsEventsNameMap[allowed.DestinationId] = append(destIdsEventsNameMap[allowed.DestinationId], event.Name)
					destIdEventFileNameMap[allowed.DestinationId+event.Name] = event.DefinitionFileName

					// ----------------------------------------------------------------------
					// definitions, in v1 there is only one definition schema for each destination-event
					// ----------------------------------------------------------------------
					defFilePath, err := filepath.Abs(appConfig.DefinitionsPath + event.DefinitionFileName)
					if err != nil {
						logger.Error("error reading definition file from path", "DefinitionsPath", appConfig.DefinitionsPath)
						return nil, err
					} // .if

					defFileContent, err := os.ReadFile(defFilePath)
					if err != nil {
						logger.Error("error loading definition from file", "error", err, "defFilePath", defFilePath)
						return nil, err
					} // .err

					var definition []Definition
					err = json.Unmarshal(defFileContent, &definition)
					if err != nil {
						logger.Error("error unmarshal to json definition from file", "error", err, "defFilePath", defFilePath)
						return nil, err
					} // .err

					allDefinitions[event.DefinitionFileName] = definition

					// ----------------------------------------------------------------------
					// upload configs
					// ----------------------------------------------------------------------
					updConfFileName := allowed.DestinationId + "-" + event.Name
					updConfFileNameExt := updConfFileName + ".json"
					updConfFilePath, err := filepath.Abs(appConfig.UploadConfigPath + updConfFileNameExt)
					if err != nil {
						logger.Error("error reading file from path", "UploadConfigPath", appConfig.UploadConfigPath)
						return nil, err
					} // .if

					updConfFileContent, err := os.ReadFile(updConfFilePath)
					if err != nil {
						logger.Error("error loading upload config from file", "error", err, "updConfFileNameExt", updConfFileNameExt)
						return nil, err
					} // .err

					var updConfig UploadConfig
					err = json.Unmarshal(updConfFileContent, &updConfig)
					if err != nil {
						logger.Error("error unmarshal to json upload config from file", "error", err, "updConfFileNameExt", updConfFileNameExt)
						return nil, err
					} // .err

					allUploadConfigs[updConfFileName] = updConfig

					// ----------------------------------------------------------------------
					// hydrate v1 to v2 there is only one definition schema for each destination-event
					// ----------------------------------------------------------------------
					hydrateFilePath, err := filepath.Abs(appConfig.HydrateV1ConfigPath + updConfFileNameExt)
					if err != nil {
						logger.Error("error reading hydrate file from path", "HydrateV1ConfigPath", appConfig.HydrateV1ConfigPath, "error", err)
						return nil, err
					} // .if

					hydrateFileContent, err := os.ReadFile(hydrateFilePath)
					if err != nil {
						logger.Error("error loading hydrate config from file", "error", err, "hydrateFilePath", hydrateFilePath)
						return nil, err
					} // .err

					var hydrateV1Conf HydrateV1Config
					err = json.Unmarshal(hydrateFileContent, &hydrateV1Conf)
					if err != nil {
						logger.Error("error unmarshal to json hydrate config from file", "error", err, "hydrateFilePath", hydrateFilePath)
						return nil, err
					} // .err
					// use key same as upload config destination-event
					hydrateV1ConfigMap[updConfFileName] = hydrateV1Conf

				} // .for
			} // .for

			// ----------------------------------------------------------------------
			// set instance of loaded metadata
			// ----------------------------------------------------------------------
			metaV1Instance = &MetadataV1{

				AllowedDestAndEvents:   allowedDestAndEvents,
				Definitions:            allDefinitions,
				UploadConfigs:          allUploadConfigs,
				DestIdsEventsNameMap:   destIdsEventsNameMap,
				DestIdEventFileNameMap: destIdEventFileNameMap,
				HydrateV1ConfigsMap:    hydrateV1ConfigMap,
			} // .metaV1Instance

			logger.Info("loaded metadata v1")
		} // .if

	} // .if

	return metaV1Instance, nil
} // .Get
