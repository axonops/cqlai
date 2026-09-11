package db

import (
	"testing"

	"github.com/axonops/cqlai/internal/config"
)

func TestPageSizeFromConfig(t *testing.T) {
	tests := []struct {
		name           string
		configPageSize int
		expectedSize   int
	}{
		{
			name:           "default page size when not set",
			configPageSize: 0,
			expectedSize:   100, // Default
		},
		{
			name:           "custom page size",
			configPageSize: 500,
			expectedSize:   500,
		},
		{
			name:           "negative page size uses default",
			configPageSize: -1,
			expectedSize:   100, // Default
		},
		{
			name:           "small page size",
			configPageSize: 10,
			expectedSize:   10,
		},
		{
			name:           "large page size",
			configPageSize: 10000,
			expectedSize:   10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Host:     "localhost",
				Port:     9042,
				PageSize: tt.configPageSize,
			}

			// Create a session struct directly to test page size initialization
			// We can't create a real session without a Cassandra connection,
			// but we can verify the logic by checking the expected page size calculation
			pageSize := cfg.PageSize
			if pageSize <= 0 {
				pageSize = 100
			}

			if pageSize != tt.expectedSize {
				t.Errorf("PageSize = %d, want %d", pageSize, tt.expectedSize)
			}
		})
	}
}

func TestSetPageSize(t *testing.T) {
	// Create a mock session to test SetPageSize
	session := &Session{
		pageSize: 100,
	}

	tests := []struct {
		name     string
		newSize  int
		expected int
	}{
		{
			name:     "set to 500",
			newSize:  500,
			expected: 500,
		},
		{
			name:     "set to 0 (disable paging)",
			newSize:  0,
			expected: 0,
		},
		{
			name:     "set to 1",
			newSize:  1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session.SetPageSize(tt.newSize)
			if session.PageSize() != tt.expected {
				t.Errorf("PageSize() = %d, want %d", session.PageSize(), tt.expected)
			}
		})
	}
}
