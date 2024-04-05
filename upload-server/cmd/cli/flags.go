package cli

import (
	"errors"
	"flag"
	"slices"
) // .import

var Flags struct {
	RunMode          string // local, azure, aws
	AppConfigPath    string // if override
	UsePrebuiltHooks bool
	FileHooksDir     string
} // .flags

var runModes = []string{"local", "azure", "aws", "local_to_azure"}

const RUN_MODE_LOCAL = "local"
const RUN_MODE_LOCAL_TO_AZURE = "local_to_azure"
const RUN_MODE_AZURE = "azure"
const RUN_MODE_AWS = "aws"

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() error {

	flag.StringVar(&Flags.RunMode, "env", "local", "used to set app run mode: local, local_to_azure, azure, or aws")
	flag.StringVar(&Flags.AppConfigPath, "appconf", "./configs/local/local.env", "used to override the app configuration file path")
	flag.StringVar(&Flags.FileHooksDir, "file-hooks-dir", "", "the path to a directory containing file hooks")

	flag.Parse()

	if !slices.Contains(runModes, Flags.RunMode) {
		return errors.New("cli flag run mode not recognized")
	} // if

	return nil
} // .ParseFlags
