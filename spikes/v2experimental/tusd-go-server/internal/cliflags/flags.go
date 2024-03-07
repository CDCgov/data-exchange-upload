package cliflags

import (
	"errors"
	"flag"
	"slices"
) // .import

type Flags struct {
	Environment string // local, azure, aws

} // .flags

var environments = []string{"local", "azure", "aws"}

// ParseFlags read cli flags into an Flags struct which is returned
func ParseFlags() (Flags, error) {

	env := flag.String("env", "local", "used to set app run environment local, azure, aws")

	flag.Parse()

	if !slices.Contains(environments, *env) {
		return Flags{}, errors.New("cli flag environment not recognized")
	} // if

	flags := Flags{
		Environment: *env,
	} // .flags

	return flags, nil
} // .ParseFlags
