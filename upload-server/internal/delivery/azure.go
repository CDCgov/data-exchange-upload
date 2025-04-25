package delivery

import (
	"context"
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

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

func (ad *AzureSource) GetSize(ctx context.Context, tuid string) (float64, error) {
	return 0, errors.New("not implemented")
}

func (ad *AzureSource) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure source"
	rsp.Status = models.STATUS_UP

	if ad.FromContainerClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Azure source not configured"
	}

	_, err := ad.FromContainerClient.GetProperties(ctx, nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}

type AzureDestination struct {
	toClient          *container.Client
	Name              string `yaml:"name"`
	StorageAccount    string `yaml:"storage_account"`
	StorageKey        string `yaml:"storage_key"`
	PathTemplate      string `yaml:"path_template"`
	ContainerEndpoint string `yaml:"endpoint"`
	TenantId          string `yaml:"tenant_id"`
	ClientId          string `yaml:"client_id"`
	ClientSecret      string `yaml:"client_secret"`
	ContainerName     string `yaml:"container_name"`
}

func (ad *AzureDestination) Client() (*container.Client, error) {
	if ad.toClient == nil {
		containerClient, err := storeaz.NewContainerClient(storeaz.Credentials{
			StorageName:       ad.StorageAccount,
			StorageKey:        ad.StorageKey,
			ContainerEndpoint: ad.ContainerEndpoint,
			TenantId:          ad.TenantId,
			ClientId:          ad.ClientId,
			ClientSecret:      ad.ClientSecret,
		}, ad.ContainerName)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = storeaz.CreateContainerIfNotExists(ctx, containerClient)
		if err != nil {
			return nil, err
		}
		ad.toClient = containerClient
	}
	return ad.toClient, nil
}

func (ad *AzureDestination) Upload(ctx context.Context, path string, r io.Reader, m map[string]string) (string, error) {
	c, err := ad.Client()
	if err != nil {
		return path, err
	}
	client := c.NewBlockBlobClient(path)

	decodedUrl, err := url.QueryUnescape(client.URL())
	if err != nil {
		return client.URL(), err
	}

	_, err = client.UploadStream(ctx, r, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(m),
	})
	if err != nil {
		return decodedUrl, err
	}

	return decodedUrl, nil
}

func (ad *AzureDestination) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure deliver target " + ad.Name
	rsp.Status = models.STATUS_UP

	c, err := ad.Client()
	if err != nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Azure deliverer target " + ad.Name + " not configured"
		return rsp
	}

	if _, err := c.GetProperties(ctx, nil); err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}
