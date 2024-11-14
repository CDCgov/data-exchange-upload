package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

var pageCount int32
var totalSearched int

func main() {
	startTime := time.Now()
	fmt.Println("Start Time:", startTime)

	var serviceClient *azblob.Client
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", storageAccount)

	if storageKey != "" {
		cred, err := azblob.NewSharedKeyCredential(storageAccount, storageKey)
		if err != nil {
			log.Fatalf("Failed to create shared key credential: %v", err)
		}

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
	} else {
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatalf("Failed to get credential with default identity: %v", err)
		}
		serviceClient, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			log.Fatalf("Failed to create service client: %v", err)
		}
	}

	if serviceClient == nil {
		log.Fatalf("failed to init service client")
	}

	containerClient := serviceClient.ServiceClient().NewContainerClient(containerName)

	var o <-chan matchResult
	ctx := context.Background()

	// This currently isn't working.  Getting permission issue.  Need to see if we have the action/filter permission
	if searchTagsOnly {
		outChan := make(chan matchResult)
		go func() {
			defer close(outChan)
			searchUploadsMatchingIndexTags(ctx, metadataCriteria, serviceClient, containerClient, outChan)
		}()
		o = outChan
		// TODO implement checkpoint file saving
	} else if checkpointFile != "" {
		outChan := make(chan *searchResult)
		go func() {
			defer close(outChan)
			err := loadCheckpointFile(checkpointFile, outChan)
			if err != nil {
				log.Fatalf("error loading checkpoint file %v", err)
			}
		}()
	} else {
		c := make(chan pageItemResult)
		o = initWorkers(ctx, c, serviceClient, containerClient, metadataCriteria)

		go func() {
			defer close(c)
			searchUploadsByMetadata(ctx, serviceClient, containerName, blobPrefix, c)
		}()
	}

	searchSummary := searchResult{}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		printSummary(searchSummary, startTime)
		os.Exit(1)
	}()

	logTicker := time.NewTicker(30 * time.Second)
	defer logTicker.Stop()
	go func() {
		for range logTicker.C {
			slog.Info("heartbeat log", "total searched", totalSearched, "total matched", searchSummary.totalMatched, "total bytes", searchSummary.totalMatchedBytes)
		}
	}()

	for r := range o {
		if r.err != nil {
			slog.Error("error parsing file", "uid", r.uid, "error", r.err)
			continue
		}

		if r.bytesDeleted > 0 {
			searchSummary.totalMatched++
			searchSummary.totalMatchedBytes += r.bytesDeleted
			slog.Info("found match", "uid", r.uid, "total matched", searchSummary.totalMatched, "total searched", totalSearched)
			continue
		}

		slog.Debug("no match on file", "uid", r.uid)
	}

	printSummary(searchSummary, startTime)
}

func initWorkers(ctx context.Context, c <-chan pageItemResult, serviceClient *azblob.Client, containerClient *container.Client, criteria map[string]string) <-chan matchResult {
	o := make(chan matchResult)
	go func() {
		defer close(o)
		var wg sync.WaitGroup
		slog.Debug(fmt.Sprintf("starting %d workers", parallelism))
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(ctx, c, o, serviceClient, containerClient, criteria)
			}()
		}
		wg.Wait()
	}()
	return o
}

type searchResult struct {
	totalMatched      int
	totalMatchedBytes int64
	errors            []error
}

type pageItemResult struct {
	item *container.BlobItem
	tags *container.BlobTags
}

type matchResult struct {
	foundMatch   bool
	uid          string
	bytesDeleted int64
	err          error
}

func searchUploadsByMetadata(ctx context.Context, serviceClient *azblob.Client, containerName string, folderPrefix string, c chan<- pageItemResult) {
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

		for _, b := range page.Segment.BlobItems {
			totalSearched++
			slog.Debug("searching file", "file", *b.Name, "total searched", totalSearched)

			c <- pageItemResult{
				item: b,
				tags: b.BlobTags,
			}
		}
	}
}

func searchUploadsMatchingIndexTags(ctx context.Context, metadata map[string]string, serviceClient *azblob.Client, containerClient *container.Client, o chan<- matchResult) {
	w := buildWhereClause(metadata)
	rsp, err := containerClient.FilterBlobs(ctx, w, &container.FilterBlobsOptions{
		MaxResults: to.Ptr(int32(maxResults)),
	})
	if err != nil {
		slog.Error("error fetching filtered blobs", "error", err)
		return
	}

	for _, res := range rsp.Blobs {
		totalSearched++
		uid := *res.Name
		if strings.Contains(*res.Name, ".info") {
			continue
		}
		handleMatchResult(ctx, matchResult{
			foundMatch: true,
			uid:        uid,
			err:        nil,
		}, o, serviceClient, containerClient)
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

func worker(ctx context.Context, c <-chan pageItemResult, o chan<- matchResult, serviceClient *azblob.Client, containerClient *container.Client, criteria map[string]string) {
	for r := range c {
		if strings.Contains(*r.item.Name, ".info") {
			// Skip .info files to prevent duplicate matches.  Result handling will take care of .info files.
			continue
		}

		uid := *r.item.Name

		// First, check if matched on tags
		if r.tags != nil && len(r.tags.BlobTagSet) > 0 {
			slog.Debug("attempting to match on tags", "uid", uid, "tags", r.tags)
			if matchesTags(r.tags, criteria) {
				slog.Debug("matched on tags", "uid", uid)

				handleMatchResult(ctx, matchResult{
					foundMatch: true,
					uid:        uid,
					err:        nil,
				}, o, serviceClient, containerClient)
				continue
			}
		}

		// Couldn't match on tags, so fallback to match on metadata
		fileClient := containerClient.NewBlobClient(uid)
		rsp, err := fileClient.GetProperties(ctx, nil)
		if err != nil {
			slog.Debug("error getting file metadata", "error", err)
			handleMatchResult(ctx, matchResult{
				foundMatch: false,
				uid:        uid,
				err:        err,
			}, o, serviceClient, containerClient)
			continue
		}

		slog.Debug("attempting to match on metadata", "uid", uid, "metadata", rsp.Metadata)
		if matchesMetadata(depointerizeMap(rsp.Metadata), criteria) {
			handleMatchResult(ctx, matchResult{
				foundMatch: false,
				uid:        uid,
				err:        nil,
			}, o, serviceClient, containerClient)
			continue
		}

		// Didn't find metadata on file.  Fallback to info file content
		infoClient := containerClient.NewBlobClient(uid + ".info")
		infoRsp, err := infoClient.DownloadStream(ctx, nil)
		if err != nil {
			slog.Debug("error downloading info file", "error", err)
			handleMatchResult(ctx, matchResult{
				foundMatch: false,
				uid:        uid,
				err:        err,
			}, o, serviceClient, containerClient)
			continue
		}

		// Deserialize info file
		body, err := io.ReadAll(infoRsp.Body)
		if err != nil {
			handleMatchResult(ctx, matchResult{
				uid: uid,
				err: err,
			}, o, serviceClient, containerClient)
			continue
		}
		defer infoRsp.Body.Close()

		var data map[string]any
		err = json.Unmarshal(body, &data)
		if err != nil {
			handleMatchResult(ctx, matchResult{
				foundMatch: false,
				uid:        uid,
				err:        err,
			}, o, serviceClient, containerClient)
			continue
		}

		metadata, ok := data["MetaData"].(map[string]any)
		if !ok {
			slog.Warn("found info file with no metadata; skipping", "uid", uid)
			continue
		}

		ms := convertMap(metadata)
		slog.Debug("attempting to match on info file metadata", "uid", uid, "metadata", ms)
		if matchesMetadata(ms, criteria) {
			handleMatchResult(ctx, matchResult{
				foundMatch: true,
				uid:        uid,
				err:        nil,
			}, o, serviceClient, containerClient)
			continue
		}

		// Blob didn't match criteria.  Tag for future inspection
		_, err = fileClient.SetTags(ctx, ms, nil)
		if err != nil {
			slog.Debug("error while tagging file", "uid", uid, "error", err)
		}
		_, err = infoClient.SetTags(ctx, ms, nil)
		if err != nil {
			slog.Debug("error while tagging info file", "uid", uid, "error", err)
		}
	}
}

func matchesTags(tags *container.BlobTags, criteria map[string]string) bool {
	if tags == nil {
		return false
	}
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

func handleMatchResult(ctx context.Context, result matchResult, o chan<- matchResult, serviceClient *azblob.Client, containerClient *container.Client) {
	if result.err != nil {
		o <- result
	}

	if result.foundMatch {
		bytes, err := deleteUpload(ctx, result.uid, serviceClient, containerClient)
		if err != nil {
			result.err = fmt.Errorf("error deleting upload; %w", err)
		}
		result.bytesDeleted = bytes
	}

	o <- result
}

func convertMap(m map[string]any) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = v.(string)
	}

	return out
}

func depointerizeMap(m map[string]*string) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		out[k] = *v
	}

	return out
}

func loadCheckpointFile(filename string, o chan<- *searchResult) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var checkpointSummary searchResult
	err = json.Unmarshal(b, &checkpointSummary)
	if err != nil {
		return err
	}

	o <- &checkpointSummary
	return nil
}

func deleteUpload(ctx context.Context, uid string, serviceClient *azblob.Client, containerClient *container.Client) (int64, error) {
	var bytesDeleted int64
	infoClient := containerClient.NewBlobClient(uid + ".info")
	fileClient := containerClient.NewBlobClient(uid)

	rsp, err := infoClient.GetProperties(ctx, nil)
	if err != nil {
		slog.Warn("failed to get info file for deletion", "uid", uid, "error", err)
	} else {
		if !smoke {
			_, err := serviceClient.DeleteBlob(ctx, containerName, uid+".info", nil)
			if err != nil {
				return 0, err
			}
		}
		bytesDeleted += *rsp.ContentLength
	}

	rsp, err = fileClient.GetProperties(ctx, nil)
	if err != nil {
		slog.Warn("failed to get upload file for deletion", "uid", uid, "error", err)
	} else {
		if !smoke {
			_, err := serviceClient.DeleteBlob(ctx, containerName, uid, nil)
			if err != nil {
				return bytesDeleted, err
			}
		}
		bytesDeleted += *rsp.ContentLength
	}

	slog.Debug("successfully deleted upload", "uid", uid)
	return bytesDeleted, nil
}

func printSummary(summary searchResult, startTime time.Time) {
	fmt.Println("**********")
	fmt.Printf("searched %d blobs; matched on %d (%d total bytes)\r\n", totalSearched, summary.totalMatched, summary.totalMatchedBytes)
	fmt.Printf("Duration: %v\n", time.Since(startTime))

	if smoke {
		fmt.Println("skipped deletion due to smoke flag")
	}

	if summary.totalMatched == 0 {
		slog.Info("found no matching files")
	}
	fmt.Println("**********")
}
