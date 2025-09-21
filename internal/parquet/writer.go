package parquet

import (
	"fmt"
	"io"
	"os"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/axonops/cqlai/internal/logger"
)

// ParquetCaptureWriter handles writing query results to Parquet format
type ParquetCaptureWriter struct {
	writer        io.Writer
	schema        *arrow.Schema
	builder       *array.RecordBuilder
	allocator     memory.Allocator
	chunkSize     int64
	rowCount      int64
	totalRows     int64
	props         *parquet.WriterProperties
	arrowProps    pqarrow.ArrowWriterProperties
	records       []arrow.RecordBatch
	typeMapper    *TypeMapper
	isClosed      bool
	firstWrite    bool
	outputPath    string // for debugging/logging
}

// WriterOptions configures the Parquet writer
type WriterOptions struct {
	ChunkSize   int64
	Compression compress.Compression
	// Additional options can be added here
}

// DefaultWriterOptions returns default writer options
func DefaultWriterOptions() WriterOptions {
	return WriterOptions{
		ChunkSize:   10000, // Default 10k rows per chunk
		Compression: compress.Codecs.Snappy,
	}
}

// NewParquetCaptureWriter creates a new Parquet capture writer
func NewParquetCaptureWriter(output string, columnNames []string, columnTypes []string, options WriterOptions) (*ParquetCaptureWriter, error) {
	// Create output writer
	var writer io.Writer
	var err error

	// Check if output is a file path or stdout
	if output == "" || output == "-" || output == "STDOUT" {
		writer = os.Stdout
	} else {
		file, err := os.Create(output) // #nosec G304 - output path comes from user input but is validated
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		writer = file
	}

	// Create type mapper and schema
	typeMapper := NewTypeMapper()
	schema, err := typeMapper.CreateArrowSchema(columnNames, columnTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to create Arrow schema: %w", err)
	}

	// Create allocator
	allocator := memory.NewGoAllocator()

	// Create record builder
	builder := array.NewRecordBuilder(allocator, schema)

	// Configure Parquet writer properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(options.Compression),
		parquet.WithDictionaryDefault(false),
		parquet.WithDataPageSize(1024*1024),      // 1MB data pages
		parquet.WithMaxRowGroupLength(100000),     // 100k rows per group
		parquet.WithCreatedBy("CQLAI Parquet Writer"),
	)

	// Configure Arrow writer properties
	arrowProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	return &ParquetCaptureWriter{
		writer:     writer,
		schema:     schema,
		builder:    builder,
		allocator:  allocator,
		chunkSize:  options.ChunkSize,
		props:      props,
		arrowProps: arrowProps,
		records:    make([]arrow.RecordBatch, 0),
		typeMapper: typeMapper,
		firstWrite: true,
		outputPath: output,
	}, nil
}

// NewParquetCaptureWriterWithTypeInfo creates a new Parquet capture writer with TypeInfo support
// This allows proper handling of complex types like UDTs with native STRUCT representation
func NewParquetCaptureWriterWithTypeInfo(output string, columnNames []string, columnTypes []string, columnTypeInfos []gocql.TypeInfo, options WriterOptions) (*ParquetCaptureWriter, error) {
	// Create output writer
	var writer io.Writer
	var err error

	// Check if output is a file path or stdout
	if output == "" || output == "-" || output == "STDOUT" {
		writer = os.Stdout
	} else {
		file, err := os.Create(output) // #nosec G304 - output path comes from user input but is validated
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		writer = file
	}

	// Create type mapper and schema with TypeInfo support
	typeMapper := NewTypeMapper()
	schema, err := typeMapper.CreateArrowSchemaWithTypeInfo(columnNames, columnTypes, columnTypeInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to create Arrow schema: %w", err)
	}

	// Create allocator
	allocator := memory.NewGoAllocator()

	// Create record builder
	builder := array.NewRecordBuilder(allocator, schema)

	// Configure Parquet writer properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(options.Compression),
		parquet.WithDictionaryDefault(false),
		parquet.WithDataPageSize(1024*1024),      // 1MB data pages
		parquet.WithMaxRowGroupLength(100000),     // 100k rows per group
		parquet.WithCreatedBy("CQLAI Parquet Writer"),
	)

	// Configure Arrow writer properties
	arrowProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	return &ParquetCaptureWriter{
		writer:     writer,
		schema:     schema,
		builder:    builder,
		allocator:  allocator,
		chunkSize:  options.ChunkSize,
		props:      props,
		arrowProps: arrowProps,
		records:    make([]arrow.RecordBatch, 0),
		typeMapper: typeMapper,
		firstWrite: true,
		outputPath: output,
	}, nil
}

// WriteHeader is a no-op for Parquet (schema is written with the data)
func (w *ParquetCaptureWriter) WriteHeader() error {
	// Parquet files include schema metadata, no separate header needed
	return nil
}

// WriteRow writes a single row to the Parquet file
func (w *ParquetCaptureWriter) WriteRow(row map[string]interface{}) error {
	if w.isClosed {
		return fmt.Errorf("writer is closed")
	}

	// Append values to builders for each column
	for i := 0; i < w.schema.NumFields(); i++ {
		field := w.schema.Field(i)
		value := row[field.Name]

		// Get the specific builder for this column
		columnBuilder := w.builder.Field(i)

		// Append the value using the type mapper
		logger.DebugfToFile("ParquetWriter", "Appending column %s: value type=%T, arrow type=%v", field.Name, value, field.Type)
		if err := w.typeMapper.AppendValueToBuilder(columnBuilder, value, field.Type); err != nil {
			logger.DebugfToFile("ParquetWriter", "Error appending value for column %s: %v", field.Name, err)
			// Continue with other columns even if one fails
		}
	}

	w.rowCount++
	w.totalRows++

	// Check if we need to flush the chunk
	if w.rowCount >= w.chunkSize {
		return w.flushChunk()
	}

	return nil
}

// WriteRows writes multiple rows to the Parquet file
func (w *ParquetCaptureWriter) WriteRows(rows []map[string]interface{}) error {
	for _, row := range rows {
		if err := w.WriteRow(row); err != nil {
			return err
		}
	}
	return nil
}

// WriteRawRows writes rows with raw values (already typed, not string)
func (w *ParquetCaptureWriter) WriteRawRows(headers []string, rows [][]interface{}) error {
	if w.isClosed {
		return fmt.Errorf("writer is closed")
	}

	for _, row := range rows {
		// Convert row array to map
		rowMap := make(map[string]interface{})
		for i, header := range headers {
			if i < len(row) {
				rowMap[header] = row[i]
			}
		}
		if err := w.WriteRow(rowMap); err != nil {
			return err
		}
	}

	return nil
}

// WriteStringRows writes rows where all values are strings (need conversion)
func (w *ParquetCaptureWriter) WriteStringRows(headers []string, rows [][]string) error {
	if w.isClosed {
		return fmt.Errorf("writer is closed")
	}

	// Convert string rows to typed rows
	for _, row := range rows {
		rowMap := make(map[string]interface{})
		for i, header := range headers {
			if i < len(row) {
				// For now, keep as string - proper type conversion would happen here
				// based on the schema field type
				rowMap[header] = row[i]
			}
		}
		if err := w.WriteRow(rowMap); err != nil {
			return err
		}
	}

	return nil
}

// flushChunk writes the current chunk to the Parquet file
func (w *ParquetCaptureWriter) flushChunk() error {
	if w.rowCount == 0 {
		return nil
	}

	// Create a record batch from the current builder state
	record := w.builder.NewRecordBatch()
	// Don't defer release here - the record is stored and will be released in Close()

	// Store the record for final write
	// The record is retained for writing later
	w.records = append(w.records, record)

	// Reset the builder for the next chunk
	w.builder = array.NewRecordBuilder(w.allocator, w.schema)
	w.rowCount = 0

	logger.DebugfToFile("ParquetWriter", "Flushed chunk with %d total rows", w.totalRows)

	return nil
}

// Close finalizes the Parquet file
func (w *ParquetCaptureWriter) Close() error {
	if w.isClosed {
		return nil
	}

	// Flush any remaining data
	if err := w.flushChunk(); err != nil {
		return fmt.Errorf("failed to flush final chunk: %w", err)
	}

	// Only write if we have data
	if len(w.records) > 0 || (w.builder != nil && w.builder.Field(0).Len() > 0) {
		// If there's data in the builder but not yet in records, create final record
		if w.builder != nil && w.builder.Field(0).Len() > 0 {
			record := w.builder.NewRecordBatch()
			w.records = append(w.records, record)
		}

		// Create table from all records
		if len(w.records) > 0 {
			table := array.NewTableFromRecords(w.schema, w.records)
			defer table.Release()

			// Write the entire table to Parquet
			if err := pqarrow.WriteTable(table, w.writer, w.chunkSize, w.props, w.arrowProps); err != nil {
				return fmt.Errorf("failed to write Parquet table: %w", err)
			}

			logger.DebugfToFile("ParquetWriter", "Wrote Parquet file with %d total rows", w.totalRows)
		}
	}

	// Release all stored records - they are retained by the table
	for _, record := range w.records {
		if record != nil {
			record.Release()
		}
	}
	w.records = nil

	// Release the builder
	if w.builder != nil {
		w.builder.Release()
		w.builder = nil
	}

	// Close the underlying writer if it's a file (but not stdout)
	// Note: pqarrow.WriteTable may already close the writer in some cases,
	// so we check if it's still valid
	if !w.isClosed {
		if closer, ok := w.writer.(io.Closer); ok && w.writer != os.Stdout {
			// Try to close but don't fail if already closed
			_ = closer.Close()
		}
	}

	w.isClosed = true
	return nil
}

// GetRowCount returns the total number of rows written
func (w *ParquetCaptureWriter) GetRowCount() int64 {
	return w.totalRows
}

// IsStreaming returns true if the writer supports streaming
func (w *ParquetCaptureWriter) IsStreaming() bool {
	return true
}

// SetCompression sets the compression codec
func (w *ParquetCaptureWriter) SetCompression(compression string) error {
	if w.totalRows > 0 {
		return fmt.Errorf("cannot change compression after writing has started")
	}

	var codec compress.Compression
	switch compression {
	case "SNAPPY", "snappy":
		codec = compress.Codecs.Snappy
	case "GZIP", "gzip":
		codec = compress.Codecs.Gzip
	case "LZ4", "lz4":
		codec = compress.Codecs.Lz4
	case "ZSTD", "zstd":
		codec = compress.Codecs.Zstd
	case "NONE", "none", "":
		codec = compress.Codecs.Uncompressed
	default:
		return fmt.Errorf("unsupported compression: %s", compression)
	}

	// Recreate properties with new compression
	w.props = parquet.NewWriterProperties(
		parquet.WithCompression(codec),
		parquet.WithDictionaryDefault(false),
		parquet.WithDataPageSize(1024*1024),
		parquet.WithMaxRowGroupLength(100000),
		parquet.WithCreatedBy("CQLAI Parquet Writer"),
	)

	return nil
}

// Flush forces a write of any buffered data (creates a new row group)
func (w *ParquetCaptureWriter) Flush() error {
	return w.flushChunk()
}