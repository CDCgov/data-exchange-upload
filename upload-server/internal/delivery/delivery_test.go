package delivery

import (
	"context"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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

func getAzureSource() *AzureSource {
	azConfig := &appconfig.AzureStorageConfig{
		StorageName:       os.Getenv("AZURE_STORAGE_ACCOUNT"),
		StorageKey:        os.Getenv("AZURE_STORAGE_KEY"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ContainerEndpoint: os.Getenv("AZURE_ENDPOINT"),
	}
	tusContainerClient, err := storeaz.NewContainerClient(*azConfig, "container1")
	printError(err)
	src := &AzureSource{
		FromContainerClient: tusContainerClient,
		StorageContainer:    "container1",
		Prefix:              "upload",
	}
	return src
}

func getS3Source() *S3Source {
	config := &appconfig.S3StorageConfig{
		Endpoint:   os.Getenv("S3_ENDPOINT"),
		BucketName: os.Getenv("S3_BUCKET_NAME"),
	}
	client, err := stores3.New(context.TODO(), config)
	printError(err)
	return &S3Source{
		FromClient: client,
		BucketName: os.Getenv("S3_BUCKET_NAME"),
		Prefix:     "upload",
	}
}

func getAzureDestination(folder string) *AzureDestination {
	return &AzureDestination{
		Name:              "test",
		StorageAccount:    os.Getenv("AZURE_STORAGE_ACCOUNT"),
		StorageKey:        os.Getenv("AZURE_STORAGE_KEY"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ContainerName:     "container2",
		ContainerEndpoint: os.Getenv("AZURE_ENDPOINT"),
		PathTemplate:      folder + "/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
	}
}

func getS3Destination(folder string) *S3Destination {
	return &S3Destination{
		BucketName:      os.Getenv("S3_BUCKET_NAME"),
		Name:            "test",
		PathTemplate:    folder + "/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("AWS_REGION"),
	}
}

func getFileSource() *FileSource {
	fromPathStr := "../../testing/test/uploads/"
	fromPath := os.DirFS(fromPathStr)
	return &FileSource{
		FS: fromPath,
	}
}

func getFileDestination(folder string) *FileDestination {
	return &FileDestination{
		ToPath:       fmt.Sprintf("../../testing/test/%s/", folder),
		Name:         "test",
		PathTemplate: "",
	}
}

func TestDeliverAzureToAzure(t *testing.T) {
	src := getAzureSource()
	dest := getAzureDestination("test-deliver-azure")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)

}

func TestDeliverAzureToS3(t *testing.T) {
	src := getAzureSource()
	dest := getS3Destination("test-deliver-azure")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverS3toS3(t *testing.T) {
	src := getS3Source()
	dest := getS3Destination("test-deliver-s3")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverS3toAzure(t *testing.T) {
	src := getS3Source()
	dest := getAzureDestination("test-deliver-s3")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverFileToAzure(t *testing.T) {
	src := getFileSource()
	dest := getAzureDestination("test-deliver_file")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverFileToS3(t *testing.T) {
	src := getFileSource()
	dest := getS3Destination("test-deliver-file")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}

func TestDeliverFileToFile(t *testing.T) {
	src := getFileSource()
	dest := getFileDestination("test-deliver-file")
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}
