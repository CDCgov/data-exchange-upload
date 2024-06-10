package storeaz

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
)

var (
	logger *slog.Logger
)

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func CreateContainerIfNotExists(ctx context.Context, containerClient *container.Client) error {
	_, err := containerClient.GetProperties(ctx, nil)
	if err != nil {
		var storageErr *azcore.ResponseError
		if errors.As(err, &storageErr) {
			if storageErr.StatusCode == http.StatusNotFound {
				logger.Info("creating routing checkpoint container", "container", containerClient.URL())
				_, err := containerClient.Create(ctx, nil)
				if err != nil {
					logger.Error("failed to create routing checkpoint container", "container", containerClient.URL())
					return err
				}
			}
		}
	}

	return nil
}

func PointerizeMetadata(metadata map[string]string) map[string]*string {
	p := make(map[string]*string)
	for k, v := range metadata {
		value := v
		p[k] = &value
	}

	return p
}
