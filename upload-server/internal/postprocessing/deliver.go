package postprocessing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	// "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var ErrBadTarget = fmt.Errorf("bad delivery target")
var ErrSrcFileNotExist = fmt.Errorf("source file does not exist")

var targets = map[string]Deliverer{}

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

// Eventually, this can take a more generic list of deliverer configuration object
func RegisterAllTargets(ctx context.Context, appConfig appconfig.AppConfig) error {
	var edavDeliverer Deliverer
	edavDeliverer, err := NewFileDeliverer(ctx, "edav", &appConfig)
	if err != nil {
		return err
	}
	var routingDeliverer Deliverer
	routingDeliverer, err = NewFileDeliverer(ctx, "routing", &appConfig)
	if err != nil {
		return err
	}

	if appConfig.EdavConnection != nil {
		edavDeliverer, err = NewAzureDeliverer(ctx, "edav", &appConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to edav deliverer target %w", err)
		}

		health.Register(edavDeliverer)
	}
	if appConfig.RoutingConnection != nil {
		routingDeliverer, err = NewAzureDeliverer(ctx, "routing", &appConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to routing deliverer target %w", err)
		}

		health.Register(routingDeliverer)
	}

	if appConfig.EdavS3Connection != nil {
		edavDeliverer, err = NewS3Deliverer(ctx, "edav", appConfig.S3Connection, appConfig.EdavS3Connection, appConfig.TusUploadPrefix)
		if err != nil {
			return fmt.Errorf("failed to connect to edav deliverer target %w", err)
		}

		health.Register(edavDeliverer)
	}

	if appConfig.RoutingS3Connection != nil {
		routingDeliverer, err = NewS3Deliverer(ctx, "routing", appConfig.S3Connection, appConfig.RoutingS3Connection, appConfig.TusUploadPrefix)
		if err != nil {
			return fmt.Errorf("failed to connect to routing deliverer target for S3 %w", err)
		}
		health.Register(routingDeliverer)
	}

	RegisterTarget("edav", edavDeliverer)
	RegisterTarget("routing", routingDeliverer)

	return nil
}

func RegisterTarget(name string, d Deliverer) {
	targets[name] = d
}

func GetTarget(name string) (Deliverer, bool) {
	d, ok := targets[name]
	return d, ok
}

type Deliverer interface {
	health.Checkable
	Deliver(ctx context.Context, tuid string, metadata map[string]string) error
	GetMetadata(ctx context.Context, tuid string) (map[string]string, error)
	GetSrcUrl(ctx context.Context, tuid string) (string, error)
	GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error)
}

func NewFileDeliverer(_ context.Context, target string, appConfig *appconfig.AppConfig) (*FileDeliverer, error) {
	localConfig, err := appconfig.LocalStoreConfig(target, appConfig)
	if err != nil {
		return nil, err
	}

	return &FileDeliverer{
		LocalStorageConfig: *localConfig,
		Target:             target,
	}, nil
}

func NewAzureDeliverer(ctx context.Context, target string, appConfig *appconfig.AppConfig) (*AzureDeliverer, error) {
	config, err := appconfig.GetAzureContainerConfig(target)
	if err != nil {
		return nil, err
	}
	// TODO Can the tus container client be singleton?
	tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
	if err != nil {
		return nil, err
	}
	checkpointContainerClient, err := storeaz.NewContainerClient(config.AzureStorageConfig, config.ContainerName)
	if err != nil {
		return nil, err
	}
	checkpointClient, err := storeaz.NewBlobClient(config.AzureStorageConfig)
	if err != nil {
		return nil, err
	}
	err = storeaz.CreateContainerIfNotExists(ctx, checkpointContainerClient)
	if err != nil {
		return nil, err
	}

	return &AzureDeliverer{
		FromContainerClient: tusContainerClient,
		ToClient:            checkpointClient,
		ToContainer:         config.ContainerName,
		ToContainerClient:   checkpointContainerClient,
		TusPrefix:           appConfig.TusUploadPrefix,
		Target:              target,
	}, nil
}

func NewS3Deliverer(ctx context.Context, target string, srcS3Connection *appconfig.S3StorageConfig, destS3Connection *appconfig.S3StorageConfig, tusPrefix string) (*S3Deliverer, error) {
	s3SrcClientSrc, err := stores3.New(ctx, srcS3Connection)
	if err != nil {
		return nil, err
	}

	s3SrcClientDest, err := stores3.New(ctx, destS3Connection)
	if err != nil {
		return nil, err
	}

	// Return the initialized S3Deliverer
	return &S3Deliverer{
		SrcBucket:  srcS3Connection.BucketName,
		DestBucket: destS3Connection.BucketName,
		SrcClient:  s3SrcClientSrc,
		DestClient: s3SrcClientDest,
		TusPrefix:  tusPrefix,
		Target:     target,
	}, nil
}

// target may end up being a type
func Deliver(ctx context.Context, tuid string, target string) error {
	d, ok := targets[target]
	if !ok {
		return ErrBadTarget
	}

	rb := reports.NewBuilder[reports.FileCopyContent](
		"1.0.0",
		reports.StageFileCopy,
		tuid,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC())
	logger.Info("***getting metadata")
	manifest, err := d.GetMetadata(ctx, tuid)
	if err != nil {
		return err
	}
	logger.Info("***got metadata", "m", manifest)
	rb.SetManifest(manifest)

	srcUrl, err := d.GetSrcUrl(ctx, tuid)
	if err != nil {
		return err
	}
	destUrl, err := d.GetDestUrl(ctx, tuid, manifest)
	if err != nil {
		return err
	}
	rb.SetContent(reports.FileCopyContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageFileCopy,
		},
		FileSourceBlobUrl:      srcUrl,
		FileDestinationBlobUrl: destUrl,
		DestinationName:        target,
	})

	defer func() {
		if err != nil {
			logger.Error("failed to copy file", "target", target)
			rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
				Level:   reports.IssueLevelError,
				Message: err.Error(),
			})
		}
		report := rb.Build()
		logger.Info("File Copy Report", "report", report)
		reports.Publish(ctx, report)
	}()
	err = d.Deliver(ctx, tuid, manifest)
	rb.SetEndTime(time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

type FileDeliverer struct {
	appconfig.LocalStorageConfig
	Target string
}

type AzureDeliverer struct {
	FromContainerClient *container.Client
	ToContainerClient   *container.Client
	ToClient            *azblob.Client
	ToContainer         string
	TusPrefix           string
	Target              string
}

// S3Deliverer handles the delivery of files to an S3 bucket.
type S3Deliverer struct {
	SrcBucket  string
	DestBucket string
	SrcClient  *s3.Client
	DestClient *s3.Client
	TusPrefix  string
	Target     string
}

func (fd *FileDeliverer) Deliver(_ context.Context, tuid string, _ map[string]string) error {
	f, err := fd.FromPath.Open(tuid)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrSrcFileNotExist
		}
		return err
	}
	defer f.Close()
	os.Mkdir(fd.ToPath, 0755)
	dest, err := os.Create(filepath.Join(fd.ToPath, tuid))
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, f); err != nil {
		return err
	}

	return err
}

func (fd *FileDeliverer) GetMetadata(_ context.Context, tuid string) (map[string]string, error) {
	f, err := fd.FromPath.Open(tuid + ".meta")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrSrcFileNotExist
		}
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (fd *FileDeliverer) GetSrcUrl(_ context.Context, tuid string) (string, error) {
	_, err := fs.Stat(fd.FromPath, tuid)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", ErrSrcFileNotExist
		}
		return "", err
	}
	return fd.FromPathStr + tuid, nil
}

func (fd *FileDeliverer) GetDestUrl(_ context.Context, tuid string, _ map[string]string) (string, error) {
	return fd.ToPath + "/" + tuid, nil
}

func (fd *FileDeliverer) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Deliver Target " + fd.Target
	info, err := os.Stat(fd.ToPath)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", fd.ToPath)
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (ad *AzureDeliverer) Deliver(ctx context.Context, tuid string, manifest map[string]string) error {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	blobName, err := getDeliveredFilename(ctx, ad.Target, tuid, manifest)
	if err != nil {
		return err
	}
	destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
	s, err := srcBlobClient.DownloadStream(ctx, nil)
	defer s.Body.Close()
	if s.ErrorCode != nil && *s.ErrorCode == string(bloberror.BlobNotFound) {
		return ErrSrcFileNotExist
	}
	if err != nil {
		return err
	}

	logger.Info("starting copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

	_, err = ad.ToClient.UploadStream(ctx, ad.ToContainer, blobName, s.Body, &azblob.UploadStreamOptions{
		Metadata: storeaz.PointerizeMetadata(manifest),
	})
	if err != nil {
		return err
	}

	logger.Info("successful copy from", "src", srcBlobClient.URL(), "to dest", destBlobClient.URL())

	return nil
}

func (ad *AzureDeliverer) GetMetadata(ctx context.Context, tuid string) (map[string]string, error) {
	// Get blob src blob client.
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	resp, err := srcBlobClient.GetProperties(ctx, nil)
	if err != nil {
		return nil, err
	}
	return storeaz.DepointerizeMetadata(resp.Metadata), nil
}

func (ad *AzureDeliverer) GetSrcUrl(_ context.Context, tuid string) (string, error) {
	srcBlobClient := ad.FromContainerClient.NewBlobClient(ad.TusPrefix + "/" + tuid)
	return srcBlobClient.URL(), nil
}

func (ad *AzureDeliverer) GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error) {
	blobName, err := getDeliveredFilename(ctx, ad.Target, tuid, manifest)
	if err != nil {
		return "", err
	}
	destBlobClient := ad.ToContainerClient.NewBlobClient(blobName)
	return destBlobClient.URL(), nil
}

func (ad *AzureDeliverer) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure deliver target " + ad.Target
	rsp.Status = models.STATUS_UP

	if ad.ToContainerClient == nil {
		// Running in azure, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Azure deliverer target " + ad.Target + " not configured"
	}

	_, err := ad.ToContainerClient.GetProperties(ctx, nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}

	return rsp
}

func (sd *S3Deliverer) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "AWS S3 deliver target " + sd.Target
	rsp.Status = models.STATUS_UP

	if sd.SrcBucket == "" {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "AWS S3 deliverer Source Bucket not configured"
	}

	if sd.DestBucket == "" {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "AWS S3 deliverer Destination Bucket not configured"
	}

	if sd.SrcClient == nil {
		// Running in aws, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "AWS S3 deliverer src client not configured"
	}

	if sd.DestClient == nil {
		// Running in aws, but deliverer not set up.
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "AWS S3 deliverer target client not configured"
	}

	_, err := sd.DestClient.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: &sd.DestBucket,
	})
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Error pinging destination bucket " + err.Error()
	}

	return rsp
}

type writeAtWrapper struct {
	writer io.Writer
}

func (w *writeAtWrapper) WriteAt(p []byte, _ int64) (int, error) {
	// Ignoring offset because we force sequential writing
	return w.writer.Write(p)
}

func (sd *S3Deliverer) Deliver(ctx context.Context, tuid string, manifest map[string]string) error {
	logger.Info("***in deliverer")
	id := strings.Split(tuid, "+")[0]
	srcFilename := sd.TusPrefix + "/" + id
	destFileName, err := getDeliveredFilename(ctx, sd.Target, tuid, manifest)
	if err != nil {
		return err
	}
	logger.Info("***deliver filename", "filename", destFileName)

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

// TODO get from client
func (sd *S3Deliverer) GetSrcUrl(_ context.Context, tuid string) (string, error) {
	// Construct the S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.SrcBucket, sd.TusPrefix+"/"+tuid)
	return s3URL, nil
}

// TODO get from client
func (sd *S3Deliverer) GetDestUrl(ctx context.Context, tuid string, manifest map[string]string) (string, error) {
	objectKey, err := getDeliveredFilename(ctx, sd.Target, tuid, manifest)
	if err != nil {
		return "", err
	}

	// Construct the S3 URL
	s3URL := fmt.Sprintf("https://%s.s3.us-east-1.amazonaws.com/%s", sd.DestBucket, objectKey)
	return s3URL, nil
}

func getDeliveredFilename(ctx context.Context, target string, tuid string, manifest map[string]string) (string, error) {
	// First, build the filename from the manifest and config.  This will be the default.
	filename := metadataPkg.GetFilename(manifest)
	extension := filepath.Ext(filename)
	filenameWithoutExtension := strings.TrimSuffix(filename, extension)

	suffix, err := metadata.GetFilenameSuffix(ctx, manifest, tuid)
	if err != nil {
		return "", err
	}
	blobName := filenameWithoutExtension + suffix + extension

	// Next, need to set the filename prefix based on config and target.
	// edav, routing, s3 -> use config
	prefix := ""

	switch target {
	case "routing", "edav", "s3":
		prefix, err = metadata.GetFilenamePrefix(ctx, manifest)
		if err != nil {
			return "", err
		}
	}
	return prefix + "/" + blobName, nil
}
