package postprocessing

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

var targets = map[string]Deliverer{}

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
	health.Checkable
	Deliver(ctx context.Context, tuid string, metadata map[string]string) error
}

func NewFileDeliverer(_ context.Context, target string, appConfig *appconfig.AppConfig) (*FileDeliverer, error) {
	localConfig, err := appconfig.LocalStoreConfig(target, appConfig)
	if err != nil {
		return nil, err
	}

	return &FileDeliverer{
		LocalStorageConfig: *localConfig,
		Target:             target,
	}, nil
}

func NewAzureDeliverer(ctx context.Context, target string, appConfig *appconfig.AppConfig) (*AzureDeliverer, error) {
	config, err := appconfig.GetAzureContainerConfig(target)
	if err != nil {
		return nil, err
	}
	// TODO Can the tus container client be singleton?
	tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
	if err != nil {
		return nil, err
	}
	checkpointContainerClient, err := storeaz.NewContainerClient(config.AzureStorageConfig, config.ContainerName)
	if err != nil {
		return nil, err
	}
	checkpointClient, err := storeaz.NewBlobClient(config.AzureStorageConfig)
	if err != nil {
		return nil, err
	}
	err = storeaz.CreateContainerIfNotExists(ctx, checkpointContainerClient)
	if err != nil {
		return nil, err
	}

	return &AzureDeliverer{
		FromContainerClient: tusContainerClient,
		ToClient:            checkpointClient,
		ToContainer:         config.ContainerName,
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

	rb := reports.NewBuilder[reports.FileCopyContent](
		"1.0.0",
		reports.StageFileCopy,
		tuid,
		manifest,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC()).SetContent(reports.FileCopyContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageFileCopy,
		},
		FileSourceBlobUrl:      "", // TODO
		FileDestinationBlobUrl: "", // TODO
		Timestamp:              "", // TODO.  Does PS API do this for us?
	})

	err := d.Deliver(ctx, tuid, manifest)
	rb.SetEndTime(time.Now().UTC())
	if err != nil {
		logger.Error("failed to copy file", "target", target)
		rb.SetStatus(reports.StatusFailed).AppendIssue(err.Error())
	}

	report := rb.Build()
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
	ToClient            *azblob.Client
	ToContainer         string
	TusPrefix           string
	Target              string
}

func (fd *FileDeliverer) Deliver(_ context.Context, tuid string, _ map[string]string) error {
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
	s, err := srcBlobClient.DownloadStream(ctx, nil)
	defer s.Body.Close()
	if err != nil {
		return err
	}

	logger.Info("starting copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

	// TODO include src blob metadata
	_, err = ad.ToClient.UploadStream(ctx, ad.ToContainer, blobName, s.Body, nil)
	if err != nil {
		return err
	}

	logger.Info("successful copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

	return nil
}

func (ad *AzureDeliverer) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure deliver target " + ad.Target
	rsp.Status = models.STATUS_UP

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
	filename := metadataPkg.GetFilename(manifest)
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
