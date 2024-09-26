package loaders

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
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
	defer func() {
		if output.Body != nil {
			output.Body.Close()
		}
	}()
	if err != nil {
		var notExist *types.NoSuchKey
		if errors.As(err, &notExist) {
			return nil, errors.Join(err, validation.ErrNotFound)
		}
		return nil, err
	}

	return io.ReadAll(output.Body)
}
