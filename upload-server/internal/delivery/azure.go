package delivery

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
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
		containerClient, err := storeaz.NewContainerClient(appconfig.AzureStorageConfig{
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
	blobName, err := getDeliveredFilename(ctx, path, ad.PathTemplate, m)
	if err != nil {
		return blobName, err
	}

	c, err := ad.Client()
	if err != nil {
		return blobName, err
	}
	client := c.NewBlockBlobClient(blobName)

	_, err = client.UploadStream(ctx, r, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(m),
	})
	return client.URL(), err
}

type azDestinationWriter struct {
	baseID   string
	chunks   []int
	client   *blockblob.Client
	ctx      context.Context
	metadata map[string]string
}

type reader struct {
	*bytes.Reader
}

func (r *reader) Close() error {
	return nil
}

func (adw *azDestinationWriter) WriteAt(p []byte, off int64) (n int, err error) {
	id := []byte(fmt.Sprintf("%d:%s", off, adw.baseID))
	if len(id) > 64 {
		id = id[:64]
	}
	if len(id) != 64 {
		return 0, errors.New("Length for id must be 64")
	}
	chunkID := base64.StdEncoding.EncodeToString(id)
	b := bytes.NewReader(p)
	r := &reader{b}
	rsp, err := adw.client.StageBlock(adw.ctx, chunkID, r, &blockblob.StageBlockOptions{})
	if err != nil {
		return n, err
	}
	//TODO does this need to be thread safe?
	adw.chunks = append(adw.chunks, int(off))
	slog.Debug("Staged block", "response", rsp, "chunkID", chunkID, "size", len(p), "at", off)
	return len(p), nil
}

func (adw *azDestinationWriter) Chunks() (chunks []string) {
	sort.Ints(adw.chunks)
	for _, c := range adw.chunks {
		chunks = append(chunks, fmt.Sprintf("%d:%s", c, adw.baseID))
	}
	return chunks
}
func (adw *azDestinationWriter) Close() error {
	_, err := adw.client.CommitBlockList(adw.ctx, adw.Chunks(), &blockblob.CommitBlockListOptions{
		Metadata: storeaz.PointerizeMetadata(adw.metadata),
	})
	return err
}

type Writable interface {
	Writer(ctx context.Context, id string, m map[string]string) (io.WriterAt, error)
}

func (ad *AzureDestination) Writer(ctx context.Context, id string, m map[string]string) (io.WriterAt, error) {
	blobName, err := getDeliveredFilename(ctx, id, ad.PathTemplate, m)
	if err != nil {
		return nil, err
	}
	c, err := ad.Client()
	if err != nil {
		return nil, err
	}
	client := c.NewBlockBlobClient(blobName)
	adw := &azDestinationWriter{
		baseID:   id,
		ctx:      ctx,
		client:   client,
		metadata: m,
	}
	return adw, nil
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
