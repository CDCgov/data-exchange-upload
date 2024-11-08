package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// TODO move out to config.go
// COMMAND LINE INPUTS
var (
	TARGET_ENV  *string = flag.String("target_env", "", "Target Environment")
	DATA_STREAM *string = flag.String("data_stream", "", "The data stream name")
	ROUTE       *string = flag.String("route", "", "The route name")
)

// GITHUB ACTION INPUTS
var ACCOUNT_KEY = ""

// FIXED VALUES
var CONTAINER = "bulkuploads"
var FOLDER = "tus-prefix/"

// Concurrency
const MAX_CONCURRENCY = 10

// Delete List
var blobsToDelete []string
var mutex sync.Mutex

var readCount = big.NewInt(0)

func main() {
	startTime := time.Now()
	fmt.Println("Start Time:", startTime)

	flag.Parse()

	if *TARGET_ENV == "" {
		log.Fatalf("TARGET_ENV is not set")
	}

	if *DATA_STREAM == "" {
		log.Fatalf("DATA_STREAM is not set")
	}

	if *ROUTE == "" {
		log.Fatalf("ROUTE is not set")
	}

	ACCOUNT_KEY = os.Getenv("ACCOUNT_KEY")
	if ACCOUNT_KEY == "" {
		log.Fatalf("ACCOUNT_KEY is not set")
	}

	fmt.Println("------- SCRIPT INPUTS -------")
	fmt.Println("TARGET_ENV:", *TARGET_ENV)
	fmt.Println("DATA_STREAM:", *DATA_STREAM)
	fmt.Println("ROUTE:", *ROUTE)
	fmt.Println("CONTAINER:", CONTAINER)
	fmt.Println("FOLDER:", FOLDER)
	fmt.Println("------- SCRIPT INPUTS -------")

	STORAGE_ACCOUNT := getStorageAccountName(*TARGET_ENV)
	if STORAGE_ACCOUNT == "" {
		log.Fatalf("STORAGE_ACCOUNT is not set")
	}

	fmt.Println("STORAGE_ACCOUNT: ", STORAGE_ACCOUNT)

	cred, err := azblob.NewSharedKeyCredential(STORAGE_ACCOUNT, ACCOUNT_KEY)
	if err != nil {
		log.Fatalf("Failed to create shared key credential: %v", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", STORAGE_ACCOUNT)
	serviceClient, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create service client: %v", err)
	}

	//containerClient := serviceClient.ServiceClient().NewContainerClient(CONTAINER)

	// TODO use data_stream_id and data_stream_route
	criteria := map[string]string{
		"meta_destination_id": *DATA_STREAM,
		"meta_ext_event":      *ROUTE,
	}

	c := make(chan searchPage)
	defer close(c)
	ctx := context.Background()
	o := initWorkers(ctx, c, serviceClient)

	go func() {
		searchUploadsByMetadata(ctx, criteria, serviceClient, CONTAINER, FOLDER, c)
	}()

	searchSummary := searchResult{}
	for r := range o {
		// may need to use atomic here
		searchSummary.totalSearched += r.totalSearched
		searchSummary.totalMatched += r.totalMatched
		fmt.Printf("searched %d blobs; matched on %d\r\n", searchSummary.totalSearched, searchSummary.totalMatched)
	}

	//listBlobstoDelete(serviceClient, containerClient)
	//
	//fmt.Println("------- DELETE SUMMARY -------")
	//log.Printf("Delete List Count: %d", len(blobsToDelete))
	//fmt.Println("------- DELETE SUMMARY -------")
	//
	//deleteBlobs(serviceClient)

	fmt.Printf("Duration: %v\n", time.Since(startTime))
}

func initWorkers(ctx context.Context, c <-chan searchPage, serviceClient *azblob.Client) <-chan *searchResult {
	o := make(chan *searchResult)
	go func() {
		defer close(o)
		var wg sync.WaitGroup
		slog.Info("starting 10 workers")
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(ctx, c, o, serviceClient)
			}()
		}
		wg.Wait()
	}()
	return o
}

type searchPage struct {
	page             *container.BlobFlatListSegment
	metadataCriteria map[string]string
}

type searchResult struct {
	matchingUploads []string
	totalMatched    int
	totalSearched   int
	errors          []error
}

func searchUploadsByMetadata(ctx context.Context, metadata map[string]string, serviceClient *azblob.Client, containerName string, folderPrefix string, c chan<- searchPage) {
	// Loop through all blobs
	// TODO add to config
	//var maxResults int32 = 10
	pager := serviceClient.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		//MaxResults: &maxResults,
		Prefix: &folderPrefix,
		Include: azblob.ListBlobsInclude{
			Tags: true,
		},
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			slog.Error("error getting page", "error", err)
			continue
		}

		go func() {
			c <- searchPage{
				page:             page.Segment,
				metadataCriteria: metadata,
			}
		}()
	}
}

func worker(ctx context.Context, c <-chan searchPage, o chan<- *searchResult, serviceClient *azblob.Client) {
	for p := range c {
		result := &searchResult{
			totalMatched:  0,
			totalSearched: len(p.page.BlobItems),
		}
		for _, blob := range p.page.BlobItems {
			if strings.Contains(*blob.Name, ".info") {
				// Found an upload
				uid := strings.Split(*blob.Name, ".")[0]
				slog.Info("found info file for upload", "upload id", uid)

				// First, check if matched on tags
				if blob.BlobTags != nil && len(blob.BlobTags.BlobTagSet) > 0 {
					if matchesTags(blob.BlobTags, p.metadataCriteria) {
						result.matchingUploads = append(result.matchingUploads, uid)
						continue
					}
				}

				// Download, unmarshal, and match on the metadata criteria
				rsp, err := serviceClient.DownloadStream(ctx, CONTAINER, *blob.Name, nil)
				if err != nil {
					result.errors = append(result.errors, err)
					continue
				}

				body, err := io.ReadAll(rsp.Body)
				if err != nil {
					result.errors = append(result.errors, err)
					continue
				}
				defer rsp.Body.Close()

				var data map[string]any
				err = json.Unmarshal(body, &data)
				if err != nil {
					result.errors = append(result.errors, err)
					continue
				}

				metadata, ok := data["MetaData"].(map[string]any)
				if !ok {
					slog.Warn("found info file with no metadata; skipping", "filename", *blob.Name)
					continue
				}

				ms := convertMap(metadata)
				if matchesMetadata(ms, p.metadataCriteria) {
					result.matchingUploads = append(result.matchingUploads, uid)
					continue
				}

				// Tag blob and info blob for all criteria metadata fields
				containerClient := serviceClient.ServiceClient().NewContainerClient(CONTAINER)
				infoBlobClient := containerClient.NewBlobClient(*blob.Name)
				uploadBlobClient := containerClient.NewBlobClient(uid)
				_, err = infoBlobClient.SetTags(ctx, ms, nil)
				if err != nil {
					slog.Warn("error while tagging file", "filename", blob.Name, "error", err)
				}
				_, err = uploadBlobClient.SetTags(ctx, ms, nil)
				if err != nil {
					slog.Warn("error while tagging file", "filename", blob.Name, "error", err)
				}
			}
		}

		result.totalMatched = len(result.matchingUploads)
		o <- result
	}
}

func matchesTags(tags *container.BlobTags, criteria map[string]string) bool {
	tagMap := make(map[string]string)
	for _, t := range tags.BlobTagSet {
		tagMap[*t.Key] = *t.Value
	}

	return matchesMetadata(tagMap, criteria)
}

func matchesMetadata(metadata map[string]string, criteria map[string]string) bool {
	for k, v := range criteria {
		super, ok := metadata[k]
		if !ok || super != v {
			return false
		}
	}
	return true
}

func convertMap(m map[string]any) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = v.(string)
	}

	return out
}

func listBlobstoDelete(serviceClient *azblob.Client, containerClient *container.Client) {

	options := azblob.ListBlobsFlatOptions{
		Prefix: &FOLDER,
	}

	pager := serviceClient.NewListBlobsFlatPager(CONTAINER, &options)

	sem := make(chan struct{}, MAX_CONCURRENCY)

	var wg sync.WaitGroup

	for pager.More() {

		resp, err := pager.NextPage(context.Background())
		if err != nil {

			log.Fatalf("Failed to list blobs: %v", err)
		}

		for _, blobItem := range resp.Segment.BlobItems {

			if strings.HasSuffix(*blobItem.Name, "/") {
				continue // Skip directories
			}

			wg.Add(1)

			blobName := *blobItem.Name

			go func(blobName string) {

				defer wg.Done()

				sem <- struct{}{}
				defer func() { <-sem }()

				if strings.Contains(blobName, ".info") {

					readBlob(serviceClient, containerClient, blobName)
				}
			}(blobName)
		}
	}

	wg.Wait()
}

func readBlob(serviceClient *azblob.Client, containerClient *container.Client, blobName string) {

	readCount.Add(readCount, big.NewInt(1))
	log.Printf("Inspecting Blob: %s  Count: %d", blobName, readCount)

	resp, err := serviceClient.DownloadStream(context.Background(), CONTAINER, blobName, nil)

	if err != nil {

		log.Printf("Failed to download blob: %v", err)
	} else {

		defer func() {
			if err := resp.Body.Close(); err != nil {

				log.Printf("Failed to close response body: %v", err)
			}
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {

			log.Printf("Failed to read blob content: %v", err)
		}

		var jsonData map[string]interface{}

		if err := json.Unmarshal(body, &jsonData); err != nil {

			log.Printf("Blob content is not JSON: %s\nContent:\n%s\n", blobName, string(body))
			return

		} else {

			metadata, metadataExists := jsonData["MetaData"].(map[string]interface{})

			if metadataExists {

				data_stream := metadata["meta_destination_id"]
				route := metadata["meta_ext_event"]

				// Define tags
				tags := map[string]string{
					"data_stream": fmt.Sprintf("%v", data_stream),
					"route":       fmt.Sprintf("%v", route),
				}

				addTags(containerClient, blobName, tags)

				if data_stream == DATA_STREAM && route == ROUTE {

					mutex.Lock()
					blobsToDelete = append(blobsToDelete, strings.TrimSuffix(blobName, ".info"))
					blobsToDelete = append(blobsToDelete, blobName)
					mutex.Unlock()
				}
			}
		}
	}
}

func getStorageAccountName(targetEnv string) string {

	switch targetEnv {
	case "dev":
		return "ocioededataexchangedev"
	case "tst":
		return "ocioededataexchangetst"
	case "stg":
		return "ocioededataexchangestg"
	case "prd":
		return "ocioededataexchangeprd"
	default:
		return ""
	}
}

func addTags(containerClient *container.Client, blobName string, tags map[string]string) {

	blobClient := containerClient.NewBlobClient(blobName)
	_, err := blobClient.SetTags(context.Background(), tags, nil)
	if err != nil {
		log.Printf("Failed to set tags on blob %s: %v", blobName, err)
	}
}

func deleteBlobs(serviceClient *azblob.Client) {

	for _, blobName := range blobsToDelete {
		log.Printf("Deleting Blob - %s", blobName)
		_, err := serviceClient.DeleteBlob(context.Background(), CONTAINER, blobName, nil)
		if err != nil {
			log.Printf("Failed to delete blob: %v", err)
		}
	}
}
