package delivery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/tus/tusd/v2/pkg/handler"

	"gopkg.in/yaml.v3"

	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
)

var UploadSrc = "upload"

var ErrSrcFileNotExist = fmt.Errorf("source file does not exist")

var Groups map[string]Group
var Targets map[string]Destination

func GetTarget(target string) (Destination, bool) {
	d, ok := Targets[target]
	return d, ok
}

func FindGroupFromMetadata(meta handler.MetaData) (Group, bool) {
	dataStreamId, dataStreamRoute := meta["data_stream_id"], meta["data_stream_route"]
	group := Group{
		DataStreamId:    dataStreamId,
		DataStreamRoute: dataStreamRoute,
	}
	g, ok := Groups[group.Key()]
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

type Source interface {
	Reader(context.Context, string) (io.Reader, error)
	GetMetadata(context.Context, string) (map[string]string, error)
	GetSize(context.Context, string) (int64, error)
}

type Destination interface {
	Upload(context.Context, string, io.Reader, map[string]string) (string, error)
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

func UnmarshalDeliveryConfig(confBody string) (*Config, error) {
	confStr := os.ExpandEnv(confBody)
	c := &Config{}

	err := yaml.Unmarshal([]byte(confStr), &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// target may end up being a type
func Deliver(ctx context.Context, id string, path string, s Source, d Destination) (string, error) {

	manifest, err := s.GetMetadata(ctx, id)
	if err != nil {
		return "", err
	}

	r, err := s.Reader(ctx, id)
	if err != nil {
		return "", err
	}
	if rc, ok := r.(io.Closer); ok {
		defer rc.Close()
	}
	return d.Upload(ctx, path, r, manifest)
}

var ErrBadIngestTimestamp = errors.New("bad ingest timestamp")

func GetDeliveredFilename(ctx context.Context, tuid string, pathTemplate string, manifest map[string]string) (string, error) {
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
		DataStreamID:    manifest["data_stream_id"],
		DataStreamRoute: manifest["data_stream_route"],
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
