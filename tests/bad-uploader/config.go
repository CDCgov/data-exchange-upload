package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"math/rand"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var (
	url         string
	infoUrl     string
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
		"data_stream_id":    "dextesting",
		"data_stream_route": "testevent1",
		"received_filename": "test",
		"sender_id":         "dex simulation harness",
		"data_producer_id":  "dex simulation harness",
		"jurisdiction":      "test",
	}
	manifestTargets = []string{"edav"}

	testcase TestCase
	cases    TestCases

	templatePath string
	repetitions  int

	reportsURL string

	duration time.Duration
	conf     *config
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

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {

	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		dd := Duration(time.Duration(value))
		*d = dd
		return nil
	case string:
		var err error
		dd, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(dd)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type TestCase struct {
	Chunk                   float64
	Size                    float64
	Manifest                map[string]string
	TemplateFile            string
	Templates               []SubTemplate
	Repetitions             int
	TimeLimit               Duration `json:"time_limit"`
	ExpectedDeliveryTargets []string `json:"expected_delivery_targets"`
	ExpectedReports         []Report `json:"expected_reports"`
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
	cases  []TestCase
	i      int
	random bool
}

func (t *TestCases) Next() TestCase {
	c := t.cases[t.i]
	t.i = (t.i + 1) % len(t.cases)
	if t.random {
		t.i = rand.Intn(len(t.cases))
	}
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

func fromEnv[T any](key string, backup T, conv func(string) (T, error)) T {
	if val, ok := os.LookupEnv(key); ok {
		result, err := conv(val)
		if err != nil {
			return result
		}
		slog.Error("Failed to convert env var, falling back to default", "error", err, "env", key)
	}
	return backup
}
func passthroughString(s string) (string, error) {
	return s, nil
}

func init() {
	flag.Float64Var(&size, "size", 5, "the size of the file to upload, in MB")
	flag.StringVar(&url, "url", fromEnv("UPLOAD_URL", "http://localhost:8080/files", passthroughString), "the upload url for the tus server")
	flag.StringVar(&infoUrl, "info-url", "", "the url for the info endpoint")
	flag.StringVar(&reportsURL, "reports-url", fromEnv("DEX_REPORTS_URL", "", passthroughString), "the url for the reports graphql server")
	flag.IntVar(&parallelism, "parallelism", fromEnv("UPLOAD_PARALLELISM", runtime.NumCPU(), strconv.Atoi), "the number of parallel threads to use, defaults to MAXGOPROC when set to < 1.")
	flag.IntVar(&load, "load", fromEnv("UPLOAD_LOAD", 0, strconv.Atoi), "set the number of files to load, defaults to 0 and adjusts based on benchmark logic")
	flag.Float64Var(&chunk, "chunk", 2, "set the chunk size to use when uploading files in MB")
	flag.StringVar(&samsURL, "sams-url", fromEnv("UPLOAD_SAMS_OAUTH", "", passthroughString), "use sams to authenticate to the upload server")
	flag.StringVar(&username, "username", fromEnv("UPLOAD_USERNAME", "", passthroughString), "username for sams")
	flag.StringVar(&password, "password", fromEnv("UPLOAD_PASSWORD", "", passthroughString), "password for sams")
	flag.BoolVar(&verbose, "v", false, "turn on debug logs")
	flag.Var(&manifest, "manifest", "The manifest to use for the load test.")
	flag.Var(&testcase, "case-file", "A json file describing the test case to use.")
	flag.Var(&cases, "case-dir", "A directory of test cases.")
	flag.StringVar(&templatePath, "template", "", "The path to a template file to use to generate test files")
	flag.IntVar(&repetitions, "repetitions", 1, "The number of times to repeat a template when building a file")
	flag.DurationVar(&duration, "duration", 0, "the duration to run load for.")
	flag.StringVar(&patchURL, "patch-url", "", "Override the base url to use to upload the file itself after upload creation.")
	flag.BoolVar(&cases.random, "random", false, "Randomly select the next test case to run, only affects anything if multiple test cases are used.")
	flag.Parse()
	chunk = chunk * 1024 * 1024
	size = size * 1024 * 1024
	programLevel := new(slog.LevelVar) // Info by default
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel, AddSource: true})
	slog.SetDefault(slog.New(h))
	if verbose {
		programLevel.Set(slog.LevelDebug)
	}

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
	if !flagset["case-file"] {
		testcase = TestCase{
			Chunk:                   chunk,
			Size:                    size,
			Manifest:                manifest,
			TimeLimit:               Duration(10 * time.Second),
			ExpectedDeliveryTargets: manifestTargets,
		}
		if templatePath != "" {
			testcase.TemplateFile = templatePath
			testcase.Repetitions = repetitions
		}
	}
	if !flagset["case-dir"] {
		cases.cases = []TestCase{testcase}
	}
	if !flagset["info-url"] {
		serverUrl, _ := path.Split(url)
		infoUrl, _ = neturl.JoinPath(serverUrl, "info")
	}
	slog.Debug("testing with cases", "cases", cases)
	conf = resultOrFatal(buildConfig())
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

func getExecutor() Executor {
	if duration > 0 {
		return DurationExecutor{
			Duration: duration,
		}
	}
	if load > 0 {
		return SimpleLoadExecutor{
			Load: load,
		}
	}
	return BenchmarkExecutor{}
}

func resultOrFatal[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
