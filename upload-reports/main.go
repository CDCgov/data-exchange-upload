package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	// "net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	// "github.com/shurcooL/graphql"
	// "golang.org/x/oauth2"
)

type ReportConfig struct {
	DataStreams []string
	StartDate   string
	EndDate     string
	TargetEnv   string
	PsApiUrl    string
}

type ReportDataRow struct {
	DataStream           string
	Route                string
	StartDate            string
	EndDate              string
	UploadCount          string
	DeliverySuccessCount string
	DeliveryEndCount     string
}

func main() {
	// Load environment variables and config
	config := getConfig()
	fmt.Printf("Target Env: %s, DataStreams: %v, Start Date: %s, End Date: %s\n", config.TargetEnv, config.DataStreams, config.StartDate, config.EndDate)

	csvData := getCsvData(config)

	fmt.Printf("Fetched CSV Data: %v\n", csvData)

	// // Create CSV
	// csvBytes, err := createCSV(csvData)
	// if err != nil {
	// 	log.Fatalf("Error creating CSV: %v", err)
	// }
	//
	// // Upload CSV to S3
	// bucketName := "upload-file-count-reports"
	// key := fmt.Sprintf("file-counts-report-%s.csv", targetEnv)
	// if err := uploadCsvToS3(bucketName, key, csvBytes); err != nil {
	// 	log.Fatalf("Error uploading CSV to S3: %v", err)
	// }
}

func getEnvVar(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s environment variable not set", key)
	}
	return val
}

func getConfig() ReportConfig {
	dataStreams := strings.Split(getEnvVar("DATASTREAMS"), ",")
	startDate := getEnvVar("START_DATE")
	endDate := getEnvVar("END_DATE")
	targetEnv := getEnvVar("ENV")
	psApiUrl := getEnvVar("PS_API_ENDPOINT")

	config := ReportConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
		PsApiUrl:    psApiUrl,
	}

	return config
}

func fetchDataForDataStream(apiURL string, datastream string, route string, startDate string, endDate string) (ReportDataRow, error) {
	fmt.Printf("PS-API graphql endpoint: %s\n", apiURL)

	// src := oauth2.StaticTokenSource(
	// 	&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
	// )
	// httpClient := oauth2.NewClient(context.Background(), src)
	//
	// client := graphql.NewClient("https://example.com/graphql", httpClient)
	//
	// req := graphql.NewRequest(`
	// 	query {
	// 		# Your GraphQL query here
	// 		uploads {
	// 			id
	// 			filename
	// 			timestamp
	// 		}
	// 	}
	// `)
	//
	// // Add headers if required
	// req.Header.Set("Authorization", "Bearer "+getEnvVar("API_TOKEN"))

	var uploadResponse struct {
		Datestream string
		Route      string
		Count      string
	}

	var deliveryResponse struct {
		Datestream   string
		Route        string
		SuccessCount string
		FailCount    string
	}

	// Perform the request
	// if err := client.Run(context.Background(), req, &response); err != nil {
	// 	return nil, err
	// }

	reportRow := ReportDataRow{
		DataStream:           datastream,
		Route:                route,
		StartDate:            startDate,
		EndDate:              endDate,
		UploadCount:          uploadResponse.Count,
		DeliverySuccessCount: deliveryResponse.SuccessCount,
		DeliveryEndCount:     deliveryResponse.FailCount,
	}

	return reportRow, nil
}

func getCsvData(config ReportConfig) [][]string {

	// Prepare data for CSV
	var csvData [][]string
	csvData = append(csvData, []string{"Data Stream", "Route", "Start Date", "End Date", "Upload Count", "Delivery Success Count", "Delivery Fail Count"})

	for _, ds := range config.DataStreams {
		streamAndRoute := strings.Split(ds, "-")
		if len(streamAndRoute) != 2 {
			log.Fatalf("Data stream passed in does not have correct formatting: %s", ds)
		}

		datastream := streamAndRoute[0]
		route := streamAndRoute[1]

		rowData, err := fetchDataForDataStream(config.PsApiUrl, datastream, route, config.StartDate, config.EndDate)
		if err != nil {
			log.Fatalf("Error fetching data from GraphQL API: %v", err)
		}

		csvData = append(csvData, []string{
			rowData.DataStream,
			rowData.Route,
			rowData.StartDate,
			rowData.EndDate,
			rowData.UploadCount,
			rowData.DeliverySuccessCount,
			rowData.DeliveryEndCount,
		})
	}

	return csvData

}

func createCSV(data [][]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func uploadCsvToS3(bucketName, key string, csvData []byte) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	putInput := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(csvData),
	}

	_, err = s3Client.PutObject(context.TODO(), putInput)
	if err != nil {
		return fmt.Errorf("failed to upload CSV to S3: %v", err)
	}

	log.Printf("Successfully uploaded %s to S3 bucket %s\n", key, bucketName)
	return nil
}
