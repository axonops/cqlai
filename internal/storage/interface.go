package storage

import (
	"context"
	"io"
)

// Provider represents a storage provider
type Provider string

const (
	ProviderLocal Provider = "local"
)

// Config holds configuration for storage
type Config struct {
	Provider Provider
	Path     string

	// Additional options
	Metadata map[string]string // Custom metadata
}

// Writer provides a unified interface for writing to storage
type Writer interface {
	io.WriteCloser
	// URL returns the full URL/path of the resource being written
	URL() string
}

// Reader provides a unified interface for reading from storage
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
	// Always use local backend
	return NewLocalBackend(config)
}