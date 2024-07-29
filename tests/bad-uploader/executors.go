package main

import (
	"fmt"
	"log/slog"
	"testing"
	"time"
)

type TestCaseIterator func() TestCase

type Executor interface {
	Run(chan<- TestCase, TestCaseIterator)
}

type DurationExecutor struct {
	Duration time.Duration
}

func (de DurationExecutor) Run(c chan<- TestCase, next TestCaseIterator) {
	tStart := time.Now()
	slog.Info("Running duration test", "duration", duration)
	i := 0
	for {
		if time.Since(tStart) > duration {
			break
		}
		c <- next()
		i++
	}
	slog.Info("uploads over time", "uploads", i, "duration", duration)
	slog.Info("roughly in 24 hours", "uploads", int((24*time.Hour)/duration)*i)
}

type SimpleLoadExecutor struct {
	Load int
}

func (sle SimpleLoadExecutor) Run(c chan<- TestCase, next TestCaseIterator) {
	slog.Info("Running load test", "uploads", load)
	for range sle.Load {
		c <- next()
	}
}

type BenchmarkExecutor struct{}

func (b BenchmarkExecutor) Run(c chan<- TestCase, next TestCaseIterator) {
	slog.Info("Running benchmark")
	test := func(b *testing.B) {
		slog.Debug("benchmarking", "runs", b.N)
		for i := 0; i < b.N; i++ {
			c <- next()
		}
	}
	result := testing.Benchmark(test)
	fmt.Printf("Benchmarking results: %f seconds/op\n", float64(result.NsPerOp())/float64(time.Second))
}
