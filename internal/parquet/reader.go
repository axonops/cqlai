package parquet

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

// ParquetReader reads data from Parquet files
type ParquetReader struct {
	file         *os.File
	reader       *file.Reader
	arrowReader  *pqarrow.FileReader
	schema       *arrow.Schema
	columnNames  []string
	columnTypes  []string
	rowGroupIdx  int
	currentBatch arrow.RecordBatch
	batchIdx     int
	totalRows    int64
}

// NewParquetReader creates a new Parquet reader
func NewParquetReader(filename string) (*ParquetReader, error) {
	// Open the file
	f, err := os.Open(filename) // #nosec G304 - filename comes from user input but is validated
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Create Parquet file reader
	reader, err := file.NewParquetReader(f)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}

	// Create Arrow reader for easier data access
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		_ = reader.Close()
		_ = f.Close()
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	// Get schema
	schema, err := arrowReader.Schema()
	if err != nil {
		_ = reader.Close()
		_ = f.Close()
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	// Extract column names and types
	columnNames := make([]string, schema.NumFields())
	columnTypes := make([]string, schema.NumFields())

	mapper := NewTypeMapper()
	for i := 0; i < schema.NumFields(); i++ {
		field := schema.Field(i)
		columnNames[i] = field.Name

		// Map Arrow type back to Cassandra type
		cassandraType := mapper.ArrowToCassandraType(field.Type)
		columnTypes[i] = cassandraType
	}


	return &ParquetReader{
		file:        f,
		reader:      reader,
		arrowReader: arrowReader,
		schema:      schema,
		columnNames: columnNames,
		columnTypes: columnTypes,
		totalRows:   reader.NumRows(),
	}, nil
}

// GetSchema returns the column names and types
func (r *ParquetReader) GetSchema() ([]string, []string) {
	return r.columnNames, r.columnTypes
}

// GetColumnNames returns just the column names
func (r *ParquetReader) GetColumnNames() []string {
	return r.columnNames
}

// GetRowCount returns the total number of rows in the file
func (r *ParquetReader) GetRowCount() int64 {
	return r.totalRows
}

// NumRowGroups returns the number of row groups in the file
func (r *ParquetReader) NumRowGroups() int {
	return r.reader.NumRowGroups()
}

// ReadBatch reads the next batch of rows
func (r *ParquetReader) ReadBatch(batchSize int) ([]map[string]any, error) {
	ctx := context.Background()

	// If we don't have a current batch or we've consumed it, get the next batch
	if r.currentBatch == nil || r.batchIdx >= int(r.currentBatch.NumRows()) {
		// For the first batch, read the entire table
		// Note: Row group reading seems to have issues with nested types
		if r.rowGroupIdx == 0 {
			table, err := r.arrowReader.ReadTable(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to read table: %w", err)
			}
			r.rowGroupIdx++ // Mark that we've read the table

			if table.NumRows() == 0 {
				table.Release()
				return nil, io.EOF
			}

			// Create a record batch from the table
			chunkSize := int64(batchSize)
			if chunkSize <= 0 || chunkSize > table.NumRows() {
				chunkSize = table.NumRows()
			}
			reader := array.NewTableReader(table, chunkSize)

			if reader.Next() {
				r.currentBatch = reader.RecordBatch()
				r.currentBatch.Retain() // Keep the record alive
				r.batchIdx = 0
			} else {
				table.Release()
				reader.Release()
				return nil, fmt.Errorf("failed to read record batch from table with %d rows", table.NumRows())
			}
			reader.Release()
			table.Release()
		} else {
			return nil, io.EOF
		}
	}

	// Extract rows from the current batch
	rows := make([]map[string]any, 0, batchSize)

	for i := 0; i < batchSize && r.batchIdx < int(r.currentBatch.NumRows()); i++ {
		row := make(map[string]any)

		for colIdx, colName := range r.columnNames {
			col := r.currentBatch.Column(colIdx)
			value := extractValue(col, r.batchIdx)
			row[colName] = value
		}

		rows = append(rows, row)
		r.batchIdx++
	}

	if len(rows) == 0 {
		return nil, io.EOF
	}

	return rows, nil
}

// ReadAll reads all rows from the file
func (r *ParquetReader) ReadAll() ([]map[string]any, error) {
	ctx := context.Background()

	// Read the entire table at once
	table, err := r.arrowReader.ReadTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	rows := make([]map[string]any, 0, table.NumRows())

	// Create a table reader to iterate through records
	reader := array.NewTableReader(table, 10000) // Process in chunks of 10000
	defer reader.Release()

	for reader.Next() {
		record := reader.RecordBatch()

		for rowIdx := 0; rowIdx < int(record.NumRows()); rowIdx++ {
			row := make(map[string]any)

			for colIdx, colName := range r.columnNames {
				col := record.Column(colIdx)
				value := extractValue(col, rowIdx)
				row[colName] = value
			}

			rows = append(rows, row)
		}
	}

	return rows, nil
}

// Close closes the Parquet reader
func (r *ParquetReader) Close() error {
	if r.currentBatch != nil {
		r.currentBatch.Release()
	}
	if r.reader != nil {
		_ = r.reader.Close()
	}
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// CreateReader creates an appropriate reader based on the input path
// It supports local files and stdin
func CreateReader(ctx context.Context, input string) (io.ReadCloser, error) {
	// Check for special inputs
	if input == "" || input == "-" || input == "STDIN" {
		return io.NopCloser(os.Stdin), nil
	}

	// Create local file reader
	return os.Open(input) // #nosec G304 - input path is validated by caller
}
