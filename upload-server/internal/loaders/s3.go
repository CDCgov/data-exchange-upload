package loaders

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type S3ConfigLoader struct {
	Client     *s3.Client
	BucketName string
}

func (l *S3ConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	output, err := l.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &l.BucketName,
		Key:    &path,
	})
	if err != nil {
		return nil, err
	}

	return io.ReadAll(output.Body)
}
