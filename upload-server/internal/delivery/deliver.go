package delivery

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/configs"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"gopkg.in/yaml.v3"
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
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

var ErrSrcFileNotExist = fmt.Errorf("source file does not exist")

var destinations = map[string]map[string]Destination{}

func RegisterDestination(name string, targetName string, d Destination) {
	if _, ok := destinations[name]; ok {
		destinations[name][targetName] = d
	} else {
		destinations[name] = map[string]Destination{targetName: d}
	}
}

func GetDestinationTargetNames(dataStreamId string, dataStreamRoute string) []string {
	targets, ok := destinations[dataStreamId+"-"+dataStreamRoute]
	if !ok {
		return []string{}
	}
	targetNames := make([]string, 0, len(targets))
	for t := range targets {
		targetNames = append(targetNames, t)
	}
	return targetNames
}

func GetDestinationTarget(dataStreamId string, dataStreamRoute string, target string) (Destination, bool) {
	d, ok := destinations[dataStreamId+"-"+dataStreamRoute][target]
	return d, ok
}

func getTargetHealthChecks() []any {
	targetSet := map[string]Destination{}
	var dests []any

	for _, destination := range destinations {
		for name, t := range destination {
			targetSet[name] = t
		}
	}

	for _, dest := range targetSet {
		dests = append(dests, dest)
	}

	return dests
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
	UploadId string
	Filename string
}

type Config struct {
	Programs []Program `yaml:"programs"`
}

type Program struct {
	DataStreamId    string   `yaml:"data_stream_id"`
	DataStreamRoute string   `yaml:"data_stream_route"`
	PathTemplate    string   `yaml:"path_template"`
	DeliveryTargets []Target `yaml:"delivery_targets"`
}

type Target struct {
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	PathTemplate string `yaml:"path_template"`
	Endpoint     string `yaml:"endpoint"`

	BucketName         string `yaml:"bucket_name"`
	AwsAccessKeyId     string `yaml:"access_key_id"`
	AwsSecretAccessKey string `yaml:"secret_access_key"`
	AwsRegion          string `yaml:"aws_region"`

	ContainerName    string `yaml:"container_name"`
	TenantID         string `yaml:"tenant_id"`
	ClientID         string `yaml:"client_id"`
	ClientSecret     string `yaml:"client_secret"`
	ConnectionString string `yaml:"connection_string"`
	SasToken         string `yaml:"sas_token"`

	Directory string `yaml:"directory"`
}

func (t Target) createS3Destination(ctx context.Context, _ appconfig.AppConfig) (Destination, error) {
	storageConfig := appconfig.S3StorageConfig{
		Endpoint:   t.Endpoint,
		BucketName: t.BucketName,
	}
	return NewS3Destination(ctx, t.Name, t.PathTemplate, &storageConfig)
}

func (t Target) createAzBlobDestination(ctx context.Context, _ appconfig.AppConfig) (Destination, error) {
	// TODO config from connection string and SAS token
	// see storeaz.azblobclientnew.NewContainerClient
	storageConfig := appconfig.AzureContainerConfig{
		AzureStorageConfig: appconfig.AzureStorageConfig{
			ContainerEndpoint: t.Endpoint,
			TenantId:          t.TenantID,
			ClientId:          t.ClientID,
			ClientSecret:      t.ClientSecret,
		},
		ContainerName: t.ContainerName,
	}
	return NewAzureDestination(ctx, t.Name, t.PathTemplate, &storageConfig)

}

func (t Target) createLocalDestination(ctx context.Context, appConfig appconfig.AppConfig) (Destination, error) {
	storageConfig := appconfig.LocalUploadStoreConfig(&appConfig)
	storageConfig.ToPath = t.Directory
	return NewFileDestination(ctx, t.Name, t.PathTemplate, storageConfig)
}

type opFunc = func(t Target, ctx context.Context, appConfig appconfig.AppConfig) (Destination, error)

var opMap = map[string]opFunc{
	"s3":      Target.createS3Destination,
	"az-blob": Target.createAzBlobDestination,
	"file":    Target.createLocalDestination,
}

func unmarshalDeliverYML() (*Config, error) {
	confStr := os.ExpandEnv(string(configs.DeliverYML))
	c := &Config{}

	err := yaml.Unmarshal([]byte(confStr), &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	var src Source
	fromPathStr := appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix
	fromPath := os.DirFS(fromPathStr)
	src = &FileSource{
		FS: fromPath,
	}

	cfg, err := unmarshalDeliverYML()
	if err != nil {
		return err
	}
	for _, p := range cfg.Programs {
		for _, t := range p.DeliveryTargets {
			if t.Name[0] == '-' {
				continue
			}
			if t.PathTemplate == "" {
				t.PathTemplate = p.PathTemplate
			}
			op, ok := opMap[t.Type]
			if ok {
				destination, err := op(t, ctx, appConfig)
				if err != nil {
					return err
				}
				name := p.DataStreamId + "-" + p.DataStreamRoute
				RegisterDestination(name, t.Name, destination)
			}
		}
	}

	targets := getTargetHealthChecks()
	if len(targets) == 0 {
		return fmt.Errorf("failed to register destination targets")
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
	RegisterSource("upload", src)

	targets = append(targets, src)
	if err := health.Register(targets...); err != nil {
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

func getDeliveredFilename(ctx context.Context, tuid string, pathTemplate string, manifest map[string]string) (string, error) {
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadataPkg.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	if pathTemplate != "" {
		// Use path template to form the full name.
		t := time.Now().UTC()
		pathInfo := &PathInfo{
			Year:     strconv.Itoa(t.Year()),
			Month:    strconv.Itoa(int(t.Month())),
			Day:      strconv.Itoa(t.Day()),
			Filename: filenameWithoutExtension,
			UploadId: tuid,
		}
		tmpl, err := template.New("path").Parse(pathTemplate)
		if err != nil {
			return "", err
		}
		b := new(bytes.Buffer)
		err = tmpl.Execute(b, pathInfo)
		if err != nil {
			return "", err
		}

		if extension != "" {
			return fmt.Sprintf("%s.%s", b.String(), extension), nil
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
