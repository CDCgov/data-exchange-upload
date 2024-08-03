package metadata

func GetFilename(manifest map[string]string) string {

	keys := []string{
		"filename",
		"original_filename",
		"meta_ext_filename",
		"received_filename",
	}

	for _, key := range keys {
		if name, ok := manifest[key]; ok {
			return name
		}
	}
	return ""
}

func GetSenderId(manifest map[string]string) string {
	switch manifest["version"] {
	case "2.0":
		return manifest["sender_id"]
	default:
		return manifest["meta_ext_source"]
	}
}

func GetDataStreamID(manifest map[string]string) string {
	switch manifest["version"] {
	case "2.0":
		return manifest["data_stream_id"]
	default:
		return manifest["meta_destination_id"]
	}
}

func GetDataStreamRoute(manifest map[string]string) string {
	switch manifest["version"] {
	case "2.0":
		return manifest["data_stream_route"]
	default:
		return manifest["meta_ext_event"]
	}

}

func GetJurisdiction(manifest map[string]string) string {
	return manifest["jurisdiction"]
}

func GetDexIngestDatetime(manifest map[string]string) string {
	return manifest["dex_ingest_datetime"]
}
