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

func ParseConfig() (Config, error) {// TODO: does this need to return and error, if not refactor signature and call

	return Config{
		ServerPort: ":8080", //TODO dynamic from env/file
	}, nil 

} // .ParseConfig