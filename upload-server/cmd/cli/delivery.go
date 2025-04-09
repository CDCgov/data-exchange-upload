package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/prometheus/client_golang/prometheus"
)

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	delivery.Targets = make(map[string]delivery.Destination)
	delivery.Groups = make(map[string]delivery.Group)
	var src delivery.Source

	fromPathStr := filepath.Join(appConfig.LocalFolderUploadsTus, appConfig.TusUploadPrefix)
	fromPath := os.DirFS(fromPathStr)
	src = &delivery.FileSource{
		FS: fromPath,
	}

	dat, err := os.ReadFile(appConfig.DeliveryConfigFile)
	if err != nil {
		return err
	}
	cfg, err := delivery.UnmarshalDeliveryConfig(string(dat))
	if err != nil {
		return err
	}

	for _, t := range cfg.Targets {
		delivery.Targets[t.Name] = t.Destination
		if err := health.Register(t.Destination); err != nil {
			slog.Error("failed to register destination", "destination", t)
		}
		metrics.ActiveDeliveries.With(prometheus.Labels{"target": t.Name}).Set(0)
	}
	slog.Info("registering destinations", "targets", delivery.Targets)

	for _, g := range cfg.Groups {
		delivery.Groups[g.Key()] = g
		if g.DeliveryTargets == nil {
			slog.Warn(fmt.Sprintf("no targets configured for group %s", g.Key()))
		}
	}

	if appConfig.AzureConnection != nil {
		// TODO Can the tus container client be singleton?
		tusContainerClient, err := storeaz.NewContainerClient(appConfig.AzureConnection.Credentials(), appConfig.AzureUploadContainer)
		if err != nil {
			return err
		}
		src = &delivery.AzureSource{
			FromContainerClient: tusContainerClient,
			Prefix:              appConfig.TusUploadPrefix,
		}
	}
	if appConfig.S3Connection != nil {
		s3Client, err := stores3.NewWithEndpoint(ctx, appConfig.S3Connection.Endpoint)
		if err != nil {
			return err
		}
		src = &delivery.S3Source{
			FromClient: s3Client,
			BucketName: appConfig.S3Connection.BucketName,
			Prefix:     appConfig.TusUploadPrefix,
		}
	}
	delivery.RegisterSource(delivery.UploadSrc, src)

	if err := health.Register(src); err != nil {
		slog.Error("failed to register some health checks", "error", err)
	}
	return nil
}
