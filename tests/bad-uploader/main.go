package main

import (
	"crypto/rand"
	"flag"
	"io"
	"log"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/eventials/go-tus"
)

var (
	url         string
	size        int
	parallelism int
	load        int
	chunk       int64
)

func init() {
	flag.IntVar(&size, "size", 250*10000, "the size of the file to upload, in bytes")
	flag.StringVar(&url, "url", "http://localhost:8080/files/", "the upload url for the tus server")
	flag.IntVar(&parallelism, "parallelism", runtime.NumCPU(), "the number of parallel threads to use, defaults to MAXGOPROC when set to < 1.")
	flag.IntVar(&load, "load", 0, "set the number of files to load, defaults to 0 and adjusts based on benchmark logic")
	flag.Int64Var(&chunk, "chunk", 2, "set the chunk size to use when uploading files in MB")
	flag.Parse()
	chunk = chunk * 1024 * 1024
}

func buildConfig() (*config, error) {
	tconf := tus.DefaultConfig()
	tconf.ChunkSize = chunk
	return &config{
		url: url,
		tus: tconf,
	}, nil
}

func resultOrFatal[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
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
	log.Printf("Starting %d threads\n", parallelism)
	for i := 0; i < parallelism; i++ {
		go worker(c, conf)
	}

	tStart := time.Now()
	if load > 0 {
		log.Printf("Running load test of %d files\n", load)
		for i := 0; i < load; i++ {
			c <- struct{}{}
		}
	} else {
		log.Println("Running benchmark")
		result := testing.Benchmark(asPallelBenchmark(c))
		log.Println(result.String())
	}
	wg.Wait()
	log.Println("Benchmarking took ", time.Since(tStart).Seconds(), " seconds")
}

type config struct {
	tus *tus.Config
	url string
}

func worker(c <-chan struct{}, conf *config) {
	for range c {
		wg.Add(1)
		f := &BadHL7{
			Size:           size,
			DummyGenerator: &RandomBytesReader{},
		}
		if err := runTest(f, conf); err != nil {
			log.Println("ERROR: ", err)
		}
		wg.Done()
	}
}

func asPallelBenchmark(c chan struct{}) func(*testing.B) {
	return func(b *testing.B) {
		log.Println("N is ", b.N)
		//b.SetParallelism(parallelism)
		//log.Println("with parallelism ", parallelism, " or default ", runtime.NumCPU())
		for i := 0; i < b.N; i++ {
			c <- struct{}{}
		}
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
