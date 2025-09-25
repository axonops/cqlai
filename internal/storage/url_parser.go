package storage

import (
	"fmt"
	"strings"
)

// ParseURL parses a storage URL and returns the appropriate configuration
func ParseURL(rawURL string) (*Config, error) {
	// Check if this looks like a cloud URL that we no longer support
	if looksLikeCloudURL(rawURL) {
		return nil, fmt.Errorf("cloud storage URLs (s3://, gs://, az://) are no longer supported. Please mount your cloud storage as a local filesystem using rclone or similar tools. See: https://github.com/axonops/cqlai/blob/main/docs/cloud-storage.md")
	}

	// Everything is treated as a local path
	return &Config{
		Provider: ProviderLocal,
		Path:     rawURL,
		Metadata: make(map[string]string),
	}, nil
}

// IsCloudURL returns true if the URL points to cloud storage
// Deprecated: Always returns false as cloud URLs are no longer supported
func IsCloudURL(rawURL string) bool {
	return false
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