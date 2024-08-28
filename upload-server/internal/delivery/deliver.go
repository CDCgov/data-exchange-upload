package delivery

import (
	"context"
	"fmt"
	"io"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
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
	Reader(string) (io.Reader, error)
	GetMetadata(context.Context, string) (map[string]string, error)
}

type Destination interface {
	Upload(context.Context, string, io.Reader, map[string]string) error
}

// target may end up being a type
func Deliver(ctx context.Context, path string, s Source, d Destination) error {

	manifest, err := s.GetMetadata(ctx, path)
	if err != nil {
		return err
	}

	/*
		//TODO pull reports up if we can
		rb := reports.NewBuilder[reports.FileCopyContent](
			"1.0.0",
			reports.StageFileCopy,
			tuid,
			reports.DispositionTypeAdd).SetStartTime(time.Now().UTC())

		rb.SetManifest(manifest)

		rb.SetContent(reports.FileCopyContent{
			ReportContent: reports.ReportContent{
				ContentSchemaVersion: "1.0.0",
				ContentSchemaName:    reports.StageFileCopy,
			},
			FileSourceBlobUrl:      srcUrl,
			FileDestinationBlobUrl: destUrl,
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
	*/

	//NOTE could use ctx to store the manifest
	r, err := s.Reader(path)
	if err != nil {
		return err
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
