// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package psApi

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

// GetUploadStatsGetUploadStats includes the requested fields of the GraphQL type UploadStats.
// The GraphQL type's documentation follows.
//
// Collection of various uploads statistics
type GetUploadStatsGetUploadStats struct {
	// Number of uploads that have been completed.  This means, not only did the upload start, but according to the upload status reports we have received 100% of the expected chunks.
	CompletedUploadsCount int64 `json:"completedUploadsCount"`
	// Number of uploads where we have received at least one chunk of data, but not all of them.
	InProgressUploadsCount int64 `json:"inProgressUploadsCount"`
	// Provides a list of all the uploads that are pending. This means, the upload started, but according to the upload status reports we did not receive 100% of the expected chunks.
	PendingUploads GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts `json:"pendingUploads"`
	// Provides a list of all the uploads that have not been delivered. This means, the upload started, but according to the upload status reports we did not receive 100% of the expected chunks.
	UnDeliveredUploads GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts `json:"unDeliveredUploads"`
}

// GetCompletedUploadsCount returns GetUploadStatsGetUploadStats.CompletedUploadsCount, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStats) GetCompletedUploadsCount() int64 {
	return v.CompletedUploadsCount
}

// GetInProgressUploadsCount returns GetUploadStatsGetUploadStats.InProgressUploadsCount, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStats) GetInProgressUploadsCount() int64 {
	return v.InProgressUploadsCount
}

// GetPendingUploads returns GetUploadStatsGetUploadStats.PendingUploads, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStats) GetPendingUploads() GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts {
	return v.PendingUploads
}

// GetUnDeliveredUploads returns GetUploadStatsGetUploadStats.UnDeliveredUploads, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStats) GetUnDeliveredUploads() GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts {
	return v.UnDeliveredUploads
}

// GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts includes the requested fields of the GraphQL type PendingUploadCounts.
// The GraphQL type's documentation follows.
//
// Collection of undelivered uploads found
type GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts struct {
	// Total number of undelivered uploads.
	TotalCount int64 `json:"totalCount"`
}

// GetTotalCount returns GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts.TotalCount, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStatsPendingUploadsPendingUploadCounts) GetTotalCount() int64 {
	return v.TotalCount
}

// GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts includes the requested fields of the GraphQL type UnDeliveredUploadCounts.
// The GraphQL type's documentation follows.
//
// Collection of undelivered uploads found
type GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts struct {
	// Total number of undelivered uploads.
	TotalCount int64 `json:"totalCount"`
}

// GetTotalCount returns GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts.TotalCount, and is useful for accessing the field via an interface.
func (v *GetUploadStatsGetUploadStatsUnDeliveredUploadsUnDeliveredUploadCounts) GetTotalCount() int64 {
	return v.TotalCount
}

// GetUploadStatsResponse is returned by GetUploadStats on success.
type GetUploadStatsResponse struct {
	// Return various uploads statistics
	GetUploadStats GetUploadStatsGetUploadStats `json:"getUploadStats"`
}

// GetGetUploadStats returns GetUploadStatsResponse.GetUploadStats, and is useful for accessing the field via an interface.
func (v *GetUploadStatsResponse) GetGetUploadStats() GetUploadStatsGetUploadStats {
	return v.GetUploadStats
}

// __GetUploadStatsInput is used internally by genqlient
type __GetUploadStatsInput struct {
	Datastream string `json:"datastream"`
	Route      string `json:"route"`
	DateStart  string `json:"dateStart"`
	DateEnd    string `json:"dateEnd"`
}

// GetDatastream returns __GetUploadStatsInput.Datastream, and is useful for accessing the field via an interface.
func (v *__GetUploadStatsInput) GetDatastream() string { return v.Datastream }

// GetRoute returns __GetUploadStatsInput.Route, and is useful for accessing the field via an interface.
func (v *__GetUploadStatsInput) GetRoute() string { return v.Route }

// GetDateStart returns __GetUploadStatsInput.DateStart, and is useful for accessing the field via an interface.
func (v *__GetUploadStatsInput) GetDateStart() string { return v.DateStart }

// GetDateEnd returns __GetUploadStatsInput.DateEnd, and is useful for accessing the field via an interface.
func (v *__GetUploadStatsInput) GetDateEnd() string { return v.DateEnd }

// The query or mutation executed by GetUploadStats.
const GetUploadStats_Operation = `
query GetUploadStats ($datastream: String!, $route: String!, $dateStart: String, $dateEnd: String) {
	getUploadStats(dataStreamId: $datastream, dataStreamRoute: $route, dateStart: $dateStart, dateEnd: $dateEnd) {
		completedUploadsCount
		inProgressUploadsCount
		pendingUploads {
			totalCount
		}
		unDeliveredUploads {
			totalCount
		}
	}
}
`

func GetUploadStats(
	ctx_ context.Context,
	client_ graphql.Client,
	datastream string,
	route string,
	dateStart string,
	dateEnd string,
) (*GetUploadStatsResponse, error) {
	req_ := &graphql.Request{
		OpName: "GetUploadStats",
		Query:  GetUploadStats_Operation,
		Variables: &__GetUploadStatsInput{
			Datastream: datastream,
			Route:      route,
			DateStart:  dateStart,
			DateEnd:    dateEnd,
		},
	}
	var err_ error

	var data_ GetUploadStatsResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}