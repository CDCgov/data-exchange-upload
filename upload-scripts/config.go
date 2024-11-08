package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime"
	"strings"
)

var (
	storageAccount   string
	storageKey       string
	containerName    string
	blobPrefix       string
	metadataCriteria Criteria
	parallelism      int
	maxPageSize      int
	maxPages         int
	verbose          bool
	searchTagsOnly   bool
	nonInteractive   bool
)

type Criteria map[string]string

func (c *Criteria) Set(s string) error {
	m := make(map[string]string)
	entries := strings.Split(s, ",")

	for _, e := range entries {
		kv := strings.Split(e, ":")
		m[kv[0]] = kv[1]
	}

	c = (*Criteria)(&m)
	return nil
}

func (c *Criteria) String() string {
	return fmt.Sprintf("%+v", *c)
}

func fromEnv[T any](key string, backup T, conv func(string) (T, error)) T {
	if val, ok := os.LookupEnv(key); ok {
		result, err := conv(val)
		if err == nil {
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
	flag.StringVar(&storageAccount, "storageAccount", "", "name of the storage account to search")
	flag.StringVar(&storageKey, "storageKey", fromEnv("STORAGE_KEY", "", passthroughString), "access token for the storage account")
	flag.StringVar(&containerName, "container", "bulkuploads", "name of the blob container to search")
	flag.StringVar(&blobPrefix, "blobPrefix", "tus-prefix", "subfolder prefix for target files")
	flag.Var(&metadataCriteria, "metadataCriteria", "metadata fields to match on when searching blobs")
	flag.IntVar(&parallelism, "parallelism", runtime.NumCPU(), "number of parallel threads to use; represents number of search pages to process in parallel")
	flag.IntVar(&maxPageSize, "maxPageSize", 5000, "limit of number of search results for a page")
	flag.IntVar(&maxPages, "maxPages", math.MaxInt32, "limit of total number of pages to search; if zero, fetches and searches all pages")
	flag.BoolVar(&verbose, "v", false, "turn on debug logging")
	flag.BoolVar(&searchTagsOnly, "searchTagsOnly", false, "search using blob index tags instead of pages; parallelism is ignored in this mode as all search results are fetched in a single call")
	flag.BoolVar(&nonInteractive, "yes", false, "prompt user before deleting files")
	flag.Parse()

	programLevel := new(slog.LevelVar)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel, AddSource: true})
	slog.SetDefault(slog.New(h))
	if verbose {
		programLevel.Set(slog.LevelDebug)
	}
}
