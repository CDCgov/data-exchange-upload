package metadatav1

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

var metaV1Instance *MetadataV1

// avoids racing for metadata loads if program changes
var lock = &sync.Mutex{}

// load metadata once from files, this should be called only once in main
func LoadOnce(appConfig appconfig.AppConfig) (*MetadataV1, error) {

	// only to be called once from main
	// if metaV1Instance == nil checks would not be needed if is this is only called once but things can change

	if metaV1Instance == nil { // check cheap once outside lock, this should usually be false

		lock.Lock()
		defer lock.Unlock()
		if metaV1Instance == nil { // check one more time inside lock because this can be accessed concurrent

			logger := pkgLogger(appConfig)

			// ----------------------------------------------------------------------
			// allowed destination and events
			// ----------------------------------------------------------------------

			fileContent, err := os.ReadFile(appConfig.AllowedDestAndEventsPath)
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


			for _, allowed := range allowedDestAndEvents {

				destIdsEventsNameMap[allowed.DestinationId] = []string{}

				extEvents := allowed.ExtEvents
				for _, event := range extEvents {

					destIdsEventsNameMap[allowed.DestinationId] = append(destIdsEventsNameMap[allowed.DestinationId], event.Name)
					destIdEventFileNameMap[allowed.DestinationId + event.Name] = event.DefinitionFileName

					// ----------------------------------------------------------------------
					// definitions, in v1 there is only one definition schema for each destination-event
					// ----------------------------------------------------------------------
					defFilePath := appConfig.DefinitionsPath + event.DefinitionFileName

					defFileContent, err := os.ReadFile(defFilePath)
					if err != nil {
						logger.Error("error loading definition from file", "error", err, "event.DefinitionFileName", event.DefinitionFileName)
						return nil, err
					} // .err
					
					var definition []Definition
					err = json.Unmarshal(defFileContent, &definition)
					if err != nil {
						logger.Error("error unmarshal to json definition from file", "error", err, "event.DefinitionFileName", event.DefinitionFileName)
						return nil, err
					} // .err

					allDefinitions[event.DefinitionFileName] = definition

					// ----------------------------------------------------------------------
					// upload configs
					// ----------------------------------------------------------------------
					updConfFileName := allowed.DestinationId + "-" + event.Name
					updConfFileNameExt := updConfFileName + ".json"
					updConfFilePath := appConfig.UploadConfigPath + updConfFileNameExt

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

				} // .for
			} // .for

			// ----------------------------------------------------------------------
			// set instance of loaded metadata
			// ----------------------------------------------------------------------
			metaV1Instance = &MetadataV1{

				AllowedDestAndEvents: allowedDestAndEvents,
				Definitions:          allDefinitions,
				UploadConfigs:        allUploadConfigs,
				DestIdsEventsNameMap: destIdsEventsNameMap,
				DestIdEventFileNameMap: destIdEventFileNameMap, 

			} // .metaV1Instance

			logger.Info("loaded metadata v1")
		} // .if

	} // .if

	return metaV1Instance, nil
} // .Get
