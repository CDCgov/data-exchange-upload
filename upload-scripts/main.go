package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

var pageCount int32

func main() {
	startTime := time.Now()
	fmt.Println("Start Time:", startTime)

	var serviceClient *azblob.Client

	if storageKey != "" {
		cred, err := azblob.NewSharedKeyCredential(storageAccount, storageKey)
		if err != nil {
			log.Fatalf("Failed to create shared key credential: %v", err)
		}

		serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccount)
		serviceClient, err = azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			log.Fatalf("Failed to create service client: %v", err)
		}
	} else if storageConnectionString != "" {
		var err error
		serviceClient, err = azblob.NewClientFromConnectionString(storageConnectionString, nil)
		if err != nil {
			log.Fatalf("Failed to create service client: %v", err)
		}
	}

	if serviceClient == nil {
		log.Fatalf("failed to init service client")
	}

	var o <-chan *searchResult
	ctx := context.Background()

	if searchTagsOnly {
		outChan := make(chan *searchResult)
		go func() {
			defer close(outChan)
			searchUploadsMatchingIndexTags(ctx, metadataCriteria, serviceClient, containerName, outChan)
		}()
		o = outChan
	} else {
		c := make(chan searchPage)
		o = initWorkers(ctx, c, serviceClient)

		go func() {
			defer close(c)
			searchUploadsByMetadata(ctx, metadataCriteria, serviceClient, containerName, blobPrefix, c)
		}()
	}

	searchSummary := searchResult{}
	for r := range o {
		searchSummary.totalSearched += r.totalSearched
		searchSummary.totalMatched += r.totalMatched
		searchSummary.totalMatchedBytes += r.totalMatchedBytes
		searchSummary.matchingUploads = append(searchSummary.matchingUploads, r.matchingUploads...)
	}

	fmt.Printf("searched %d blobs; matched on %d (%d total bytes)\r\n", searchSummary.totalSearched, searchSummary.totalMatched, searchSummary.totalMatchedBytes)
	fmt.Printf("Duration: %v\n", time.Since(startTime))

	if searchSummary.totalMatched == 0 {
		slog.Info("found no matching files")
		return
	}

	var ans string
	if !nonInteractive {
		r := bufio.NewReader(os.Stdin)
		fmt.Printf("%d uploads marked for deletion.  Proceed? (y/n) ", searchSummary.totalMatched)
		ans, _ = r.ReadString('\n')
		ans = strings.TrimSpace(ans)
	}

	if ans == "y" || ans == "yes" || nonInteractive {
		err := deleteUploads(ctx, searchSummary.matchingUploads, serviceClient)
		if err != nil {
			slog.Error("error deleting uploads", "error", err)
		}
	}
}

func initWorkers(ctx context.Context, c <-chan searchPage, serviceClient *azblob.Client) <-chan *searchResult {
	o := make(chan *searchResult)
	go func() {
		defer close(o)
		var wg sync.WaitGroup
		slog.Debug(fmt.Sprintf("starting %d workers", parallelism))
		for i := 0; i < parallelism; i++ {
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
	matchingUploads   []string
	totalMatched      int
	totalMatchedBytes int64
	totalSearched     int
	errors            []error
}

func searchUploadsByMetadata(ctx context.Context, metadata map[string]string, serviceClient *azblob.Client, containerName string, folderPrefix string, c chan<- searchPage) {
	pageSize := int32(maxPageSize)
	pager := serviceClient.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		MaxResults: &pageSize,
		Prefix:     &folderPrefix,
		Include: azblob.ListBlobsInclude{
			Tags: true,
		},
	})

	for pager.More() && pageCount < int32(maxPages) {
		page, err := pager.NextPage(ctx)
		if err != nil {
			slog.Error("error getting page", "error", err)
			continue
		}

		atomic.AddInt32(&pageCount, 1)

		c <- searchPage{
			page:             page.Segment,
			metadataCriteria: metadata,
		}
	}
}

func searchUploadsMatchingIndexTags(ctx context.Context, metadata map[string]string, serviceClient *azblob.Client, containerName string, o chan<- *searchResult) {
	cc := serviceClient.ServiceClient().NewContainerClient(containerName)

	w := buildWhereClause(metadata)
	rsp, err := cc.FilterBlobs(ctx, w, &container.FilterBlobsOptions{
		MaxResults: to.Ptr(int32(maxResults)),
	})
	if err != nil {
		slog.Error("error fetching filtered blobs", "error", err)
		return
	}

	uids, bytes := parseIndexResult(ctx, rsp, cc)

	o <- &searchResult{
		totalMatched:      len(rsp.Blobs),
		totalMatchedBytes: bytes,
		matchingUploads:   uids,
	}
}

func buildWhereClause(criteria map[string]string) string {
	out := ""
	for k, v := range criteria {
		out += fmt.Sprintf("\"%s\"='%s' AND ", k, v)
	}
	// Remove trailing AND
	out = strings.TrimRightFunc(out, unicode.IsSpace)
	out = strings.TrimSuffix(out, "AND")
	slog.Debug(fmt.Sprintf("using WHERE clause %s", out))
	return out
}

func parseIndexResult(ctx context.Context, res container.FilterBlobsResponse, containerClient *container.Client) ([]string, int64) {
	uids := make([]string, len(res.Blobs))
	var bytes int64

	for _, b := range res.Blobs {
		// Get blob size
		blobClient := containerClient.NewBlobClient(*b.Name)
		props, err := blobClient.GetProperties(ctx, nil)
		if err != nil {
			// TODO better error handling here
			slog.Warn("error getting blob properties", "error", err)
			continue
		}
		bytes += *props.ContentLength

		// Get uid
		if strings.Contains(*b.Name, ".info") {
			// Skip info files to avoid duplicates
			continue
		}
		uids = append(uids, *b.Name)
	}

	return uids, bytes
}

func worker(ctx context.Context, c <-chan searchPage, o chan<- *searchResult, serviceClient *azblob.Client) {
	for p := range c {
		result := &searchResult{}
		for _, blob := range p.page.BlobItems {
			result.totalSearched++

			if strings.Contains(*blob.Name, ".info") {
				// Found an upload
				uid := strings.Split(*blob.Name, ".")[0]
				slog.Debug("found info file for upload", "upload id", uid)

				// First, check if matched on tags
				if blob.BlobTags != nil && len(blob.BlobTags.BlobTagSet) > 0 {
					if matchesTags(blob.BlobTags, p.metadataCriteria) {
						result.matchingUploads = append(result.matchingUploads, uid)
						result.totalMatchedBytes += *blob.Properties.ContentLength

						uploadBlob := serviceClient.ServiceClient().NewContainerClient(containerName).NewBlobClient(uid)
						rsp, err := uploadBlob.GetProperties(ctx, nil)
						if err == nil {
							result.totalMatchedBytes += *rsp.ContentLength
						}
						continue
					}
				}

				// Download, unmarshal, and match on the metadata criteria
				rsp, err := serviceClient.DownloadStream(ctx, containerName, *blob.Name, nil)
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
					result.totalMatchedBytes += *blob.Properties.ContentLength

					uploadBlob := serviceClient.ServiceClient().NewContainerClient(containerName).NewBlobClient(uid)
					rsp, err := uploadBlob.GetProperties(ctx, nil)
					if err == nil {
						result.totalMatchedBytes += *rsp.ContentLength
					}
					continue
				}

				// Tag blob and info blob for all criteria metadata fields
				containerClient := serviceClient.ServiceClient().NewContainerClient(containerName)
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

// TODO use threads for this
func deleteUploads(ctx context.Context, files []string, serviceClient *azblob.Client) error {
	for _, f := range files {
		infoFile := f + ".info"
		uploadFile := f

		_, err := serviceClient.DeleteBlob(ctx, containerName, infoFile, nil)
		if err != nil {
			slog.Warn("failed to delete info file", "filename", infoFile, "error", err)
		} else {
			slog.Debug(fmt.Sprintf("successfully deleted file %s", infoFile))
		}

		_, err = serviceClient.DeleteBlob(ctx, containerName, uploadFile, nil)
		if err != nil {
			slog.Warn("failed to delete info file", "filename", uploadFile, "error", err)
		} else {
			slog.Debug(fmt.Sprintf("successfully deleted file %s", uploadFile))
		}
	}

	return nil
}
