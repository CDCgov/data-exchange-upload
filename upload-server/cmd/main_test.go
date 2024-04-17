//go:build integration
// +build integration

package main

import (
	"os"
	"testing"
	"time"

	dexTesting "github.com/cdcgov/data-exchange-upload/upload-server/testing"
)

func TestTus(t *testing.T) {
	url := "http://localhost:8080/files/"
	for name, c := range dexTesting.Cases {
		if err := dexTesting.RunTusTestCase(url, "../testing/test/test.txt", c); err != nil {
			t.Error(name, err)
		} else {
			t.Log("test case", name, "passed")
		}
	}
}

func TestMain(m *testing.M) {
	//check integration test env var
	if _, ok := os.LookupEnv("PODMAN_CI_TESTS"); ok {
		go main()
		// wait for main to start (should make this more resilient)
		time.Sleep(5 * time.Second)
		os.Exit(m.Run())
	}
}
