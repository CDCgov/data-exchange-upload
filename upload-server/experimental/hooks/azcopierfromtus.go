package hooks

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
) // .import

type CopierAzTusToDex struct {
	SrcTusAzBlobClient    *azblob.Client
	SrcTusAzBlobName      string
	SrcTusAzContainerName string
	//
	DstAzContainerName string
	DstAzBlobName      string
	Manifest           map[string]*string
} // .CopierAzTusToDex

// CopyTusSrcToDst copies a file in azure from tus upload container to the dex container including adding manifest as file metadata
func (cd CopierAzTusToDex) CopyTusSrcToDst() error {

	ctx := context.TODO()

	get, err := cd.SrcTusAzBlobClient.DownloadStream(ctx, cd.SrcTusAzContainerName, cd.SrcTusAzBlobName, nil) // &azblob.DownloadStreamOptions{}
	if err != nil {
		return err
	} // .if

	_, err = cd.SrcTusAzBlobClient.UploadStream(ctx, cd.DstAzContainerName, cd.DstAzBlobName, get.Body, &azblob.UploadStreamOptions{
		Metadata: cd.Manifest,
	})
	if err != nil {
		return err
	} // .if

	return nil
} // .CopyTusSrcToDst
