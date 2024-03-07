package cliflags

import (
	"errors"
	"flag"
	"slices"
) // .import

type Flags struct {
	Environment        string // local, azure, aws
	AppLocalConfigPath string // if wanting to override

} // .flags

var environments = []string{"local", "azure", "aws"}

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() (Flags, error) {

	env := flag.String("env", "local", "used to set app run environment: local, azure, or aws")
	lcp := flag.String("conf", "../configs/local.env", "used to override the configuration file path")

	flag.Parse()

	if !slices.Contains(environments, *env) {
		return Flags{}, errors.New("cli flag environment not recognized")
	} // if

	flags := Flags{
		Environment:        *env,
		AppLocalConfigPath: *lcp,
	} // .flags

	return flags, nil
} // .ParseFlags
