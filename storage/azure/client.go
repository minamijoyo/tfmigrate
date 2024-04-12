package azure

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Client interface {
	// Read an object from an Azure blob.
	Read(ctx context.Context, container, blob string) ([]byte, error)

	// Write an object onto an Azure blob.
	Write(ctx context.Context, container, blob string, p []byte) error
}

type client struct {
	BlobAPI *azblob.Client
}

// newClient returns a new instance of Client.
func newClient(config *Config) (Client, error) {
	// If the access key isn't defined in the configuration, try to read it from the environment.
	if config.AccessKey == "" {
		config.AccessKey = os.Getenv("TFMIGRATE_AZURE_STORAGE_ACCESS_KEY")
	}

	cred, err := azblob.NewSharedKeyCredential(config.AccountName, config.AccessKey)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s.blob.core.windows.net/", config.AccountName)
	c, err := azblob.NewClientWithSharedKeyCredential(url, cred, nil)

	return &client{c}, err
}

// Read an object from an Azure blob.
func (c *client) Read(ctx context.Context, container, blob string) ([]byte, error) {
	resp, err := c.BlobAPI.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, err
	}

	bs := bytes.Buffer{}
	r := resp.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	defer r.Close()

	_, err = bs.ReadFrom(r)

	return bs.Bytes(), err
}

// Write an object onto an Azure blob.
func (c *client) Write(ctx context.Context, container, blob string, p []byte) error {
	_, err := c.BlobAPI.UploadBuffer(ctx, container, blob, p, nil)

	return err
}
