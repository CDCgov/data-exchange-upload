package utils

import (
	"os"
)

type AppConfig struct {
	DataStreams string
	StartDate   string
	EndDate     string
	TargetEnv   string
	PsApiUrl    string
	S3Config    *S3StorageConfig
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
	dataStreams := (GetEnvVar("DATASTREAMS"))
	startDate := GetEnvVar("START_DATE")
	endDate := GetEnvVar("END_DATE")
	targetEnv := GetEnvVar("TARGET_ENV")
	psApiUrl := GetEnvVar("PS_API_ENDPOINT")
	s3BucketName := GetEnvVar("S3_BUCKET_NAME")
	s3Endpoint := GetEnvVar("S3_ENDPOINT")

	config := AppConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
		PsApiUrl:    psApiUrl,
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
