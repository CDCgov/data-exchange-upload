package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders"
	azureloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/azure"
	fileloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/google/uuid"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

const FolderStructureDate = "date_YYYY_MM_DD"
const FolderStructureRoot = "root"
const FilenameSuffixUploadId = "upload_id"
const ErrNoUploadId = "no upload ID defined"

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type PreCreateResponse struct {
	UploadId         string   `json:"upload_id"`
	ValidationErrors []string `json:"validation_errors"`
}

var Cache *ConfigCache

type ConfigCache struct {
	sync.Map
	Loader validation.ConfigLoader
}

func InitConfigCache(ctx context.Context, appConfig appconfig.AppConfig) error {
	Cache = &ConfigCache{
		Loader: &fileloader.FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	if appConfig.AzureConnection != nil && appConfig.S3Connection != nil {
		return errors.New("cannot load metadata config from multiple locations")
	}

	if appConfig.AzureConnection != nil && appConfig.AzureManifestConfigContainer != "" {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return err
		}
		Cache.Loader = &azureloader.AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.AzureManifestConfigContainer,
		}
	}

	if appConfig.S3Connection != nil && appConfig.S3ManifestConfigFolder != "" {
		client, err := stores3.New(ctx, appConfig.S3Connection)
		bucket := appConfig.S3Connection.BucketName
		if appConfig.S3ManifestConfigBucket != "" {
			bucket = appConfig.S3ManifestConfigBucket
		}
		if err != nil {
			return err
		}
		Cache.Loader = &loaders.S3ConfigLoader{
			Client:     client,
			BucketName: bucket,
			Folder:     appConfig.S3ManifestConfigFolder,
		}
	}

	return nil
}

func (c *ConfigCache) GetConfig(ctx context.Context, key string) (*validation.ManifestConfig, error) {
	conf, ok := c.Load(key)
	if !ok {
		if c.Loader == nil {
			return nil, errors.New("misconfigured config cache, set a loader")
		}
		b, err := c.Loader.LoadConfig(ctx, key)
		if err != nil {
			return nil, err
		}

		// Expand config string to substitute any env var placeholders within.
		expandedConf := os.ExpandEnv(string(b))
		mc := &validation.ManifestConfig{}
		if err := json.Unmarshal([]byte(expandedConf), mc); err != nil {
			return nil, err
		}
		c.SetConfig(key, mc)
		return mc, nil
	}
	config, ok := conf.(*validation.ManifestConfig)
	if !ok {
		return nil, errors.New("manifest not found")
	}
	return config, nil
}

func (c *ConfigCache) SetConfig(key any, config *validation.ManifestConfig) {
	c.Store(key, config)
}

func GetConfigFromManifest(ctx context.Context, manifest handler.MetaData) (*validation.ManifestConfig, error) {
	path, err := NewFromManifest(manifest)
	if err != nil {
		return nil, err
	}
	config, err := Cache.GetConfig(ctx, path.Path())
	if err != nil {
		return nil, err
	}
	return config, nil
}

func Uid() string {
	return uuid.NewString()
}

type SenderManifestVerification struct {
	Configs *ConfigCache
}

func (v *SenderManifestVerification) verify(ctx context.Context, manifest handler.MetaData) error {
	path, err := NewFromManifest(manifest) //GetConfigIdentifierByVersion(manifest)
	if err != nil {
		return err
	}
	c, err := v.Configs.GetConfig(ctx, strings.ToLower(path.Path()))
	if err != nil {
		return err
	}
	config := c.Metadata
	logger.Info("checking config", "config", config)

	var errs error
	for _, field := range config.Fields {
		err := field.Validate(manifest)
		errs = errors.Join(errs, err)
	}
	return errs
}

func (v *SenderManifestVerification) Verify(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	manifest := event.Upload.MetaData
	logger.Info("checking the sender manifest:", "manifest", manifest)
	tuid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tuid = resp.ChangeFileInfo.ID
	}
	if tuid == "" {
		return resp, errors.New("no Upload ID defined")
	}

	rb := reports.NewBuilderWithManifest[reports.MetaDataVerifyContent](
		"1.0.0",
		reports.StageMetadataVerify,
		tuid,
		manifest,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC()).SetContent(reports.MetaDataVerifyContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageMetadataVerify,
		},
		Filename: metadata.GetFilename(manifest),
		Metadata: manifest,
	})

	defer func() {
		rb.SetEndTime(time.Now().UTC())
		report := rb.Build()
		logger.Info("REPORT", "report", report)
		reports.Publish(event.Context, report)
	}()

	if err := v.verify(event.Context, manifest); err != nil {
		logger.Error("validation errors and warnings", "errors", err)

		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})

		if errors.Is(err, validation.ErrFailure) {
			resp.RejectUpload = true

			respBody := PreCreateResponse{
				UploadId:         tuid,
				ValidationErrors: strings.Split(err.Error(), "\n"),
			}
			b, err := json.Marshal(respBody)
			if err != nil {
				return resp, err
			}
			resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       string(b),
			})
			return resp, nil
		}
		return resp, err
	}

	rb.SetStatus(reports.StatusSuccess)
	return resp, nil
}

type FileMetadataAppender struct {
	Path string
}

type AzureMetadataAppender struct {
	ContainerClient *container.Client
	TusPrefix       string
}

type Appender interface {
	Append(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error)
}

func (fa *FileMetadataAppender) Append(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tuid = resp.ChangeFileInfo.ID
	}
	if tuid == "" {
		return resp, errors.New(ErrNoUploadId)
	}

	manifest := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		manifest = resp.ChangeFileInfo.MetaData
	}

	m, err := json.Marshal(manifest)
	if err != nil {
		return resp, err
	}
	err = os.WriteFile(filepath.Join(fa.Path, tuid+".meta"), m, 0666)

	return resp, nil
}

func (aa *AzureMetadataAppender) Append(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tuid = resp.ChangeFileInfo.ID
	}
	if tuid == "" {
		return resp, errors.New("no Upload ID defined")
	}

	manifest := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		manifest = resp.ChangeFileInfo.MetaData
	}

	// Get blob client.
	blobClient := aa.ContainerClient.NewBlobClient(aa.TusPrefix + "/" + tuid)
	_, err := blobClient.SetMetadata(event.Context, storeaz.PointerizeMetadata(manifest), nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func WithPreCreateManifestTransforms(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := Uid()
	resp.ChangeFileInfo.ID = tuid

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	logger.Info("adding global timestamp", "timestamp", timestamp)

	manifest := event.Upload.MetaData
	manifest["dex_ingest_datetime"] = timestamp
	manifest["upload_id"] = tuid
	resp.ChangeFileInfo.MetaData = manifest

	report := reports.NewBuilderWithManifest[reports.BulkMetadataTransformReportContent](
		"1.0.0",
		reports.StageMetadataTransform,
		tuid,
		manifest,
		reports.DispositionTypeAdd).SetContent(reports.BulkMetadataTransformReportContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageMetadataTransform,
		},
		Transforms: []reports.MetadataTransformContent{
			{Action: "update",
				Field: "ID",
				Value: tuid}, {
				Action: "append",
				Field:  "dex_ingest_datetime",
				Value:  timestamp,
			}, {
				Action: "append",
				Field:  "upload_id",
				Value:  tuid,
			}},
	}).Build()

	logger.Info("METADATA TRANSFORM REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}
