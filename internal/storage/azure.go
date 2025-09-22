package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// AzureBackend implements Backend for Azure Blob Storage
type AzureBackend struct {
	config    *Config
	client    *azblob.Client
	container string
}

// NewAzureBackend creates a new Azure Blob Storage backend
func NewAzureBackend(cfg *Config) (*AzureBackend, error) {
	var client *azblob.Client
	var err error

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net", cfg.AccountName)

	// Determine authentication method
	switch {
	case cfg.SASToken != "":
		// Use SAS token
		serviceURL = serviceURL + "?" + cfg.SASToken
		client, err = azblob.NewClientWithNoCredential(serviceURL, nil)
	case cfg.AccountKey != "":
		// Use account key
		cred, err := azblob.NewSharedKeyCredential(cfg.AccountName, cfg.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create shared key credential: %w", err)
		}
		client, err = azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure client: %w", err)
		}
	default:
		// Use default Azure credentials (managed identity, environment vars, etc.)
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default credential: %w", err)
		}
		client, err = azblob.NewClient(serviceURL, cred, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure client: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &AzureBackend{
		config:    cfg,
		client:    client,
		container: cfg.Bucket, // Container name is stored in Bucket field
	}, nil
}

// azureWriter implements Writer for Azure Blob Storage
type azureWriter struct {
	buffer    *bytes.Buffer
	backend   *AzureBackend
	blobPath  string
	ctx       context.Context
	url       string
}

func (w *azureWriter) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}

func (w *azureWriter) Close() error {
	// Upload the buffer content to Azure
	containerClient := w.backend.client.ServiceClient().NewContainerClient(w.backend.container)
	blobClient := containerClient.NewBlockBlobClient(w.blobPath)

	// Upload using UploadBuffer which handles the buffer directly
	_, err := blobClient.UploadBuffer(w.ctx, w.buffer.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("failed to upload to Azure: %w", err)
	}

	return nil
}

func (w *azureWriter) URL() string {
	return w.url
}

// NewWriter creates a new writer for Azure Blob Storage
func (b *AzureBackend) NewWriter(ctx context.Context, path string) (Writer, error) {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	url := fmt.Sprintf("az://%s/%s", b.container, path)

	return &azureWriter{
		buffer:   bytes.NewBuffer(nil),
		backend:  b,
		blobPath: path,
		ctx:      ctx,
		url:      url,
	}, nil
}

// azureReader implements Reader for Azure Blob Storage
type azureReader struct {
	reader io.ReadCloser
	url    string
}

func (r *azureReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *azureReader) Close() error {
	return r.reader.Close()
}

func (r *azureReader) URL() string {
	return r.url
}

// NewReader creates a new reader for Azure Blob Storage
func (b *AzureBackend) NewReader(ctx context.Context, path string) (Reader, error) {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	containerClient := b.client.ServiceClient().NewContainerClient(b.container)
	blobClient := containerClient.NewBlobClient(path)

	response, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download from Azure: %w", err)
	}

	url := fmt.Sprintf("az://%s/%s", b.container, path)

	return &azureReader{
		reader: response.Body,
		url:    url,
	}, nil
}

// Exists checks if a blob exists
func (b *AzureBackend) Exists(ctx context.Context, path string) (bool, error) {
	path = strings.TrimPrefix(path, "/")

	containerClient := b.client.ServiceClient().NewContainerClient(b.container)
	blobClient := containerClient.NewBlobClient(path)

	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "BlobNotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// List lists blobs with the given prefix
func (b *AzureBackend) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimPrefix(prefix, "/")

	containerClient := b.client.ServiceClient().NewContainerClient(b.container)

	var files []string
	pager := containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Azure blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if blob.Name != nil {
				files = append(files, *blob.Name)
			}
		}
	}

	return files, nil
}

// Delete removes a blob
func (b *AzureBackend) Delete(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/")

	containerClient := b.client.ServiceClient().NewContainerClient(b.container)
	blobClient := containerClient.NewBlobClient(path)

	_, err := blobClient.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete Azure blob: %w", err)
	}

	return nil
}

// Provider returns the provider type
func (b *AzureBackend) Provider() Provider {
	return ProviderAzure
}