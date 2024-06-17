package postprocessing

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	Deliver(ctx context.Context, tuid string, metadata map[string]string) error
}

func NewFileDeliverer(_ context.Context, target string) (*FileDeliverer, error) {
	localConfig, err := appconfig.LocalStoreConfig(target)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(localConfig.ToPath)
	if err != nil {
		os.Mkdir(localConfig.ToPath, 0755)
	}

	return &FileDeliverer{
		LocalStorageConfig: *localConfig,
		Target:             target,
	}, nil
}

func NewAzureDeliverer(ctx context.Context, target string, appConfig *appconfig.AppConfig) (*AzureDeliverer, error) {
	config, err := appconfig.AzureStoreConfig(target)
	if err != nil {
		return nil, err
	}
	// TODO Can the tus container client be singleton?
	tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
	if err != nil {
		return nil, err
	}
	checkpointContainerClient, err := storeaz.NewContainerClient(*config, config.ContainerName)
	if err != nil {
		return nil, err
	}
	err = storeaz.CreateContainerIfNotExists(ctx, checkpointContainerClient)
	if err != nil {
		return nil, err
	}

	return &AzureDeliverer{
		FromContainerClient: tusContainerClient,
		ToContainerClient:   checkpointContainerClient,
		TusPrefix:           appConfig.TusUploadPrefix,
		Target:              target,
	}, nil
}

// target may end up being a type
func Deliver(ctx context.Context, tuid string, manifest map[string]string, target string) error {
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

	err := d.Deliver(ctx, tuid, manifest)
	if err != nil {
		logger.Error("failed to copy file", "target", target)
		content.Result = "failed"
		content.ErrorDescription = err.Error()
	}

	logger.Info("File Copy Report", "report", report)
	reports.Publish(ctx, report)

	return err
}

type FileDeliverer struct {
	appconfig.LocalStorageConfig
	Target string
}

type AzureDeliverer struct {
	FromContainerClient *container.Client
	ToContainerClient   *container.Client
	TusPrefix           string
	Target              string
}

func (fd *FileDeliverer) Deliver(_ context.Context, tuid string, manifest map[string]string) error {
	f, err := fd.FromPath.Open(tuid)
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

	return err
}

func (fd *FileDeliverer) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Deliver Target " + fd.Target
	info, err := os.Stat(fd.ToPath)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", fd.ToPath)
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (ad *AzureDeliverer) Deliver(ctx context.Context, tuid string, manifest map[string]string) error {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	blobName, err := getDeliveredFilename(ctx, ad.Target, tuid, manifest)

	destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
	logger.Info("starting copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())
	resp, err := destBlobClient.StartCopyFromURL(ctx, srcBlobClient.URL(), nil)
	if err != nil {
		return err
	}

	status := *resp.CopyStatus
	var statusDescription string
	for status == blob.CopyStatusTypePending {
		getPropResp, err := destBlobClient.GetProperties(ctx, nil)
		if err != nil {
			return err
		}
		status = *getPropResp.CopyStatus
		statusDescription = *getPropResp.CopyStatusDescription
		logger.Info("Copy progress", "status", fmt.Sprintf("%s", status))
	}

	if status != blob.CopyStatusTypeSuccess {
		return fmt.Errorf("copy to target %s unsuccessful with status %s and description %s", ad.Target, status, statusDescription)
	}

	logger.Info("Copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL(), "status", status)

	return nil
}

func (ad *AzureDeliverer) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure deliver target " + ad.Target

	if ad.ToContainerClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Azure deliverer target " + ad.Target + " not configured"
	}

	_, err := ad.ToContainerClient.GetProperties(ctx, nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}

func getDeliveredFilename(ctx context.Context, target string, tuid string, manifest map[string]string) (string, error) {
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadata.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	suffix, err := metadata.GetFilenameSuffix(ctx, manifest, tuid)
	blobName := filenameWithoutExtension + suffix + extension

	// Next, need to set the filename prefix based on config and target.
	// edav, routing -> use config
	prefix := ""

	switch target {
	case "edav":
	case "routing":
		prefix, err = metadata.GetFilenamePrefix(ctx, manifest)
		if err != nil {
			return "", err
		}
	}

	return prefix + blobName, nil
}
