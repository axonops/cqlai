package parquet

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/parquet/compress"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// PartitionedParquetWriter writes Parquet files organized in partitioned directories
type PartitionedParquetWriter struct {
	baseDir       string
	partitionCols []string
	partitionIdx  map[string]int // Column name to index in data
	writers       map[string]*partitionWriter
	writerOrder   *list.List // LRU tracking
	writerElements map[string]*list.Element
	schema        *arrow.Schema
	options       WriterOptions
	maxOpenFiles  int
	maxFileSize   int64
	mu            sync.Mutex
	isClosed      bool
	// Store column info for creating new partition files
	columnNames     []string
	columnTypes     []string
	columnTypeInfos []gocql.TypeInfo
}

// partitionWriter wraps a ParquetCaptureWriter for a specific partition
type partitionWriter struct {
	writer      *ParquetCaptureWriter
	partitionKey string
	dirPath     string
	filePath    string
	rowCount    int64
	fileSize    int64
	partNum     int
	maxFileSize int64
}

// PartitionedWriterOptions extends WriterOptions with partitioning config
type PartitionedWriterOptions struct {
	WriterOptions
	PartitionColumns []string
	MaxOpenFiles     int
	MaxFileSize      int64 // Maximum size per partition file in bytes
}

// ParseCompression parses compression string to codec
func ParseCompression(compression string) compress.Compression {
	switch strings.ToUpper(compression) {
	case "SNAPPY":
		return compress.Codecs.Snappy
	case "GZIP":
		return compress.Codecs.Gzip
	case "LZ4":
		return compress.Codecs.Lz4
	case "ZSTD":
		return compress.Codecs.Zstd
	case "NONE", "":
		return compress.Codecs.Uncompressed
	default:
		return compress.Codecs.Snappy // Default to Snappy
	}
}

// DefaultPartitionedOptions returns default partitioned writer options
func DefaultPartitionedOptions() PartitionedWriterOptions {
	return PartitionedWriterOptions{
		WriterOptions:    DefaultWriterOptions(),
		PartitionColumns: []string{},
		MaxOpenFiles:     10,
		MaxFileSize:      100 * 1024 * 1024, // 100MB default
	}
}

// NewPartitionedParquetWriterWithTypeInfo creates a new partitioned Parquet writer with detailed type info
func NewPartitionedParquetWriterWithTypeInfo(baseDir string, columnNames []string, columnTypes []string, columnTypeInfos []gocql.TypeInfo, options PartitionedWriterOptions) (*PartitionedParquetWriter, error) {
	// Validate partition columns - allow virtual columns (with dots)
	partitionIdx := make(map[string]int)
	for _, partCol := range options.PartitionColumns {
		// Check if it's a virtual column (contains a dot)
		if strings.Contains(partCol, ".") {
			// Virtual column - extract base column name
			baseCol := strings.SplitN(partCol, ".", 2)[0]
			// Check if base column exists
			found := false
			for i, colName := range columnNames {
				if colName == baseCol {
					partitionIdx[baseCol] = i
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("base column '%s' for virtual partition column '%s' not found in schema", baseCol, partCol)
			}
		} else {
			// Regular column
			found := false
			for i, colName := range columnNames {
				if colName == partCol {
					partitionIdx[partCol] = i
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("partition column '%s' not found in schema", partCol)
			}
		}
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Store the type infos for later use when creating partition files
	pw := &PartitionedParquetWriter{
		baseDir:        baseDir,
		partitionCols:  options.PartitionColumns,
		partitionIdx:   partitionIdx,
		writers:        make(map[string]*partitionWriter),
		writerOrder:    list.New(),
		writerElements: make(map[string]*list.Element),
		options:        options.WriterOptions,
		maxOpenFiles:   options.MaxOpenFiles,
		maxFileSize:    options.MaxFileSize,
		columnNames:    columnNames,
		columnTypes:    columnTypes,
		columnTypeInfos: columnTypeInfos,
	}

	if pw.maxOpenFiles <= 0 {
		pw.maxOpenFiles = 10
	}
	if pw.maxFileSize <= 0 {
		pw.maxFileSize = 100 * 1024 * 1024
	}

	return pw, nil
}

// NewPartitionedParquetWriter creates a new partitioned Parquet writer
func NewPartitionedParquetWriter(baseDir string, columnNames []string, columnTypes []string, options PartitionedWriterOptions) (*PartitionedParquetWriter, error) {
	// Validate partition columns - allow virtual columns (with dots)
	partitionIdx := make(map[string]int)
	for _, partCol := range options.PartitionColumns {
		// Check if it's a virtual column (contains a dot)
		if strings.Contains(partCol, ".") {
			// Virtual column - extract base column name
			baseCol := strings.SplitN(partCol, ".", 2)[0]
			// Check if base column exists
			found := false
			for i, colName := range columnNames {
				if colName == baseCol {
					partitionIdx[baseCol] = i
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("base column '%s' for virtual partition column '%s' not found in schema", baseCol, partCol)
			}
		} else {
			// Regular column
			found := false
			for i, colName := range columnNames {
				if colName == partCol {
					partitionIdx[partCol] = i
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("partition column '%s' not found in schema", partCol)
			}
		}
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create Arrow schema (will be reused for all partition files)
	typeMapper := NewTypeMapper()
	schema, err := typeMapper.CreateArrowSchema(columnNames, columnTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to create Arrow schema: %w", err)
	}

	return &PartitionedParquetWriter{
		baseDir:        baseDir,
		partitionCols:  options.PartitionColumns,
		partitionIdx:   partitionIdx,
		writers:        make(map[string]*partitionWriter),
		writerOrder:    list.New(),
		writerElements: make(map[string]*list.Element),
		schema:         schema,
		options:        options.WriterOptions,
		maxOpenFiles:   options.MaxOpenFiles,
		columnNames:    columnNames,
		columnTypes:    columnTypes,
	}, nil
}

// WriteRows writes rows to appropriate partitions
func (pw *PartitionedParquetWriter) WriteRows(rows []map[string]interface{}) error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if pw.isClosed {
		return fmt.Errorf("writer is closed")
	}

	// Group rows by partition
	partitions := pw.groupByPartition(rows)

	// Write to each partition
	for partitionKey, partitionRows := range partitions {
		// Filter out virtual partition columns from the data
		cleanRows := pw.removeVirtualColumns(partitionRows)
		// Debug: log if virtual columns were removed
		if len(partitionRows) > 0 && len(cleanRows) > 0 {
			originalKeys := make([]string, 0)
			cleanKeys := make([]string, 0)
			for k := range partitionRows[0] {
				originalKeys = append(originalKeys, k)
			}
			for k := range cleanRows[0] {
				cleanKeys = append(cleanKeys, k)
			}
			if len(originalKeys) != len(cleanKeys) {
				logger.DebugfToFile("PartitionedWriter", "Removed virtual columns: original=%v, clean=%v", originalKeys, cleanKeys)
			}
		}

		columnNames := make([]string, 0)
		if len(cleanRows) > 0 {
			for key := range cleanRows[0] {
				columnNames = append(columnNames, key)
			}
		}

		writer, err := pw.getOrCreateWriter(partitionKey, columnNames)
		if err != nil {
			return fmt.Errorf("failed to get writer for partition %s: %w", partitionKey, err)
		}

		// Write rows to partition (without virtual columns)
		if err := writer.writer.WriteRows(cleanRows); err != nil {
			return fmt.Errorf("failed to write to partition %s: %w", partitionKey, err)
		}

		writer.rowCount += int64(len(partitionRows))

		// Check if we need to rotate the file
		if writer.maxFileSize > 0 && writer.fileSize > writer.maxFileSize {
			if err := pw.rotatePartitionFile(partitionKey); err != nil {
				return fmt.Errorf("failed to rotate partition file: %w", err)
			}
		}
	}

	return nil
}

// removeVirtualColumns removes virtual partition columns from rows
func (pw *PartitionedParquetWriter) removeVirtualColumns(rows []map[string]interface{}) []map[string]interface{} {
	if len(rows) == 0 {
		return rows
	}

	// Check if any partition columns are virtual (contain dots)
	hasVirtualCols := false
	virtualCols := make(map[string]bool)
	for _, col := range pw.partitionCols {
		if strings.Contains(col, ".") {
			hasVirtualCols = true
			virtualCols[col] = true
		}
	}

	// If no virtual columns, return rows as-is
	if !hasVirtualCols {
		return rows
	}

	// Create clean rows without virtual columns
	cleanRows := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		cleanRow := make(map[string]interface{})
		for key, value := range row {
			// Skip virtual columns
			if !virtualCols[key] {
				cleanRow[key] = value
			}
		}
		cleanRows[i] = cleanRow
	}

	return cleanRows
}

// groupByPartition groups rows by their partition values
func (pw *PartitionedParquetWriter) groupByPartition(rows []map[string]interface{}) map[string][]map[string]interface{} {
	partitions := make(map[string][]map[string]interface{})

	for _, row := range rows {
		partitionKey := pw.buildPartitionKey(row)
		partitions[partitionKey] = append(partitions[partitionKey], row)
	}

	return partitions
}

// buildPartitionKey builds the partition key from row values
func (pw *PartitionedParquetWriter) buildPartitionKey(row map[string]interface{}) string {
	if len(pw.partitionCols) == 0 {
		return "default"
	}

	parts := make([]string, len(pw.partitionCols))
	for i, col := range pw.partitionCols {
		value := pw.extractPartitionValue(row, col)
		formattedValue := formatPartitionValue(value)
		parts[i] = fmt.Sprintf("%s=%s", col, formattedValue)
	}

	return strings.Join(parts, "/")
}

// extractPartitionValue extracts the partition value, handling special cases like TimeUUID
func (pw *PartitionedParquetWriter) extractPartitionValue(row map[string]interface{}, partitionCol string) interface{} {
	// Check for virtual partition columns derived from TimeUUID
	// Format: column_name.year, column_name.month, column_name.day
	if strings.Contains(partitionCol, ".") {
		parts := strings.SplitN(partitionCol, ".", 2)
		sourceCol := parts[0]
		datePart := parts[1]

		// Get the source column value
		sourceValue := row[sourceCol]
		if sourceValue == nil {
			return nil
		}

		// Extract timestamp from TimeUUID or use timestamp directly
		var timestamp time.Time
		switch v := sourceValue.(type) {
		case gocql.UUID:
			// Extract timestamp from TimeUUID (Type 1 UUID)
			timestamp = v.Time()
		case string:
			// Try to parse as UUID string
			if uuid, err := gocql.ParseUUID(v); err == nil {
				timestamp = uuid.Time()
			} else {
				// Not a valid UUID, return nil
				return nil
			}
		case time.Time:
			timestamp = v
		default:
			// Cannot extract timestamp
			return nil
		}

		// Extract the requested date part
		switch datePart {
		case "year":
			return timestamp.Year()
		case "month":
			return int(timestamp.Month())
		case "day":
			return timestamp.Day()
		case "hour":
			return timestamp.Hour()
		case "date":
			return timestamp.Format("2006-01-02")
		default:
			return nil
		}
	}

	// Check if this is a date-related partition column but the actual column might be TimeUUID
	// Try common patterns: if partitioning by "year", look for "timestamp" or "created_at" TimeUUID
	if partitionCol == "year" || partitionCol == "month" || partitionCol == "day" {
		// Look for TimeUUID columns that might contain the time data
		for colName, colValue := range row {
			// Check common time-related column names
			if strings.Contains(strings.ToLower(colName), "time") ||
			   strings.Contains(strings.ToLower(colName), "created") ||
			   strings.Contains(strings.ToLower(colName), "updated") ||
			   strings.Contains(strings.ToLower(colName), "date") {

				switch v := colValue.(type) {
				case gocql.UUID:
					// Extract timestamp from TimeUUID
					timestamp := v.Time()
					switch partitionCol {
					case "year":
						return timestamp.Year()
					case "month":
						return int(timestamp.Month())
					case "day":
						return timestamp.Day()
					}
				case string:
					// Try to parse as UUID string
					if uuid, err := gocql.ParseUUID(v); err == nil {
						timestamp := uuid.Time()
						switch partitionCol {
						case "year":
							return timestamp.Year()
						case "month":
							return int(timestamp.Month())
						case "day":
							return timestamp.Day()
						}
					}
				}
			}
		}
	}

	// Regular column value
	return row[partitionCol]
}

// formatPartitionValue formats a value for use in partition path
func formatPartitionValue(value interface{}) string {
	if value == nil {
		return "__NULL__"
	}

	switch v := value.(type) {
	case string:
		// Escape special characters
		escaped := strings.ReplaceAll(v, "/", "__SLASH__")
		escaped = strings.ReplaceAll(escaped, "=", "__EQ__")
		return escaped
	default:
		return fmt.Sprintf("%v", v)
	}
}

// buildPartitionPath builds the full directory path for a partition
func (pw *PartitionedParquetWriter) buildPartitionPath(partitionKey string) string {
	if partitionKey == "default" {
		return pw.baseDir
	}
	return filepath.Join(pw.baseDir, partitionKey)
}

// getOrCreateWriter gets an existing writer or creates a new one
func (pw *PartitionedParquetWriter) getOrCreateWriter(partitionKey string, columnNames []string) (*partitionWriter, error) {
	// Check if writer exists
	if writer, exists := pw.writers[partitionKey]; exists {
		// Move to front of LRU
		pw.writerOrder.MoveToFront(pw.writerElements[partitionKey])
		return writer, nil
	}

	// Check if we need to close an old writer
	if len(pw.writers) >= pw.maxOpenFiles {
		// Close least recently used writer
		oldest := pw.writerOrder.Back()
		if oldest != nil {
			oldKey := oldest.Value.(string)
			if err := pw.closeWriter(oldKey); err != nil {
				logger.DebugfToFile("PartitionedWriter", "Failed to close LRU writer: %v", err)
			}
		}
	}

	// Create new writer
	dirPath := pw.buildPartitionPath(partitionKey)
	if err := os.MkdirAll(dirPath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create partition directory: %w", err)
	}

	// Generate filename
	fileName := fmt.Sprintf("part-%05d.parquet", 0)
	filePath := filepath.Join(dirPath, fileName)

	// Create Parquet writer with TypeInfo if available
	var parquetWriter *ParquetCaptureWriter
	var err error

	switch {
	case len(pw.columnTypeInfos) > 0:
		// Use the TypeInfo-aware constructor
		// Use columnNames parameter (which has virtual columns removed), not pw.columnNames
		// Need to get the corresponding types and typeinfos for the filtered columns
		filteredTypes := make([]string, len(columnNames))
		filteredTypeInfos := make([]gocql.TypeInfo, len(columnNames))
		for i, name := range columnNames {
			// Find the type and typeinfo for this column
			for j, origName := range pw.columnNames {
				if origName == name {
					if j < len(pw.columnTypes) {
						filteredTypes[i] = pw.columnTypes[j]
					}
					if j < len(pw.columnTypeInfos) {
						filteredTypeInfos[i] = pw.columnTypeInfos[j]
					}
					break
				}
			}
		}
		parquetWriter, err = NewParquetCaptureWriterWithTypeInfo(filePath, columnNames, filteredTypes, filteredTypeInfos, pw.options)
	case len(pw.columnTypes) > 0:
		// Fall back to standard constructor
		// Use columnNames parameter (which has virtual columns removed), not pw.columnNames
		// Need to get the corresponding column types for the filtered columns
		filteredTypes := make([]string, len(columnNames))
		for i, name := range columnNames {
			// Find the type for this column from pw.columnNames/pw.columnTypes
			for j, origName := range pw.columnNames {
				if origName == name && j < len(pw.columnTypes) {
					filteredTypes[i] = pw.columnTypes[j]
					break
				}
			}
		}
		parquetWriter, err = NewParquetCaptureWriter(filePath, columnNames, filteredTypes, pw.options)
	default:
		// Get column types from schema (fallback)
		columnTypes := make([]string, len(columnNames))
		for i, name := range columnNames {
			for j, field := range pw.schema.Fields() {
				if field.Name == name {
					columnTypes[i] = pw.schema.Field(j).Type.String()
					break
				}
			}
		}
		parquetWriter, err = NewParquetCaptureWriter(filePath, columnNames, columnTypes, pw.options)
	}

	if err != nil {
		return nil, err
	}

	writer := &partitionWriter{
		writer:       parquetWriter,
		partitionKey: partitionKey,
		dirPath:      dirPath,
		filePath:     filePath,
		rowCount:     0,
		fileSize:     0,
		partNum:      0,
		maxFileSize:  pw.maxFileSize,
	}

	// Add to cache
	pw.writers[partitionKey] = writer
	elem := pw.writerOrder.PushFront(partitionKey)
	pw.writerElements[partitionKey] = elem

	return writer, nil
}

// rotatePartitionFile closes current file and creates a new one for the partition
func (pw *PartitionedParquetWriter) rotatePartitionFile(partitionKey string) error {
	writer, exists := pw.writers[partitionKey]
	if !exists {
		return fmt.Errorf("no writer for partition %s", partitionKey)
	}

	// Close current writer
	if err := writer.writer.Close(); err != nil {
		return fmt.Errorf("failed to close current writer: %w", err)
	}

	// Increment part number
	writer.partNum++

	// Generate new filename
	fileName := fmt.Sprintf("part-%05d.parquet", writer.partNum)
	filePath := filepath.Join(writer.dirPath, fileName)

	// Get column names and types from schema
	columnNames := make([]string, 0, len(pw.schema.Fields()))
	columnTypes := make([]string, 0, len(pw.schema.Fields()))
	for _, field := range pw.schema.Fields() {
		columnNames = append(columnNames, field.Name)
		columnTypes = append(columnTypes, field.Type.String())
	}

	// Create new Parquet writer
	parquetWriter, err := NewParquetCaptureWriter(filePath, columnNames, columnTypes, pw.options)
	if err != nil {
		return err
	}

	writer.writer = parquetWriter
	writer.filePath = filePath
	writer.fileSize = 0

	return nil
}

// closeWriter closes a specific partition writer
func (pw *PartitionedParquetWriter) closeWriter(partitionKey string) error {
	writer, exists := pw.writers[partitionKey]
	if !exists {
		return nil
	}

	if err := writer.writer.Close(); err != nil {
		return err
	}

	// Remove from cache
	delete(pw.writers, partitionKey)
	if elem, exists := pw.writerElements[partitionKey]; exists {
		pw.writerOrder.Remove(elem)
		delete(pw.writerElements, partitionKey)
	}

	return nil
}

// Flush flushes all open writers
func (pw *PartitionedParquetWriter) Flush() error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	for _, writer := range pw.writers {
		if err := writer.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush partition %s: %w", writer.partitionKey, err)
		}
	}

	return nil
}

// Close closes all partition writers
func (pw *PartitionedParquetWriter) Close() error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if pw.isClosed {
		return nil
	}

	var firstErr error
	for partitionKey := range pw.writers {
		if err := pw.closeWriter(partitionKey); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	pw.isClosed = true
	return firstErr
}

// GetPartitionInfo returns information about written partitions
func (pw *PartitionedParquetWriter) GetPartitionInfo() map[string]PartitionInfo {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	info := make(map[string]PartitionInfo)
	for key, writer := range pw.writers {
		info[key] = PartitionInfo{
			Path:     writer.dirPath,
			RowCount: writer.rowCount,
			FileSize: writer.fileSize,
			NumFiles: writer.partNum + 1,
		}
	}
	return info
}

// PartitionInfo contains metadata about a partition
type PartitionInfo struct {
	Path     string
	RowCount int64
	FileSize int64
	NumFiles int
}