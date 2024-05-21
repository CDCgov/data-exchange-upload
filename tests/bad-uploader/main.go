package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/eventials/go-tus"
	"golang.org/x/oauth2"
)

var (
	url         string
	size        int
	parallelism int
	load        int
	chunk       int64
	username    string
	password    string
	samsURL     string
	verbose     bool
)

func init() {
	flag.IntVar(&size, "size", 250*10000, "the size of the file to upload, in bytes")
	flag.StringVar(&url, "url", "http://localhost:8080/files/", "the upload url for the tus server")
	flag.IntVar(&parallelism, "parallelism", runtime.NumCPU(), "the number of parallel threads to use, defaults to MAXGOPROC when set to < 1.")
	flag.IntVar(&load, "load", 0, "set the number of files to load, defaults to 0 and adjusts based on benchmark logic")
	flag.Int64Var(&chunk, "chunk", 2, "set the chunk size to use when uploading files in MB")
	flag.StringVar(&samsURL, "sams-url", "", "use sams to authenticate to the upload server")
	flag.StringVar(&username, "username", "", "username for sams")
	flag.StringVar(&password, "password", "", "password for sams")
	flag.BoolVar(&verbose, "v", false, "turn on debug logs")
	flag.Parse()
	chunk = chunk * 1024 * 1024
	programLevel := new(slog.LevelVar) // Info by default
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	if verbose {
		programLevel.Set(slog.LevelDebug)
	}
}

func buildConfig() (*config, error) {
	var tokenSource *SAMSTokenSource
	tokenSource = nil

	if samsURL != "" {
		tokenSource = &SAMSTokenSource{
			username: username,
			password: password,
			url:      samsURL,
		}
	}

	return &config{
		url:         url,
		tokenSource: tokenSource,
	}, nil
}

func resultOrFatal[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

var wg sync.WaitGroup

/*
so we need to be able to create an arbirary number of test uploads
should have a context on them
let them run conncurrently
and get some meaningful output that we can go through after the fact maybe in a few channels
this will cover a use case for a single bad sender, so only one cred needed
*/

func main() {

	conf := resultOrFatal(buildConfig())

	c := make(chan struct{}, parallelism)
	slog.Info("Starting threads", "parallelism", parallelism)
	for i := 0; i < parallelism; i++ {
		go worker(c, conf)
	}

	tStart := time.Now()
	if load > 0 {
		slog.Info("Running load test", "uploads", load)
		for i := 0; i < load; i++ {
			wg.Add(1)
			c <- struct{}{}
		}
	} else {
		slog.Info("Running benchmark")
		result := testing.Benchmark(asPallelBenchmark(c))
		defer fmt.Println(result.String())
	}
	wg.Wait()
	fmt.Println("Benchmarking took ", time.Since(tStart).Seconds(), " seconds")
}

type config struct {
	url         string
	tokenSource *SAMSTokenSource
}

func worker(c <-chan struct{}, conf *config) {
	for range c {
		f := &BadHL7{
			Size:           size,
			DummyGenerator: &RandomBytesReader{},
		}
		if err := runTest(f, conf); err != nil {
			slog.Error("ERROR: ", "error", err)
		}
		wg.Done()
	}
}

func asPallelBenchmark(c chan struct{}) func(*testing.B) {
	return func(b *testing.B) {
		slog.Info("benchmarking", "runs", b.N)
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			c <- struct{}{}
		}
	}
}

func runTest(f *BadHL7, conf *config) error {
	//not great
	defer f.Close()
	// create the tus client.
	tusConf := tus.DefaultConfig()
	tusConf.ChunkSize = chunk
	if conf.tokenSource != nil {
		tusConf.HttpClient = oauth2.NewClient(context.TODO(), conf.tokenSource)
	}
	client, err := tus.NewClient(conf.url, tusConf)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// create an upload from a file.
	upload := tus.NewUpload(f, int64(f.Size), f.Metadata(), f.Fingerprint())

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return fmt.Errorf("failed to create upload: %w", err)
	}

	for uploader.Offset() < upload.Size() && !uploader.IsAborted() {
		err := uploader.UploadChunck()

		if err != nil {
			return err
		}
		slog.Debug("uploaded", "offset", uploader.Offset(), "size", upload.Size())
	}

	return nil
}

type Generator interface {
	io.Reader
}

type RandomBytesReader struct{}

func (rb *RandomBytesReader) Read(b []byte) (int, error) {
	return rand.Read(b)
}

type BadHL7 struct {
	Size           int
	offset         int
	DummyGenerator Generator
}

func (b *BadHL7) Metadata() map[string]string {
	return map[string]string{
		"meta_destination_id": "dextesting",
		"meta_ext_event":      "testevent1",
		"filename":            "test",
	}
}

func (b *BadHL7) Fingerprint() string {
	return ""
}

func (b *BadHL7) Read(p []byte) (int, error) {

	// needs to limit size read to size eventually
	i, err := b.DummyGenerator.Read(p)
	if err != nil {
		return i, err
	}

	if b.offset+i > b.Size {
		return b.Size - b.offset, nil
	}
	b.offset += i
	return i, nil
}

func (b *BadHL7) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (b *BadHL7) Close() error {
	return nil
}

type SAMSTokenSource struct {
	username string
	password string
	url      string
	token    *oauth2.Token
	lock     sync.Mutex
}

type SAMSToken struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        string   `json:"scope"`
	Resource     []string `json:"resource"`
}

func (sts *SAMSTokenSource) Token() (*oauth2.Token, error) {
	sts.lock.Lock()
	defer sts.lock.Unlock()

	if sts.token != nil && time.Now().Before(sts.token.Expiry) {
		return sts.token, nil
	}

	tStart := time.Now()
	defer func(tStart time.Time) { fmt.Println("Auth took ", time.Since(tStart).Seconds(), " seconds") }(tStart)

	body := neturl.Values{
		"username": []string{sts.username},
		"password": []string{sts.password},
	}

	resp, err := http.PostForm(sts.url, body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	t := &SAMSToken{}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, err
	}

	sts.token = &oauth2.Token{
		AccessToken:  t.AccessToken,
		TokenType:    t.TokenType,
		RefreshToken: t.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(t.ExpiresIn) * time.Second),
	}
	return sts.token, nil
}
