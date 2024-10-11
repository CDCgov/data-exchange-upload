package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cdcgov/data-exchange-upload/upload-reports/psApi"
	"github.com/cdcgov/data-exchange-upload/upload-reports/utils"
)

type ReportDataRow struct {
	DataStream           string
	Route                string
	StartDate            string
	EndDate              string
	UploadCount          int64
	DeliverySuccessCount int64
	DeliveryEndCount     int64
}

func main() {
	config := utils.GetConfig()
	fmt.Printf("Target Env: %s, DataStreams: %v, Start Date: %s, End Date: %s\n", config.TargetEnv, config.DataStreams, config.StartDate, config.EndDate)

	csvData := getCsvData(config)

	csvBytes, err := createCSV(csvData)
	if err != nil {
		log.Fatalf("Error creating CSV: %v", err)
	}

	fmt.Printf("CSV Data: %v\n", csvBytes)

	saveCsvToFile(csvBytes)
	if err != nil {
		log.Fatalf("Error saving CSV: %v", err)
	}

	if config.S3Config != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		key := fmt.Sprintf("upload-report-%s-%s.csv", config.TargetEnv, config.StartDate)

		client, err := createS3Client(ctx, "us-east-1", config.S3Config.Endpoint)
		if err != nil {
			log.Fatalf("failed to create S3 client: %v", err)
		}

		if err := uploadCsvToS3(ctx, client, config.S3Config.BucketName, key, csvBytes); err != nil {
			log.Fatalf("Error uploading CSV to S3: %v", err)
		}
	}
}

func fetchDataForDataStream(apiURL string, datastream string, route string, startDate string, endDate string) (ReportDataRow, error) {
	ctx := context.Background()
	client := graphql.NewClient(apiURL, http.DefaultClient)

	resp, err := psApi.GetUploadStats(ctx, client, datastream, route, startDate, endDate)

	if err != nil {
		fmt.Printf("There was an issue reaching graphql: %v\n", err)
	}

	reportRow := ReportDataRow{
		DataStream:           datastream,
		Route:                route,
		StartDate:            startDate,
		EndDate:              endDate,
		UploadCount:          resp.GetGetUploadStats().CompletedUploadsCount,
		DeliverySuccessCount: resp.GetGetUploadStats().PendingUploads.TotalCount,
		DeliveryEndCount:     resp.GetGetUploadStats().UnDeliveredUploads.TotalCount,
	}

	return reportRow, nil
}

func getCsvData(config utils.AppConfig) [][]string {
	var csvData [][]string
	csvData = append(csvData, []string{"Data Stream", "Route", "Start Date", "End Date", "Upload Count", "Delivery Success Count", "Delivery Fail Count"})

	cleanedStartDate, err := utils.FormatDateString(config.StartDate)
	if err != nil {
		log.Fatalf("Start date is in incorrect format: %v", err)
	}

	cleanedEndDate, err := utils.FormatDateString(config.EndDate)
	if err != nil {
		log.Fatalf("End date is in incorrect format: %v", err)
	}

	datastreams := strings.Split(config.DataStreams, ",")
	var wg sync.WaitGroup
	dataChan := make(chan ReportDataRow, len(datastreams))

	for _, ds := range datastreams {
		wg.Add(1)
		go func(ds string) {
			defer wg.Done()

			datastream, route, err := utils.FormatStreamAndRoute(ds)
			if err != nil {
				log.Printf("There was an issue parsing the datastream and route: %v", err)
				return
			}

			rowData, err := fetchDataForDataStream(config.PsApiUrl, datastream, route, cleanedStartDate, cleanedEndDate)
			if err != nil {
				log.Printf("Error fetching data from GraphQL API: %v", err)
				return
			}

			dataChan <- rowData
		}(ds)
	}

	wg.Wait()
	close(dataChan)

	for rowData := range dataChan {
		csvData = append(csvData, []string{
			rowData.DataStream,
			rowData.Route,
			rowData.StartDate,
			rowData.EndDate,
			strconv.FormatInt(rowData.UploadCount, 10),
			strconv.FormatInt(rowData.DeliverySuccessCount, 10),
			strconv.FormatInt(rowData.DeliveryEndCount, 10),
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

func saveCsvToFile(csvData []byte) error {
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get the current working directory: %v", err)
	}

	fullPath := filepath.Join(workingDir, "upload-report.csv")

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create the file %v: %v", file, err)
	}
	defer file.Close()

	_, err = file.Write(csvData)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", fullPath, err)
	}

	fmt.Printf("CSV successfully saved to file: %s\n", fullPath)
	return nil
}

func createS3Client(ctx context.Context, region string, endpoint string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.UsePathStyle = true
			o.BaseEndpoint = &endpoint
		}
	})

	return client, nil
}

func uploadCsvToS3(ctx context.Context, client *s3.Client, bucketName string, key string, csvData []byte) error {
	putInput := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(csvData),
	}

	_, err := client.PutObject(ctx, putInput)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("upload to S3 timed out")
		}
		fmt.Printf("Error putting to S3: %v\n", err)
		return fmt.Errorf("failed to upload CSV to S3: %v", err)
	}

	fmt.Printf("Successfully uploaded %s to S3 bucket %s\n", key, bucketName)
	return nil
}
