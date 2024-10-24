package delivery

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/google/uuid"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

const maxPartsAzure = 50000

type AzureSource struct {
	FromContainerClient *container.Client
	StorageContainer    string
	Prefix              string
}

func (ad *AzureSource) SourceType() string {
	return storageTypeAzureBlob
}

func (ad *AzureSource) Container() string {
	return ad.StorageContainer
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
	props := storeaz.DepointerizeMetadata(resp.Metadata)
	props["last_modified"] = resp.LastModified.Format(time.RFC3339Nano)
	props["content_length"] = strconv.FormatInt(*resp.ContentLength, 10)
	return props, nil
}

func (ad *AzureSource) GetSignedObjectURL(ctx context.Context, containerName string, objectPath string) (string, error) {
	return "", nil
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
func (ad *AzureDestination) Copy(ctx context.Context, path string, source *Source, length int64, concurrency int) (string, error) {

	return "url", nil
}

func (ad *AzureDestination) copyWholeFromSignedURL(ctx context.Context, sourceSignedURL string, destPath string,
	metadata map[string]string) (string, error) {
	destBlob := ad.toClient.NewBlockBlobClient(destPath)

	_, e := destBlob.UploadBlobFromURL(ctx, sourceSignedURL,
		&blockblob.UploadBlobFromURLOptions{
			Metadata: storeaz.PointerizeMetadata(metadata),
		})

	if e != nil {
		return "", fmt.Errorf("unable to copy blob: %v", e)
	}
	return destBlob.URL(), nil
}

func (ad *AzureDestination) copyBlocksFromSignedURL(ctx context.Context, sourceSignedURL string, destPath string,
	length int64, concurrency int, metadata map[string]string) (string, error) {
	var partSize int64 = size5MB
	if length > size5MB*maxPartsAzure {
		// we need to increase the Part size
		partSize = length / maxPartsAzure
	}
	numChunks := length / partSize
	blockBlobClient := ad.toClient.NewBlockBlobClient(destPath)
	blockBase := uuid.New()
	blockIDs := make([]string, numChunks)
	var chunkNum int64
	var start int64 = 0
	var count = partSize
	chunkIdMap := make(map[string]azblob.HTTPRange)
	for chunkNum = 0; chunkNum < numChunks; chunkNum++ {
		end := start + count
		if chunkNum == numChunks-1 {
			count = 0
		}
		chunkId := base64.StdEncoding.EncodeToString([]byte(blockBase.String() + fmt.Sprintf("%05d", chunkNum)))
		blockIDs[chunkNum] = chunkId
		chunkIdMap[chunkId] = azblob.HTTPRange{
			Offset: start,
			Count:  count,
		}
		start = end
	}
	wg := sync.WaitGroup{}
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	routines := 0
	defer cancel()
	for id := range chunkIdMap {
		wg.Add(1)
		routines++
		go func(chunkId string) {
			defer wg.Done()
			_, err := blockBlobClient.StageBlockFromURL(ctx, chunkId, sourceSignedURL, &blockblob.StageBlockFromURLOptions{
				Range: chunkIdMap[chunkId],
			})

			if err != nil {
				select {
				case errCh <- err:
					// error was set
				default:
					// some other error is already set
				}
				cancel()
			}
		}(id)
		if routines >= concurrency {
			wg.Wait()
			routines = 0
		}
	}
	wg.Wait()
	select {
	case err := <-errCh:
		// there was an error during staging
		return "", fmt.Errorf("error staging blocks: %v", err)
	default:
		// no error was encountered
	}
	_, err := blockBlobClient.CommitBlockList(ctx, blockIDs,
		&blockblob.CommitBlockListOptions{Metadata: storeaz.PointerizeMetadata(metadata)})
	if err != nil {
		return "", fmt.Errorf("unable to commit blocks: %v", err)
	}

	return blockBlobClient.URL(), nil
}

func (ad *AzureDestination) copyFromStream() (string, error) {
	return "url", nil
}

func (ad *AzureDestination) DestinationType() string {
	return storageTypeAzureBlob
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
