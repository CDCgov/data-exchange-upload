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

func TestTus(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s", os.Getenv("SERVER_PORT"))
	var wg sync.WaitGroup
	for name, c := range dexTesting.Cases {
		wg.Add(1)
		go func(t *testing.T) {
			defer wg.Done()
			if _, err := dexTesting.RunTusTestCase(url, "../testing/test/test.txt", c); err != nil {
				t.Error(name, err)
			} else {
				t.Log("test case", name, "passed")
			}
		}(t)
	}
	wg.Wait()
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

func init() {
	// clear the environment to prevent anything exciting
	os.Clearenv()
	port, err := GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))
	os.Setenv("UPLOAD_CONFIG_PATH", "../../upload-configs")
	os.Setenv("S3_ENDPOINT", "http://minio:8000")
	os.Setenv("S3_BUCKET_NAME", "test-bucket")
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY_ID", "minioadmin")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "minioadmin")

	os.Setenv("EDAV_STORAGE_ACCOUNT", "devstoreaccount1")
	os.Setenv("EDAV_STORAGE_KEY", "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==")
	os.Setenv("EDAV_ENDPOINT", "http://azurite:10000/devstoreaccount1")
	os.Setenv("EDAV_CHECKPOINT_CONTAINER_NAME", "edav")
}

func TestMain(m *testing.M) {

	go main()
	// wait for main to start (should make this more resilient)
	time.Sleep(1 * time.Second)
	result := m.Run()
	syscall.Kill(os.Getpid(), syscall.SIGINT)
	os.Exit(result)
}
