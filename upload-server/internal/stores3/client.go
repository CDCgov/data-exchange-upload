package stores3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewWithEndpoint(ctx context.Context, endpoint string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		// For non-AWS S3 backends or custom endpoints
		if endpoint != "" {
			o.UsePathStyle = true
			o.BaseEndpoint = &endpoint
		}
	})
	return client, nil
}
