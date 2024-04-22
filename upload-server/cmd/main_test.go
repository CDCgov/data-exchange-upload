//go:build integration
// +build integration

package main

import (
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	dexTesting "github.com/cdcgov/data-exchange-upload/upload-server/testing"
)

func TestTus(t *testing.T) {
	url := "http://localhost:8080/files/"
	var wg sync.WaitGroup
	for name, c := range dexTesting.Cases {
		wg.Add(1)
		go func(t *testing.T) {
			defer wg.Done()
			if err := dexTesting.RunTusTestCase(url, "../testing/test/test.txt", c); err != nil {
				t.Error(name, err)
			} else {
				t.Log("test case", name, "passed")
			}
		}(t)
	}
	wg.Wait()
}

func TestMain(m *testing.M) {
	//check integration test env var
	if _, ok := os.LookupEnv("PODMAN_CI_TESTS"); ok {
		go main()
		// wait for main to start (should make this more resilient)
		time.Sleep(5 * time.Second)
		result := m.Run()
		syscall.Kill(os.Getpid(), syscall.SIGINT)
		os.Exit(result)
	}
}
