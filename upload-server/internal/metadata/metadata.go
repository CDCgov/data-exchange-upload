package metadata

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	v1 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v1"
	v2 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v2"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

var registeredVersions = map[string]func(handler.MetaData) (validation.ConfigLocation, error){
	"1.0": v1.NewFromManifest,
	"2.0": v2.NewFromManifest,
}

var Cache *ConfigCache

type ConfigCache struct {
	sync.Map
	Loader validation.ConfigLoader
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
		mc := &validation.ManifestConfig{}
		if err := json.Unmarshal(b, mc); err != nil {
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

func GetConfigIdentifierByVersion(ctx context.Context, manifest handler.MetaData) (string, error) {
	version := manifest["version"]
	if version == "" {
		version = "1.0"
	}
	configLocationBuilder, ok := registeredVersions[version]
	if !ok {
		return "", fmt.Errorf("unsupported version %s %w", version, validation.ErrFailure)
	}
	configLoc, err := configLocationBuilder(manifest)
	if err != nil {
		return "", err
	}
	return configLoc.Path(), nil
}

func GetFilename(manifest map[string]string) string {

	keys := []string{
		"filename",
		"original_filename",
		"meta_ext_filename",
		"received_filename",
	}

	for _, key := range keys {
		if name, ok := manifest[key]; ok {
			return name
		}
	}
	return ""
}

func GetDataStreamID(manifest map[string]string) string {
	switch manifest["version"] {
	case "2.0":
		return manifest["data_stream_id"]
	default:
		return manifest["meta_destination_id"]
	}
}

func GetDataStreamRoute(manifest map[string]string) string {
	switch manifest["version"] {
	case "2.0":
		return manifest["data_stream_route"]
	default:
		return manifest["meta_ext_event"]
	}

}

func Uid() string {
	id := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		// This is probably an appropriate way to handle errors from our source
		// for random bits.
		panic(err)
	}
	return hex.EncodeToString(id)
}

type SenderManifestVerification struct {
	Configs *ConfigCache
}

func (v *SenderManifestVerification) verify(ctx context.Context, manifest map[string]string) error {
	path, err := GetConfigIdentifierByVersion(ctx, manifest)
	if err != nil {
		return err
	}
	c, err := v.Configs.GetConfig(ctx, strings.ToLower(path))
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

	content := &models.MetaDataVerifyContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "0.0.1",
			SchemaName:    "dex-metadata-verify",
		},
		Filename: GetFilename(manifest),
		Metadata: manifest,
	}

	report := &models.Report{
		UploadID:        tuid,
		DataStreamID:    GetDataStreamID(manifest),
		DataStreamRoute: GetDataStreamRoute(manifest),
		StageName:       "dex-metadata-verify",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}

	defer func() {
		logger.Info("REPORT", "report", report)
		reports.Publish(event.Context, report)
	}()

	if err := v.verify(event.Context, manifest); err != nil {
		logger.Error("validation errors and warnings", "errors", err)

		content.Issues = &validation.ValidationError{Err: err}

		if errors.Is(err, validation.ErrFailure) {
			resp.RejectUpload = true
			resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       err.Error(),
			})
			return resp, nil
		}
		return resp, err
	}

	return resp, nil
}

func (v *SenderManifestVerification) getHydrationConfig(ctx context.Context, manifest map[string]string) (*validation.ManifestConfig, error) {
	path, err := GetConfigIdentifierByVersion(ctx, manifest)
	if err != nil {
		return nil, err
	}
	c, err := v.Configs.GetConfig(ctx, strings.ToLower(path))
	if err != nil {
		return nil, err
	}
	if c.CompatConfigFilename != "" {
		return v.Configs.GetConfig(ctx, c.CompatConfigFilename)
	}

	//TODO: don't trigger this this way, it's a weird sideaffect
	manifest["version"] = "2.0"
	manifest["data_stream_id"] = manifest["meta_destination_id"]
	manifest["data_stream_route"] = manifest["meta_ext_event"]
	path, err = GetConfigIdentifierByVersion(ctx, manifest)
	if err != nil {
		return nil, err
	}
	return v.Configs.GetConfig(ctx, strings.ToLower(path))
}

func (v *SenderManifestVerification) Hydrate(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	// TODO: this could be the event context...but honestly we don't want this to stop
	// we do need graceful shutdown, so maybe we need a custom context here somehow
	ctx := context.TODO()

	manifest := event.Upload.MetaData
	if v, ok := manifest["version"]; ok && v == "2.0" {
		return resp, nil
	}

	c, err := v.getHydrationConfig(ctx, manifest)
	if err != nil {
		return resp, err
	}

	v2Manifest, transforms := v1.Hydrate(manifest, c)
	resp.ChangeFileInfo.MetaData = v2Manifest

	// Report new metadata
	content := &models.BulkMetaDataTransformContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "metadata-transform",
		},
		Transforms: transforms,
	}
	report := &models.Report{
		UploadID:        event.Upload.ID,
		DataStreamID:    GetDataStreamID(manifest),
		DataStreamRoute: GetDataStreamRoute(manifest),
		StageName:       "dex-metadata-transform",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}
	logger.Info("Metadata Hydration Report", "report", report)
	reports.Publish(ctx, report)

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

type MetadataAppender struct {
	Appender Appender
}

func (fa *FileMetadataAppender) Append(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tuid = resp.ChangeFileInfo.ID
	}
	if tuid == "" {
		return resp, errors.New("no Upload ID defined")
	}

	metadata := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		metadata = resp.ChangeFileInfo.MetaData
	}

	m, err := json.Marshal(metadata)
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

	metadata := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		metadata = resp.ChangeFileInfo.MetaData
	}

	// Get blob client.
	blobClient := aa.ContainerClient.NewBlobClient(aa.TusPrefix + "/" + tuid)
	_, err := blobClient.SetMetadata(event.Context, storeaz.PointerizeMetadata(metadata), nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func WithUploadID(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := Uid()
	resp.ChangeFileInfo.ID = tuid

	if sloger.DefaultLogger != nil {
		logger = sloger.DefaultLogger.With(models.TGUID_KEY, tuid)
	}
	logger.Info("Generated UUID", "UUID", tuid)

	content := &models.MetaDataTransformContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "metadata-transform",
		},
		Action: "update",
		Field:  "ID",
		Value:  tuid,
	}

	manifest := event.Upload.MetaData
	report := &models.Report{
		UploadID:        tuid,
		DataStreamID:    GetDataStreamID(manifest),
		DataStreamRoute: GetDataStreamRoute(manifest),
		StageName:       "dex-metadata-transform",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}

	logger.Info("METADATA TRANSFORM REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil

}

func WithTimestamp(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tguid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tguid = resp.ChangeFileInfo.ID
	}
	if tguid == "" {
		return resp, errors.New("no Upload ID defined")
	}

	timestamp := time.Now().Format(time.RFC3339)
	logger.Info("adding global timestamp", "timestamp", timestamp)

	manifest := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		manifest = resp.ChangeFileInfo.MetaData
	}

	fieldname := "dex_ingest_datetime"
	manifest[fieldname] = timestamp
	resp.ChangeFileInfo.MetaData = manifest

	content := &models.MetaDataTransformContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "metadata-transform",
		},
		Action: "append",
		Field:  fieldname,
		Value:  timestamp,
	}

	report := &models.Report{
		UploadID:        tguid,
		DataStreamID:    GetDataStreamID(manifest),
		DataStreamRoute: GetDataStreamRoute(manifest),
		StageName:       "dex-metadata-transform",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}
	logger.Info("METADATA TRANSFORM REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}
