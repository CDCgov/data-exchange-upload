package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/server"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"

)



func main() {

	// TODO: structured logging, decide if slog is used and config at global level with default outputs
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))


	// TODO: context object, decide if custom slog is to be passed using the go context object

	flags := flags.ParseFlags()
	config := config.ParseConfig()

	var httpServer = server.New(flags, config)

	logger.Info("starting server", "port", config.ServerPort)

	go func() {
		err := httpServer.Serve()
		if err != nil {
			panic(fmt.Errorf("unable to listen: %s", err))
		} // .if
	}()

	// ---------------------------------
	// 			Block for Exit
	// ---------------------------------
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	// close connections, TODO if needed
	logger.Info("closing server by os signal", "port", config.ServerPort)



} // .main