package cliflags

import (
	"errors"
	"flag"
	"slices"
) // .import

type Flags struct {
	Environment        string // local, azure, aws
	AppLocalConfigPath string // if override
	AzLocalConfigPath  string // if override

} // .flags

var environments = []string{"local", "azure", "aws"}

const ENV_LOCAL = "local"
const ENV_AZURE = "azure"
const ENV_AWS = "aws"

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() (Flags, error) {

	env := flag.String("env", "local", "used to set app run environment: local, azure, or aws")
	appLcp := flag.String("appconf", "../configs/local/local.env", "used to override the app configuration file path")
	azLcp := flag.String("azconf", "../configs/local/az.env", "used to override the azure configuration file path")

	flag.Parse()

	if !slices.Contains(environments, *env) {
		return Flags{}, errors.New("cli flag environment not recognized")
	} // if

	flags := Flags{
		Environment:        *env,
		AppLocalConfigPath: *appLcp,
		AzLocalConfigPath:  *azLcp,
	} // .flags

	return flags, nil
} // .ParseFlags
