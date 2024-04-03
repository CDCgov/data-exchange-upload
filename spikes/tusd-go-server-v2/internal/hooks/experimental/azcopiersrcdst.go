package experimental

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
) // .import

type CopierAzSrcDst struct {
	SrcTusAzBlobClient    *azblob.Client
	SrcTusAzBlobName      string
	SrcTusAzContainerName string
	//
	DstAzBlobClient    *azblob.Client
	DstAzContainerName string
	DstAzBlobName      string
	Manifest           map[string]*string
} // .CopierAzSrcDst

// CopierAzSrcDst copies a file in azure from tus upload container to the dex container including adding manifest as file metadata
func (csd CopierAzSrcDst) CopyAzSrcToDst() error {

	ctx := context.TODO()

	get, err := csd.SrcTusAzBlobClient.DownloadStream(ctx, csd.SrcTusAzContainerName, csd.SrcTusAzBlobName, nil) // &azblob.DownloadStreamOptions{}
	if err != nil {
		return err
	} // .if

	_, err = csd.DstAzBlobClient.UploadStream(ctx, csd.DstAzContainerName, csd.DstAzBlobName, get.Body, &azblob.UploadStreamOptions{
		Metadata: csd.Manifest,
	})
	if err != nil {
		return err
	} // .if

	return nil
} // .CopierAzSrcDst
