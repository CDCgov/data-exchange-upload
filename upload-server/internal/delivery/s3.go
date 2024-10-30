package delivery

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

const size5MB = 5 * 1024 * 1024
const s3MaxCopySize = size5MB * 1024
const s3MaxParts = 10000

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
	srcFileName := ss.GetSourceFilePath(path)
	downloader := manager.NewDownloader(ss.FromClient, func(d *manager.Downloader) {
		d.Concurrency = 1
	})

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

func (ss *S3Source) SourceType() string {
	return storageTypeS3
}

func (ss *S3Source) Container() string {
	return ss.BucketName
}

func (ss *S3Source) GetSourceFilePath(tuid string) string {
	// Temp workaround for getting the real upload ID without the hash.  See https://github.com/tus/tusd/pull/1167
	id := strings.Split(tuid, "+")[0]
	return ss.Prefix + "/" + id
}

func (ss *S3Source) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get the object from S3
	srcFilename := ss.GetSourceFilePath(tuid)
	output, err := ss.FromClient.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(ss.BucketName),
		Key:    aws.String(srcFilename),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve object: %w", err)
	}

	props := output.Metadata
	props["last_modified"] = output.LastModified.Format(time.RFC3339Nano)
	props["content_length"] = strconv.FormatInt(*output.ContentLength, 10)
	return props, nil
}

func (ss *S3Source) GetSignedObjectURL(ctx context.Context, containerName string, objectPath string) (string, error) {
	presignClient := s3.NewPresignClient(ss.FromClient)
	sourceFile := ss.GetSourceFilePath(objectPath)
	request, err := presignClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(containerName),
			Key:    aws.String(sourceFile),
		},
		func(options *s3.PresignOptions) {
			options.Expires = time.Hour
		},
	)
	if err != nil {
		return "", fmt.Errorf("could not obtain presigned url: %s", err.Error())
	}
	return request.URL, nil
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

func (sd *S3Destination) DestinationType() string {
	return storageTypeS3
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

func (sd *S3Destination) Copy(ctx context.Context, path string, source *Source, metadata map[string]string,
	length int64, concurrency int) (string, error) {
	s := *source
	if s.SourceType() == sd.DestinationType() {
		// copy s3 to s3
		// going to assume we have correct IAM permissions
		return sd.copyFromLocalStorage(ctx, source, path, metadata, length, concurrency)
	} else {
		// stream
		reader, err := s.Reader(ctx, path)
		if err != nil {
			return "", fmt.Errorf("unable to get source stream reader: %v", err)
		}
		return sd.Upload(ctx, path, reader, metadata)
	}

}

func (sd *S3Destination) copyFromLocalStorage(ctx context.Context, source *Source, path string,
	sourceMetadata map[string]string, sourceLength int64, concurrency int) (string, error) {
	s := *source
	sourceFile := s.GetSourceFilePath(path)
	sourceContainer := s.Container()
	sourcePath := fmt.Sprintf("%s/%s", sourceContainer, sourceFile)
	destFile, err := getDeliveredFilename(ctx, sourceFile, sd.PathTemplate, sourceMetadata)
	if err != nil {
		return "", fmt.Errorf("unable to determine destination object name: %v", err)
	}
	client := sd.Client()
	if sourceLength < s3MaxCopySize {
		_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
			CopySource: aws.String(sourcePath),
			Bucket:     aws.String(sd.BucketName),
			Key:        aws.String(destFile),
		})
		if err != nil {
			return "", fmt.Errorf("unable to copy object to S3 bucket: %v", err)
		}
		return fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.BucketName, destFile), nil
	} else {
		lengthInt := int(sourceLength)
		upload, err := client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
			Bucket:   aws.String(sd.BucketName),
			Key:      aws.String(destFile),
			Metadata: sourceMetadata,
		})
		if err != nil {
			return "", fmt.Errorf("unable to create multipart upload: %v", err)
		}
		uploadId := *upload.UploadId
		var partSize = size5MB
		if lengthInt > partSize*s3MaxParts {
			// we need to increase the Part size
			partSize = lengthInt / s3MaxParts
		}
		numChunks := lengthInt / partSize
		var chunkNum int
		var start = 0
		var end = 0
		chunkIdMap := make(map[int]string)
		for chunkNum = 1; chunkNum <= numChunks; chunkNum++ {
			end = start + partSize - 1
			if chunkNum == numChunks {
				end = lengthInt - 1
			}
			chunkIdMap[chunkNum] = fmt.Sprintf("bytes=%d-%d", start, end)
			start = end + 1
		}
		wg := sync.WaitGroup{}
		errCh := make(chan error, 1)
		responseCh := make(chan types.CompletedPart, numChunks)
		completedParts := make([]types.CompletedPart, numChunks)

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		routines := 0
		for chunkId, rangeHeader := range chunkIdMap {
			wg.Add(1)
			routines++
			go func(chunkId int, rangeHeader string) {
				defer wg.Done()
				uploadPartResp, err := client.UploadPartCopy(ctx, &s3.UploadPartCopyInput{
					Bucket:          aws.String(sd.BucketName),
					CopySource:      aws.String(sourcePath),
					CopySourceRange: aws.String(rangeHeader),
					Key:             aws.String(destFile),
					PartNumber:      aws.Int32(int32(chunkId)),
					UploadId:        aws.String(uploadId),
				})
				if err != nil {
					select {
					case errCh <- err:
						// error was set
					default:
						// some other error is already set
					}
					cancel()
				} else {
					responseCh <- types.CompletedPart{
						ETag:       uploadPartResp.CopyPartResult.ETag,
						PartNumber: aws.Int32(int32(chunkId)),
					}
				}

			}(chunkId, rangeHeader)
			if routines >= concurrency {
				wg.Wait()
				routines = 0
			}
		}
		wg.Wait()
		close(responseCh)
		select {
		case err = <-errCh:
			// there was an error during staging
			_, _ = client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
				Bucket:   aws.String(sd.BucketName),
				Key:      aws.String(destFile),
				UploadId: aws.String(uploadId),
			})
			return "", fmt.Errorf("error staging blocks; copy aborted: %v", err)
		default:
			// no error was encountered
		}

		// arrange parts in ordered list
		for part := range responseCh {
			partNum := *part.PartNumber
			completedParts[partNum-1] = part
		}

		completion, err := client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
			Bucket:   aws.String(sd.BucketName),
			Key:      aws.String(destFile),
			UploadId: aws.String(uploadId),
			MultipartUpload: &types.CompletedMultipartUpload{
				Parts: completedParts,
			},
		})
		if err != nil {
			return "", fmt.Errorf("unable to complete multipart upload: %v", err)
		}
		return *completion.Location, nil
	}

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
