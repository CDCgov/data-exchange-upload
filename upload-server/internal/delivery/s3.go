package delivery

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
)

func NewS3Destination(ctx context.Context, target string, conn *appconfig.S3StorageConfig) (*S3Destination, error) {
	c, err := stores3.New(ctx, conn)
	if err != nil {
		return nil, err
	}

	return &S3Destination{
		ToClient:   c,
		BucketName: conn.BucketName,
		Target:     target,
	}, nil
}

type writeAtWrapper struct {
	writer io.Writer
}

func (w *writeAtWrapper) WriteAt(p []byte, _ int64) (int, error) {
	// Ignoring offset because we force sequential writing
	return w.writer.Write(p)
}

type S3Source struct {
	FromClient *s3.Client
	BucketName string
	Prefix     string
}

func (ss *S3Source) Reader(ctx context.Context, path string) (io.Reader, error) {
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(path, "+")[0]
	srcFileName := ss.Prefix + "/" + id
	downloader := manager.NewDownloader(ss.FromClient)
	downloader.Concurrency = 1

	r, w := io.Pipe()

	go func() {
		defer w.Close()

		_, err := downloader.Download(ctx, &writeAtWrapper{w}, &s3.GetObjectInput{
			Bucket: &ss.BucketName,
			Key:    &srcFileName,
		})
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	return r, nil
}

func (ss *S3Source) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get the object from S3
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(tuid, "+")[0]
	srcFilename := ss.Prefix + "/" + id
	output, err := ss.FromClient.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(ss.BucketName),
		Key:    aws.String(srcFilename),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object: %w", err)
	}

	return output.Metadata, nil
}

func (sd *S3Source) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "S3 source"
	rsp.Status = models.STATUS_UP

	if sd.FromClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "S3 source not configured"
	}

	_, err := sd.FromClient.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &sd.BucketName,
	})
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}

type S3Destination struct {
	ToClient   *s3.Client
	BucketName string
	Target     string
}

func (sd *S3Destination) Upload(ctx context.Context, path string, r io.Reader, m map[string]string) (string, error) {
	destFileName, err := getDeliveredFilename(ctx, path, m)
	if err != nil {
		return "", err
	}

	uploader := manager.NewUploader(sd.ToClient)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:   &sd.BucketName,
		Key:      &destFileName,
		Body:     r,
		Metadata: m,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.BucketName, destFileName)
	return s3URL, nil
}

func (sd *S3Destination) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "S3 deliver target " + sd.Target
	rsp.Status = models.STATUS_UP

	if sd.ToClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "S3 deliverer target " + sd.Target + " not configured"
	}

	_, err := sd.ToClient.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &sd.BucketName,
	})
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}
