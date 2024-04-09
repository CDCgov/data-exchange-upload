package metadata

import (
	"errors"
	"log/slog"
	"reflect"
	"strings"

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

var RequiredManifestFields = map[string][]string{
	"1.0": {"meta_destination_id", "meta_ext_event"},
	"2.0": {"data_stream_id", "data_stream_route"},
}

func getVersionFromManifest(manifest handler.MetaData, loader validation.ConfigLoader) (*validation.MetadataConfig, error) {
	version, ok := manifest["version"]
	if !ok {
		version = "1.0"
	}

	switch version {
	case "1.0":
		getter, err := v1.NewFromManifest(manifest)
		if err != nil {
			return nil, err
		}
		return getter.GetConfig(loader)
	case "2.0":
		getter, err := v2.NewFromManifest(manifest)
		if err != nil {
			return nil, err
		}
		return getter.GetConfig(loader)
	default:
		return nil, errors.New("Unsupported metadata version")
	}
}

type SenderManifestVerification struct {
	Loader validation.ConfigLoader
}

func (v *SenderManifestVerification) Verify(event handler.HookEvent) (hooks.HookResponse, error) {
	resp := hooks.HookResponse{}

	manifest := event.Upload.MetaData

	config, err := getVersionFromManifest(manifest, v.Loader)
	if err != nil {
		// TODO: does this fail the upload if an error is returned
		return resp, err
	}

	logger.Info("checking config", "config", config)
	logger.Info("checking the sender manifest:", "manifest", manifest)
	return resp, nil
}
