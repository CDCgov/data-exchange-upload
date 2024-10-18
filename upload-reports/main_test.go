package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockGraphQLServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"getUploadStats": map[string]interface{}{
					"CompletedUploadsCount": 10,
					"PendingUploads": map[string]interface{}{
						"TotalCount": 5,
						"PendingUploads": []map[string]interface{}{
							{"UploadId": "pending1", "Filename": "file1.csv"},
							{"UploadId": "pending2", "Filename": "file2.csv"},
						},
					},
					"UndeliveredUploads": map[string]interface{}{
						"TotalCount": 2,
						"UndeliveredUploads": []map[string]interface{}{
							{"UploadId": "undelivered1", "Filename": "file3.csv"},
							{"UploadId": "undelivered2", "Filename": "file4.csv"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(handler)
}

func TestFetchDataForDataStream(t *testing.T) {
	server := mockGraphQLServer()
	defer server.Close()

	reportRow, anomalousData, err := fetchDataForDataStream(server.URL, "stream1", "route1", "2024-01-01", "2024-01-31")

	// Assertions for SummaryRow
	assert.NoError(t, err)
	assert.Equal(t, "stream1", reportRow.DataStream)
	assert.Equal(t, "route1", reportRow.Route)
	assert.Equal(t, "2024-01-01", reportRow.StartDate)
	assert.Equal(t, "2024-01-31", reportRow.EndDate)
	assert.Equal(t, int64(10), reportRow.TotalUploadCount)
	assert.Equal(t, int64(5), reportRow.PendingUploadCount)
	assert.Equal(t, int64(2), reportRow.UndeliveredUploadCount)

	// Assertions for AnomalousItemRow
	assert.Len(t, anomalousData, 4)

	// Check Pending Uploads
	assert.Equal(t, "pending1", anomalousData[0].UploadId)
	assert.Equal(t, "file1.csv", anomalousData[0].Filename)
	assert.Equal(t, "stream1", anomalousData[0].DataStream)
	assert.Equal(t, "route1", anomalousData[0].Route)
	assert.Equal(t, Pending, anomalousData[0].Category)

	assert.Equal(t, "pending2", anomalousData[1].UploadId)
	assert.Equal(t, "file2.csv", anomalousData[1].Filename)
	assert.Equal(t, "stream1", anomalousData[1].DataStream)
	assert.Equal(t, "route1", anomalousData[1].Route)
	assert.Equal(t, Pending, anomalousData[1].Category)

	// Check Undelivered Uploads
	assert.Equal(t, "undelivered1", anomalousData[2].UploadId)
	assert.Equal(t, "file3.csv", anomalousData[2].Filename)
	assert.Equal(t, "stream1", anomalousData[2].DataStream)
	assert.Equal(t, "route1", anomalousData[2].Route)
	assert.Equal(t, Undelivered, anomalousData[2].Category)

	assert.Equal(t, "undelivered2", anomalousData[3].UploadId)
	assert.Equal(t, "file4.csv", anomalousData[3].Filename)
	assert.Equal(t, "stream1", anomalousData[3].DataStream)
	assert.Equal(t, "route1", anomalousData[3].Route)
	assert.Equal(t, Undelivered, anomalousData[3].Category)
}
