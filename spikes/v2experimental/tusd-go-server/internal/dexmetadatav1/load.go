package dexmetadatav1

import (
	"os"
	"encoding/json"
	"log/slog"
) // .import


func Load() (ConfigMetaV1, error) {

	// TODO move to singleton, only read once at entry in main

	// TODO: get file path from config
	content, err := os.ReadFile("../../../../tus/file-hooks/metadata-verify/allowed_destination_and_events.json")
	if err != nil {
		slog.Error("dexmetadatav1, error reading allowed_destination_and_events file", "error", err)
		return ConfigMetaV1{}, err  
	} // .err

	var destAndEvents []AllowedDestAndEvents
	err = json.Unmarshal(content, &destAndEvents)
	if err != nil {
		slog.Error("dexmetadatav1, error reading allowed_destination_and_events into typed object", "error", err)
		return ConfigMetaV1{}, err 
	} // .err

	return ConfigMetaV1{
		AllowedDestAndEvents: destAndEvents,
	}, nil

} // .Load