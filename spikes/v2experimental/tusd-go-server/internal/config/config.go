package config

type Config struct {
	
	ServerPort string

	// TODO

	// AzStorage                        string
	// AzContainerAccessType            string
	// AzBlobAccessTier                 string
	// AzObjectPrefix                   string
	// AzEndpoint                       string

} // .flags

func ParseConfig() Config {

	return Config{
		ServerPort: ":8080", //TODO dynamic from env/file
	}

} // .ParseConfig