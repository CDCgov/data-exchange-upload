package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
					},
					"UnDeliveredUploads": map[string]interface{}{
						"TotalCount": 2,
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
	// Start the mock server
	server := mockGraphQLServer()
	defer server.Close() // Ensure the server is closed after the test

	// Call the function with the mock server URL
	reportRow, err := fetchDataForDataStream(server.URL, "stream1", "route1", "2024-01-01", "2024-01-31")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "stream1", reportRow.DataStream)
	assert.Equal(t, "route1", reportRow.Route)
	assert.Equal(t, "2024-01-01", reportRow.StartDate)
	assert.Equal(t, "2024-01-31", reportRow.EndDate)
	assert.Equal(t, int64(10), reportRow.UploadCount)
	assert.Equal(t, int64(5), reportRow.DeliverySuccessCount)
	assert.Equal(t, int64(2), reportRow.DeliveryEndCount)
}

func TestCreateCSV(t *testing.T) {
	data := [][]string{
		{"Data Stream", "Route", "Start Date", "End Date", "Upload Count", "Delivery Success Count", "Delivery Fail Count"},
		{"TestStream", "TestRoute", "2024-01-01", "2024-01-02", "10", "5", "2"},
	}

	csvBytes, err := createCSV(data)
	assert.NoError(t, err)

	expected := "Data Stream,Route,Start Date,End Date,Upload Count,Delivery Success Count,Delivery Fail Count\nTestStream,TestRoute,2024-01-01,2024-01-02,10,5,2\n"
	assert.Equal(t, expected, string(csvBytes))
}

func TestSaveCsvToFile(t *testing.T) {
	testData := []byte("Test CSV Data")
	err := saveCsvToFile(testData)
	assert.NoError(t, err)

	// Read back the file to check contents
	data, err := os.ReadFile("upload-report.csv")
	assert.NoError(t, err)
	assert.Equal(t, "Test CSV Data", string(data))

	// Clean up the file
	os.Remove("upload-report.csv")
}

func TestGetNewS3Client_Success(t *testing.T) {
	client, err := getNewS3Client("us-east-1", "")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
