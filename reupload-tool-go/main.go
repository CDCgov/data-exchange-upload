package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

var (
	uploadAPIUrl string
	inputFiles   CSVFiles
	replays      []Replay
	verbose      bool
)

type CSVFiles []string
type Replay struct {
	Id     string `json:"-"`
	Target string `json:"target"`
}

func (ids CSVFiles) String() string {
	return strings.Join(inputFiles, ",")
}

func (ids CSVFiles) Set(value string) error {
	inputFiles = strings.Split(value, ",")
	return nil
}

func init() {
	flag.StringVar(&uploadAPIUrl, "url", "", "URL of the upload API service")
	flag.Var(&inputFiles, "csvFiles", "file1.csv,file2.csv")
	flag.BoolVar(&verbose, "v", false, "turn on debug logs")
	flag.Parse()

	programLevel := new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	if verbose {
		programLevel.Set(slog.LevelDebug)
	}
}

func main() {
	// First, load upload IDs from CSV
	slog.Debug("reading input files", inputFiles)
	for _, fname := range inputFiles {
		f, err := os.Open(fname)
		if err != nil {
			slog.Error("failed to open input CSV", "filename", fname, "error", err.Error())
			return
		}
		defer f.Close()

		r := csv.NewReader(f)
		// Read header
		if _, err := r.Read(); err != nil {
			slog.Error("error parsing CSV", "file", fname, "error", err.Error())
		}
		// Read all IDs
		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				slog.Error("error parsing CSV", "file", fname, "error", err.Error())
			}
			replays = append(replays, Replay{
				Id:     record[0],
				Target: record[1], // TODO handle index out of bounds
			})
		}
	}
	slog.Info(fmt.Sprintf("replaying %d files", len(replays)))
	slog.Debug("File IDs: %v", replays)

	// Next, send request to replay service.
	for _, replay := range replays {
		url := uploadAPIUrl + "/route" + "/" + replay.Id
		b, err := json.Marshal(replay)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		slog.Debug(fmt.Sprintf("replaying %s for target %s", replay.Id, replay.Target))
		resp, err := http.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			slog.Error(err.Error())
			continue
		}
		if resp.StatusCode != http.StatusOK {
			slog.Error("replay attempt unsuccessful", "response", resp.Status)
		}
	}
	// TODO
	// Count success and fail responses
	// Use a channel and go routines to fan out
	// Smoke flag
}
