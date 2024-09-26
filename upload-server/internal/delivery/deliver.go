package delivery

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"

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

type PathInfo struct {
	Year     string
	Month    string
	Day      string
	Hour     string
	UploadId string
	Filename string
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	var src Source
	fromPathStr := appConfig.LocalFolderUploadsTus
	fromPathStr = filepath.Join(fromPathStr, appConfig.TusUploadPrefix)

	fromPath := os.DirFS(fromPathStr)
	src = &FileSource{
		FS:   fromPath,
		Path: fromPathStr,
	}

	var edavDeliverer Destination
	edavDeliverer, err = NewFileDestination(ctx, appconfig.DeliveryTargetEdav, &appConfig)
	if err != nil {
		return err
	}
	var routingDeliverer Destination
	routingDeliverer, err = NewFileDestination(ctx, "routing", &appConfig)
	if err != nil {
		return err
	}
	var ehdiDeliverer Destination
	ehdiDeliverer, err = NewFileDestination(ctx, appconfig.DeliveryTargetEhdi, &appConfig)
	if err != nil {
		return err
	}
	var eicrDeliverer Destination
	eicrDeliverer, err = NewFileDestination(ctx, appconfig.DeliveryTargetEicr, &appConfig)
	if err != nil {
		return err
	}
	var ncirdDeliverer Destination
	ncirdDeliverer, err = NewFileDestination(ctx, appconfig.DeliveryTargetNcird, &appConfig)
	if err != nil {
		return err
	}

	if appConfig.EdavConnection != nil {
		edavDeliverer, err = NewAzureDestination(ctx, appconfig.DeliveryTargetEdav)
		if err != nil {
			return fmt.Errorf("failed to connect to edav deliverer target %w", err)
		}
	}
	if appConfig.RoutingConnection != nil {
		routingDeliverer, err = NewAzureDestination(ctx, "routing")
		if err != nil {
			return fmt.Errorf("failed to connect to routing deliverer target %w", err)
		}
	}
	if appConfig.EhdiConnection != nil {
		ehdiDeliverer, err = NewAzureDestination(ctx, appconfig.DeliveryTargetEhdi)
		if err != nil {
			return fmt.Errorf("failed to connect to ehdi deliverer target %w", err)
		}
	}
	if appConfig.EicrConnection != nil {
		eicrDeliverer, err = NewAzureDestination(ctx, appconfig.DeliveryTargetEicr)
		if err != nil {
			return fmt.Errorf("failed to connect to eicr deliverer target %w", err)
		}
	}
	if appConfig.NcirdConnection != nil {
		ncirdDeliverer, err = NewAzureDestination(ctx, appconfig.DeliveryTargetNcird)
		if err != nil {
			return fmt.Errorf("failed to connect to ncird deliverer target %w", err)
		}
	}

	if appConfig.EdavS3Connection != nil {
		edavDeliverer, err = NewS3Destination(ctx, appconfig.DeliveryTargetEdav, appConfig.EdavS3Connection)
		if err != nil {
			return fmt.Errorf("failed to connect to edav deliverer target %w", err)
		}
	}
	if appConfig.NcirdS3Connection != nil {
		ncirdDeliverer, err = NewS3Destination(ctx, appconfig.DeliveryTargetNcird, appConfig.NcirdS3Connection)
		if err != nil {
			return fmt.Errorf("failed to connect to ncird deliverer target %w", err)
		}
	}
	if appConfig.RoutingS3Connection != nil {
		routingDeliverer, err = NewS3Destination(ctx, "routing", appConfig.RoutingS3Connection)
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

	if appConfig.S3Connection != nil {
		s3Client, err := stores3.New(ctx, appConfig.S3Connection)
		if err != nil {
			return err
		}
		src = &S3Source{
			FromClient: s3Client,
			BucketName: appConfig.S3Connection.BucketName,
			Prefix:     appConfig.TusUploadPrefix,
		}
	}

	RegisterDestination(appconfig.DeliveryTargetEdav, edavDeliverer)
	RegisterDestination("routing", routingDeliverer)
	RegisterDestination(appconfig.DeliveryTargetEhdi, ehdiDeliverer)
	RegisterDestination(appconfig.DeliveryTargetEicr, eicrDeliverer)
	RegisterDestination(appconfig.DeliveryTargetNcird, ncirdDeliverer)

	RegisterSource("upload", src)

	if err := health.Register(edavDeliverer, routingDeliverer, ehdiDeliverer, eicrDeliverer, ncirdDeliverer, src); err != nil {
		slog.Error("failed to register some health checks", "error", err)
	}

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

func getDeliveredFilename(ctx context.Context, tuid string, manifest map[string]string) (string, error) {
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadataPkg.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	c, err := metadata.GetConfigFromManifest(ctx, manifest)

	if c.Copy.PathTemplate != "" {
		// Use path template to form the full name.
		t := time.Now().UTC()
		m := fmt.Sprintf("%02d", t.Month())
		d := fmt.Sprintf("%02d", t.Day())
		h := fmt.Sprintf("%02d", t.Hour())
		pathInfo := &PathInfo{
			Year:     strconv.Itoa(t.Year()),
			Month:    m,
			Day:      d,
			Hour:     h,
			Filename: filenameWithoutExtension,
			UploadId: tuid,
		}
		tmpl, err := template.New("path").Parse(c.Copy.PathTemplate)
		if err != nil {
			return "", err
		}
		b := new(bytes.Buffer)
		err = tmpl.Execute(b, pathInfo)
		if err != nil {
			return "", err
		}

		if extension != "" {
			return b.String() + extension, nil
		}

		return b.String(), nil
	}

	// Otherwise, use the suffix and folder structure values
	suffix, err := metadata.GetFilenameSuffix(ctx, manifest, tuid)
	if err != nil {
		return "", err
	}
	blobName := filenameWithoutExtension + suffix + extension

	// Next, need to set the filename prefix based on config and target.
	prefix := ""

	prefix, err = metadata.GetFilenamePrefix(ctx, manifest)
	if err != nil {
		return "", err
	}

	return prefix + "/" + blobName, nil
}
