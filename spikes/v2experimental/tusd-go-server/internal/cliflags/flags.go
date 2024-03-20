package cliflags

import (
	"errors"
	"flag"
	"slices"
) // .import

type Flags struct {
	RunMode            string // local, azure, aws
	AppLocalConfigPath string // if override
	AzLocalConfigPath  string // if override

} // .flags

var runModes = []string{"local", "azure", "aws", "local_to_azure"}

const ENV_LOCAL = "local"
const ENV_LOCAL_TO_AZURE = "local_to_azure"
const ENV_AZURE = "azure"
const ENV_AWS = "aws"

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() (Flags, error) {

	runMode := flag.String("env", "local", "used to set app run mode: local, local_to_azure, azure, or aws")
	appLcp := flag.String("appconf", "./configs/local/local.env", "used to override the app configuration file path")
	azLcp := flag.String("azconf", "./configs/local/az.env", "used to override the azure configuration file path")

	flag.Parse()

	if !slices.Contains(runModes, *runMode) {
		return Flags{}, errors.New("cli flag run mode not recognized")
	} // if

	flags := Flags{
		RunMode:            *runMode,
		AppLocalConfigPath: *appLcp,
		AzLocalConfigPath:  *azLcp,
	} // .flags

	return flags, nil
} // .ParseFlags
