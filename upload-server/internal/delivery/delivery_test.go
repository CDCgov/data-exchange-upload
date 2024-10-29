package delivery

import (
	"context"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
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

func TestDeliverAzureToAzure(t *testing.T) {
	src := getAzureSource()
	dest := &AzureDestination{
		Name:              "test",
		StorageAccount:    os.Getenv("AZURE_STORAGE_ACCOUNT"),
		StorageKey:        os.Getenv("AZURE_STORAGE_KEY"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ContainerName:     "container2",
		ContainerEndpoint: os.Getenv("AZURE_ENDPOINT"),
		PathTemplate:      "test-folder/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
	}
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)

}

func TestDeliverAzureToS3(t *testing.T) {
	src := getAzureSource()
	dest := &S3Destination{
		BucketName:      os.Getenv("S3_BUCKET_NAME"),
		Name:            "test",
		PathTemplate:    "test-folder/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("AWS_REGION"),
	}
	url, err := Deliver(context.TODO(), "test.HL7", src, dest)
	printError(err)
	assert.True(t, url != "")
	assert.True(t, err == nil)
}
