package postprocessing

import (
	"context"
	"encoding/json"
	"errors"
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
	// lets make this really dumb, it should take a file uri and take the rest from there. That makes it pretty recoverable.
	// root -> ""
	// date -> default pattern
	// "" -> ""

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
}

func (fd *FileDeliverer) Deliver(tuid string, manifest map[string]string) error {
	//dir := "./uploads/tus-prefix"
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
	// TODO get tus prefix from app config.  Place in azure deliverer.
	srcBlobClient := ad.FromContainerClient.NewBlobClient("tus-prefix/" + tuid)
	// Get filename from metadata.
	filename := metadata.GetFilename(manifest)
	tokens := strings.Split(filename, ".")
	extension := ""
	if len(tokens) > 1 {
		extension = "." + tokens[len(tokens)-1]
	}

	// Load config from metadata.
	// TODO create a utility function from this.
	path, err := metadata.GetConfigIdentifierByVersion(context.TODO(), manifest)
	if err != nil {
		return err
	}
	config, err := metadata.Cache.GetConfig(context.TODO(), path)
	if err != nil {
		return err
	}
	suffix := ""
	if config.Copy.FilenameSuffix == "upload_id" {
		suffix = "_" + tuid
	}
	blobName := filename + suffix + extension

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
	for status == "pending" {
		getPropResp, err := destBlobClient.GetProperties(context.TODO(), nil)
		if err != nil {
			return err
		}
		status = *getPropResp.CopyStatus
		logger.Info("Copy progress", status)
	}

	logger.Info("Copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL(), "status", status)

	return nil
}
