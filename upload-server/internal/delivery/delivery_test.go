package delivery_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println("Unable to load .env")
	}
}

func printError(err error) {
	if err != nil {
		println(err.Error())
	}
}

func getAzureSource() *delivery.AzureSource {
	azConfig := &appconfig.AzureStorageConfig{
		StorageName:       os.Getenv("AZURE_STORAGE_ACCOUNT"),
		StorageKey:        os.Getenv("AZURE_STORAGE_KEY"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ContainerEndpoint: os.Getenv("AZURE_ENDPOINT"),
	}
	tusContainerClient, err := storeaz.NewContainerClient(*azConfig, "container1")
	printError(err)
	src := &delivery.AzureSource{
		FromContainerClient: tusContainerClient,
		StorageContainer:    "container1",
		Prefix:              "upload",
	}
	return src
}

func getS3Source() *delivery.S3Source {
	config := &appconfig.S3StorageConfig{
		Endpoint:   os.Getenv("S3_ENDPOINT"),
		BucketName: os.Getenv("S3_BUCKET_NAME"),
	}
	client, err := stores3.New(context.TODO(), config)
	printError(err)
	return &delivery.S3Source{
		FromClient: client,
		BucketName: os.Getenv("S3_BUCKET_NAME"),
		Prefix:     "upload",
	}
}

func getAzureDestination(folder string) *delivery.AzureDestination {
	return &delivery.AzureDestination{
		Name:              "test",
		StorageAccount:    os.Getenv("AZURE_STORAGE_ACCOUNT"),
		StorageKey:        os.Getenv("AZURE_STORAGE_KEY"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ContainerName:     "container2",
		ContainerEndpoint: os.Getenv("AZURE_ENDPOINT"),
		PathTemplate:      folder + "/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
	}
}

func getS3Destination(folder string) *delivery.S3Destination {
	return &delivery.S3Destination{
		BucketName:      os.Getenv("S3_BUCKET_NAME"),
		Name:            "test",
		PathTemplate:    folder + "/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("AWS_REGION"),
	}
}

func getFileSource() *delivery.FileSource {
	fromPathStr := "../../testing/test/uploads/"
	fromPath := os.DirFS(fromPathStr)
	return &delivery.FileSource{
		FS: fromPath,
	}
}

func getFileDestination(folder string) *delivery.FileDestination {
	return &delivery.FileDestination{
		ToPath:       fmt.Sprintf("../../testing/test/%s/", folder),
		Name:         "test",
		PathTemplate: "",
	}
}

func runDeliveryTest(t *testing.T, src delivery.Source, dest delivery.Destination, template string) {
	srcFile := "test.HL7"
	metadata, _ := src.GetMetadata(context.TODO(), srcFile)
	destPath, _ := delivery.GetDeliveredFilename(srcFile, template, metadata)
	url, err := delivery.Deliver(context.TODO(), srcFile, destPath, src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverAzureToAzure(t *testing.T) {
	src := getAzureSource()
	dest := getAzureDestination("test-deliver-azure")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverAzureToS3(t *testing.T) {
	src := getAzureSource()
	dest := getS3Destination("test-deliver-azure")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverS3toS3(t *testing.T) {
	src := getS3Source()
	dest := getS3Destination("test-deliver-s3")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverS3toAzure(t *testing.T) {
	src := getS3Source()
	dest := getAzureDestination("test-deliver-s3")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverFileToAzure(t *testing.T) {
	src := getFileSource()
	dest := getAzureDestination("test-deliver_file")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverFileToS3(t *testing.T) {
	src := getFileSource()
	dest := getS3Destination("test-deliver-file")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestDeliverFileToFile(t *testing.T) {
	src := getFileSource()
	dest := getFileDestination("test-deliver-file")
	runDeliveryTest(t, src, dest, dest.PathTemplate)
}

func TestGetDeliveredFilename(t *testing.T) {
	type testCase struct {
		ctx          context.Context
		tuid         string
		pathTemplate string
		manifest     map[string]string
		err          error
		result       string
	}
	ctx := context.Background()
	testTime := time.Date(2020, time.April, 11, 15, 12, 30, 50, time.UTC)

	testCases := []testCase{
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":             "2.0",
				"data_stream_id":      "dextesting",
				"data_stream_route":   "testevent1",
				"dex_ingest_datetime": testTime.Format(time.RFC3339Nano),
				"filename":            "test.txt",
			},
			nil,
			"routine-immunization-zip/2020/04/11/test.txt",
		},
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":             "2.0",
				"data_stream_id":      "dextesting",
				"data_stream_route":   "testevent1",
				"dex_ingest_datetime": "bogus time",
				"filename":            "test.txt",
			},
			delivery.ErrBadIngestTimestamp,
			"",
		},
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":           "2.0",
				"data_stream_id":    "dextesting",
				"data_stream_route": "testevent1",
				"filename":          "test.txt",
			},
			delivery.ErrBadIngestTimestamp,
			"",
		},
	}

	for i, c := range testCases {
		res, err := delivery.GetDeliveredFilename(c.tuid, c.pathTemplate, c.manifest)
		if res != c.result {
			t.Errorf("missmatched results for test case %d: got %s expected %s", i, res, c.result)
		}
		if !errors.Is(err, c.err) || (c.err == nil && err != nil) {
			t.Errorf("missmatched errors for test case %d: got %+v expected %+v", i, err, c.err)
		}
	}
}
