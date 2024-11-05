package delivery

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
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

const maxPartsAzure = 50000 // maximum number of parts per block blob in Azure Storage

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

func (ad *AzureSource) GetSignedObjectURL(_ context.Context, _, objectPath string) (string, error) {
	path := ad.GetSourceFilePath(objectPath)
	sourceBlob := ad.FromContainerClient.NewBlockBlobClient(path)
	sourceURL, er := sourceBlob.GetSASURL(sas.BlobPermissions{
		Read:   true,
		Add:    true,
		Create: true,
		Write:  true,
	}, time.Now().Add(time.Hour), nil)
	if er != nil {
		return "", fmt.Errorf("unable to get signed url for source object: %v", er)
	}
	return sourceURL, nil
}

func (ad *AzureSource) GetSourceFilePath(path string) string {
	return ad.Prefix + "/" + path
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
	LargeObjectSize   int    `env:"AZURE_LARGE_OBJECT_SIZE, default=52,428,800"`
}

type azureBlobChunk struct {
	BlockId string
	Range   azblob.HTTPRange
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

func (ad *AzureDestination) Copy(ctx context.Context, id string, path string, source *Source, metadata map[string]string, length int64, concurrency int) (string, error) {
	s := *source
	cSource, ok := s.(CloudSource)
	if ok {
		sourceUrl, err := cSource.GetSignedObjectURL(ctx, cSource.Container(), id)
		if err != nil {
			return "", fmt.Errorf("unable to obtain signed url: %v", err)
		}
		if int(length) < ad.LargeObjectSize {
			return ad.copyWholeFromSignedURL(ctx, sourceUrl, path, metadata)
		}
		return ad.copyBlocksFromSignedURL(ctx, sourceUrl, path, length, concurrency, metadata)
	}
	r, err := s.Reader(ctx, id)
	if err != nil {
		return "", fmt.Errorf("unable to obtain source reader: %v", err)
	}
	return ad.Upload(ctx, r, path, metadata)
}

func (ad *AzureDestination) copyWholeFromSignedURL(ctx context.Context, sourceSignedURL string, destPath string,
	metadata map[string]string) (string, error) {
	client, e := ad.Client()
	if e != nil {
		return "", fmt.Errorf("unable to obtain Azure container client: %v", e)
	}
	destBlob := client.NewBlockBlobClient(destPath)

	_, err := destBlob.UploadBlobFromURL(ctx, sourceSignedURL,
		&blockblob.UploadBlobFromURLOptions{
			Metadata: storeaz.PointerizeMetadata(metadata),
		})

	if err != nil {
		return "", fmt.Errorf("unable to copy blob: %v", err)
	}
	return destBlob.URL(), nil
}

func (ad *AzureDestination) copyBlocksFromSignedURL(ctx context.Context, sourceSignedURL string, destPath string,
	length int64, concurrency int, metadata map[string]string) (string, error) {
	client, e := ad.Client()
	if e != nil {
		return "", fmt.Errorf("unable to obtain Azure container client: %v", e)
	}
	var partSize int64 = size5MB
	if length > size5MB*maxPartsAzure {
		// we need to increase the Part size
		partSize = length / maxPartsAzure
	}
	numChunks := length / partSize
	blockBlobClient := client.NewBlockBlobClient(destPath)
	blockBase := uuid.New()
	// channel of block jobs to do
	jobsCh := make(chan azureBlobChunk, numChunks)
	// ordered list of block IDs, needed for commit at the end
	blockIDs := make([]string, numChunks)

	var chunkNum int64
	var start int64 = 0
	var count = partSize

	// fill the jobs channel
	for chunkNum = 0; chunkNum < numChunks; chunkNum++ {
		end := start + count
		if chunkNum == numChunks-1 {
			count = 0
		}
		chunkId := base64.StdEncoding.EncodeToString([]byte(blockBase.String() + fmt.Sprintf("%05d", chunkNum)))
		jobsCh <- azureBlobChunk{
			BlockId: chunkId,
			Range: azblob.HTTPRange{
				Offset: start,
				Count:  count,
			},
		}
		blockIDs[chunkNum] = chunkId
		start = end
	}
	close(jobsCh)
	// use worker pool to pull jobs from jobsCh
	var wg sync.WaitGroup
	wg.Add(concurrency)
	errCh := make(chan error, 1)
	// send jobs to the workers
	for worker := 0; worker < concurrency; worker++ {
		go ad.copyPartWorker(ctx, blockBlobClient, sourceSignedURL, jobsCh, errCh, &wg)
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

func (ad *AzureDestination) copyPartWorker(ctx context.Context, blockBlobClient *blockblob.Client, sourceSignedURL string,
	jobsCh <-chan azureBlobChunk, errCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobsCh {
		_, err := blockBlobClient.StageBlockFromURL(ctx, job.BlockId, sourceSignedURL,
			&blockblob.StageBlockFromURLOptions{
				Range: job.Range,
			})

		if err != nil {
			select {
			case errCh <- err:
				// error was set
			default:
				// some other error is already set
			}
		}
	}
}

func (ad *AzureDestination) Upload(ctx context.Context, r io.Reader, path string,
	metadata map[string]string) (string, error) {
	c, err := ad.Client()
	if err != nil {
		return "", fmt.Errorf("unable to obtain Azure container client: %v", err)

	}
	client := c.NewBlockBlobClient(path)

	_, err = client.UploadStream(ctx, r, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(metadata),
	})
	if err != nil {
		return "", fmt.Errorf("unable to upload stream: %v", err)
	}
	return client.URL(), nil
}

func (ad *AzureDestination) DestinationType() string {
	return storageTypeAzureBlob
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
