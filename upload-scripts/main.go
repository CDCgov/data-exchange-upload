package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// COMMAND LINE INPUTS
var (
	TARGET_ENV      *string = flag.String("target_env", "", "Target Environment")
	DATA_STREAM     *string = flag.String("data_stream", "", "The data stream name")
	ROUTE           *string = flag.String("route", "", "The route name")
)

// GITHUB ACTION INPUTS
var ACCOUNT_KEY string = ""

// FIXED VALUES
var CONTAINER string = "bulkuploads"
var FOLDER string = "tus-prefix/"

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

	if(ACCOUNT_KEY == "") {
		// IF not set for local runs pull from the environment
		envKey := fmt.Sprintf("ACCOUNT_KEY_%s", strings.ToUpper(*TARGET_ENV))
		ACCOUNT_KEY := os.Getenv(envKey)
		if ACCOUNT_KEY == "" {
			log.Fatalf("ACCOUNT_KEY is not set")
		}
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

	containerClient := serviceClient.ServiceClient().NewContainerClient(CONTAINER)

    listBlobstoDelete(serviceClient, containerClient)

	fmt.Println("------- DELETE SUMMARY -------")
	log.Printf("Delete List Count: %d", len(blobsToDelete))
	fmt.Println("------- DELETE SUMMARY -------")

	deleteBlobs(serviceClient)

	endTime := time.Now()
    fmt.Println("End Time:", endTime)

    duration := endTime.Sub(startTime)
    fmt.Printf("Duration: %v\n", duration)
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
			
				data_stream :=  metadata["meta_destination_id"]
				route := metadata["meta_ext_event"]

				// Define tags
				tags := map[string]string{
					"data_stream": fmt.Sprintf("%v", data_stream),
					"route": fmt.Sprintf("%v", route),
				}

				addTags(containerClient, blobName, tags)

				if(data_stream == DATA_STREAM && route == ROUTE) {

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