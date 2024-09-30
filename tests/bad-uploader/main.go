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
				limit = 1 * time.Minute
			}
			checkTimeout, cancel := context.WithTimeout(ctx, limit)
			defer cancel()

			wg.Add(len(PostUploadChecks))
			for _, check := range PostUploadChecks {
				check := check
				go func(r *Result) {
					defer wg.Done()
					// return a specific error and/or check result.  Specific error can have check specific info like upload id and reports
					err := WithRetry(checkTimeout, r.testCase, uid, check.DoCase)
					if err != nil {
						slog.Error("failed post upload check", "error", err, "test case", r.testCase)
						errs = errors.Join(errs, err)
					} else {
						check.OnSuccess()
					}
				}(r)
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

	if validationErrors != nil {
		fmt.Println("Validation Failures!")
		for {
			err := errors.Unwrap(validationErrors)
			if err == nil {
				break
			}

			slog.Error("", "", err)

			// TODO get upload ID out of err
			//filename := "output/" + err.() + "_check_failures"
			//f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
			//if err != nil {
			//	os.Exit(1)
			//}
			//defer f.Close()
			//je := json.NewEncoder(f)
			//if err := je.Encode(err); err != nil {
			//	os.Exit(1)
			//}
		}
	}

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
		}
		atomic.AddInt32(&testResult.SuccessfulUploads, 1)
		o <- res
	}
}
