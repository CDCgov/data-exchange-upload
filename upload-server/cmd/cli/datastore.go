package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filestore"
)

type S3Store struct {
	store s3store.S3Store
}

type S3StoreUpload struct {
	handler.Upload
}

func (su *S3StoreUpload) GetInfo(ctx context.Context) (handler.FileInfo, error) {
	info, err := su.Upload.GetInfo(ctx)
	info.ID, _, _ = strings.Cut(info.ID, "+")
	return info, err
}

func (s *S3Store) NewUpload(ctx context.Context, info handler.FileInfo) (handler.Upload, error) {
	u, err := s.store.NewUpload(ctx, info)
	return &S3StoreUpload{
		u,
	}, err
}

func (s *S3Store) metadataKeyWithPrefix(key string) *string {
	prefix := s.store.MetadataObjectPrefix
	if prefix == "" {
		prefix = s.store.ObjectPrefix
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return aws.String(prefix + key)
}

func (s *S3Store) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	if !strings.Contains(id, "+") {
		log.Println("UPLOAD ID", id)
		c, err := stores3.NewWithEndpoint(ctx, appconfig.LoadedConfig.S3Connection.Endpoint)
		if err != nil {
			return nil, err
		}
		rsp, err := c.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.store.Bucket),
			Key:    s.metadataKeyWithPrefix(id + ".info"),
		})
		if err != nil {
			return nil, err
		}
		info := &handler.FileInfo{}
		if err := json.NewDecoder(rsp.Body).Decode(info); err != nil {
			return nil, err
		}
		id = info.ID
	}
	u, err := s.store.GetUpload(ctx, id)
	return &S3StoreUpload{
		u,
	}, err
}

func (s S3Store) AsTerminatableUpload(upload handler.Upload) handler.TerminatableUpload {
	u := upload.(*S3StoreUpload)
	return s.store.AsTerminatableUpload(u.Upload)
}

func (s S3Store) AsLengthDeclarableUpload(upload handler.Upload) handler.LengthDeclarableUpload {
	u := upload.(*S3StoreUpload)
	return s.store.AsLengthDeclarableUpload(u.Upload)
}

func (s S3Store) AsConcatableUpload(upload handler.Upload) handler.ConcatableUpload {
	u := upload.(*S3StoreUpload)
	return s.store.AsConcatableUpload(u.Upload)
}

func (s *S3Store) UseIn(composer *handler.StoreComposer) {
	composer.UseCore(s)
	composer.UseTerminater(s)
	composer.UseConcater(s)
	composer.UseLengthDeferrer(s)
}

func GetDataStore(ctx context.Context, appConfig appconfig.AppConfig) (handlertusd.Store, health.Checkable, error) {
	// ------------------------------------------------------------------
	// Load Az dependencies, needed for the DEX handler paths
	// ------------------------------------------------------------------
	if appConfig.AzureConnection != nil {
		if err := appConfig.AzureConnection.Check(); err != nil {
			return nil, nil, err
		}

		accountName := appConfig.AzureConnection.StorageName

		azureEndpoint := appConfig.AzureConnection.ContainerEndpoint
		logger.Info("Using Azure endpoint", "endpoint", azureEndpoint)

		azConfig := &azurestore.AzConfig{
			AccountName:   accountName,
			AccountKey:    appConfig.AzureConnection.StorageKey,
			ContainerName: appConfig.AzureUploadContainer,
			Endpoint:      azureEndpoint,
		} // .azConfig

		azService, err := azurestore.NewAzureService(azConfig)
		if err != nil {
			return nil, nil, err
		} // azService

		hc, err := storeaz.NewAzureHealthCheck(azConfig)

		if err != nil {
			return nil, nil, err
		} // azService

		store := azurestore.New(azService)
		store.ObjectPrefix = appConfig.TusUploadPrefix
		store.Container = appConfig.AzureUploadContainer
		return store, hc, nil
	} // .if
	if appConfig.S3Connection != nil {
		client, err := stores3.NewWithEndpoint(ctx, appConfig.S3Connection.Endpoint)
		if err != nil {
			return nil, nil, err
		}
		hc := &stores3.S3HealthCheck{
			Client:     client,
			BucketName: appConfig.S3Connection.BucketName,
		}
		store := s3store.New(appConfig.S3Connection.BucketName, client)
		store.ObjectPrefix = appConfig.TusUploadPrefix

		logger.Info("using S3 bucket", "bucket", appConfig.S3Connection.BucketName)
		return &S3Store{
			store: store,
		}, hc, nil
	}

	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	path := appConfig.LocalFolderUploadsTus

	return filestore.FileStore{
		Path: filepath.Join(path, appConfig.TusUploadPrefix),
	}, &FileStoreHealthCheck{path: path}, nil // .store
}

type FileStoreHealthCheck struct {
	path string
}

func (c *FileStoreHealthCheck) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Storage"
	info, err := os.Stat(c.path)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", c.path)
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}
