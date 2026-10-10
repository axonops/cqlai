package parquet

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet"
	"github.com/apache/arrow-go/v18/parquet/compress"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// ParquetCaptureWriter handles writing query results to Parquet format
type ParquetCaptureWriter struct {
	writer     io.Writer
	schema     *arrow.Schema
	builder    *array.RecordBuilder
	allocator  memory.Allocator
	chunkSize  int64
	rowCount   int64
	totalRows  int64
	props      *parquet.WriterProperties
	arrowProps pqarrow.ArrowWriterProperties
	// file writes each chunk out as it fills, as a row group of its own.
	// The chunks used to be kept until Close and written as one table, so an
	// export held every row it had read in memory until the end.
	file       *pqarrow.FileWriter
	typeMapper *TypeMapper
	isClosed   bool
	firstWrite bool
	outputPath string // for debugging/logging
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
	// Create output writer using the new cloud-aware function
	writer, err := CreateWriter(context.Background(), output)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
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
		parquet.WithDataPageSize(1024*1024),   // 1MB data pages
		parquet.WithMaxRowGroupLength(100000), // 100k rows per group
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
		parquet.WithDataPageSize(1024*1024),   // 1MB data pages
		parquet.WithMaxRowGroupLength(100000), // 100k rows per group
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
			// TODO: Fix underlying type conversion issues and then surface this error
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

	record := w.builder.NewRecordBatch()
	defer record.Release()
	if err := w.openFile(); err != nil {
		return err
	}
	if err := w.file.Write(record); err != nil {
		return fmt.Errorf("failed to write Parquet rows: %w", err)
	}

	// Reset the builder for the next chunk
	w.builder.Release()
	w.builder = array.NewRecordBuilder(w.allocator, w.schema)
	w.rowCount = 0

	logger.DebugfToFile("ParquetWriter", "Flushed chunk with %d total rows", w.totalRows)
	return nil
}

// openFile starts the Parquet file, once: the writer properties are final
// by the first chunk.
func (w *ParquetCaptureWriter) openFile() error {
	if w.file != nil {
		return nil
	}
	file, err := pqarrow.NewFileWriter(w.schema, w.writer, w.props, w.arrowProps)
	if err != nil {
		return fmt.Errorf("failed to start the Parquet file: %w", err)
	}
	w.file = file
	return nil
}

// Close writes what is left, the footer, and closes the file. A file with no
// rows still has its schema.
func (w *ParquetCaptureWriter) Close() error {
	if w.isClosed {
		return nil
	}
	w.isClosed = true

	err := w.flushChunk()
	if err == nil {
		err = w.openFile()
	}
	if w.file != nil {
		// Closes the file under it too: until the footer is written, it
		// cannot be read, so a failure here is a failure of the export.
		if cerr := w.file.Close(); err == nil && cerr != nil {
			err = fmt.Errorf("failed to finish the Parquet file: %w", cerr)
		}
	} else if closer, ok := w.writer.(io.Closer); ok && w.writer != os.Stdout {
		_ = closer.Close()
	}
	if w.builder != nil {
		w.builder.Release()
		w.builder = nil
	}
	logger.DebugfToFile("ParquetWriter", "Wrote Parquet file with %d total rows", w.totalRows)
	return err
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

// CreateWriter creates an appropriate writer based on the output path
// It supports local files and stdout
func CreateWriter(ctx context.Context, output string) (io.WriteCloser, error) {
	// Check for special outputs
	if output == "" || output == "-" || output == "STDOUT" {
		return nopCloser{os.Stdout}, nil
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
