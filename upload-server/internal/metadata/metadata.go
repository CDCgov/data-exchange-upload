package metadata

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	v1 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v1"
	v2 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v2"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
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

var cachedConfigs = &configCache{}

type configCache struct {
	sync.Map
}

func (c *configCache) GetConfig(key any) (*validation.MetadataConfig, bool) {
	config, ok := c.Load(key)
	if !ok {
		return nil, ok
	}
	metaConfig, ok := config.(*validation.MetadataConfig)
	return metaConfig, ok
}

func (c *configCache) SetConfig(key any, config *validation.MetadataConfig) {
	c.Store(key, config)
}

func loadConfig(ctx context.Context, path string, loader validation.ConfigLoader) (*validation.MetadataConfig, error) {
	config, ok := cachedConfigs.GetConfig(path)
	if !ok {
		b, err := loader.LoadConfig(ctx, path)
		if err != nil {
			return nil, err
		}
		c := &validation.UploadConfig{}
		if err := json.Unmarshal(b, c); err != nil {
			return nil, err
		}
		config = &c.Metadata
		cachedConfigs.SetConfig(path, config)
	}
	return config, nil
}

func getVersionFromManifest(ctx context.Context, manifest handler.MetaData, loader validation.ConfigLoader) (*validation.MetadataConfig, error) {
	version, ok := manifest["version"]
	if version == "" {
		version = "1.0"
	}
	configLocationBuilder, ok := registeredVersions[version]
	if !ok {
		return nil, fmt.Errorf("unsupported version %s %w", version, validation.ErrFailure)
	}
	configLoc, err := configLocationBuilder(manifest)
	if err != nil {
		return nil, err
	}
	return loadConfig(ctx, configLoc.Path(), loader)
}

type SenderManifestVerification struct {
	Loader   validation.ConfigLoader
	Reporter reporters.Reporter
}

func (v *SenderManifestVerification) verify(ctx context.Context, manifest map[string]string) error {
	config, err := getVersionFromManifest(ctx, manifest, v.Loader)
	if err != nil {
		return err
	}

	logger.Info("checking config", "config", config)

	var errs error
	for _, field := range config.Fields {
		err := field.Validate(manifest)
		errs = errors.Join(errs, err)
	}
	return errs
}

type Report struct {
	UploadID        string `json:"upload_id"`
	StageName       string `json:"stage_name"`
	DataStreamID    string `json:"data_stream_id"`
	DataStreamRoute string `json:"data_stream_route"`
	ContentType     string `json:"content_type"`
	Content         any    `json:"content"` // TODO: Can we limit this to a specific type (i.e. ReportContent or UploadStatusTYpe type?
}

func (r *Report) Identifier() string {
	return r.UploadID
}

type MetaDataVerifyContent struct {
	SchemaVersion string `json:"schema_version"`
	SchemaName    string `json:"schema_name"`
	Filename      string `json:"filename"`
	Metadata      any    `json:"metadata"`
	Issues        error  `json:"issues"`
}

type UploadStatusContent struct {
	SchemaVersion string `json:"schema_version"`
	SchemaName    string `json:"schema_name"`
	Filename      string `json:"filename"`
	Metadata      any    `json:"metadata"`
	// Additional postReceive values:
	Tguid  string `json:"tguid"`
	Offset string `json:"offset"`
	Size   string `json:"size"`
}
type ValidationError struct {
	Err error
}

func (v *ValidationError) Error() string {
	return v.Err.Error()
}

func unwrap(e error) []error {
	errs := []error{}
	u, ok := e.(interface {
		Unwrap() []error
	})
	if ok {
		for _, err := range u.Unwrap() {
			errs = append(errs, unwrap(err)...)
		}
	} else {
		errs = append(errs, e)
		err := errors.Unwrap(e)
		if err != nil {
			errs = append(errs, unwrap(err)...)
		}
	}
	return errs
}

func (v *ValidationError) MarshalJSON() ([]byte, error) {
	errs := unwrap(v.Err)
	res := make([]any, len(errs))
	for i, e := range errs {
		res[i] = e.Error() // Fallback to the error string
	}
	return json.Marshal(res)
}

/*
   payload = {
       'schema_version': '0.0.1',
       'schema_name': 'dex-metadata-verify',
       'filename': filename,
       'metadata': meta_json,
       'issues': messages
   }
            "upload_id": tguid,
            "stage_name": "dex-upload",
            "data_stream_id": metadata["data_stream_id"],
            "data_stream_route": metadata["data_stream_route"],
            "content_type": "json",
            "content": {
                        "schema_name": "upload",
                        "schema_version": "1.0",
                        "tguid": tguid,
                        "offset": offset,
                        "size": size,
                        "filename": filename,
                        "data_stream_id": metadata["data_stream_id"],
                        "data_stream_route": metadata["data_stream_route"],
                        "metadata": metadata
            },
*/

func getFilename(manifest map[string]string) string {

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

func getDataStreamID(manifest map[string]string) string {
	switch manifest["version"] {
	case "v2":
		return manifest["data_stream_id"]
	default:
		return manifest["meta_destination_id"]
	}
}

func getDataStreamRoute(manifest map[string]string) string {
	switch manifest["version"] {
	case "v2":
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

	content := &MetaDataVerifyContent{
		SchemaVersion: "0.0.1",
		SchemaName:    "dex-metadata-verify",
		Filename:      getFilename(manifest),
		Metadata:      manifest,
	}

	report := &Report{
		UploadID:        tuid,
		DataStreamID:    getDataStreamID(manifest),
		DataStreamRoute: getDataStreamRoute(manifest),
		StageName:       "dex-metadata-verify",
		ContentType:     "json",
		Content:         content,
	}

	defer func() {
		logger.Info("REPORT", "report", report)
		if err := v.Reporter.Publish(event.Context, report); err != nil {
			logger.Error("Failed to report", "report", report, "reporter", v.Reporter, "UUID", tuid, "err", err)
		}
	}()

	if err := v.verify(event.Context, manifest); err != nil {
		logger.Error("validation errors and warnings", "errors", err)

		content.Issues = &ValidationError{err}

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

func WithUploadID(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {

	tuid := Uid()
	resp.ChangeFileInfo.ID = tuid

	logger.Info("Generated UUID", "UUID", tuid)

	return resp, nil

}

func WithTimestamp(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	timestamp := time.Now().Format(time.RFC3339)
	logger.Info("adding global timestamp", "timestamp", timestamp)

	manifest := event.Upload.MetaData

	if resp.ChangeFileInfo.MetaData != nil {
		manifest = resp.ChangeFileInfo.MetaData
	}

	manifest["dex_ingest_datetime"] = timestamp
	resp.ChangeFileInfo.MetaData = manifest

	return resp, nil
}

type HookEventHandler struct {
	Reporter reporters.Reporter
}

func (v *HookEventHandler) postReceive(tguid string, offset int64, size int64, manifest map[string]string, ctx context.Context) error {

	logger.Info("go version", "version", runtime.Version())
	logger.Info("metadata values", "manifest", manifest)

	filename := getFilename(manifest)

	logger.Info("file info", "filename", filename)

	content := &UploadStatusContent{
		SchemaVersion: "1.0",
		SchemaName:    "upload",
		Filename:      getFilename(manifest),
		Metadata:      manifest,
		Tguid:         tguid,
		Offset:        strconv.FormatInt(offset, 10),
		Size:          strconv.FormatInt(size, 10),
	}

	report := &Report{
		UploadID:        tguid,
		DataStreamID:    getDataStreamID(manifest),
		DataStreamRoute: getDataStreamRoute(manifest),
		StageName:       "dex-upload-status",
		ContentType:     "json",
		Content:         content,
	}

	defer func() {
		logger.Info("REPORT", "report", report)
		if err := v.Reporter.Publish(ctx, report); err != nil {
			logger.Error("Failed to report", "report", report, "reporter", v.Reporter, "UUID", tguid, "err", err)
		}
	}()

	return nil
}

func (v *HookEventHandler) PostReceive(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {

	logger.Info("------resp-------", "resp", resp)

	// Get values from event
	uploadId := event.Upload.ID
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	uploadMetadata := event.Upload.MetaData

	logger.Info(
		"[PostReceive]: event.Upload values",
		"uploadMetadata", uploadMetadata,
		"uploadId", uploadId,
		"uploadSize", uploadSize,
		"uploadOffset", uploadOffset,
	)

	if err := v.postReceive(uploadId, uploadOffset, uploadSize, uploadMetadata, event.Context); err != nil {
		//logger.Error("postReceive errors and warnings", "errors", err)
		logger.Error("postReceive errors and warnings", "err", err)

		//		content.Issues = &ValidationError{err}

		//		if errors.Is(err, validation.ErrFailure) {
		//			resp.RejectUpload = true
		//			resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
		//				StatusCode: http.StatusBadRequest,
		//				Body:       err.Error(),
		//			})
		//			return resp, nil
		//		}
		//		return resp, err

	}

	return resp, nil

}
