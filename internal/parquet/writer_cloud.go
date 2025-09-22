package parquet

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/axonops/cqlai/internal/storage"
)

// CreateWriter creates an appropriate writer based on the output path
// It supports local files and cloud storage (S3, GCS, Azure)
func CreateWriter(ctx context.Context, output string) (io.WriteCloser, error) {
	// Check for special outputs
	if output == "" || output == "-" || output == "STDOUT" {
		return nopCloser{os.Stdout}, nil
	}

	// Check if it's a cloud URL
	if storage.IsCloudURL(output) {
		return createCloudWriter(ctx, output)
	}

	// Default to local file
	return os.Create(output) // #nosec G304 - output path is validated by caller
}

// createCloudWriter creates a writer for cloud storage
func createCloudWriter(ctx context.Context, output string) (io.WriteCloser, error) {
	// Parse the URL to get configuration
	config, err := storage.ParseURL(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse storage URL: %w", err)
	}

	// Load credentials from environment if not set
	if config.Provider == storage.ProviderS3 {
		loadS3CredentialsFromEnv(config)
	}

	// Create storage backend
	backend, err := storage.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage backend: %w", err)
	}

	// Extract path from config
	path := config.Path
	if config.Provider == storage.ProviderLocal {
		// For local storage, the full path is in config.Path
		path = ""
	}

	// Create writer
	writer, err := backend.NewWriter(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage writer: %w", err)
	}

	return writer, nil
}

// loadS3CredentialsFromEnv loads S3 credentials from environment variables
func loadS3CredentialsFromEnv(config *storage.Config) {
	// AWS SDK will automatically use these, but we can also set them explicitly
	if config.AccessKeyID == "" {
		config.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if config.SecretAccessKey == "" {
		config.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if config.SessionToken == "" {
		config.SessionToken = os.Getenv("AWS_SESSION_TOKEN")
	}
	if config.Region == "" {
		config.Region = os.Getenv("AWS_REGION")
		if config.Region == "" {
			config.Region = os.Getenv("AWS_DEFAULT_REGION")
		}
	}

	// For S3-compatible services
	if config.Endpoint == "" {
		config.Endpoint = os.Getenv("S3_ENDPOINT")
	}
}

// nopCloser wraps an io.Writer to add a no-op Close method
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// CreateReader creates an appropriate reader based on the input path
// It supports local files and cloud storage (S3, GCS, Azure)
func CreateReader(ctx context.Context, input string) (io.ReadCloser, error) {
	// Check if it's a cloud URL
	if storage.IsCloudURL(input) {
		return createCloudReader(ctx, input)
	}

	// Default to local file
	return os.Open(input) // #nosec G304 - input path is validated by caller
}

// createCloudReader creates a reader for cloud storage
func createCloudReader(ctx context.Context, input string) (io.ReadCloser, error) {
	// Parse the URL to get configuration
	config, err := storage.ParseURL(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse storage URL: %w", err)
	}

	// Load credentials from environment if not set
	if config.Provider == storage.ProviderS3 {
		loadS3CredentialsFromEnv(config)
	}

	// Create storage backend
	backend, err := storage.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage backend: %w", err)
	}

	// Extract path from config
	path := config.Path
	if config.Provider == storage.ProviderLocal {
		// For local storage, the full path is in config.Path
		path = ""
	}

	// Create reader
	reader, err := backend.NewReader(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage reader: %w", err)
	}

	return reader, nil
}