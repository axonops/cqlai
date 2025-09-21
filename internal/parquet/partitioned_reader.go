package parquet

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
)

// GetFileInfo returns file info for a path
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// PartitionedParquetReader reads from partitioned Parquet datasets
type PartitionedParquetReader struct {
	baseDir          string
	partitionFiles   []PartitionFile
	currentFileIndex int
	currentReader    *ParquetReader
	partitionColumns []string
	totalRowCount    int64
}

// PartitionFile represents a single Parquet file in a partitioned dataset
type PartitionFile struct {
	Path            string
	PartitionValues map[string]string
}

// NewPartitionedParquetReader creates a reader for partitioned Parquet datasets
func NewPartitionedParquetReader(path string) (*PartitionedParquetReader, error) {
	// Check if path is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		// Single file - not partitioned
		return nil, fmt.Errorf("path is not a directory; use regular ParquetReader for single files")
	}

	pr := &PartitionedParquetReader{
		baseDir:          path,
		partitionFiles:   []PartitionFile{},
		currentFileIndex: -1,
	}

	// Discover all Parquet files in the directory tree
	if err := pr.discoverPartitionFiles(); err != nil {
		return nil, fmt.Errorf("failed to discover partition files: %w", err)
	}

	if len(pr.partitionFiles) == 0 {
		return nil, fmt.Errorf("no Parquet files found in directory: %s", path)
	}

	// Sort files for consistent ordering
	sort.Slice(pr.partitionFiles, func(i, j int) bool {
		return pr.partitionFiles[i].Path < pr.partitionFiles[j].Path
	})

	// Open first file to get schema
	if err := pr.nextFile(); err != nil {
		return nil, fmt.Errorf("failed to open first file: %w", err)
	}

	return pr, nil
}

// discoverPartitionFiles walks the directory tree and finds all Parquet files
func (pr *PartitionedParquetReader) discoverPartitionFiles() error {
	return filepath.Walk(pr.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's a Parquet file
		if !strings.HasSuffix(strings.ToLower(path), ".parquet") {
			return nil
		}

		// Parse partition values from path
		relPath, err := filepath.Rel(pr.baseDir, filepath.Dir(path))
		if err != nil {
			return err
		}

		partitionValues := pr.parsePartitionPath(relPath)

		pr.partitionFiles = append(pr.partitionFiles, PartitionFile{
			Path:            path,
			PartitionValues: partitionValues,
		})

		// Update partition columns if not set
		if len(pr.partitionColumns) == 0 && len(partitionValues) > 0 {
			for key := range partitionValues {
				pr.partitionColumns = append(pr.partitionColumns, key)
			}
			sort.Strings(pr.partitionColumns)
		}

		return nil
	})
}

// parsePartitionPath extracts partition values from a Hive-style path
func (pr *PartitionedParquetReader) parsePartitionPath(relPath string) map[string]string {
	if relPath == "." {
		return map[string]string{}
	}

	values := make(map[string]string)
	parts := strings.Split(relPath, string(os.PathSeparator))

	for _, part := range parts {
		// Look for key=value pattern
		if idx := strings.Index(part, "="); idx > 0 {
			key := part[:idx]
			value := part[idx+1:]

			// Unescape special values
			value = strings.ReplaceAll(value, "__NULL__", "")
			value = strings.ReplaceAll(value, "__SLASH__", "/")
			value = strings.ReplaceAll(value, "__EQ__", "=")

			values[key] = value
		}
	}

	return values
}

// nextFile advances to the next Parquet file in the partition
func (pr *PartitionedParquetReader) nextFile() error {
	// Close current reader if exists
	if pr.currentReader != nil {
		// Ignore close errors - the reader might already be closed
		_ = pr.currentReader.Close()
		pr.currentReader = nil
	}

	pr.currentFileIndex++
	if pr.currentFileIndex >= len(pr.partitionFiles) {
		return io.EOF
	}

	file := pr.partitionFiles[pr.currentFileIndex]

	logger.DebugfToFile("PartitionedReader", "Opening partition file: %s", file.Path)

	reader, err := NewParquetReader(file.Path)
	if err != nil {
		return fmt.Errorf("failed to open Parquet file %s: %w", file.Path, err)
	}

	pr.currentReader = reader
	pr.totalRowCount += reader.GetRowCount()

	return nil
}

// GetSchema returns the schema from the current file
func (pr *PartitionedParquetReader) GetSchema() ([]string, []string) {
	if pr.currentReader == nil {
		return nil, nil
	}

	columns, types := pr.currentReader.GetSchema()

	// Add partition columns to schema if not already present
	// Skip virtual partition columns (those with dots in the name)
	columnSet := make(map[string]bool)
	for _, col := range columns {
		columnSet[col] = true
	}

	for _, partCol := range pr.partitionColumns {
		// Skip virtual columns (e.g., event_id.year, event_id.month)
		if strings.Contains(partCol, ".") {
			continue
		}
		if !columnSet[partCol] {
			columns = append(columns, partCol)
			types = append(types, "string") // Partition columns are always strings
		}
	}

	return columns, types
}

// ReadAll reads all rows from all partition files
func (pr *PartitionedParquetReader) ReadAll() ([]map[string]interface{}, error) {
	var allRows []map[string]interface{}

	// Reset to first file
	pr.currentFileIndex = -1
	pr.currentReader = nil

	for {
		if err := pr.nextFile(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Read rows from current file
		rows, err := pr.currentReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read from partition file: %w", err)
		}

		// Add partition values to each row (only if not already present)
		partitionValues := pr.partitionFiles[pr.currentFileIndex].PartitionValues
		for i := range rows {
			for key, value := range partitionValues {
				// Only add partition column if it doesn't already exist in the row
				if _, exists := rows[i][key]; !exists {
					// Convert string partition value to appropriate type based on existing data
					if value != "" {
						// Try to parse as number if possible
						if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
							rows[i][key] = intVal
						} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
							rows[i][key] = floatVal
						} else {
							rows[i][key] = value
						}
					} else {
						rows[i][key] = nil // Empty string means NULL
					}
				}
			}
		}

		allRows = append(allRows, rows...)
	}

	return allRows, nil
}

// ReadBatch reads a batch of rows, potentially spanning multiple files
func (pr *PartitionedParquetReader) ReadBatch(batchSize int) ([]map[string]interface{}, error) {
	var batch []map[string]interface{}
	remaining := batchSize

	for remaining > 0 {
		if pr.currentReader == nil {
			if err := pr.nextFile(); err != nil {
				if err == io.EOF && len(batch) > 0 {
					// Return partial batch at end of dataset
					return batch, nil
				}
				return batch, err
			}
		}

		// Try to read from current file
		rows, err := pr.currentReader.ReadBatch(remaining)
		if err != nil {
			if err == io.EOF {
				// Current file exhausted, close it and try next file
				if pr.currentReader != nil {
					_ = pr.currentReader.Close()
					pr.currentReader = nil
				}
				continue
			}
			return nil, err
		}

		// Add partition values to rows (only if not already present)
		// Skip virtual partition columns (those with dots in the name)
		if pr.currentFileIndex >= 0 && pr.currentFileIndex < len(pr.partitionFiles) {
			partitionValues := pr.partitionFiles[pr.currentFileIndex].PartitionValues
			for i := range rows {
				for key, value := range partitionValues {
					// Skip virtual partition columns (e.g., event_id.year, event_id.month)
					if strings.Contains(key, ".") {
						continue
					}
					// Only add partition column if it doesn't already exist in the row
					if _, exists := rows[i][key]; !exists {
						// Convert string partition value to appropriate type
						if value != "" {
							// Try to parse as number if possible
							if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
								rows[i][key] = intVal
							} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
								rows[i][key] = floatVal
							} else {
								rows[i][key] = value
							}
						} else {
							rows[i][key] = nil // Empty string means NULL
						}
					}
				}
			}
		}

		batch = append(batch, rows...)
		remaining -= len(rows)

		// If no rows were read, close current file and move to next
		if len(rows) == 0 {
			if pr.currentReader != nil {
				_ = pr.currentReader.Close()
				pr.currentReader = nil
			}
		}
	}

	return batch, nil
}

// GetRowCount returns the total row count across all partitions
func (pr *PartitionedParquetReader) GetRowCount() int64 {
	return pr.totalRowCount
}

// GetPartitionColumns returns the partition column names
func (pr *PartitionedParquetReader) GetPartitionColumns() []string {
	return pr.partitionColumns
}

// GetPartitionFiles returns information about all partition files
func (pr *PartitionedParquetReader) GetPartitionFiles() []PartitionFile {
	return pr.partitionFiles
}

// Close closes the reader and releases resources
func (pr *PartitionedParquetReader) Close() error {
	if pr.currentReader != nil {
		return pr.currentReader.Close()
	}
	return nil
}