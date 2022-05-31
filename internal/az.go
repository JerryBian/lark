package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func getContainerClient(create bool) (*azblob.ContainerClient, context.Context, error) {
	connStr, ok := os.LookupEnv(AzureConnectionString)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("Env %s not set.", AzureConnectionString))
	}

	containerName, ok := os.LookupEnv(AzureBlobContainer)
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("Env %s not set.", AzureBlobContainer))
	}

	serviceClient, err := azblob.NewServiceClientFromConnectionString(connStr, nil)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	container, err := serviceClient.NewContainerClient(containerName)
	if create {
		_, err = container.Create(ctx, &azblob.ContainerCreateOptions{  })
		if err != nil {
			return nil, nil, err
		}
	}
	
	return container, ctx, nil
}

func CreateContainerIfNotExists() error {
	_, _, err := getContainerClient(true)
	return err
}

func Save(suffix string, data []byte) error {
	blobName, ok := os.LookupEnv(AzureBlob)
	if !ok {
		return errors.New(fmt.Sprintf("Env %s not set.", AzureBlob))
	}

	containerClient, ctx, err := getContainerClient(false)
	if err!= nil {
		return err
	}

	if len(suffix) > 0 {
		fileExt := filepath.Ext(blobName)
		fileNameWithoutExt := strings.TrimSuffix(blobName, fileExt)
		blobName  = fileNameWithoutExt + suffix + fileExt
	}

	blockBlob, err := containerClient.NewBlockBlobClient(blobName)
	_, err = blockBlob.UploadBuffer(ctx, data, azblob.UploadOption{ })
	if err != nil {
		return err
	}

	return nil
}

func Load() ([]byte, error) {
	blobName, ok := os.LookupEnv(AzureBlob)
	if !ok {
		return nil, errors.New(fmt.Sprintf("Env %s not set.", AzureBlob))
	}

	containerClient, ctx, err := getContainerClient(false)
	if err!= nil {
		return nil, err
	}

	blockBlob, err := containerClient.NewBlockBlobClient(blobName)
	res, err := blockBlob.Download(ctx, &azblob.BlobDownloadOptions{})
	if err != nil {
		return nil, err
	}

	data := &bytes.Buffer{}
	reader:= res.Body(&azblob.RetryReaderOptions{})
	_, err = data.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}