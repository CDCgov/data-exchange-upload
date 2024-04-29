package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	v1 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v1"
	v2 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v2"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
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
	Loader validation.ConfigLoader
}

func (v *SenderManifestVerification) Verify(event handler.HookEvent) (hooks.HookResponse, error) {
	resp := hooks.HookResponse{}

	manifest := event.Upload.MetaData
	logger.Info("checking the sender manifest:", "manifest", manifest)

	config, err := getVersionFromManifest(event.Context, manifest, v.Loader)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, validation.ErrFailure) {
			resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Body:       err.Error(),
			})
			resp.RejectUpload = true

			return resp, nil
		}
		return resp, err
	}
	logger.Info("checking config", "config", config)

	var errs error
	for _, field := range config.Fields {
		err := field.Validate(manifest)
		errs = errors.Join(errs, err)
	}

	if errs != nil {
		logger.Error("validation errors and warnings", "errors", errs)
	}

	if errors.Is(errs, validation.ErrFailure) {
		resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       errs.Error(),
		})
		resp.RejectUpload = true
	}

	return resp, nil
}
