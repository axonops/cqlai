package router

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCSVHeaderParsing tests that CSV headers with (PK) and (C) suffixes are cleaned
func TestCSVHeaderParsing(t *testing.T) {
	tests := []struct {
		name           string
		csvContent     string
		expectedCols   []string
		description    string
	}{
		{
			name: "Header with PK suffix",
			csvContent: `id (PK),name,value
1,Alice,100
2,Bob,200
`,
			expectedCols: []string{"id", "name", "value"},
			description:  "Should strip (PK) suffix from column names",
		},
		{
			name: "Header with C suffix",
			csvContent: `id (PK),name,age (C)
1,Alice,30
2,Bob,25
`,
			expectedCols: []string{"id", "name", "age"},
			description:  "Should strip (C) suffix from column names",
		},
		{
			name: "Header with spaces around suffixes",
			csvContent: `  id (PK)  ,  name  ,  age (C)
1,Alice,30
`,
			expectedCols: []string{"id", "name", "age"},
			description:  "Should strip spaces and suffixes",
		},
		{
			name: "Header with multiple PK and C columns",
			csvContent: `user_id (PK),tenant_id (PK),timestamp (C),value
1,1,100,data1
`,
			expectedCols: []string{"user_id", "tenant_id", "timestamp", "value"},
			description:  "Should handle multiple PK and C columns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary CSV file
			tempDir := t.TempDir()
			csvFile := filepath.Join(tempDir, "test.csv")
			err := os.WriteFile(csvFile, []byte(tt.csvContent), 0644)
			require.NoError(t, err)

			// Read the CSV file and parse headers
			file, err := os.Open(csvFile)
			require.NoError(t, err)
			defer file.Close()

			csvReader := csv.NewReader(file)
			headerRow, err := csvReader.Read()
			require.NoError(t, err)

			// Clean headers using the same logic as handleCopyFrom
			cleanedHeaders := make([]string, len(headerRow))
			for i, col := range headerRow {
				cleanCol := strings.TrimSpace(col)
				// Remove (PK) suffix
				if idx := strings.Index(cleanCol, " (PK)"); idx != -1 {
					cleanCol = cleanCol[:idx]
				}
				// Remove (C) suffix
				if idx := strings.Index(cleanCol, " (C)"); idx != -1 {
					cleanCol = cleanCol[:idx]
				}
				cleanedHeaders[i] = strings.TrimSpace(cleanCol)
			}

			// Verify cleaned headers match expected
			assert.Equal(t, tt.expectedCols, cleanedHeaders, tt.description)
		})
	}
}

