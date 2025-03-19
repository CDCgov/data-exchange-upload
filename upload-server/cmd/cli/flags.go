package cli

import (
	"flag"
) // .import

var Flags struct {
	AppConfigPath string // if override
} // .flags

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() error {

	flag.StringVar(&Flags.AppConfigPath, "appconf", "", "used to override the app configuration file path")

	flag.Parse()

	return nil
} // .ParseFlags
