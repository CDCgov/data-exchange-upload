package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/appconfig/azappconfig"
	"github.com/joho/godotenv"
)

var tguid string

func init() {
	flag.StringVar(&tguid, "id", "", "Upload ID")
	flag.Parse()

	if tguid == "" {
		log.Fatal("No tguid provided")
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func getFeatureFlag(flagName string) bool {
	connStr := os.Getenv("FEATURE_MANAGER_CONNECTION_STRING")
	client, err := azappconfig.NewClientFromConnectionString(connStr, nil)
	if err != nil {
		log.Fatalf("Failed to initialize Azure App Configuration: %v", err)
		return false
	}
    
	setting, err := client.GetSetting(context.Background(), flagName, nil)
	if err != nil {
		log.Printf("Error fetching feature flag %s: %v", flagName, err)
		return false
	}

	// Check if setting.Value is not nil and then dereference it to compare
	if setting.Value != nil && *setting.Value == "true" {
		return true
	}

	return false
}

func post_finish(uploadID string) {
	controller := NewProcStatController(os.Getenv("PS_API_URL"), 2*time.Second)
	traceID, spanID, err := controller.GetSpanByUploadID(uploadID, "dex-upload")
	if err != nil {
		controller.Logger.Printf("Failed to get span for upload %s: %v", uploadID, err)
		return
	}
	controller.Logger.Printf("Got span for upload %s with trace ID %s and span ID %s", uploadID, traceID, spanID)

	err = controller.StopSpanForTrace(traceID, spanID)
	if err != nil {
		controller.Logger.Printf("Failed to stop child span for parent span %s with stage name of dex-upload: %v", spanID, err)
		return
	}
	controller.Logger.Printf("Stopped child span for parent span %s with stage name of dex-upload", spanID)
}


func main() {	
	//processingStatusReportsEnabled := getFeatureFlag("PROCESSING_STATUS_REPORTS")
	processingStatusTracesEnabled := getFeatureFlag("PROCESSING_STATUS_TRACES")	

	if processingStatusTracesEnabled {
		fmt.Println("Processing for ID:", tguid)
		post_finish(tguid)
	}
}

