package utils

import (
	"log"
	"os"
	"strings"
)

type ReportConfig struct {
	DataStreams []string
	StartDate   string
	EndDate     string
	TargetEnv   string
	PsApiUrl    string
	S3Bucket    string
}

func GetEnvVar(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return val
}

func GetConfig() ReportConfig {
	dataStreams := strings.Split(GetEnvVar("DATASTREAMS"), ",")
	startDate := GetEnvVar("START_DATE")
	endDate := GetEnvVar("END_DATE")
	targetEnv := GetEnvVar("ENV")
	psApiUrl := GetEnvVar("PS_API_ENDPOINT")
	s3Bucket := GetEnvVar("S3_BUCKET")

	config := ReportConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
		PsApiUrl:    psApiUrl,
		S3Bucket:    s3Bucket,
	}

	return config
}
