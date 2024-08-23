package delivery

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Destinations []Destination

type Destination struct {
	Name               string
	System             string
	ManifestConditions ManifestCondition `json:"manifest_conditions"`
	Connection         *json.RawMessage  `json:"connection"`
}

type ManifestCondition struct {
	Key   string
	Value string
}

type FileConnection struct {
	Folder string
	f      fs.File
}

func (fc *FileConnection) Write(p []byte) (n int, err error) {
	if fc.f == nil {
		os.Mkdir(fd.ToPath, 0755)
		dest, err := os.Create(filepath.Join(fd.ToPath, tuid))
		if err != nil {
			return err
		}
	}
	return fc.f.Write(p)
}

type S3Connection struct {
	ConnectionString string `json:"connection_string`
	BucketName       string `json:"bucket_name"`
}

type Connection interface {
	// TODO: should return more info than just the error
	Upload(io.ReadSeeker) error
}

func ParseConnection[T io.Writer](raw []byte) (io.Writer, error) {
	var w T
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return w, nil
}

var connectioniTypes = map[string]func([]byte) (io.WriteCloser, error){
	"edav": ParseConnection[FileConnection],
	"s3":   ParseConnection[S3Connection],
}
