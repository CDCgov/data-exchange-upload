package storeaz

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"log/slog"
	"net/http"
)

var (
	logger *slog.Logger
)

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
