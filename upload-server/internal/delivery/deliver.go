package delivery

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

var destinations sync.Map //= map[string]Destination{}

func RegisterDestination(name string, d Destination) {
	destinations.Store(name, d) //destinations[name] = d
}

func GetDestination(name string) (Destination, bool) {
	value, ok := destinations.Load(name) //destinations[name]
	if ok {
		return value.(Destination), ok
	}
	return nil, ok
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

//go:embed delivery.yml
var targetsFile []byte
var config Targets

type S3Target struct {
	Name       string `yaml:"name"`
	Endpoint   string `yaml:"endpoint"`
	BucketName string `yaml:"bucket_name"`
}
type AzTarget struct {
	Name       string `yaml:"name"`
	Endpoint   string `yaml:"endpoint"`
	BucketName string `yaml:"bucket_name"`
}
type Targets struct {
	S3    []S3Target `yaml:"s3"`
	Azure []AzTarget `yaml:"azure"`
}

// register destination from postprocessor.go
func PostRegisterDestination(ctx context.Context, target string) (Destination, error) {
	for _, t := range config.S3 {
		if t.Name == target {
			d, err := createS3Destination(ctx, t)
			if err == nil {
				RegisterDestination(t.Name, d)
			}
			return d, err
		}
	}
	for _, t := range config.Azure {
		if t.Name == target {
			d, err := createAzureDestination(ctx, t)
			if err == nil {
				RegisterDestination(t.Name, d)
			}
			return d, err
		}
	}
	return nil, fmt.Errorf("failed to register destination for %s", target)
}
func createS3Destination(ctx context.Context, target S3Target) (Destination, error) {
	storageConfig := appconfig.S3StorageConfig{
		Endpoint:   target.Endpoint,
		BucketName: target.BucketName,
	}

	deliverer, err := NewS3Destination(ctx, target.Name, &storageConfig)
	return deliverer, err
}
func registerS3Destinations(ctx context.Context, appConfig appconfig.AppConfig, registrations *[]interface{}, config Targets) error {
	var d Destination
	var err error
	for _, t := range config.S3 {
		if appConfig.RunLocal == nil {
			d, err = createS3Destination(ctx, t)
		} else {
			d, err = NewFileDestination(ctx, t.Name, &appConfig)
		}
		if err == nil {
			*registrations = append(*registrations, d)
			RegisterDestination(t.Name, d)
		}
	}
	return nil
}

// TODO
func createAzureDestination(ctx context.Context, target AzTarget) (Destination, error) {
	return nil, nil
}
func registerAzureDestinations(ctx context.Context, appConfig appconfig.AppConfig, registrations *[]interface{}, config Targets) error {
	for _, _ = range config.Azure {
	}
	return nil
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	var src Source
	fromPathStr := appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix
	fromPath := os.DirFS(fromPathStr)
	src = &FileSource{
		FS: fromPath,
	}

	// replace $environment variables
	conf := os.ExpandEnv(string(targetsFile))
	if err := yaml.Unmarshal([]byte(conf), &config); err != nil {
		return err
	}

	var registrations []interface{}
	if err := registerS3Destinations(ctx, appConfig, &registrations, config); err != nil {
		return err
	}
	if err := registerAzureDestinations(ctx, appConfig, &registrations, config); err != nil {
		return err
	}
	if len(registrations) == 0 {
		return fmt.Errorf("failed to register destinations")
	}

	/*
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
	*/
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
	/*
		RegisterDestination(appconfig.DeliveryTargetEdav, edavDeliverer)
		RegisterDestination("routing", routingDeliverer)
		RegisterDestination(appconfig.DeliveryTargetEhdi, ehdiDeliverer)
		RegisterDestination(appconfig.DeliveryTargetEicr, eicrDeliverer)
		RegisterDestination(appconfig.DeliveryTargetNcird, ncirdDeliverer)
	*/
	RegisterSource("upload", src)
	registrations = append(registrations, src)

	//if err := health.Register(edavDeliverer, routingDeliverer, ehdiDeliverer, eicrDeliverer, ncirdDeliverer, src); err != nil {
	if err := health.Register(registrations...); err != nil {
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
