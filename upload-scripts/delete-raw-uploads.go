package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)


// GITHUB ACTION INPUTS
var STORAGE_ACCOUNT = "storage-accont-name"
var ACCOUNT_KEY = "account-key"
var CONTAINER string = "bulkuploads"
var FOLDER string = "tus-prefix/"
var DESTINATION string = "data-stream-name"
var EVENT string = "route"

// Delete List
var blobsToDelete []string

func main() {

	setParameters()
	
	cred, err := azblob.NewSharedKeyCredential(STORAGE_ACCOUNT, ACCOUNT_KEY)
	if err != nil {

		log.Fatalf("Failed to create shared key credential: %v", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", STORAGE_ACCOUNT)
	serviceClient, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {

		log.Fatalf("Failed to create service client: %v", err)
	}

    listBlobstoDelete(serviceClient)

	deleteBlobs(serviceClient)
}

func setParameters() {

	if (STORAGE_ACCOUNT == "") {
		STORAGE_ACCOUNT := os.Getenv("STORAGE_ACCOUNT")
		if STORAGE_ACCOUNT == "" {
			log.Fatalf("STORAGE_ACCOUNT is not set")
		}
	}

	if (ACCOUNT_KEY == "") {
		ACCOUNT_KEY := os.Getenv("ACCOUNT_KEY")
		if ACCOUNT_KEY == "" {
			log.Fatalf("ACCOUNT_KEY is not set")
		}
	}

	if (CONTAINER == "") {
		CONTAINER := os.Getenv("CONTAINER")
		if CONTAINER == "" {
			log.Fatalf("CONTAINER is not set")
		}
	}

	if (FOLDER == "") {
		FOLDER := os.Getenv("FOLDER")
		if FOLDER == "" {
			log.Fatalf("FOLDER is not set")
		}
	}

	if (DESTINATION == "") {
		DESTINATION := os.Getenv("DESTINATION")
		if DESTINATION == "" {
			log.Fatalf("DESTINATION is not set")
		}
	}

	if (EVENT == "") {
		EVENT := os.Getenv("EVENT")
		if EVENT == "" {
			log.Fatalf("EVENT is not set")
		}
	}
}

func listBlobstoDelete(serviceClient *azblob.Client) {

	pager := serviceClient.NewListBlobsFlatPager(CONTAINER, &azblob.ListBlobsFlatOptions{
		Prefix: &FOLDER,
	})

	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {

			log.Fatalf("Failed to list blobs: %v", err)
		}

		for _, blobItem := range resp.Segment.BlobItems {

			blobName := *blobItem.Name

			if strings.HasSuffix(blobName, "/") {

				continue
			}

			if strings.Contains(blobName, ".info") {
				
				readBlob(serviceClient, blobName)
			}
		}
	}
}
 
func readBlob(serviceClient *azblob.Client, blobName string) {

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
			
				destination :=  metadata["meta_destination_id"]
				event := metadata["meta_ext_event"]

				if(destination == DESTINATION && event == EVENT) {

					blobsToDelete = append(blobsToDelete, strings.TrimSuffix(blobName, ".info"))
					blobsToDelete = append(blobsToDelete, blobName)

					log.Printf("Delete List Count: %d", len(blobsToDelete))
				} 
			}	
		}
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