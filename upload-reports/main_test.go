package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type S3ClientMock struct {
	mock.Mock
}

func (m *S3ClientMock) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
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

// func TestUploadCsvToS3(t *testing.T) {
// 	// Setup the mock S3 client
// 	mockS3Client := new(S3ClientMock)
// 	// awsConfig := aws.Config{}
// 	// s3Client := s3.NewFromConfig(awsConfig)
//
// 	// Replace with your S3 upload logic
// 	t.Run("Successful Upload", func(t *testing.T) {
// 		mockS3Client.On("PutObject", mock.Anything, mock.Anything).Return(&s3.PutObjectOutput{}, nil)
//
// 		err := uploadCsvToS3("test-bucket", "http://mock-endpoint", "test-key", []byte("test data"))
// 		assert.NoError(t, err)
// 		mockS3Client.AssertExpectations(t)
// 	})
//
// 	t.Run("Failed Upload", func(t *testing.T) {
// 		mockS3Client.On("PutObject", mock.Anything, mock.Anything).Return(nil, errors.New("upload error"))
//
// 		err := uploadCsvToS3("test-bucket", "http://mock-endpoint", "test-key", []byte("test data"))
// 		assert.Error(t, err)
// 		mockS3Client.AssertExpectations(t)
// 	})
// }
//
// func TestFetchDataForDataStream(t *testing.T) {
// 	// Mock the GraphQL client and response as needed.
// 	// For this test, focus on the output structure.
// }
