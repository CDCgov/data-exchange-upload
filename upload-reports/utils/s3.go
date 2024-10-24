package utils

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreateS3Client(ctx context.Context, region string, endpoint string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true
			o.BaseEndpoint = &endpoint
		}
	})

	return client, nil
}

func UploadCsvToS3(ctx context.Context, client *s3.Client, bucketName string, key string, csvData *bytes.Buffer) error {
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(csvData.Bytes()),
		ContentType: aws.String("application/zip"),
	}

	_, err := client.PutObject(ctx, putInput)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("upload to S3 timed out")
		}
		fmt.Printf("Error putting to S3: %v\n", err)
		return fmt.Errorf("failed to upload CSV to S3: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to S3 bucket %s\n", key, bucketName)
	return nil
}
