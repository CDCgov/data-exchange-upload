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
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

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

func (ss *S3Source) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "S3 source " + ss.BucketName
	rsp.Status = models.STATUS_UP

	if ss.FromClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "S3 source not configured"
	}

	_, err := ss.FromClient.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &ss.BucketName,
	})

	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}

type S3Destination struct {
	toClient        *s3.Client
	BucketName      string `yaml:"bucket_name"`
	Name            string `yaml:"name"`
	PathTemplate    string `yaml:"path_template"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Endpoint        string `yaml:"endpoint"`
	Region          string `yaml:"region"`
}

func (sd *S3Destination) Retrieve(_ context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     sd.AccessKeyID,
		SecretAccessKey: sd.SecretAccessKey,
	}, nil
}

func (sd *S3Destination) Client() *s3.Client {
	options := s3.Options{
		Credentials: sd,
		Region:      sd.Region,
	}

	if sd.Endpoint != "" {
		options.UsePathStyle = true
		options.BaseEndpoint = &sd.Endpoint
	}
	// Create a Session with a custom region
	sd.toClient = s3.New(options)
	return sd.toClient
}

func (sd *S3Destination) Upload(ctx context.Context, path string, r io.Reader, m map[string]string) (string, error) {

	uploader := manager.NewUploader(sd.Client())
	if _, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:   &sd.BucketName,
		Key:      &path,
		Body:     r,
		Metadata: m,
	}); err != nil {
		return "", fmt.Errorf("failed to upload file to %s %s: %w", sd.BucketName, path, err)
	}

	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.BucketName, path)
	return s3URL, nil
}

func (sd *S3Destination) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "S3 deliver target " + sd.Name
	rsp.Status = models.STATUS_UP

	if sd.Client() == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "S3 deliverer target " + sd.Name + " not configured"
	}

	_, err := sd.Client().HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &sd.BucketName,
	})
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}
