package parquet

import (
	"context"
	"io"
	"os"

	"github.com/axonops/cqlai/internal/storage"
)

// CreateWriter creates an appropriate writer based on the output path
// It supports local files and stdout
func CreateWriter(ctx context.Context, output string) (io.WriteCloser, error) {
	// Check for special outputs
	if output == "" || output == "-" || output == "STDOUT" {
		return nopCloser{os.Stdout}, nil
	}

	// Parse URL to check if it's a cloud URL (will return error for cloud URLs)
	config, err := storage.ParseURL(output)
	if err != nil {
		return nil, err // This will include the helpful error message about cloud URLs
	}

	// Create local file writer
	return os.Create(config.Path) // #nosec G304 - output path is validated by caller
}

// nopCloser wraps an io.Writer to add a no-op Close method
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error {
	return nil
}

// CreateReader creates an appropriate reader based on the input path
// It supports local files and stdin
func CreateReader(ctx context.Context, input string) (io.ReadCloser, error) {
	// Check for special inputs
	if input == "" || input == "-" || input == "STDIN" {
		return io.NopCloser(os.Stdin), nil
	}

	// Parse URL to check if it's a cloud URL (will return error for cloud URLs)
	config, err := storage.ParseURL(input)
	if err != nil {
		return nil, err // This will include the helpful error message about cloud URLs
	}

	// Create local file reader
	return os.Open(config.Path) // #nosec G304 - input path is validated by caller
}