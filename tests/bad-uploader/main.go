package main

import (
	"crypto/rand"
	"flag"
	"io"
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/eventials/go-tus"
)

var (
	url         string
	size        int
	parallelism int
	load        int
)

func init() {
	flag.IntVar(&size, "size", 250*10000, "the size of the file to upload, in bytes")
	flag.StringVar(&url, "url", "http://localhost:8080/files/", "the upload url for the tus server")
	flag.IntVar(&parallelism, "parallelism", 0, "the number of parallel threads to use, defaults to MAXGOPROC when set to < 1.")
	flag.IntVar(&load, "load", 0, "set the number of files to load, defaults to 0 and adjusts based on benchmark logic")
	flag.Parse()
}

func buildConfig() (*config, error) {
	return &config{
		url: url,
		tus: tus.DefaultConfig(),
	}, nil
}

func resultOrFatal[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

/*
so we need to be able to create an arbirary number of test uploads
should have a context on them
let them run conncurrently
and get some meaningful output that we can go through after the fact maybe in a few channels
this will cover a use case for a single bad sender, so only one cred needed
*/

func main() {

	conf := resultOrFatal(buildConfig())

	tStart := time.Now()
	result := testing.Benchmark(asPallelBenchmark(conf))
	log.Println(result.String())
	log.Println("Benchmarking took ", time.Since(tStart).Seconds(), " seconds")
}

type config struct {
	tus *tus.Config
	url string
}

func asPallelBenchmark(conf *config) func(*testing.B) {
	return func(b *testing.B) {
		b.SetParallelism(parallelism)
		log.Println("with parallelism ", parallelism, " or default ", runtime.NumCPU())
		if load > 0 {
			b.N = load
		}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				f := &BadHL7{
					Size:           size,
					DummyGenerator: &RandomBytesReader{},
				}
				if err := runTest(f, conf); err != nil {
					log.Println("ERROR: ", err)
				}
			}
		})
	}
}

func runTest(f *BadHL7, conf *config) error {
	//not great
	defer f.Close()
	// create the tus client.
	client, err := tus.NewClient(conf.url, conf.tus)
	if err != nil {
		return err
	}

	// create an upload from a file.
	upload := tus.NewUpload(f, int64(f.Size), f.Metadata(), f.Fingerprint())

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return err
	}

	// start the uploading process.
	return uploader.Upload()
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
