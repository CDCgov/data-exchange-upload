package postprocessing

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strings"
)

type S3Deliverer struct {
	SrcBucket  string
	DestBucket string
	SrcClient  *s3.Client
	DestClient *s3.Client
	TusPrefix  string
	Target     string
}

type writeAtWrapper struct {
	writer io.Writer
}

func (w *writeAtWrapper) WriteAt(p []byte, _ int64) (int, error) {
	// Ignoring offset because we force sequential writing
	return w.writer.Write(p)
}

func (sd *S3Deliverer) Deliver(ctx context.Context, tuid string, manifest map[string]string) error {
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(tuid, "+")[0]
	srcFilename := sd.TusPrefix + "/" + id
	destFileName, err := getDeliveredFilename(ctx, sd.Target, tuid, manifest)
	if err != nil {
		return err
	}

	// Create a downloader and uploader
	downloader := manager.NewDownloader(sd.SrcClient)
	downloader.Concurrency = 1
	uploader := manager.NewUploader(sd.DestClient)

	r, w := io.Pipe()

	go func() {
		defer w.Close()

		_, err := downloader.Download(ctx, &writeAtWrapper{w}, &s3.GetObjectInput{
			Bucket: &sd.SrcBucket,
			Key:    &srcFilename,
		})
		if err != nil {
			logger.Error(err.Error())
		}
	}()

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:   &sd.DestBucket,
		Key:      &destFileName,
		Body:     r,
		Metadata: manifest,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (sd *S3Deliverer) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get the object from S3
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(tuid, "+")[0]
	srcFilename := sd.TusPrefix + "/" + id
	output, err := sd.SrcClient.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(sd.SrcBucket),
		Key:    aws.String(srcFilename),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object: %w", err)
	}

	return output.Metadata, nil
}

func (sd *S3Deliverer) GetSrcUrl(_ context.Context, tuid string) (string, error) {
	// Construct the S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.SrcBucket, sd.TusPrefix+"/"+tuid)
	return s3URL, nil
}

func (sd *S3Deliverer) GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error) {
	objectKey, err := getDeliveredFilename(ctx, sd.Target, tuid, manifest)
	if err != nil {
		return "", err
	}

	// Construct the S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.DestBucket, objectKey)
	return s3URL, nil
}
