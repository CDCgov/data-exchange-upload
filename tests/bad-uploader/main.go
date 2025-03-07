package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

type LoadTestResult struct {
	SuccessfulUploads    int32
	SuccessfulDeliveries int32
	SuccessfulEventSets  int32
	TotalDuration        time.Duration
}

var testResult LoadTestResult

func main() {
	testResult = LoadTestResult{}

	err := os.Mkdir("output", 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		slog.Error("error making output dir", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	InitChecks(ctx, conf)

	tStart := time.Now()
	c := InitiateTests(getExecutor())
	o := StartWorkers(c)
	err = ValidateResults(ctx, o)
	testResult.TotalDuration = time.Since(tStart)
	PrintFinalReport(err)
	exitError(err)
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
			uid := path.Base(r.url)
			limit := time.Duration(r.testCase.TimeLimit)
			if limit == 0*time.Second {
				limit = 11 * time.Second
			}
			checkTimeout, cancel := context.WithTimeout(ctx, limit)
			defer cancel()

			wg.Add(len(PostUploadChecks))
			for _, check := range PostUploadChecks {
				go func(r *Result, check Checker) {
					defer wg.Done()
					// return a specific error and/or check result.  Specific error can have check specific info like upload id and reports
					err := WithRetry(checkTimeout, r.testCase, uid, check.DoCase)
					if err != nil {
						slog.Error("failed post upload check", "error", err, "test case", r.testCase, "upload ID", uid)
						logErrors(err, uid)
						errs = errors.Join(errs, err)
					} else {
						slog.Info("Pass", "upload ID", uid)
						check.OnSuccess()
					}
				}(r, check)
			}
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

func PrintFinalReport(validationErrors error) {
	fmt.Println("**********************************")
	printValidationErrors(validationErrors)
	fmt.Printf(`
RESULTS:
Files uploaded: %d/%d
Files delivered: %d/%d
`,
		testResult.SuccessfulUploads,
		load,
		testResult.SuccessfulDeliveries,
		testResult.SuccessfulUploads)

	if reportsURL != "" {
		fmt.Printf("Successful event sets generated: %d/%d\r\n", testResult.SuccessfulEventSets, testResult.SuccessfulUploads)
	} else {
		fmt.Println("Skipped event generation check")
	}

	fmt.Printf("Duration: %f seconds\r\n", testResult.TotalDuration.Seconds())
	fmt.Println("**********************************")
}

func worker(c <-chan TestCase, o chan<- *Result, conf *config) {
	for e := range c {
		res, err := runTest(e, conf)
		if err != nil {
			slog.Error("ERROR: ", "error", err, "case", e)
		} else {
			atomic.AddInt32(&testResult.SuccessfulUploads, 1)
		}
		o <- res
	}
}

func printValidationErrors(errs error) {
	if errs != nil {
		u, ok := errs.(interface {
			Unwrap() []error
		})
		if !ok {
			fmt.Printf("validation error %s\n", errs)
			return
		}

		for _, err := range u.Unwrap() {
			printValidationErrors(err)
		}
	}
}

func logErrors(err error, uploadId string) {
	if err == nil {
		return
	}
	filename := "output/" + uploadId + "_failures"
	f, e := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if _, e = f.WriteString(err.Error()); e != nil {
		panic(e)
	}
}

func exitError(validationErrors error) {
	unsuccessfulUploads := load > 0 && testResult.SuccessfulUploads < int32(load)
	unsuccessfulDeliveries := testResult.SuccessfulDeliveries < testResult.SuccessfulUploads
	if validationErrors != nil || unsuccessfulUploads || unsuccessfulDeliveries {
		fmt.Println("FAILED")
		os.Exit(1)
	}
}
