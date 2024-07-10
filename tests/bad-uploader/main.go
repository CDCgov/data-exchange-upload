package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/eventials/go-tus"
	"golang.org/x/oauth2"
)

var (
	url         string
	size        float64
	parallelism int
	load        int
	chunk       float64
	username    string
	password    string
	samsURL     string
	verbose     bool
	patchURL    string

	manifest = JSONVar{
		"meta_destination_id": "dextesting",
		"meta_ext_event":      "testevent1",
		"filename":            "test",
	}

	testcase TestCase
	cases    TestCases

	templatePath string
	repetitions  int

	duration time.Duration
)

type JSONVar map[string]string

func (f *JSONVar) String() string {
	s, err := json.Marshal(f)
	if err != nil {
		log.Println("failed to create a string value", err)
	}
	return string(s)
}

func (f *JSONVar) Set(s string) error {
	*f = make(JSONVar) // reset f
	return json.Unmarshal([]byte(s), f)
}

type SubTemplate struct {
	Name        string
	Repetitions int
	Args        map[string]any
}

type TestCase struct {
	Chunk        float64
	Size         float64
	Manifest     map[string]string
	TemplateFile string
	Templates    []SubTemplate
	Repetitions  int
}

func (t *TestCase) String() string {
	s, err := json.Marshal(t)
	if err != nil {
		log.Println("failed to create a string value", err)
	}
	return string(s)
}

func (t *TestCase) Set(s string) error {
	f, err := os.Open(s)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, t); err != nil {
		return err
	}

	t.Chunk = t.Chunk * 1024 * 1024
	t.Size = t.Size * 1024 * 1024

	if t.Chunk < 1 {
		return fmt.Errorf("chunk size must be > 1 byte")
	}

	if t.Size < 1 && t.TemplateFile == "" {
		return fmt.Errorf("size of file must be > 1 byte")
	}
	return nil
}

type TestCases struct {
	cases []TestCase
	i     int
}

func (t *TestCases) Next() TestCase {
	c := t.cases[t.i]
	t.i = (t.i + 1) % len(t.cases)
	return c
}

func (t *TestCases) Set(s string) error {
	t.cases = []TestCase{}
	return filepath.WalkDir(s, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			c := TestCase{}
			if err := (&c).Set(path); err != nil {
				return err
			}
			t.cases = append(t.cases, c)
		}
		return nil
	})
}

func (t *TestCases) String() string {
	return ""
}

func init() {
	flag.Float64Var(&size, "size", 5, "the size of the file to upload, in MB")
	flag.StringVar(&url, "url", "http://localhost:8080/files/", "the upload url for the tus server")
	flag.IntVar(&parallelism, "parallelism", runtime.NumCPU(), "the number of parallel threads to use, defaults to MAXGOPROC when set to < 1.")
	flag.IntVar(&load, "load", 0, "set the number of files to load, defaults to 0 and adjusts based on benchmark logic")
	flag.Float64Var(&chunk, "chunk", 2, "set the chunk size to use when uploading files in MB")
	flag.StringVar(&samsURL, "sams-url", "", "use sams to authenticate to the upload server")
	flag.StringVar(&username, "username", "", "username for sams")
	flag.StringVar(&password, "password", "", "password for sams")
	flag.BoolVar(&verbose, "v", false, "turn on debug logs")
	flag.Var(&manifest, "manifest", "The manifest to use for the load test.")
	flag.Var(&testcase, "case-file", "A json file describing the test case to use.")
	flag.Var(&cases, "case-dir", "A directory of test cases.")
	flag.StringVar(&templatePath, "template", "", "The path to a template file to use to generate test files")
	flag.IntVar(&repetitions, "repetitions", 1, "The number of times to repeat a template when building a file")
	flag.DurationVar(&duration, "duration", 0, "the duration to run load for.")
	flag.StringVar(&patchURL, "patch-url", "", "Override the base url to use to upload the file itself after upload creation.")
	flag.Parse()
	chunk = chunk * 1024 * 1024
	size = size * 1024 * 1024
	programLevel := new(slog.LevelVar) // Info by default
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	if verbose {
		programLevel.Set(slog.LevelDebug)
	}

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	if !flagset["case-file"] {
		testcase = TestCase{
			Chunk:    chunk,
			Size:     size,
			Manifest: manifest,
		}
		if templatePath != "" {
			testcase.TemplateFile = templatePath
			testcase.Repetitions = repetitions
		}
	}
	if !flagset["case-dir"] {
		cases = TestCases{cases: []TestCase{testcase}}
	}
	slog.Debug("testing with cases", "cases", cases)
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

	c := make(chan TestCase, parallelism)
	slog.Info("Starting threads", "parallelism", parallelism)
	for i := 0; i < parallelism; i++ {
		go worker(c, conf)
	}

	tStart := time.Now()
	if duration > 0 {
		slog.Info("Running duration test", "duration", duration)
		i := 0
		for {
			if time.Since(tStart) > duration {
				break
			}
			wg.Add(1)
			c <- cases.Next()
			i++
		}
		slog.Info("uploads over time", "uploads", i, "duration", duration)
		slog.Info("roughly in 24 hours", "uploads", int((24*time.Hour)/duration)*i)
	} else if load > 0 {
		slog.Info("Running load test", "uploads", load)
		for i := 0; i < load; i++ {
			wg.Add(1)
			c <- cases.Next()
		}
	} else {
		slog.Info("Running benchmark")
		result := testing.Benchmark(asParallelBenchmark(c, cases.Next))
		defer fmt.Printf("Benchmarking results: %f seconds/op\n", float64(result.NsPerOp())/float64(time.Second))
	}
	wg.Wait()
	fmt.Println("Benchmarking took ", time.Since(tStart).Seconds(), " seconds")
}

type config struct {
	url         string
	tokenSource *SAMSTokenSource
}

func worker(c <-chan TestCase, conf *config) {
	for e := range c {
		if err := runTest(e, conf); err != nil {
			slog.Error("ERROR: ", "error", err, "case", e)
		}
		wg.Done()
	}
}

func asParallelBenchmark(c chan TestCase, next func() TestCase) func(*testing.B) {
	return func(b *testing.B) {
		slog.Info("benchmarking", "runs", b.N)
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			c <- next()
		}
	}
}

type uploadable interface {
	io.ReadSeekCloser
	Size() int64
	Metadata() map[string]string
	Fingerprint() string
}

func runTest(t TestCase, conf *config) error {

	var f uploadable
	if t.TemplateFile != "" {
		f = &TemplateGenerator{
			Repeats:   t.Repetitions,
			Path:      t.TemplateFile,
			Templates: t.Templates,
			Manifest:  t.Manifest,
		}
	} else {
		f = &BadFile{
			FileSize:       int(t.Size),
			Manifest:       t.Manifest,
			DummyGenerator: &RandomBytesReader{},
		}
	}

	// create the tus client.
	tusConf := tus.DefaultConfig()
	tusConf.ChunkSize = int64(t.Chunk)
	tusConf.HttpClient = &http.Client{}
	if conf.tokenSource != nil {
		tusConf.HttpClient = oauth2.NewClient(context.TODO(), conf.tokenSource)
	}
	tusConf.Header.Set("Upload-Defer-Length", "1")
	tusConf.Header.Set("Upload-Length", "")
	client, err := tus.NewClient(conf.url, tusConf)
	if err != nil {
		return fmt.Errorf("failed to create client: %w, %+v", err, t)
	}

	// create an upload from a file.
	upload := tus.NewUpload(f, f.Size(), f.Metadata(), f.Fingerprint())

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return fmt.Errorf("failed to create upload: %w, %+v", err, t)
	}

	if patchURL != "" {
		uploader.SetUrl(path.Join(patchURL, path.Base(uploader.Url())))
	}

	slog.Debug("UploadID", "upload_id", uploader.Url())
	c := make(chan tus.Upload)
	uploader.NotifyUploadProgress(c)
	go func(c chan tus.Upload, url string) {
		for u := range c {
			slog.Debug("Upload Progress", "url", url, "progress", u.Progress())
		}
	}(c, uploader.Url())

	for {
		if err := uploader.UploadChunck(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
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

type BadFile struct {
	FileSize       int
	offset         int
	Manifest       map[string]string
	DummyGenerator Generator
}

func (b *BadFile) Size() int64 {
	return int64(b.FileSize)
}

func (b *BadFile) Metadata() map[string]string {
	return b.Manifest
}

func (b *BadFile) Fingerprint() string {
	return ""
}

func (b *BadFile) Read(p []byte) (int, error) {

	// needs to limit size read to size eventually
	i, err := b.DummyGenerator.Read(p)
	if err != nil {
		return i, err
	}
	log.Println("reading", b.offset, b.FileSize)

	if b.offset+i > b.FileSize {
		i = b.FileSize - b.offset
	}

	b.offset += i

	if b.offset >= b.FileSize {
		return i, io.EOF
	}
	log.Println("read", b.offset, b.FileSize)
	return i, nil
}

func (b *BadFile) Seek(offset int64, whence int) (int64, error) {
	return offset, nil
}

func (b *BadFile) Close() error {
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
