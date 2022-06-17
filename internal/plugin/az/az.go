package internal

import (
	"bytes"
	"context"

	"strings"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	C "github.com/JerryBian/lark/internal/config"
)

func getContainerClient(create bool, c *C.Config) (*azblob.ContainerClient, context.Context, error) {
	serviceClient, err := azblob.NewServiceClientFromConnectionString(c.Az.ConnStr, nil)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	container, err := serviceClient.NewContainerClient(c.Az.BlobContainer)
	if create {
		_, err = container.Create(ctx, &azblob.ContainerCreateOptions{  })
		if err != nil {
			return nil, nil, err
		}
	}
	
	return container, ctx, nil
}

func CreateContainerIfNotExists(c *C.Config) error {
	_, _, err := getContainerClient(true, c)
	return err
}

func Save(suffix string, data []byte, c *C.Config) error {
	containerClient, ctx, err := getContainerClient(false, c)
	if err!= nil {
		return err
	}

	blobName := c.Az.Blob
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

func Load(c *C.Config) ([]byte, error) {
	containerClient, ctx, err := getContainerClient(false, c)
	if err!= nil {
		return nil, err
	}

	blockBlob, err := containerClient.NewBlockBlobClient(c.Az.Blob)
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