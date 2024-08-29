package s3inspector

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"io"
	"strings"
)

type S3UploadInspector struct {
	Client     *s3.Client
	BucketName string
	TusPrefix  string
}

func (sui *S3UploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {
	// temp solution for handling hash that tus s3 store puts on upload IDs.  See https://github.com/tus/tusd/pull/1167
	id = strings.Split(id, "+")[0]

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
	// temp solution for handling hash that tus s3 store puts on upload IDs.  See https://github.com/tus/tusd/pull/1167
	id = strings.Split(id, "+")[0]

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
		"updated_at": output.LastModified,
		"size_bytes": output.ObjectSize,
	}, nil
}

func (sui *S3UploadInspector) InspectFileStatus(ctx context.Context, id string) (*info.DeliveryStatus, error) {
	//TODO implement me
	panic("implement me")
}
