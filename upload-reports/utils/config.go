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

	config := ReportConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
		PsApiUrl:    psApiUrl,
	}

	return config
}
