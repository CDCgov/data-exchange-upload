package stores3

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
)

type S3Store struct {
	Store s3store.S3Store
}

type S3StoreUpload struct {
	handler.Upload
}

func (su *S3StoreUpload) GetInfo(ctx context.Context) (handler.FileInfo, error) {
	info, err := su.Upload.GetInfo(ctx)
	info.ID, _, _ = strings.Cut(info.ID, "+")
	return info, err
}

func (s *S3Store) NewUpload(ctx context.Context, info handler.FileInfo) (handler.Upload, error) {
	u, err := s.Store.NewUpload(ctx, info)
	return &S3StoreUpload{
		u,
	}, err
}

func (s *S3Store) metadataKeyWithPrefix(key string) *string {
	prefix := s.Store.MetadataObjectPrefix
	if prefix == "" {
		prefix = s.Store.ObjectPrefix
	}
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return aws.String(prefix + key)
}

func (s *S3Store) GetUpload(ctx context.Context, id string) (handler.Upload, error) {
	if !strings.Contains(id, "+") {
		c, ok := s.Store.Service.(*s3.Client)
		if !ok {
			return nil, fmt.Errorf("Bad configuration, non-standard s3 client")
		}
		rsp, err := c.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.Store.Bucket),
			Key:    s.metadataKeyWithPrefix(id + ".info"),
		})
		if err != nil {
			return nil, err
		}
		info := &handler.FileInfo{}
		if err := json.NewDecoder(rsp.Body).Decode(info); err != nil {
			return nil, err
		}
		id = info.ID
	}
	u, err := s.Store.GetUpload(ctx, id)
	return &S3StoreUpload{
		u,
	}, err
}

func (s S3Store) AsTerminatableUpload(upload handler.Upload) handler.TerminatableUpload {
	u := upload.(*S3StoreUpload)
	return s.Store.AsTerminatableUpload(u.Upload)
}

func (s S3Store) AsLengthDeclarableUpload(upload handler.Upload) handler.LengthDeclarableUpload {
	u := upload.(*S3StoreUpload)
	return s.Store.AsLengthDeclarableUpload(u.Upload)
}

func (s S3Store) AsConcatableUpload(upload handler.Upload) handler.ConcatableUpload {
	u := upload.(*S3StoreUpload)
	return s.Store.AsConcatableUpload(u.Upload)
}

func (s *S3Store) UseIn(composer *handler.StoreComposer) {
	composer.UseCore(s)
	composer.UseTerminater(s)
	composer.UseConcater(s)
	composer.UseLengthDeferrer(s)
}
