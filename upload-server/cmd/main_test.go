//go:build integration
// +build integration

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	dexTesting "github.com/cdcgov/data-exchange-upload/upload-server/testing"
)

var AZURITE_KEY = os.Getenv("AZURITE_STORAGE_KEY")

func TestTus(t *testing.T) {
	for name, c := range cases {
		log.Println("Starting case", name)
		setUp(name, c)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			main()
		}()
		// wait for main to start (should make this more resilient)
		time.Sleep(1 * time.Second)

		url := fmt.Sprintf("http://localhost:%s", os.Getenv("SERVER_PORT"))
		var twg sync.WaitGroup
		for name, c := range dexTesting.Cases {
			twg.Add(1)
			go func(t *testing.T) {
				defer twg.Done()
				if _, err := dexTesting.RunTusTestCase(url, "../testing/test.txt", c); err != nil {
					t.Error(name, err)
				} else {
					t.Log("test case", name, "passed")
				}
			}(t)
		}
		twg.Wait()
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		wg.Wait()
	}
}

// GetFreePort asks the kernel for a free open port that is ready to use.
// credit: https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func setUp(name string, c map[string]string) {
	// clear the environment to prevent anything exciting
	os.Clearenv()
	port, err := GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("LOCAL_FOLDER_UPLOADS_TUS", fmt.Sprintf("./tests/%s/uploads", name))
	os.Setenv("LOCAL_REPORTS_FOLDER", fmt.Sprintf("./tests/%s/uploads/reports", name))
	os.Setenv("LOCAL_EVENTS_FOLDER", fmt.Sprintf("./tests/%s/uploads/events", name))
	os.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))
	os.Setenv("UI_PORT", "")
	os.Setenv("REDIS_CONNECTION_STRING", "redis://redispw@cache:6379")
	os.Setenv("SQS_SUBSCRIBER_EVENT_ARN", "arn:aws:sqs:us-east-1:000000000042:test-topic")
	os.Setenv("SNS_PUBLISHER_EVENT_ARN", "arn:aws:sns:us-east-1:000000000042:test-topic")
	os.Setenv("SNS_REPORTER_EVENT_ARN", "arn:aws:sns:us-east-1:000000000042:report-topic")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY_ID", "LSIAQAAAAAAVNCBMPNSG")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "bogus")
	os.Setenv("AWS_ENDPOINT_URL", "http://localstack:4566")
	for key, val := range c {
		os.Setenv(key, val)
	}
}

var cases = map[string]map[string]string{
	"s3_to_azure": {
		"UPLOAD_CONFIG_PATH":       "../../upload-configs",
		"S3_ENDPOINT":              "http://localstack:4566",
		"S3_BUCKET_NAME":           "test-bucket",
		"AZURITE_KEY":              AZURITE_KEY,
		"DEX_DELIVERY_CONFIG_FILE": "../configs/testing/azurite_destinations.yml",
	},
	"azure_to_s3": {
		"UPLOAD_CONFIG_PATH":       "../../upload-configs",
		"AZURE_STORAGE_ACCOUNT":    "devstoreaccount1",
		"AZURE_STORAGE_KEY":        AZURITE_KEY,
		"AZURE_ENDPOINT":           "http://azurite:10000/devstoreaccount1",
		"TUS_AZURE_CONTAINER_NAME": "test",
		"DEX_DELIVERY_CONFIG_FILE": "../configs/testing/minio_destinations.yml",
	},
	"file_to_both": {
		"UPLOAD_CONFIG_PATH":       "../../upload-configs",
		"AZURITE_KEY":              AZURITE_KEY,
		"DEX_DELIVERY_CONFIG_FILE": "../configs/testing/mix_destinations.yml",
	},
}
