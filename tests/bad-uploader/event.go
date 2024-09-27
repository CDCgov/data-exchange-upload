package main

import (
	"context"
	"encoding/json"
	"github.com/hasura/go-graphql-client"
	"sort"
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

type EventChecker interface {
	Events(ctx context.Context, uploadId string) ([]Report, error)
}

type PSAPIEventChecker struct {
	GraphQLClient *graphql.Client
}

func (paec *PSAPIEventChecker) Events(ctx context.Context, uploadId string) ([]Report, error) {
	var q struct {
		GetReports Reports `graphql:"getReports(uploadId: $id, reportsSortedBy: null, sortOrder: null)"`
	}
	variables := map[string]interface{}{
		"id": uploadId,
	}

	err := paec.GraphQLClient.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	reports := q.GetReports
	// Could use sets instead of sorting
	sort.Sort(reports)

	return reports, nil
}
