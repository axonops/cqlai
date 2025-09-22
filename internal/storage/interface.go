package storage

import (
	"context"
	"io"
)

// Provider represents a cloud storage provider
type Provider string

const (
	ProviderLocal Provider = "local"
	ProviderS3    Provider = "s3"
	ProviderAzure Provider = "azure"
	ProviderGCS   Provider = "gcs"
)

// Config holds configuration for cloud storage
type Config struct {
	Provider Provider
	Bucket   string
	Path     string
	Region   string

	// Provider-specific settings
	Endpoint        string            // For S3-compatible services
	AccessKeyID     string            // AWS/S3
	SecretAccessKey string            // AWS/S3
	SessionToken    string            // AWS temporary credentials
	AccountName     string            // Azure
	AccountKey      string            // Azure
	SASToken        string            // Azure SAS token
	ProjectID       string            // GCS
	CredentialsJSON string            // GCS service account JSON

	// Additional options
	UseSSL          bool              // For S3/MinIO
	PathStyle       bool              // For S3 path-style URLs
	Metadata        map[string]string // Custom metadata
}

// Writer provides a unified interface for writing to cloud storage
type Writer interface {
	io.WriteCloser
	// URL returns the full URL/path of the resource being written
	URL() string
}

// Reader provides a unified interface for reading from cloud storage
type Reader interface {
	io.ReadCloser
	// URL returns the full URL/path of the resource being read
	URL() string
}

// Backend is the interface that storage providers must implement
type Backend interface {
	// NewWriter creates a new writer for the given path
	NewWriter(ctx context.Context, path string) (Writer, error)

	// NewReader creates a new reader for the given path
	NewReader(ctx context.Context, path string) (Reader, error)

	// Exists checks if a path exists
	Exists(ctx context.Context, path string) (bool, error)

	// List lists objects with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Delete removes an object
	Delete(ctx context.Context, path string) error

	// Provider returns the provider type
	Provider() Provider
}


// New creates a new storage backend based on the configuration
func New(config *Config) (Backend, error) {
	switch config.Provider {
	case ProviderLocal:
		return NewLocalBackend(config)
	case ProviderS3:
		return NewS3Backend(config)
	case ProviderAzure:
		return NewAzureBackend(config)
	case ProviderGCS:
		return NewGCSBackend(config)
	default:
		return NewLocalBackend(config) // Default to local
	}
}