package delivery

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

var ErrBadTarget = fmt.Errorf("bad delivery target")
var ErrSrcFileNotExist = fmt.Errorf("source file does not exist")

var destinations = map[string]Destination{}

func RegisterDestination(name string, d Destination) {
	destinations[name] = d
}

func GetDestination(name string) (Destination, bool) {
	d, ok := destinations[name]
	return d, ok
}

var sources = map[string]Source{}

func RegisterSource(name string, s Source) {
	sources[name] = s
}

func GetSource(name string) (Source, bool) {
	s, ok := sources[name]
	return s, ok
}

type Source interface {
	Reader(context.Context, string) (io.Reader, error)
	GetMetadata(context.Context, string) (map[string]string, error)
}

type Destination interface {
	Upload(context.Context, string, io.Reader, map[string]string) (string, error)
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	var src Source
	fromPathStr := appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix
	fromPath := os.DirFS(fromPathStr)
	src = &FileSource{
		FS: fromPath,
	}

	var edavDeliverer Destination
	edavDeliverer, err = NewFileDestination(ctx, "edav", &appConfig)
	if err != nil {
		return err
	}
	var routingDeliverer Destination
	routingDeliverer, err = NewFileDestination(ctx, "routing", &appConfig)
	if err != nil {
		return err
	}

	if appConfig.EdavConnection != nil {
		edavDeliverer, err = NewAzureDestination(ctx, "edav", &appConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to edav deliverer target %w", err)
		}
	}
	if appConfig.RoutingConnection != nil {
		routingDeliverer, err = NewAzureDestination(ctx, "routing", &appConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to routing deliverer target %w", err)
		}
	}

	if appConfig.AzureConnection != nil {
		// TODO Can the tus container client be singleton?
		tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return err
		}
		src = &AzureSource{
			FromContainerClient: tusContainerClient,
			Prefix:              appConfig.TusUploadPrefix,
		}
	}

	RegisterDestination("edav", edavDeliverer)
	RegisterDestination("routing", routingDeliverer)

	RegisterSource("upload", src)

	health.Register(edavDeliverer, routingDeliverer, src)

	return nil
}

// target may end up being a type
func Deliver(ctx context.Context, path string, s Source, d Destination) (string, error) {

	manifest, err := s.GetMetadata(ctx, path)
	if err != nil {
		return "", err
	}

	r, err := s.Reader(ctx, path)
	if err != nil {
		return "", err
	}
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}
	return d.Upload(ctx, path, r, manifest)
}
