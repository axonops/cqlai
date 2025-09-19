// +build !integration

package router

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/session"
	"github.com/stretchr/testify/require"
)


func BenchmarkCopyToCSV(b *testing.B) {
	b.Skip("Skipping benchmark that requires proper session mocking")
	benchmarks := []struct {
		name     string
		rowCount int
	}{
		{"100_rows", 100},
		{"1000_rows", 1000},
		{"10000_rows", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cfg := &config.Config{}
			sessionMgr := session.NewManager(cfg)

			mockSession := NewMockSessionWithRowCount(bm.rowCount)

			handler := &MetaCommandHandler{
				session:        mockSession.Session,
				sessionManager: sessionMgr,
			}

			tempDir := b.TempDir()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				outputPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.csv", i))
				command := fmt.Sprintf("COPY test_table TO '%s'", outputPath)

				result := handler.handleCopy(command)
				require.Contains(b, result, fmt.Sprintf("Exported %d rows", bm.rowCount))

				// Clean up
				os.Remove(outputPath)
			}
		})
	}
}

func BenchmarkCopyToParquet(b *testing.B) {
	b.Skip("Skipping benchmark that requires proper session mocking")
	benchmarks := []struct {
		name     string
		rowCount int
	}{
		{"100_rows", 100},
		{"1000_rows", 1000},
		{"10000_rows", 10000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cfg := &config.Config{}
			sessionMgr := session.NewManager(cfg)

			mockSession := NewMockSessionWithRowCount(bm.rowCount)

			handler := &MetaCommandHandler{
				session:        mockSession.Session,
				sessionManager: sessionMgr,
			}

			tempDir := b.TempDir()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				outputPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.parquet", i))
				command := fmt.Sprintf("COPY test_table TO '%s' WITH FORMAT='PARQUET'", outputPath)

				result := handler.handleCopy(command)
				require.Contains(b, result, fmt.Sprintf("Exported %d rows", bm.rowCount))

				// Clean up
				os.Remove(outputPath)
			}
		})
	}
}

func BenchmarkCopyToParquetWithCompression(b *testing.B) {
	b.Skip("Skipping benchmark that requires proper session mocking")
	compressions := []string{"SNAPPY", "GZIP", "ZSTD"}
	rowCount := 1000

	for _, compression := range compressions {
		b.Run(compression, func(b *testing.B) {
			cfg := &config.Config{}
			sessionMgr := session.NewManager(cfg)

			mockSession := NewMockSessionWithRowCount(rowCount)

			handler := &MetaCommandHandler{
				session:        mockSession.Session,
				sessionManager: sessionMgr,
			}

			tempDir := b.TempDir()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				outputPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.parquet", i))
				command := fmt.Sprintf("COPY test_table TO '%s' WITH FORMAT='PARQUET' AND COMPRESSION='%s'",
					outputPath, compression)

				result := handler.handleCopy(command)
				require.Contains(b, result, fmt.Sprintf("Exported %d rows", rowCount))

				// Clean up
				os.Remove(outputPath)
			}
		})
	}
}

// TestCopyToPerformanceComparison compares CSV vs Parquet performance
func TestCopyToPerformanceComparison(t *testing.T) {
	t.Skip("Skipping test that requires proper session mocking")
	if testing.Short() {
		t.Skip("Skipping performance comparison in short mode")
	}

	rowCounts := []int{100, 1000, 10000}
	cfg := &config.Config{}
	sessionMgr := session.NewManager(cfg)
	tempDir := t.TempDir()

	for _, rowCount := range rowCounts {
		t.Run(fmt.Sprintf("%d_rows", rowCount), func(t *testing.T) {
			mockSession := NewMockSessionWithRowCount(rowCount)

			handler := &MetaCommandHandler{
				session:        mockSession.Session,
				sessionManager: sessionMgr,
			}

			// Test CSV
			csvPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.csv", rowCount))
			csvStart := time.Now()
			csvCommand := fmt.Sprintf("COPY test_table TO '%s'", csvPath)
			csvResult := handler.handleCopy(csvCommand)
			csvDuration := time.Since(csvStart)
			require.Contains(t, csvResult, fmt.Sprintf("Exported %d rows", rowCount))

			csvInfo, err := os.Stat(csvPath)
			require.NoError(t, err)

			// Test Parquet
			parquetPath := filepath.Join(tempDir, fmt.Sprintf("test_%d.parquet", rowCount))
			parquetStart := time.Now()
			parquetCommand := fmt.Sprintf("COPY test_table TO '%s' WITH FORMAT='PARQUET'", parquetPath)
			parquetResult := handler.handleCopy(parquetCommand)
			parquetDuration := time.Since(parquetStart)
			require.Contains(t, parquetResult, fmt.Sprintf("Exported %d rows", rowCount))

			parquetInfo, err := os.Stat(parquetPath)
			require.NoError(t, err)

			// Compare results
			t.Logf("Row Count: %d", rowCount)
			t.Logf("CSV: Duration=%v, Size=%d bytes (%.2f bytes/row)",
				csvDuration, csvInfo.Size(), float64(csvInfo.Size())/float64(rowCount))
			t.Logf("Parquet: Duration=%v, Size=%d bytes (%.2f bytes/row)",
				parquetDuration, parquetInfo.Size(), float64(parquetInfo.Size())/float64(rowCount))
			t.Logf("Size reduction: %.1f%%",
				(1-float64(parquetInfo.Size())/float64(csvInfo.Size()))*100)
			t.Logf("Speed ratio: %.2fx",
				float64(csvDuration.Nanoseconds())/float64(parquetDuration.Nanoseconds()))
		})
	}
}