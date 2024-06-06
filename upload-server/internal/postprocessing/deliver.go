package postprocessing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

var targets = map[string]Deliverer{}
var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func RegisterTarget(name string, d Deliverer) {
	targets[name] = d
}

type Deliverer interface {
	Deliver(tuid string, metadata map[string]string) error
	GetReporter() reporters.Reporter
}

// target may end up being a type
func Deliver(tuid string, manifest map[string]string, target string) error {
	ctx := context.TODO()
	d, ok := targets[target]
	if !ok {
		return errors.New("not recoverable, bad target " + target)
	}

	content := &models.FileCopyContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "0.0.1",
			SchemaName:    "dex-file-copy",
		},
		Destination: target,
		Result:      "success",
	}

	report := &models.Report{
		UploadID:        tuid,
		DataStreamID:    metadata.GetDataStreamID(manifest),
		DataStreamRoute: metadata.GetDataStreamRoute(manifest),
		StageName:       "dex-file-copy",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}

	err := d.Deliver(tuid, manifest)
	if err != nil {
		logger.Error("failed to copy file", "target", target)
		content.Result = "failed"
		content.ErrorDescription = err.Error()
	}

	logger.Info("File Copy Report", "report", report)
	if err := d.GetReporter().Publish(ctx, report); err != nil {
		logger.Error("Failed to report", "report", report, "reporter", d.GetReporter(), "err", err)
	}

	return err
}

type FileDeliverer struct {
	From     fs.FS
	ToPath   string
	Reporter reporters.Reporter
}

type AzureDeliverer struct {
	FromContainerClient *container.Client
	ToContainerClient   *container.Client
	TusPrefix           string
	Target              string
	Reporter            reporters.Reporter
}

func (fd *FileDeliverer) Deliver(tuid string, manifest map[string]string) error {
	f, err := fd.From.Open(tuid)
	if err != nil {
		return err
	}
	defer f.Close()
	os.Mkdir(fd.ToPath, 0755)
	dest, err := os.Create(filepath.Join(fd.ToPath, tuid))
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, f); err != nil {
		return err
	}

	m, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(fd.ToPath, tuid+".meta"), m, 0666)
	return err
}

func (fd *FileDeliverer) GetReporter() reporters.Reporter {
	return fd.Reporter
}

func (ad *AzureDeliverer) GetReporter() reporters.Reporter {
	return ad.Reporter
}

func (ad *AzureDeliverer) Deliver(tuid string, manifest map[string]string) error {
	ctx := context.TODO()
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	blobName, err := getDeliveredFilename(ad.Target, tuid, manifest)

	// TODO Check copy status in some background goroutine.
	destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
	manifestPointer := make(map[string]*string)
	for k, v := range manifest {
		value := v
		manifestPointer[k] = &value
	}
	logger.Info("starting copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())
	resp, err := destBlobClient.StartCopyFromURL(ctx, srcBlobClient.URL(), &blob.StartCopyFromURLOptions{Metadata: manifestPointer})
	if err != nil {
		return err
	}

	status := *resp.CopyStatus
	for status == blob.CopyStatusTypePending {
		getPropResp, err := destBlobClient.GetProperties(ctx, nil)
		if err != nil {
			return err
		}
		status = *getPropResp.CopyStatus
		logger.Info("Copy progress", "status", fmt.Sprintf("%s", status))
	}

	logger.Info("Copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL(), "status", status)

	return nil
}

func getDeliveredFilename(target string, tuid string, manifest map[string]string) (string, error) {
	ctx := context.TODO()
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadata.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	// Load config from metadata.
	path, err := metadata.GetConfigIdentifierByVersion(ctx, manifest)
	if err != nil {
		return "", err
	}
	config, err := metadata.Cache.GetConfig(ctx, path)
	if err != nil {
		return "", err
	}
	suffix := ""
	if config.Copy.FilenameSuffix == "upload_id" {
		suffix = "_" + tuid
	}
	blobName := filenameWithoutExtension + suffix + extension

	// Next, need to set the filename prefix based on config and target.
	// edav, routing -> use config
	prefix := ""

	switch target {
	case "edav":
	case "routing":
		if config.Copy.FolderStructure == "date_YYYY_MM_DD" {
			// Get UTC year, month, and day
			t := time.Now().UTC()
			prefix = fmt.Sprintf("%d/%02d/%02d/", t.Year(), t.Month(), t.Day())
		}
	}

	return prefix + blobName, nil
}
