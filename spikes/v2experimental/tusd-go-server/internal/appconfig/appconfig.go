package appconfig

type AppConfig struct {
	System     string
	DexProduct string
	DexApp     string

	LoggerDebugOn bool

	ServerPort string

	MetadataVersions string

	AllowedDestAndEventsPath string
	DefinitionsPath          string
	UploadConfigPath         string

	// TODO

	// AzStorage                        string
	// AzContainerAccessType            string
	// AzBlobAccessTier                 string
	// AzObjectPrefix                   string
	// AzEndpoint                       string

} // .AppConfig

func ParseConfig() (AppConfig, error) { // TODO: does this need to return and error, if not refactor signature and call

	return AppConfig{
		System:                   "DEX",            //TODO dynamic from config env/file
		DexProduct:               "Upload API",     //TODO dynamic from config env/file
		DexApp:                   "tusd-go-server", //TODO dynamic from config env/file
		ServerPort:               ":8080",          //TODO dynamic from config env/file
		MetadataVersions:         "[v1]",           //TODO dynamic from config env/file
		AllowedDestAndEventsPath: "../../../../tus/file-hooks/metadata-verify/allowed_destination_and_events.json",
		DefinitionsPath:          "../../../../tus/file-hooks/metadata-verify/",
		UploadConfigPath:         "../../../../upload-configs/",
	}, nil

} // .ParseConfig
