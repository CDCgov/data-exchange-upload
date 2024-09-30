package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hasura/go-graphql-client"
	"sort"
	"sync/atomic"
	"time"
)

type Report struct {
	ID        string
	UploadId  string
	StageInfo StageInfo
	Timestamp time.Time
	Content   json.RawMessage
}

type StageInfo struct {
	Action string
}

type Reports []Report

func (r Reports) Len() int           { return len(r) }
func (r Reports) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Reports) Less(i, j int) bool { return r[i].StageInfo.Action < r[j].StageInfo.Action } // Sorting alphabetically

type EventChecker struct {
	GraphQLClient *graphql.Client
}

func (ec *EventChecker) DoCase(ctx context.Context, c TestCase, uploadId string) error {
	reports, err := ec.Events(ctx, uploadId)
	if err != nil {
		return fmt.Errorf("failed to fetch events for upload %s", uploadId)
	}

	err = compareEvents(reports, c.ExpectedReports)
	if err != nil {
		return err
	}

	return nil
}

func (ec *EventChecker) OnSuccess() {
	atomic.AddInt32(&testResult.SuccessfulEventSets, 1)
}

func (ec *EventChecker) OnFail() error {
	return nil
}

func (ec *EventChecker) Events(ctx context.Context, uploadId string) ([]Report, error) {
	var q struct {
		GetReports Reports `graphql:"getReports(uploadId: $id, reportsSortedBy: null, sortOrder: null)"`
	}
	variables := map[string]interface{}{
		"id": uploadId,
	}

	err := ec.GraphQLClient.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	reports := q.GetReports
	// Could use sets instead of sorting
	sort.Sort(reports)

	return reports, nil
}

func compareEvents(actual []Report, expected []Report) error {
	if len(actual) != len(expected) {
		return &ErrAssertion{
			Expected: len(expected),
			Actual:   len(actual),
		}
	}

	for i, a := range actual {
		if a.StageInfo.Action != expected[i].StageInfo.Action {
			return errors.Join(&ErrAssertion{
				Expected: expected[i].StageInfo.Action,
				Actual:   a.StageInfo.Action,
			}, &ErrFatalAssertion{"unexpected event"})
		}
	}

	return nil
}
