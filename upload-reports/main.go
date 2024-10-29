package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"
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
var summaryFilename = "summary-report.csv"

var anomalousItemHeaders = []string{
	"Upload ID",
	"File Name",
	"Data Stream",
	"Route",
	"Category",
}
var anomalousItemFilename = "anomalous-items.csv"

func main() {
	config := utils.GetConfig()
	fmt.Printf("Datastreams: %v, StartDate: %v, EndDate: %v, TargetEnv: %v\n", config.DataStreams, config.StartDate, config.EndDate, config.TargetEnv)

	cleanedStartDate, cleanedEndDate := getFormattedDates(config.StartDate, config.EndDate)

	datastreams := strings.Split(config.DataStreams, ",")

	summaryData, anomalousData := getCsvData(datastreams, cleanedStartDate, cleanedEndDate, config.PsApiUrl)

	summaryBytes := generateCsvBytes(summaryData, "summary")
	anomalousBytes := generateCsvBytes(anomalousData, "anomalous")

	saveCsvFiles(summaryBytes, anomalousBytes, config.CsvOutputPath)

	if config.S3Config != nil {
		uploadToS3(&config, summaryBytes, anomalousBytes)
	}
}

func getFormattedDates(startDate, endDate string) (string, string) {
	cleanedStartDate, err := utils.FormatDateString(startDate)
	if err != nil {
		log.Fatalf("Start date is in incorrect format: %v", err)
	}

	cleanedEndDate, err := utils.FormatDateString(endDate)
	if err != nil {
		log.Fatalf("End date is in incorrect format: %v", err)
	}

	return cleanedStartDate, cleanedEndDate
}

func generateCsvBytes(data [][]string, dataType string) *bytes.Buffer {
	csvBytes, err := utils.CreateCSV(data)
	if err != nil {
		log.Fatalf("Error creating CSV for %s data: %v", dataType, err)
	}
	return csvBytes
}

func fetchDataForDataStream(apiURL string, datastream string, route string, startDate string, endDate string) (*SummaryRow, []AnomalousItemRow, error) {
	ctx := context.Background()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	client := graphql.NewClient(apiURL, httpClient)

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
			anomalousChan <- anomalousData

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

func saveCsvFiles(summaryBytes, anomalousBytes *bytes.Buffer, outputPath string) {
	if err := utils.SaveCsvToFile(summaryBytes, outputPath, summaryFilename); err != nil {
		log.Fatalf("Error saving CSV for summary data: %v", err)
	}

	if err := utils.SaveCsvToFile(anomalousBytes, outputPath, anomalousItemFilename); err != nil {
		log.Fatalf("Error saving CSV for anomalous data: %v", err)
	}
}

func uploadToS3(config *utils.AppConfig, summaryBytes, anomalousBytes *bytes.Buffer) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	zipBytes, err := utils.CreateZipArchive(summaryBytes, anomalousBytes, summaryFilename, anomalousItemFilename)
	if err != nil {
		log.Fatalf("Failed to create ZIP archive: %v", err)
	}

	key := fmt.Sprintf("upload-report-%s-%s.zip", config.TargetEnv, config.StartDate)
	client, err := utils.CreateS3Client(ctx, "us-east-1", config.S3Config.Endpoint)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	if err := utils.UploadCsvToS3(ctx, client, config.S3Config.BucketName, key, zipBytes); err != nil {
		log.Fatalf("Error uploading ZIP to S3: %v", err)
	}

	fmt.Println("ZIP file successfully uploaded to S3")
}
