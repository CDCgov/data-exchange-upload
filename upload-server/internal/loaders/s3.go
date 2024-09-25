package loaders

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type S3ConfigLoader struct {
	Client     *s3.Client
	BucketName string
	Folder     string
}

func (l *S3ConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	key := path
	if l.Folder != "" {
		key = l.Folder + "/" + key
	}
	output, err := l.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &l.BucketName,
		Key:    &key,
	})
	// TODO wrap err in validation.ErrNotFound if didn't find the object
	// TODO handle nil pointer in if body is nil
	defer output.Body.Close()
	if err != nil {
		return nil, err
	}

	return io.ReadAll(output.Body)
}
