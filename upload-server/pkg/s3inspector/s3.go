package s3inspector

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type S3UploadInspector struct {
	Client     *s3.Client
	BucketName string
	TusPrefix  string
}

func NewS3UploadInspector(containerClient *s3.Client, bucketName string, tusPrefix string) *S3UploadInspector {
	return &S3UploadInspector{
		Client:     containerClient,
		BucketName: bucketName,
		TusPrefix:  tusPrefix,
	}
}

func (sui *S3UploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {

	filename := sui.TusPrefix + "/" + id + ".info"
	output, err := sui.Client.GetObject(c, &s3.GetObjectInput{
		Bucket: &sui.BucketName,
		Key:    &filename,
	})
	if err != nil {
		// TODO specifically handle not found
		return nil, err
	}

	b, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	jsonMap := &info.InfoFileData{}
	if err := json.Unmarshal(b, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (sui *S3UploadInspector) InspectUploadedFile(c context.Context, id string) (map[string]any, error) {

	filename := sui.TusPrefix + "/" + id
	output, err := sui.Client.GetObjectAttributes(c, &s3.GetObjectAttributesInput{
		Bucket:           &sui.BucketName,
		Key:              &filename,
		ObjectAttributes: []types.ObjectAttributes{"ObjectSize"},
	})
	if err != nil {
		// Support deferred uploads
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, errors.Join(err, info.ErrNotFound)
		}
		return nil, err
	}

	return map[string]any{
		"updated_at": output.LastModified.Format(time.RFC3339Nano),
		"size_bytes": output.ObjectSize,
	}, nil
}
