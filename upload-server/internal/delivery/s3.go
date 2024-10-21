package delivery

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
)

type S3Source struct {
	Connection *appconfig.S3StorageConfig
	BucketName string
	Prefix     string
}

type ReadTo interface {
	ReadTo(ctx context.Context, path string, w io.WriterAt) error
}

func (ss *S3Source) ReadTo(ctx context.Context, path string, w io.WriterAt) error {
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(path, "+")[0]
	srcFileName := ss.Prefix + "/" + id
	client, err := ss.Client()
	if err != nil {
		return err
	}
	downloader := manager.NewDownloader(client)
	downloader.BufferProvider = manager.NewPooledBufferedWriterReadFromProvider(1024 * 1024 * 5)
	//downloader.Concurrency = 8

	if _, err := downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: &ss.BucketName,
		Key:    &srcFileName,
	}); err != nil {
		return err
	}
	return nil
}

func (ss *S3Source) Client() (*s3.Client, error) {
	return stores3.New(context.TODO(), ss.Connection)
}

func (ss *S3Source) Reader(ctx context.Context, path string) (io.Reader, error) {
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(path, "+")[0]
	srcFileName := ss.Prefix + "/" + id
	client, err := ss.Client()
	if err != nil {
		return nil, err
	}
	rsp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(ss.BucketName),
		Key:    aws.String(srcFileName),
		//TODO ReadAt with Range
	})
	if err != nil {
		return nil, err
	}

	return rsp.Body, nil
}

func (ss *S3Source) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get the object from S3
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(tuid, "+")[0]
	srcFilename := ss.Prefix + "/" + id
	client, err := ss.Client()
	if err != nil {
		return nil, err
	}
	output, err := client.HeadObject(ctx, &s3.HeadObjectInput{
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

	client, err := ss.Client()
	if err != nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "S3 source not configured"
		return rsp
	}

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &ss.BucketName,
	}); err != nil {
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

func (sd *S3Destination) Retrieve(ctx context.Context) (aws.Credentials, error) {
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
	destFileName, err := getDeliveredFilename(ctx, path, sd.PathTemplate, m)
	if err != nil {
		return "", err
	}

	uploader := manager.NewUploader(sd.Client())
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:   &sd.BucketName,
		Key:      &destFileName,
		Body:     r,
		Metadata: m,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to %s %s: %w", sd.BucketName, path, err)
	}

	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.BucketName, destFileName)
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
