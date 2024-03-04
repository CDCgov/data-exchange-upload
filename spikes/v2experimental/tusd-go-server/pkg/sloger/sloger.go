package sloger

import (
	"log/slog"
	"os"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
)

func AppLogger(appConfig appconfig.AppConfig) *slog.Logger {

	// Configure debug on if needed, otherwise should be off
	var opts *slog.HandlerOptions

	if appConfig.LoggerDebugOn {

		opts = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		} // .opts
	} // .if

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	appLogger := logger.With(
		slog.Group("app_info",
			slog.String("System", appConfig.System),
			slog.String("Product", appConfig.DexProduct),
			slog.String("App", appConfig.DexApp),
		)) // .appLogger

	return appLogger
} // .AppLogger
