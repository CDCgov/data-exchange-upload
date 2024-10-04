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
}

func main() {
	// Load environment variables and config
	config := getConfig()
	fmt.Printf("Target Env: %s, DataStreams: %v, Start Date: %v, End Date: %v\n", config.TargetEnv, config.DataStreams, config.StartDate, config.EndDate)

	// Fetch data from GraphQL API
	// apiURL := getEnvVar("GRAPHQL_API_URL")
	// csvData, err := fetchDataFromGraphQL(apiURL)
	// if err != nil {
	// 	log.Fatalf("Error fetching data from GraphQL API: %v", err)
	// }
	//
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

	config := ReportConfig{
		DataStreams: dataStreams,
		StartDate:   startDate,
		EndDate:     endDate,
		TargetEnv:   targetEnv,
	}

	return config
}

// func fetchDataFromGraphQL(apiURL string) ([][]string, error) {
// 	src := oauth2.StaticTokenSource(
// 		&oauth2.Token{AccessToken: os.Getenv("GRAPHQL_TOKEN")},
// 	)
// 	httpClient := oauth2.NewClient(context.Background(), src)
//
// 	client := graphql.NewClient("https://example.com/graphql", httpClient)
//
// 	req := graphql.NewRequest(`
// 		query {
// 			# Your GraphQL query here
// 			uploads {
// 				id
// 				filename
// 				timestamp
// 			}
// 		}
// 	`)
//
// 	// Add headers if required
// 	req.Header.Set("Authorization", "Bearer "+getEnvVar("API_TOKEN"))
//
// 	// Define a response structure
// 	var response struct {
// 		Uploads []struct {
// 			ID        string
// 			Filename  string
// 			Timestamp string
// 		}
// 	}
//
// 	// Perform the request
// 	if err := client.Run(context.Background(), req, &response); err != nil {
// 		return nil, err
// 	}
//
// 	// Prepare data for CSV
// 	var csvData [][]string
// 	csvData = append(csvData, []string{"ID", "Filename", "Timestamp"})
// 	for _, upload := range response.Uploads {
// 		csvData = append(csvData, []string{upload.ID, upload.Filename, upload.Timestamp})
// 	}
//
// 	return csvData, nil
// }

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
