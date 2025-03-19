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

func GetJurisdiction(manifest map[string]string) string {
	return manifest["jurisdiction"]
}

func GetDexIngestDatetime(manifest map[string]string) string {
	return manifest["dex_ingest_datetime"]
}
