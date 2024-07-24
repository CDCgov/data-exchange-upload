package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"
)

/*
so we need to be able to create an arbirary number of test uploads
should have a context on them
let them run conncurrently
and get some meaningful output that we can go through after the fact maybe in a few channels
this will cover a use case for a single bad sender, so only one cred needed
*/

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var cwg sync.WaitGroup

	c := make(chan TestCase, parallelism)
	o := make(chan *Result)
	slog.Info("Starting threads", "parallelism", parallelism)
	for i := 0; i < parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(c, o, conf)
		}()
	}
	cwg.Add(1)
	go func() {
		defer cwg.Done()
		for r := range o {
			if r != nil {
				cwg.Add(1)
				go func(r *Result) {
					cctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
					defer cancel()
					defer cwg.Done()
					if err := Check(cctx, r.testCase, r.url, conf); err != nil {
						slog.Error("failed check", "error", err, "test case", r.testCase)
					}
				}(r)
			}
		}
	}()

	tStart := time.Now()
	if duration > 0 {
		slog.Info("Running duration test", "duration", duration)
		i := 0
		for {
			if time.Since(tStart) > duration {
				break
			}
			c <- cases.Next()
			i++
		}
		slog.Info("uploads over time", "uploads", i, "duration", duration)
		slog.Info("roughly in 24 hours", "uploads", int((24*time.Hour)/duration)*i)
	} else if load > 0 {
		slog.Info("Running load test", "uploads", load)
		for i := 0; i < load; i++ {
			c <- cases.Next()
		}
	} else {
		slog.Info("Running benchmark")
		result := testing.Benchmark(asParallelBenchmark(c, cases.Next))
		defer fmt.Printf("Benchmarking results: %f seconds/op\n", float64(result.NsPerOp())/float64(time.Second))
	}
	close(c)
	wg.Wait()
	close(o)
	cwg.Wait()
	fmt.Println("Total run took ", time.Since(tStart).Seconds(), " seconds")
}

type config struct {
	url         string
	tokenSource *SAMSTokenSource
}

func worker(c <-chan TestCase, o chan<- *Result, conf *config) {
	for e := range c {
		res, err := runTest(e, conf)
		if err != nil {
			slog.Error("ERROR: ", "error", err, "case", e)
		}
		o <- res
	}
}

func asParallelBenchmark(c chan TestCase, next func() TestCase) func(*testing.B) {
	return func(b *testing.B) {
		slog.Info("benchmarking", "runs", b.N)
		for i := 0; i < b.N; i++ {
			c <- next()
		}
	}
}
