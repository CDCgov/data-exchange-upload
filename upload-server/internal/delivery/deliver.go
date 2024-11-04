package delivery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/tus/tusd/v2/pkg/handler"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"gopkg.in/yaml.v3"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

const storageTypeLocalFile string = "file"
const storageTypeAzureBlob string = "az-blob"
const storageTypeS3 string = "s3"

var UploadSrc = "upload"

var ErrSrcFileNotExist = fmt.Errorf("source file does not exist")

var groups map[string]Group
var targets map[string]Destination

type CloudSource interface {
	GetSignedObjectURL(ctx context.Context, containerName string, objectPath string) (string, error)
	Container() string
	Source
}

type Source interface {
	Reader(context.Context, string) (io.Reader, error)
	GetMetadata(context.Context, string) (map[string]string, error)
	GetSourceFilePath(string) string
	SourceType() string
}

type Destination interface {
	Copy(ctx context.Context, id string, path string, source *Source, metadata map[string]string, length int64,
		concurrency int) (string, error)
	DestinationType() string
}

func GetTarget(target string) (Destination, bool) {
	d, ok := targets[target]
	return d, ok
}

func FindGroupFromMetadata(meta handler.MetaData) (Group, bool) {
	dataStreamId, dataStreamRoute := metadataPkg.GetDataStreamID(meta), metadataPkg.GetDataStreamRoute(meta)
	group := Group{
		DataStreamId:    dataStreamId,
		DataStreamRoute: dataStreamRoute,
	}
	g, ok := groups[group.Key()]
	return g, ok
}

var sources = map[string]Source{}

func RegisterSource(name string, s Source) {
	sources[name] = s
}

func GetSource(name string) (Source, bool) {
	s, ok := sources[name]
	return s, ok
}

type PathInfo struct {
	Year            string
	Month           string
	Day             string
	Hour            string
	UploadId        string
	Filename        string
	Suffix          string
	Prefix          string
	DataStreamID    string
	DataStreamRoute string
}

type Config struct {
	Targets map[string]Target `yaml:"targets"`
	Groups  []Group           `yaml:"routing_groups"`
}

type Group struct {
	DataStreamId    string              `yaml:"data_stream_id"`
	DataStreamRoute string              `yaml:"data_stream_route"`
	DeliveryTargets []TargetDesignation `yaml:"delivery_targets"`
}

func (g *Group) Key() string {
	return g.DataStreamId + "_" + g.DataStreamRoute
}

func (g *Group) TargetNames() []string {
	names := make([]string, len(g.DeliveryTargets))
	for i, t := range g.DeliveryTargets {
		names[i] = t.Name
	}

	return names
}

type TargetDesignation struct {
	Name         string `yaml:"name"`
	PathTemplate string `yaml:"path_template"`
}

type Target struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Destination Destination `yaml:"-"`
}

var DestinationTypes = map[string]func() Destination{
	storageTypeS3:        func() Destination { return &S3Destination{} },
	storageTypeLocalFile: func() Destination { return &FileDestination{} },
	storageTypeAzureBlob: func() Destination { return &AzureDestination{} },
}

var ErrUnknownDestinationType = errors.New("unknown destination type")

func (t *Target) UnmarshalYAML(n *yaml.Node) error {
	type alias Target
	if err := n.Decode((*alias)(t)); err != nil {
		return err
	}
	dType, ok := DestinationTypes[t.Type]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownDestinationType, t.Type)
	}
	d := dType()
	if err := n.Decode(d); err != nil {
		return err
	}
	t.Destination = d
	return nil
}

func unmarshalDeliveryConfig(confBody string) (*Config, error) {
	confStr := os.ExpandEnv(confBody)
	c := &Config{}

	err := yaml.Unmarshal([]byte(confStr), &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
	targets = make(map[string]Destination)
	groups = make(map[string]Group)
	var src Source

	fromPathStr := filepath.Join(appConfig.LocalFolderUploadsTus, appConfig.TusUploadPrefix)
	fromPath := os.DirFS(fromPathStr)
	src = &FileSource{
		FS: fromPath,
	}

	dat, err := os.ReadFile(appConfig.DeliveryConfigFile)
	if err != nil {
		return err
	}
	cfg, err := unmarshalDeliveryConfig(string(dat))
	if err != nil {
		return err
	}

	for _, t := range cfg.Targets {
		targets[t.Name] = t.Destination
		if err := health.Register(t.Destination); err != nil {
			slog.Error("failed to register destination", "destination", t)
		}
	}
	slog.Info("targets", "targets", targets)

	for _, g := range cfg.Groups {
		groups[g.Key()] = g
		if g.DeliveryTargets == nil {
			slog.Warn(fmt.Sprintf("no targets configured for group %s", g.Key()))
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
			StorageContainer:    appConfig.AzureUploadContainer,
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
	RegisterSource(UploadSrc, src)

	if err := health.Register(src); err != nil {
		slog.Error("failed to register some health checks", "error", err)
	}
	return nil
}

// target may end up being a type
func Deliver(ctx context.Context, id string, path string, s Source, d Destination) (string, error) {
	manifest, err := s.GetMetadata(ctx, id)
	if err != nil {
		return "", err
	}

	length, e := strconv.ParseInt(manifest["content_length"], 10, 64)
	if e != nil {
		length = 1
	}

	concurrency := 5
	if length > size5MB {
		// app level configuration for this
		concurrency = int(length) / (size5MB / 5)
	}
	return d.Copy(ctx, id, path, &s, manifest, length, concurrency)
}

var ErrBadIngestTimestamp = errors.New("bad ingest timestamp")

func GetDeliveredFilename(tuid string, pathTemplate string, manifest map[string]string) (string, error) {
	if pathTemplate == "" {
		pathTemplate = "{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}"
	}
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadataPkg.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	// Use path template to form the full name.
	rawTime, ok := manifest["dex_ingest_datetime"]
	if !ok {
		return "", ErrBadIngestTimestamp
	}
	t, err := time.Parse(time.RFC3339Nano, rawTime)
	if err != nil {
		return "", errors.Join(err, ErrBadIngestTimestamp)
	}
	m := fmt.Sprintf("%02d", t.Month())
	d := fmt.Sprintf("%02d", t.Day())
	h := fmt.Sprintf("%02d", t.Hour())
	pathInfo := &PathInfo{
		Year:            strconv.Itoa(t.Year()),
		Month:           m,
		Day:             d,
		Hour:            h,
		Filename:        filenameWithoutExtension,
		UploadId:        tuid,
		DataStreamID:    metadataPkg.GetDataStreamID(manifest),
		DataStreamRoute: metadataPkg.GetDataStreamID(manifest),
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
		return b.String() + extension, nil
	}

	return b.String(), nil

}
