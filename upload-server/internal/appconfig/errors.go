package appconfig

import "fmt"

type MissingConfigError struct {
	ConfigName string
}

func (e *MissingConfigError) Error() string {
	return fmt.Sprintf("missing %s configuration value", e.ConfigName)
}
