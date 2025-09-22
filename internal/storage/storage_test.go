package storage

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalBackend(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		Provider: ProviderLocal,
		Path:     tempDir,
	}

	backend, err := NewLocalBackend(config)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Write and Read", func(t *testing.T) {
		// Write data
		writer, err := backend.NewWriter(ctx, "test.txt")
		require.NoError(t, err)

		testData := []byte("Hello, World!")
		n, err := writer.Write(testData)
		require.NoError(t, err)
		assert.Equal(t, len(testData), n)

		err = writer.Close()
		require.NoError(t, err)

		// Read data back
		reader, err := backend.NewReader(ctx, "test.txt")
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, testData, data)
	})

	t.Run("Exists", func(t *testing.T) {
		// Check existing file
		exists, err := backend.Exists(ctx, "test.txt")
		require.NoError(t, err)
		assert.True(t, exists)

		// Check non-existing file
		exists, err = backend.Exists(ctx, "nonexistent.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List", func(t *testing.T) {
		// Create another file
		writer, err := backend.NewWriter(ctx, "test2.txt")
		require.NoError(t, err)
		_, err = writer.Write([]byte("test"))
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		// List files
		files, err := backend.List(ctx, "test")
		require.NoError(t, err)
		assert.Len(t, files, 2)
		assert.Contains(t, files, "test.txt")
		assert.Contains(t, files, "test2.txt")
	})

	t.Run("Delete", func(t *testing.T) {
		err := backend.Delete(ctx, "test2.txt")
		require.NoError(t, err)

		exists, err := backend.Exists(ctx, "test2.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected *Config
	}{
		{
			name: "Local path",
			url:  "/tmp/file.txt",
			expected: &Config{
				Provider: ProviderLocal,
				Path:     "/tmp/file.txt",
			},
		},
		{
			name: "S3 URL",
			url:  "s3://my-bucket/path/to/file.parquet",
			expected: &Config{
				Provider: ProviderS3,
				Bucket:   "my-bucket",
				Path:     "path/to/file.parquet",
				Metadata: make(map[string]string),
			},
		},
		{
			name: "S3 URL with region",
			url:  "s3://my-bucket/file.parquet?region=us-west-2",
			expected: &Config{
				Provider: ProviderS3,
				Bucket:   "my-bucket",
				Path:     "file.parquet",
				Region:   "us-west-2",
				Metadata: make(map[string]string),
			},
		},
		{
			name: "GCS URL",
			url:  "gs://my-bucket/path/to/file.parquet",
			expected: &Config{
				Provider: ProviderGCS,
				Bucket:   "my-bucket",
				Path:     "path/to/file.parquet",
				Metadata: make(map[string]string),
			},
		},
		{
			name: "Azure URL",
			url:  "az://container/path/to/file.parquet",
			expected: &Config{
				Provider:    ProviderAzure,
				AccountName: "container",
				Bucket:      "path",
				Path:        "to/file.parquet",
				Metadata:    make(map[string]string),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseURL(tt.url)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Provider, config.Provider)
			assert.Equal(t, tt.expected.Bucket, config.Bucket)
			assert.Equal(t, tt.expected.Path, config.Path)
			if tt.expected.Region != "" {
				assert.Equal(t, tt.expected.Region, config.Region)
			}
		})
	}
}

func TestIsCloudURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"s3://bucket/file", true},
		{"gs://bucket/file", true},
		{"gcs://bucket/file", true},
		{"az://container/file", true},
		{"azure://container/file", true},
		{"/local/path/file", false},
		{"file:///local/path", false},
		{"http://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := IsCloudURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestS3BackendIntegration tests S3 backend with MinIO or real S3
// This test is skipped by default and requires proper S3 credentials
func TestS3BackendIntegration(t *testing.T) {
	// Skip unless S3 credentials are provided
	if os.Getenv("S3_TEST_BUCKET") == "" {
		t.Skip("Skipping S3 integration test - S3_TEST_BUCKET not set")
	}

	bucket := os.Getenv("S3_TEST_BUCKET")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	config := &Config{
		Provider:        ProviderS3,
		Bucket:          bucket,
		Region:          region,
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("S3_ENDPOINT"), // For MinIO
	}

	backend, err := NewS3Backend(config)
	require.NoError(t, err)

	ctx := context.Background()
	testPath := "test/integration-test.txt"

	t.Run("Write and Read S3", func(t *testing.T) {
		// Write data
		writer, err := backend.NewWriter(ctx, testPath)
		require.NoError(t, err)

		testData := []byte("Hello from S3 integration test!")
		_, err = writer.Write(testData)
		require.NoError(t, err)

		err = writer.Close()
		require.NoError(t, err)

		// Read data back
		reader, err := backend.NewReader(ctx, testPath)
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, testData, data)
	})

	t.Run("Cleanup S3", func(t *testing.T) {
		err := backend.Delete(ctx, testPath)
		require.NoError(t, err)

		exists, err := backend.Exists(ctx, testPath)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}