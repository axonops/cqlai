package parquet

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// CreateWriter creates an appropriate writer based on the output path
// It supports local files and stdout
func CreateWriter(ctx context.Context, output string) (io.WriteCloser, error) {
	// Check for special outputs
	if output == "" || output == "-" || output == "STDOUT" {
		return nopCloser{os.Stdout}, nil
	}

	// Check if this looks like a cloud URL that we no longer support
	if looksLikeCloudURL(output) {
		return nil, fmt.Errorf("cloud storage URLs (s3://, gs://, az://) are no longer supported. Please mount your cloud storage as a local filesystem using rclone or similar tools. See: https://github.com/axonops/cqlai/blob/main/docs/cloud-storage.md")
	}

	// Create local file writer
	return os.Create(output) // #nosec G304 - output path is validated by caller
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

	// Check if this looks like a cloud URL that we no longer support
	if looksLikeCloudURL(input) {
		return nil, fmt.Errorf("cloud storage URLs (s3://, gs://, az://) are no longer supported. Please mount your cloud storage as a local filesystem using rclone or similar tools. See: https://github.com/axonops/cqlai/blob/main/docs/cloud-storage.md")
	}

	// Create local file reader
	return os.Open(input) // #nosec G304 - input path is validated by caller
}

// looksLikeCloudURL checks if a URL appears to be a cloud storage URL
// Used to provide helpful error messages to users
func looksLikeCloudURL(rawURL string) bool {
	if !strings.Contains(rawURL, "://") {
		return false
	}

	// Check for common cloud storage URL schemes
	for _, prefix := range []string{
		"s3://",
		"gs://",
		"gcs://",
		"az://",
		"azure://",
	} {
		if strings.HasPrefix(rawURL, prefix) {
			return true
		}
	}

	return false
}