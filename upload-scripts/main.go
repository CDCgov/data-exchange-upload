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
	"strings"
	"sync"
	"sync/atomic"
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
			searchUploadsMatchingIndexTags(ctx, metadataCriteria, containerClient, outChan)
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
		o = initWorkers(ctx, c, containerClient, metadataCriteria)

		go func() {
			defer close(c)
			searchUploadsByMetadata(ctx, serviceClient, containerName, blobPrefix, c)
		}()
	}

	searchSummary := searchResult{}
	for r := range o {
		if r.err != nil {
			slog.Error("error parsing file", "uid", r.uid, "error", r.err)
			continue
		}
		searchSummary.totalMatched++

		bytes, err := deleteUpload(ctx, r.uid, serviceClient, containerClient)
		if err != nil {
			slog.Error("error deleting upload", "uid", r.uid, "error", err)
		}
		searchSummary.totalMatchedBytes += bytes
		slog.Debug("successfully deleted upload", "uid", r.uid)
	}

	fmt.Printf("searched %d blobs; matched on %d (%d total bytes)\r\n", totalSearched, searchSummary.totalMatched, searchSummary.totalMatchedBytes)
	fmt.Printf("Duration: %v\n", time.Since(startTime))

	if smoke {
		fmt.Println("skipped deletion due to smoke flag")
	}

	if searchSummary.totalMatched == 0 {
		slog.Info("found no matching files")
		return
	}
}

func initWorkers(ctx context.Context, c <-chan pageItemResult, containerClient *container.Client, criteria map[string]string) <-chan matchResult {
	o := make(chan matchResult)
	go func() {
		defer close(o)
		var wg sync.WaitGroup
		slog.Debug(fmt.Sprintf("starting %d workers", parallelism))
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				worker(ctx, c, o, containerClient, criteria)
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
	uid string
	err error
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
			c <- pageItemResult{
				item: b,
				tags: b.BlobTags,
			}
		}
	}
}

func searchUploadsMatchingIndexTags(ctx context.Context, metadata map[string]string, containerClient *container.Client, o chan<- matchResult) {
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
			uid = strings.Split(*res.Name, ".")[0]
		}
		o <- matchResult{
			uid: uid,
			err: nil,
		}
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

func worker(ctx context.Context, c <-chan pageItemResult, o chan<- matchResult, containerClient *container.Client, criteria map[string]string) {
	for r := range c {
		uid := *r.item.Name

		if strings.Contains(*r.item.Name, ".info") {
			uid = strings.Split(*r.item.Name, ".")[0]
		}
		// First, check if matched on tags
		if r.tags != nil && len(r.tags.BlobTagSet) > 0 {
			if matchesTags(r.tags, criteria) {
				slog.Debug("matched on tags", "uid", uid)
				o <- matchResult{
					uid: uid,
					err: nil,
				}
				continue
			}
		}

		// Couldn't match on tags, so fallback to match on metadata
		fileClient := containerClient.NewBlobClient(uid)
		rsp, err := fileClient.GetProperties(ctx, nil)
		if err == nil {
			if matchesMetadata(depointerizeMap(rsp.Metadata), criteria) {
				o <- matchResult{
					uid: uid,
					err: nil,
				}
				continue
			}
		}

		// Didn't find metadata on file.  Fallback to info file content
		infoClient := containerClient.NewBlobClient(uid + ".info")
		infoRsp, err := infoClient.DownloadStream(ctx, nil)
		if err != nil {
			o <- matchResult{
				uid: uid,
				err: err,
			}
			continue
		}

		body, err := io.ReadAll(infoRsp.Body)
		if err != nil {
			o <- matchResult{
				uid: uid,
				err: err,
			}
			continue
		}
		defer infoRsp.Body.Close()

		var data map[string]any
		err = json.Unmarshal(body, &data)
		if err != nil {
			o <- matchResult{
				uid: uid,
				err: err,
			}
			continue
		}

		metadata, ok := data["MetaData"].(map[string]any)
		if !ok {
			slog.Warn("found info file with no metadata; skipping", "uid", uid)
			continue
		}

		ms := convertMap(metadata)
		if matchesMetadata(ms, criteria) {
			o <- matchResult{
				uid: uid,
				err: nil,
			}
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

	return bytesDeleted, nil
}
