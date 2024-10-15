package utils

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type AppConfig struct {
	DataStreams   string
	StartDate     string
	EndDate       string
	TargetEnv     string
	CsvOutputPath string
	PsApiUrl      string
	S3Config      *S3StorageConfig
}

type S3StorageConfig struct {
	BucketName string
	Endpoint   string
}

func GetEnvVar(key string) string {
	val := os.Getenv(key)
	return val
}

func GetConfig() AppConfig {
	dataStreams := flag.String("dataStreams", "", "Comma-separated list of data streams and routes in the format data-stream-name_route-name")
	startDate := flag.String("startDate", "", "Start date in UTC (YYYY-MM-DDTHH:MM:SSZ)")
	endDate := flag.String("endDate", "", "End date in UTC (YYYY-MM-DDTHH:MM:SSZ)")
	targetEnv := flag.String("targetEnv", "dev", "Target environment (default: dev)")

	defaultCsvOutputPath, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current working directory: %v\n", err)
		defaultCsvOutputPath = "."
	}

	csvOutputPath := flag.String("csvOutputPath", defaultCsvOutputPath, "Path to save the CSV file (default: current working directory)")

	flag.Parse()

	if *startDate == "" || *endDate == "" {
		now := time.Now().UTC()
		*endDate = now.Format(time.RFC3339)
		*startDate = now.Add(-24 * time.Hour).Format(time.RFC3339)
	}

	psApiUrl := GetEnvVar("PS_API_ENDPOINT")
	s3BucketName := GetEnvVar("S3_BUCKET_NAME")
	s3Endpoint := GetEnvVar("S3_ENDPOINT")

	config := AppConfig{
		DataStreams:   *dataStreams,
		StartDate:     *startDate,
		EndDate:       *endDate,
		TargetEnv:     *targetEnv,
		CsvOutputPath: *csvOutputPath,
		PsApiUrl:      psApiUrl,
	}

	if s3BucketName != "" && s3Endpoint != "" {
		s3Config := S3StorageConfig{
			BucketName: s3BucketName,
			Endpoint:   s3Endpoint,
		}

		config.S3Config = &s3Config
	}

	return config
}
