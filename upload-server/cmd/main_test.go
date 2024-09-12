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
}

func TestMain(m *testing.M) {

	go main()
	// wait for main to start (should make this more resilient)
	time.Sleep(1 * time.Second)
	result := m.Run()
	syscall.Kill(os.Getpid(), syscall.SIGINT)
	os.Exit(result)
}
