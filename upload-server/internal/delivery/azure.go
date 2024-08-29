package delivery

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
)

func NewAzureDestination(ctx context.Context, target string, appConfig *appconfig.AppConfig) (*AzureDestination, error) {
	config, err := appconfig.GetAzureContainerConfig(target)
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

	return &AzureDestination{
		ToClient:    checkpointClient,
		ToContainer: config.ContainerName,
		Target:      target,
	}, nil
}

type AzureDeliverer struct {
	FromContainerClient *container.Client
	ToContainerClient   *container.Client
	ToClient            *azblob.Client
	ToContainer         string
	TusPrefix           string
	Target              string
}

type AzureSource struct {
	FromContainerClient *container.Client
	TusPrefix           string
}

func (ad *AzureSource) Reader(ctx context.Context, path string) (io.Reader, error) {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + path)
	s, err := srcBlobClient.DownloadStream(ctx, nil)
	if s.ErrorCode != nil && *s.ErrorCode == string(bloberror.BlobNotFound) {
		return nil, ErrSrcFileNotExist
	}
	if err != nil {
		return nil, err
	}
	return s.Body, nil
}

func (ad *AzureSource) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	resp, err := srcBlobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, err
	}
	return storeaz.DepointerizeMetadata(resp.Metadata), nil
}

type AzureDestination struct {
	ToClient    *azblob.Client
	ToContainer string
	Target      string
}

func (ad *AzureDestination) Upload(ctx context.Context, path string, r io.Reader, m map[string]string) error {
	blobName, err := getDeliveredFilename(ctx, ad.Target, path, m)
	if err != nil {
		return err
	}
	_, err = ad.ToClient.UploadStream(ctx, ad.ToContainer, blobName, r, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(m),
	})
	return err
}

/*
	func (ad *AzureDeliverer) Deliver(ctx context.Context, tuid string, manifest map[string]string) error {
		// Get blob src blob client.
		srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
		blobName, err := getDeliveredFilename(ctx, ad.Target, tuid, manifest)
		if err != nil {
			return err
		}
		destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
		s, err := srcBlobClient.DownloadStream(ctx, nil)
		defer s.Body.Close()
		if s.ErrorCode != nil && *s.ErrorCode == string(bloberror.BlobNotFound) {
			return ErrSrcFileNotExist
		}
		if err != nil {
			return err
		}

		slog.Info("starting copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

		_, err = ad.ToClient.UploadStream(ctx, ad.ToContainer, blobName, s.Body, &azblob.UploadStreamOptions{
			Metadata: storeaz.PointerizeMetadata(manifest),
		})
		if err != nil {
			return err
		}

		slog.Info("successful copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

		return nil
	}

	func (ad *AzureDeliverer) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
		// Get blob src blob client.
		srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
		resp, err := srcBlobClient.GetProperties(ctx, nil)
		if err != nil {
			return nil, err
		}
		return storeaz.DepointerizeMetadata(resp.Metadata), nil
	}

	func (ad *AzureDeliverer) GetSrcUrl(_ context.Context, tuid string) (string, error) {
		srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
		return srcBlobClient.URL(), nil
	}

	func (ad *AzureDeliverer) GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error) {
		blobName, err := getDeliveredFilename(ctx, ad.Target, tuid, manifest)
		if err != nil {
			return "", err
		}
		destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
		return destBlobClient.URL(), nil
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
*/
func getDeliveredFilename(ctx context.Context, target string, tuid string, manifest map[string]string) (string, error) {
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadataPkg.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	suffix, err := metadata.GetFilenameSuffix(ctx, manifest, tuid)
	if err != nil {
		return "", err
	}
	blobName := filenameWithoutExtension + suffix + extension

	// Next, need to set the filename prefix based on config and target.
	// edav, routing -> use config
	prefix := ""

	switch target {
	case "routing", "edav":
		prefix, err = metadata.GetFilenamePrefix(ctx, manifest)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(prefix, blobName), nil
}
