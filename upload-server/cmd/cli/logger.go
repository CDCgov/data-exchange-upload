package cli

import (
	"log/slog"
	"os"

	expslog "golang.org/x/exp/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

var (
	logger *slog.Logger
)

// AppLogger, this is the custom application logger for uniformity
func AppLogger(appConfig appconfig.AppConfig) *slog.Logger {

	// Configure debug on if needed, otherwise should be off
	opts := &slog.HandlerOptions{
		AddSource: true,
	} // .opts

	if appConfig.LoggerDebugOn {
		opts.Level = slog.LevelDebug

	} // .if

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	appLogger := logger.With(
		slog.Group("app_info",
			slog.String("System", "DEX"),
			slog.String("Product", "UPLOAD API"),
			slog.String("App", "UPLOAD SERVER"),
			slog.String("Env", appConfig.Environment),
		)) // .appLogger

	return appLogger
} // .AppLogger

// AppLogger, used to config TUSD, this is the custom application logger for uniformity
// NOTE: currently tusd supports x/exp/slog and is moving to log/slog
// then this package should be removed and replaced by the app logger in sloger/sloger.go

func ExpAppLogger(appConfig appconfig.AppConfig) *expslog.Logger {

	// Configure debug on if needed, otherwise should be off
	opts := &expslog.HandlerOptions{
		AddSource: true,
	} // .opts

	if appConfig.LoggerDebugOn {
		opts.Level = expslog.LevelDebug

	} // .if

	logger := expslog.New(expslog.NewJSONHandler(os.Stdout, opts))

	appLogger := logger.With(
		expslog.Group("app_info",
			expslog.String("System", "DEX"),
			expslog.String("Product", "UPLOAD API"),
			expslog.String("App", "UPLOAD SERVER"),
			expslog.String("Env", appConfig.Environment),
		)) // .appLogger

	return appLogger
} // .AppLogger
