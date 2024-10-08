package utils

import (
	"log"
	"os"
)

type AppConfig struct {
	DataStreams string           `env:"DATASTREAMS"`
	StartDate   string           `env:"START_DATE"`
	EndDate     string           `env:"END_DATE"`
	TargetEnv   string           `env:"TARGET_ENV"`
	PsApiUrl    string           `env:"PS_API_ENDPOINT"`
	S3Config    *S3StorageConfig `env:", prefix=S3_, noinit"`
}

type S3StorageConfig struct {
	Endpoint   string `env:"ENDPOINT"`
	BucketName string `env:"BUCKET_NAME"`
}

func GetEnvVar(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return val
}

func GetConfig() AppConfig {
	dataStreams := (GetEnvVar("DATASTREAMS"))
	startDate := GetEnvVar("START_DATE")
	endDate := GetEnvVar("END_DATE")
	targetEnv := GetEnvVar("ENV")
	psApiUrl := GetEnvVar("PS_API_ENDPOINT")

	config := AppConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
		PsApiUrl:    psApiUrl,
	}

	return config
}
