package server

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
)

type Server struct {
	flags flags.Flags
	config config.Config
}

func New(flags flags.Flags, config config.Config) Server {

	return Server {
		flags: flags,
		config: config,
	} // .return

} // New