package postprocessing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
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
}

// target may end up being a type
func Deliver(tuid string, manifest map[string]string, target string) error {
	d, ok := targets[target]
	if !ok {
		return errors.New("not recoverable, bad target " + target)
	}
	return d.Deliver(tuid, manifest)
}

type FileDeliverer struct {
	From   fs.FS
	ToPath string
}

type AzureDeliverer struct {
	FromContainerClient *container.Client
	ToContainerClient   *container.Client
	TusPrefix           string
	Target              string
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

func (ad *AzureDeliverer) Deliver(tuid string, manifest map[string]string) error {
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
	resp, err := destBlobClient.StartCopyFromURL(context.TODO(), srcBlobClient.URL(), &blob.StartCopyFromURLOptions{Metadata: manifestPointer})
	if err != nil {
		return err
	}

	status := *resp.CopyStatus
	for status == blob.CopyStatusTypePending {
		getPropResp, err := destBlobClient.GetProperties(context.TODO(), nil)
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
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadata.GetFilename(manifest)
	filenameWithoutExtention := filename
	tokens := strings.Split(filename, ".")
	extension := ""
	if len(tokens) > 1 {
		extension = "." + tokens[len(tokens)-1]
		filenameWithoutExtention = strings.Join(tokens[:len(tokens)-1], ".")
	}
	// Load config from metadata.
	path, err := metadata.GetConfigIdentifierByVersion(context.TODO(), manifest)
	if err != nil {
		return "", err
	}
	config, err := metadata.Cache.GetConfig(context.TODO(), path)
	if err != nil {
		return "", err
	}
	suffix := ""
	if config.Copy.FilenameSuffix == "upload_id" {
		suffix = "_" + tuid
	}
	blobName := filenameWithoutExtention + suffix + extension

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
