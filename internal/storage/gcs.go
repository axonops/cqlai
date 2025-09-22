package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSBackend implements Backend for Google Cloud Storage
type GCSBackend struct {
	config *Config
	client *storage.Client
	bucket string
}

// NewGCSBackend creates a new Google Cloud Storage backend
func NewGCSBackend(cfg *Config) (*GCSBackend, error) {
	ctx := context.Background()
	var opts []option.ClientOption

	// Set credentials if provided
	if cfg.CredentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(cfg.CredentialsJSON)))
	}
	// Otherwise, it will use Application Default Credentials

	// Create GCS client
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSBackend{
		config: cfg,
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// gcsWriter implements Writer for Google Cloud Storage
type gcsWriter struct {
	writer *storage.Writer
	url    string
}

func (w *gcsWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *gcsWriter) Close() error {
	return w.writer.Close()
}

func (w *gcsWriter) URL() string {
	return w.url
}

// NewWriter creates a new writer for Google Cloud Storage
func (b *GCSBackend) NewWriter(ctx context.Context, path string) (Writer, error) {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	// Get bucket handle
	bucket := b.client.Bucket(b.bucket)
	obj := bucket.Object(path)

	// Create writer
	writer := obj.NewWriter(ctx)

	// Set metadata if provided
	if b.config.Metadata != nil {
		writer.Metadata = b.config.Metadata
	}

	url := fmt.Sprintf("gs://%s/%s", b.bucket, path)

	return &gcsWriter{
		writer: writer,
		url:    url,
	}, nil
}

// gcsReader implements Reader for Google Cloud Storage
type gcsReader struct {
	reader io.ReadCloser
	url    string
}

func (r *gcsReader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *gcsReader) Close() error {
	return r.reader.Close()
}

func (r *gcsReader) URL() string {
	return r.url
}

// NewReader creates a new reader for Google Cloud Storage
func (b *GCSBackend) NewReader(ctx context.Context, path string) (Reader, error) {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	// Get bucket handle
	bucket := b.client.Bucket(b.bucket)
	obj := bucket.Object(path)

	// Create reader
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS reader: %w", err)
	}

	url := fmt.Sprintf("gs://%s/%s", b.bucket, path)

	return &gcsReader{
		reader: reader,
		url:    url,
	}, nil
}

// Exists checks if an object exists
func (b *GCSBackend) Exists(ctx context.Context, path string) (bool, error) {
	path = strings.TrimPrefix(path, "/")

	bucket := b.client.Bucket(b.bucket)
	obj := bucket.Object(path)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// List lists objects with the given prefix
func (b *GCSBackend) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimPrefix(prefix, "/")

	bucket := b.client.Bucket(b.bucket)
	query := &storage.Query{
		Prefix: prefix,
	}

	var files []string
	it := bucket.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCS objects: %w", err)
		}
		files = append(files, attrs.Name)
	}

	return files, nil
}

// Delete removes an object
func (b *GCSBackend) Delete(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/")

	bucket := b.client.Bucket(b.bucket)
	obj := bucket.Object(path)

	err := obj.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete GCS object: %w", err)
	}

	return nil
}

// Provider returns the provider type
func (b *GCSBackend) Provider() Provider {
	return ProviderGCS
}