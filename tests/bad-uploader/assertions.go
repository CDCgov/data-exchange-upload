package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

type ErrAssertion struct {
	Expected any
	Actual   any
	msg      string
}

func (e *ErrAssertion) Error() string {
	return fmt.Sprintf("%s; expected %v; read %v", e.msg, e.Expected, e.Actual)
}

type ErrFatalAssertion struct {
	msg string
}

func (e *ErrFatalAssertion) Error() string {
	return e.msg
}

type ErrAssertionTimeout struct {
	Limit Duration
}

func (e *ErrAssertionTimeout) Error() string {
	return fmt.Sprintf("failed to pass assertion after %.2f seconds", (time.Duration(e.Limit)).Seconds())
}

var PostUploadChecks []Checker

func InitChecks(ctx context.Context, conf *config) {
	PostUploadChecks = []Checker{}

	httpClient := &http.Client{}
	if conf.tokenSource != nil {
		httpClient = oauth2.NewClient(ctx, conf.tokenSource)
	}

	PostUploadChecks = append(PostUploadChecks, &InfoChecker{
		Client: httpClient,
	})

	if reportsURL != "" {
		PostUploadChecks = append(PostUploadChecks, &EventChecker{
			GraphQLClient: graphql.NewClient(reportsURL, nil),
		})
	}
}

type Checker interface {
	DoCase(ctx context.Context, c TestCase, uploadId string) error
	OnSuccess()
}

type CheckFunc func(ctx context.Context, c TestCase, uploadId string) error

func WithRetry(timeout context.Context, c TestCase, uploadId string, checker CheckFunc) error {
	timer := time.NewTicker(1 * time.Second)
	var assertionError error
	for {
		select {
		case <-timer.C:
			// Perform the checkable action
			err := checker(timeout, c, uploadId)
			if err == nil {
				// Action was successful, all done
				return nil
			}
			// Did we fail assertion early?
			var fatalErr *ErrFatalAssertion
			if errors.As(err, &fatalErr) {
				return err
			}
			// Maybe failed but need to check again
			var assertionErr *ErrAssertion
			if errors.As(err, &assertionErr) {
				slog.Debug("retrying check; got unexpected value:", "error", err)
				assertionError = err
				continue
			}
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			// Unexpected error
			return err
		case <-timeout.Done():
			return errors.Join(assertionError, &ErrAssertionTimeout{
				Limit: c.TimeLimit,
			})
		}
	}
}
