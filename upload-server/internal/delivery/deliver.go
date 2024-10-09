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

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"gopkg.in/yaml.v3"

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
	targetNames := make([]string, len(targets))
	i := 0
	for t := range targets {
		targetNames[i] = t
		i++
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
	Hour     string
	UploadId string
	Filename string
}

type Config struct {
	Programs []Program `yaml:"programs"`
}

type Program struct {
	DataStreamId    string   `yaml:"data_stream_id"`
	DataStreamRoute string   `yaml:"data_stream_route"`
	DeliveryTargets []Target `yaml:"delivery_targets"`
}

type Target struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"`
	Destination Destination `yaml:"-"`
}

var DestinationTypes = map[string]func() Destination{
	"s3":      func() Destination { return &S3Destination{} },
	"file":    func() Destination { return &FileDestination{} },
	"az-blob": func() Destination { return &AzureDestination{} },
}

var ErrUnknownDestinationType = errors.New("Unknown destination type")

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

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllSourcesAndDestinations(ctx context.Context, appConfig appconfig.AppConfig) (err error) {
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
	for _, p := range cfg.Programs {
		for _, t := range p.DeliveryTargets {
			name := p.DataStreamId + "-" + p.DataStreamRoute
			RegisterDestination(name, t.Name, t.Destination)
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
