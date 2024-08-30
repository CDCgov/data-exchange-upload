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
	containerClient, err := storeaz.NewContainerClient(config.AzureStorageConfig, config.ContainerName)
	if err != nil {
		return nil, err
	}
	client, err := storeaz.NewBlobClient(config.AzureStorageConfig)
	if err != nil {
		return nil, err
	}
	err = storeaz.CreateContainerIfNotExists(ctx, containerClient)
	if err != nil {
		return nil, err
	}

	return &AzureDestination{
		ToClient:    client,
		ToContainer: config.ContainerName,
		Target:      target,
	}, nil
}

type AzureSource struct {
	FromContainerClient *container.Client
	Prefix              string
}

func (ad *AzureSource) Reader(ctx context.Context, path string) (io.Reader, error) {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.Prefix + "/" + path)
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
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.Prefix + "/" + tuid)
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

func (ad *AzureDestination) Upload(ctx context.Context, path string, r io.Reader, m map[string]string) (string, error) {
	blobName, err := getDeliveredFilename(ctx, ad.Target, path, m)
	if err != nil {
		return blobName, err
	}
	_, err = ad.ToClient.UploadStream(ctx, ad.ToContainer, blobName, r, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(m),
	})
	return blobName, err
}

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
