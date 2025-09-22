package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseURL parses a storage URL and returns the appropriate configuration
func ParseURL(rawURL string) (*Config, error) {
	// Handle local paths
	if !strings.Contains(rawURL, "://") {
		return &Config{
			Provider: ProviderLocal,
			Path:     rawURL,
		}, nil
	}

	// Parse as URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	config := &Config{
		Metadata: make(map[string]string),
	}

	// Determine provider based on scheme
	switch u.Scheme {
	case "s3":
		config.Provider = ProviderS3
		config.Bucket = u.Host
		config.Path = strings.TrimPrefix(u.Path, "/")

		// Parse query parameters for additional options
		params := u.Query()
		if region := params.Get("region"); region != "" {
			config.Region = region
		}
		if endpoint := params.Get("endpoint"); endpoint != "" {
			config.Endpoint = endpoint
		}
		if params.Get("path_style") == "true" {
			config.PathStyle = true
		}

	case "gs", "gcs":
		config.Provider = ProviderGCS
		config.Bucket = u.Host
		config.Path = strings.TrimPrefix(u.Path, "/")

		params := u.Query()
		if project := params.Get("project"); project != "" {
			config.ProjectID = project
		}

	case "az", "azure":
		config.Provider = ProviderAzure
		// For Azure, the host might be the account name or container
		parts := strings.Split(u.Host, ".")
		if len(parts) > 0 {
			config.AccountName = parts[0]
		}
		config.Path = strings.TrimPrefix(u.Path, "/")

		// Parse container from path if needed
		pathParts := strings.SplitN(config.Path, "/", 2)
		if len(pathParts) > 0 {
			config.Bucket = pathParts[0] // Container in Azure
			if len(pathParts) > 1 {
				config.Path = pathParts[1]
			} else {
				config.Path = ""
			}
		}

	case "file":
		config.Provider = ProviderLocal
		config.Path = u.Path

	default:
		return nil, fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	return config, nil
}

// IsCloudURL returns true if the URL points to cloud storage
func IsCloudURL(rawURL string) bool {
	if !strings.Contains(rawURL, "://") {
		return false
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	switch u.Scheme {
	case "s3", "gs", "gcs", "az", "azure":
		return true
	default:
		return false
	}
}