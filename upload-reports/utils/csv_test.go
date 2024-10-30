package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCSV(t *testing.T) {
	data := [][]string{
		{"Data Stream", "Route", "Start Date", "End Date", "Upload Count", "Delivery Success Count", "Delivery Fail Count"},
		{"TestStream", "TestRoute", "2024-01-01", "2024-01-02", "10", "5", "2"},
	}

	csvBytes, err := CreateCSV(data)
	assert.NoError(t, err)

	expected := "Data Stream,Route,Start Date,End Date,Upload Count,Delivery Success Count,Delivery Fail Count\nTestStream,TestRoute,2024-01-01,2024-01-02,10,5,2\n"
	assert.Equal(t, expected, csvBytes.String())
}

func TestSaveCsvToFile(t *testing.T) {
	data := [][]string{
		{"Test CSV Data"},
	}

	csvBytes, err := CreateCSV(data)
	assert.NoError(t, err)

	tmpDir := t.TempDir()
	filename := "test-upload-report.csv"

	err = SaveCsvToFile(csvBytes, tmpDir, filename)
	assert.NoError(t, err)

	fullPath := filepath.Join(tmpDir, filename)
	readData, err := os.ReadFile(fullPath)
	assert.NoError(t, err)
	assert.Equal(t, "Test CSV Data\n", string(readData))
}
