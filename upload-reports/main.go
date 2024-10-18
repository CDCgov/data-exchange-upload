package main

import (
	"archive/zip"
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

type SummaryRow struct {
	DataStream             string
	Route                  string
	StartDate              string
	EndDate                string
	TotalUploadCount       int64
	PendingUploadCount     int64
	UndeliveredUploadCount int64
}

type AnomalousItemRow struct {
	UploadId   string
	Filename   string
	DataStream string
	Route      string
	Category   Category
}

type Category string

const (
	Pending     Category = "pending"
	Undelivered Category = "undelivered"
)

var summaryHeaders = []string{
	"Data Stream",
	"Route",
	"Start Date",
	"End Date",
	"Total Upload Count",
	"Pending Upload Count",
	"Undelivered Upload Count",
}

var anomalousItemHeaders = []string{
	"Upload ID",
	"File Name",
	"Data Stream",
	"Route",
	"Category",
}

func main() {
	config := utils.GetConfig()
	fmt.Printf("Datastreams: %v, StartDate: %v, EndDate: %v, TargetEnv: %v\n", config.DataStreams, config.StartDate, config.EndDate, config.TargetEnv)

	cleanedStartDate, err := utils.FormatDateString(config.StartDate)
	if err != nil {
		log.Fatalf("Start date is in incorrect format: %v", err)
	}

	cleanedEndDate, err := utils.FormatDateString(config.EndDate)
	if err != nil {
		log.Fatalf("End date is in incorrect format: %v", err)
	}

	datastreams := strings.Split(config.DataStreams, ",")

	summaryData, anomalousData := getCsvData(datastreams, cleanedStartDate, cleanedEndDate, config.PsApiUrl)

	summaryBytes, err := createCSV(summaryData)
	if err != nil {
		log.Fatalf("Error creating CSV for summary data: %v", err)
	}

	anomalousBytes, err := createCSV(anomalousData)
	if err != nil {
		log.Fatalf("Error creating CSV for anomalous data: %v", err)
	}

	saveCsvToFile(summaryBytes, config.CsvOutputPath, "summary-report.csv")
	if err != nil {
		log.Fatalf("Error saving CSV for summary data: %v", err)
	}

	saveCsvToFile(anomalousBytes, config.CsvOutputPath, "anomalous-items.csv")
	if err != nil {
		log.Fatalf("Error saving CSV for anomalous data: %v", err)
	}

	if config.S3Config != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		zipBytes, err := createZipArchive(summaryBytes, anomalousBytes)
		if err != nil {
			log.Fatalf("failed to create ZIP archive: %v", err)
		}

		key := fmt.Sprintf("upload-report-%s-%s.zip", config.TargetEnv, config.StartDate)

		client, err := createS3Client(ctx, "us-east-1", config.S3Config.Endpoint)
		if err != nil {
			log.Fatalf("failed to create S3 client: %v", err)
		}

		if err := uploadCsvToS3(ctx, client, config.S3Config.BucketName, key, zipBytes); err != nil {
			log.Fatalf("Error uploading CSV to S3: %v", err)
		}
	}
}

func fetchDataForDataStream(apiURL string, datastream string, route string, startDate string, endDate string) (*SummaryRow, []AnomalousItemRow, error) {
	ctx := context.Background()
	client := graphql.NewClient(apiURL, http.DefaultClient)

	resp, err := psApi.GetUploadStats(ctx, client, datastream, route, startDate, endDate)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to fetch upload stats for datastream %s, route %s: %w", datastream, route, err)
	}

	fmt.Printf("PS-API -- Datastream: %v, Route: %v, UploadCount: %v\n", datastream, route, resp.GetGetUploadStats().CompletedUploadsCount)

	reportRow := SummaryRow{
		DataStream:             datastream,
		Route:                  route,
		StartDate:              startDate,
		EndDate:                endDate,
		TotalUploadCount:       resp.GetGetUploadStats().CompletedUploadsCount,
		PendingUploadCount:     resp.GetGetUploadStats().PendingUploads.TotalCount,
		UndeliveredUploadCount: resp.GetGetUploadStats().UndeliveredUploads.TotalCount,
	}

	var anomalousItems []AnomalousItemRow

	if resp.GetGetUploadStats().PendingUploads.TotalCount > 0 {
		for _, pending := range resp.GetGetUploadStats().PendingUploads.PendingUploads {
			anomalousItem := AnomalousItemRow{
				UploadId:   pending.UploadId,
				Filename:   pending.Filename,
				DataStream: datastream,
				Route:      route,
				Category:   Pending,
			}
			anomalousItems = append(anomalousItems, anomalousItem)
		}
	}

	if resp.GetGetUploadStats().UndeliveredUploads.TotalCount > 0 {
		for _, undelivered := range resp.GetGetUploadStats().UndeliveredUploads.UndeliveredUploads {
			anomalousItem := AnomalousItemRow{
				UploadId:   undelivered.UploadId,
				Filename:   undelivered.Filename,
				DataStream: datastream,
				Route:      route,
				Category:   Undelivered,
			}
			anomalousItems = append(anomalousItems, anomalousItem)
		}
	}

	return &reportRow, anomalousItems, nil
}

func getCsvData(datastreams []string, cleanedStartDate string, cleanedEndDate string, psApiUrl string) ([][]string, [][]string) {
	var summaryData [][]string
	summaryData = append(summaryData, summaryHeaders)

	var anomalousData [][]string
	anomalousData = append(anomalousData, anomalousItemHeaders)

	var wg sync.WaitGroup
	summaryChan := make(chan *SummaryRow, len(datastreams))
	anomalousChan := make(chan []AnomalousItemRow, len(datastreams))
	wg.Add(len(datastreams))

	for _, ds := range datastreams {
		go func(ds string) {
			defer wg.Done()

			datastream, route, err := utils.FormatStreamAndRoute(ds)
			if err != nil {
				log.Printf("There was an issue parsing the datastream and route: %v", err)
				return
			}

			rowData, anomalousData, err := fetchDataForDataStream(psApiUrl, datastream, route, cleanedStartDate, cleanedEndDate)
			if err != nil {
				log.Printf("Error fetching data from GraphQL API: %v", err)
				return
			}

			summaryChan <- rowData

			if len(anomalousData) > 0 {
				anomalousChan <- anomalousData
			}
		}(ds)
	}

	wg.Wait()
	close(summaryChan)
	close(anomalousChan)

	for rowData := range summaryChan {
		summaryData = append(summaryData, []string{
			rowData.DataStream,
			rowData.Route,
			rowData.StartDate,
			rowData.EndDate,
			strconv.FormatInt(rowData.TotalUploadCount, 10),
			strconv.FormatInt(rowData.PendingUploadCount, 10),
			strconv.FormatInt(rowData.UndeliveredUploadCount, 10),
		})
	}

	for items := range anomalousChan {
		for _, item := range items {
			anomalousData = append(anomalousData, []string{
				item.UploadId,
				item.Filename,
				item.DataStream,
				item.Route,
				string(item.Category),
			})
		}
	}

	return summaryData, anomalousData
}

func createCSV(data [][]string) (*bytes.Buffer, error) {
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

	return &buf, nil
}

func saveCsvToFile(csvData *bytes.Buffer, outputPath string, filename string) error {
	fullPath := filepath.Join(outputPath, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create the file %v: %v", file, err)
	}
	defer file.Close()

	_, err = file.Write(csvData.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", fullPath, err)
	}

	fmt.Printf("CSV successfully saved to file: %s\n", fullPath)
	return nil
}

func createZipArchive(overviewData, anomalousData *bytes.Buffer) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	overviewFile, err := zipWriter.Create("summary-report.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create overview file in zip: %v", err)
	}
	_, err = overviewFile.Write(overviewData.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write overview data to zip: %v", err)
	}

	anomalousFile, err := zipWriter.Create("anomalous-items.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create anomalous file in zip: %v", err)
	}
	_, err = anomalousFile.Write(anomalousData.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to write anomalous data to zip: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %v", err)
	}

	return buf, nil
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

func uploadCsvToS3(ctx context.Context, client *s3.Client, bucketName string, key string, csvData *bytes.Buffer) error {
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(csvData.Bytes()),
		ContentType: aws.String("application/zip"),
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
