package serverdex

import (
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

// ServerDex, main Upload Api server, handles requests to both tusd handler and dex handler
type ServerDex struct {
	CliFlags  cli.Flags
	AppConfig appconfig.AppConfig
	MetaV1    *metadatav1.MetadataV1

	handlerTusd *tusd.Handler
	HandlerDex  *handlerdex.HandlerDex
	logger      *slog.Logger
	Metrics     Metrics
} // .ServerDex

// New returns an custom server for DEX Upload Api ready to serve
func New(cliFlags cli.Flags, appConfig appconfig.AppConfig, metaV1 *metadatav1.MetadataV1, psSender *processingstatus.PsSender) (ServerDex, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])

	handlerTusd, err := handlertusd.New(cliFlags, appConfig, psSender)
	if err != nil {
		logger.Error("error starting tusd handler")
		return ServerDex{}, err
	} // .handlerTusd

	handlerDex := handlerdex.New(cliFlags, appConfig, psSender)

	return ServerDex{
		CliFlags:    cliFlags,
		AppConfig:   appConfig,
		MetaV1:      metaV1,
		handlerTusd: handlerTusd,
		HandlerDex:  handlerDex,
		logger:      logger,
		Metrics:     newMetricsDex(),
	}, nil // .return

} // New

// HttpServer, adds the routes for the tusd and dex handlers and can customize the server with port address
func (sd *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------
	// Route for TUSD to start listening on and accept http request
	http.Handle(sd.AppConfig.TusdHandlerBasePath, http.StripPrefix(sd.AppConfig.TusdHandlerBasePath, sd.handlerTusd))

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {

			event := <-sd.handlerTusd.CompleteUploads
			sd.logger.Info("upload finished", models.EVENT_UPLOAD_ID, event.Upload.ID)

			// --------------------------------------------------------------
			// 	pulling from metadata v1 the upload config and copy targets for this event
			// --------------------------------------------------------------
			// TODO: meta_destination_id and meta_ext_event were checked in pre-check so they would be in metadata
			// TODO: could add ok check just in case ^
			uploadConfigKey := event.Upload.MetaData[models.META_DESTINATION_ID]
			uploadConfigKey += "-"
			uploadConfigKey += event.Upload.MetaData[models.META_EXT_EVENT]
			uploadConfig := sd.MetaV1.UploadConfigs[uploadConfigKey]
			hydrateV1Config := sd.MetaV1.HydrateV1ConfigsMap[uploadConfigKey]

			var copyTargets []metadatav1.CopyTarget

		copyTargetsFound:
			for _, v := range sd.MetaV1.AllowedDestAndEvents {
				if v.DestinationId == event.Upload.MetaData[models.META_DESTINATION_ID] {

					for _, ev := range v.ExtEvents {
						if ev.Name == event.Upload.MetaData[models.META_EXT_EVENT] {
							copyTargets = ev.CopyTargets
						}
						break copyTargetsFound
					} // .for
				} // .if
			} // .for

			err := sd.onUploadComplete(uploadConfig, hydrateV1Config, copyTargets, event)
			if err != nil {

				sd.logger.Error("error copy upload", "error", err)
				// TODO: DLQ
				// TODO: DLQ
				// TODO: DLQ

			} else {
				sd.logger.Info("upload copied", models.EVENT_UPLOAD_ID, event.Upload.ID)
			} // .else

		} // .for
	}() // .go func

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	sd.setupMetrics(sd.handlerTusd)
	hooks.SetupHookMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// --------------------------------------------------------------
	// 	DEX handler for all other http requests except above
	// --------------------------------------------------------------
	http.Handle("/", sd.HandlerDex)

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: ":" + sd.AppConfig.ServerPort,

		ConnState: func(_ net.Conn, cs http.ConnState) {
			switch cs {
			case http.StateNew:
				metricsOpenConnections.Inc()
			case http.StateClosed, http.StateHijacked:
				metricsOpenConnections.Dec()
			} // .switch
		},
		// etc...

	} // .httpServer
} // .HttpServer
