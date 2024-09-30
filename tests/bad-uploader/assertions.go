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
}

func (e *ErrAssertion) Error() string {
	return fmt.Sprintf("expected %v; read %v", e.Expected, e.Actual)
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
	return fmt.Sprintf("failed to pass assertion after %v seconds", e.Limit*Duration(time.Second))
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
	OnFail() error
}

type CheckFunc func(ctx context.Context, c TestCase, uploadId string) error

func WithRetry(timeout context.Context, c TestCase, uploadId string, checker CheckFunc) error {
	timer := time.NewTicker(1 * time.Second)
	var err error
	for {
		select {
		case <-timer.C:
			// Perform the checkable action
			err = checker(timeout, c, uploadId)
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
				continue
			}
			// Unexpected error
			return err
		case <-timeout.Done():
			if err != nil {
				err = errors.Join(err, &ErrAssertionTimeout{
					Limit: c.TimeLimit,
				})
			}
			return err
		}
	}
}
