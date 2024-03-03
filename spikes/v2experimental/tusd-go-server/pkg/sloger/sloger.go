package sloger

import (
	"log/slog"
	"os"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
)


func AppLogger(appConfig appconfig.AppConfig) *slog.Logger{

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	appLogger := logger.With(		
		slog.Group("app_info",
			slog.String("System", appConfig.System),
			slog.String("Product", appConfig.DexProduct),
			slog.String("App", appConfig.DexApp),
	))

	return appLogger
} // .AppLogger

