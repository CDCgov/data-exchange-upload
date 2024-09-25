package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tStart := time.Now()
	c := InitiateTests(getExecutor())
	o := StartWorkers(c)
	if err := ValidateResults(ctx, o); err != nil {
		fmt.Println("validation failed:", err)
	}
	fmt.Println("Total run took ", time.Since(tStart).Seconds(), " seconds")
	/*
		Report output:
		duration
		files uploaded successfully
		files uploaded unsuccessfully
		files delivered successfully
		files delivered unsuccessfully (print all ids and error messages if debug enabled)
	*/
}

func StartWorkers(c <-chan TestCase) <-chan *Result {
	o := make(chan *Result)
	go func() {
		defer close(o)
		var wg sync.WaitGroup
		slog.Info("Starting threads", "parallelism", parallelism)
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(c, o, conf)
			}()
		}
		wg.Wait()
	}()
	return o
}

func ValidateResults(ctx context.Context, o <-chan *Result) error {
	var wg sync.WaitGroup

	var errs error
	for r := range o {
		if r != nil {
			wg.Add(1)
			go func(r *Result) {
				defer wg.Done()
				limit := time.Duration(r.testCase.TimeLimit)
				if limit == 0*time.Second {
					limit = 1 * time.Minute
				}
				cctx, cancel := context.WithTimeout(ctx, limit)
				defer cancel()
				check := NewCheck(ctx, conf, r.testCase, r.url)
				if err := Check(cctx, check); err != nil {
					slog.Error("failed check", "error", err, "test case", r.testCase)
					errs = errors.Join(errs, err)
				}
			}(r)
		}
	}
	wg.Wait()
	return errs
}

func InitiateTests(e Executor) <-chan TestCase {
	c := make(chan TestCase)
	go func() {
		defer close(c)
		e.Run(c, cases.Next)
	}()
	return c
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
