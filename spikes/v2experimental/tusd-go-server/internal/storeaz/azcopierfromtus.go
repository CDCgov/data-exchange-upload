package storeaz

import (
	"context"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tusd "github.com/tus/tusd/v2/pkg/handler"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
) // .import

type CopierAzTusToDex struct {
	EventUploadComplete tusd.HookEvent
	UploadConfig        metadatav1.UploadConfig
	CopyTargets         []metadatav1.CopyTarget
	//
	SrcTusAzBlobClient    *azblob.Client
	SrcTusAzBlobName      string
	SrcTusAzContainerName string
	//
	DstAzContainerName string
	DstAzBlobName      string
	IngestDt           time.Time
} // .CopierAzTusToDex

// CopyTusSrcToDst copies a file in azure from tus upload container to the dex container including adding manifest as file metadata
func (caz CopierAzTusToDex) CopyTusSrcToDst() error {

	ctx := context.TODO()

	get, err := caz.SrcTusAzBlobClient.DownloadStream(ctx, caz.SrcTusAzContainerName, caz.SrcTusAzBlobName, nil) // &azblob.DownloadStreamOptions{}
	if err != nil {
		return err
	} // .if

	manifest := make(map[string]*string)
	for mdk, mdv := range caz.EventUploadComplete.Upload.MetaData {
		manifest[mdk] = to.Ptr(mdv)
	} // .for
	manifest["dex_ingest_datetime"] = to.Ptr(caz.IngestDt.Format(time.RFC3339)) // add ingest datetime to file blob metadata

	_, err = caz.SrcTusAzBlobClient.UploadStream(context.TODO(), caz.DstAzContainerName, caz.DstAzBlobName, get.Body, &azblob.UploadStreamOptions{
		Metadata: manifest,
	})
	if err != nil {
		return err
	} // .if

	return nil
} // .CopyTusSrcToDst
