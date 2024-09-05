package stores3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

func NewContainerClient(ctx context.Context, s3Config *appconfig.S3StorageConfig) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// For non-AWS S3 backends
		if s3Config.Endpoint != "" {
			o.UsePathStyle = true
			o.BaseEndpoint = &s3Config.Endpoint
		}
	})

	return client, nil
}
