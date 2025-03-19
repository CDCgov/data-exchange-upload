package utils

import (
	"flag"
	"os"
	"testing"
)

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{"TEST_VAR", "test_value", "test_value"},
		{"NON_EXISTENT_VAR", "", ""},
	}

	for _, test := range tests {
		if test.value != "" {
			os.Setenv(test.key, test.value)
		}

		result := GetEnvVar(test.key)
		if result != test.expected {
			t.Errorf("GetEnvVar(%s) = %s; want %s", test.key, result, test.expected)
		}

		os.Unsetenv(test.key)
	}
}

func TestGetConfig(t *testing.T) {
	os.Args = []string{
		"cmd",
		"-dataStreams=celr_csv",
		"-startDate=2024-10-10T00:00:00Z",
		"-endDate=2024-10-11T00:00:00Z",
		"-targetEnv=dev",
		"-csvOutputPath=/tmp/output.csv",
	}

	os.Setenv("PS_API_ENDPOINT", "https://example.com/api")
	os.Setenv("S3_BUCKET_NAME", "my-bucket")
	os.Setenv("S3_ENDPOINT", "https://s3.us-west-2.amazonaws.com")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	config := GetConfig()

	if config.DataStreams != "celr_csv" {
		t.Errorf("Expected DataStreams 'celr_csv', got '%s'", config.DataStreams)
	}
	if config.StartDate != "2024-10-10T00:00:00Z" {
		t.Errorf("Expected StartDate '2024-10-10T00:00:00Z', got '%s'", config.StartDate)
	}
	if config.EndDate != "2024-10-11T00:00:00Z" {
		t.Errorf("Expected EndDate '2024-10-11T00:00:00Z', got '%s'", config.EndDate)
	}
	if config.TargetEnv != "dev" {
		t.Errorf("Expected TargetEnv 'dev', got '%s'", config.TargetEnv)
	}
	if config.CsvOutputPath != "/tmp/output.csv" {
		t.Errorf("Expected CsvOutputPath '/tmp/output.csv', got '%s'", config.CsvOutputPath)
	}
	if config.PsApiUrl != "https://example.com/api" {
		t.Errorf("Expected PsApiUrl 'https://example.com/api', got '%s'", config.PsApiUrl)
	}
	if config.S3Config == nil {
		t.Errorf("Expected S3Config to be set, but got nil")
	} else {
		if config.S3Config.BucketName != "my-bucket" {
			t.Errorf("Expected S3Config.BucketName 'my-bucket', got '%s'", config.S3Config.BucketName)
		}
		if config.S3Config.Endpoint != "https://s3.us-west-2.amazonaws.com" {
			t.Errorf("Expected S3Config.Endpoint 'https://s3.us-west-2.amazonaws.com', got '%s'", config.S3Config.Endpoint)
		}
	}

	os.Unsetenv("PS_API_ENDPOINT")
	os.Unsetenv("S3_BUCKET_NAME")
	os.Unsetenv("S3_ENDPOINT")
}
