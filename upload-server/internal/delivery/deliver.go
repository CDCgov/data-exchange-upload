package delivery

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
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
	Upload(context.Context, string, io.Reader, map[string]string) error
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllTargets(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
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
		//health.Register(edavDeliverer)
	}
	if appConfig.RoutingConnection != nil {
		routingDeliverer, err = NewAzureDestination(ctx, "routing", &appConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to routing deliverer target %w", err)
		}
		//health.Register(routingDeliverer)
	}

	if appConfig.AzureConnection != nil {
		// TODO Can the tus container client be singleton?
		tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return err
		}
		src = &AzureSource{
			FromContainerClient: tusContainerClient,
			TusPrefix:           appConfig.TusUploadPrefix,
		}
	}

	RegisterDestination("edav", edavDeliverer)
	RegisterDestination("routing", routingDeliverer)

	RegisterSource("upload", src)

	return nil
}

// target may end up being a type
func Deliver(ctx context.Context, path string, s Source, d Destination) error {

	manifest, err := s.GetMetadata(ctx, path)
	if err != nil {
		return err
	}

	//TODO pull reports up if we can
	rb := reports.NewBuilder[reports.FileCopyContent](
		"1.0.0",
		reports.StageFileCopy,
		path,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC())

	rb.SetManifest(manifest)

	rb.SetContent(reports.FileCopyContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageFileCopy,
		},
		//FileSourceBlobUrl:      srcUrl,
		//FileDestinationBlobUrl: destUrl,
	})

	defer func() {
		rb.SetEndTime(time.Now().UTC())
		if err != nil {
			rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
				Level:   reports.IssueLevelError,
				Message: err.Error(),
			})
		}
		report := rb.Build()
		reports.Publish(ctx, report)
	}()

	//NOTE could use ctx to store the manifest
	r, err := s.Reader(ctx, path)
	if err != nil {
		return err
	}
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}
	if err := d.Upload(ctx, path, r, manifest); err != nil {
		return err
	}

	return nil
}

type Deliverer interface {
	health.Checkable
	Deliver(ctx context.Context, tuid string, metadata map[string]string) error
	GetMetadata(ctx context.Context, tuid string) (map[string]string, error)
	GetSrcUrl(ctx context.Context, tuid string) (string, error)
	GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error)
}
