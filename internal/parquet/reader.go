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

// ParquetReader reads data from Parquet files using row-group streaming
// to avoid loading the entire file into memory
type ParquetReader struct {
	file         *os.File
	reader       *file.Reader
	arrowReader  *pqarrow.FileReader
	schema       *arrow.Schema
	columnNames  []string
	columnTypes  []string
	allColumns   []int // Explicit list of all leaf column indices; nil means "no columns" in pqarrow
	rowGroupIdx  int   // Current row group being processed
	numRowGroups int   // Total number of row groups
	currentBatch arrow.RecordBatch // Current record batch
	batchIdx     int   // Current row index within current batch
	totalRows    int64
	exhausted    bool  // True when all data has been read
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


	// Build explicit list of all leaf column indices. pqarrow.ReadRowGroups
	// treats nil as "no columns", not "all columns" — see ReadTable in
	// vendor/.../pqarrow/file_reader.go for the equivalent pattern.
	numCols := reader.MetaData().Schema.NumColumns()
	allColumns := make([]int, numCols)
	for i := range allColumns {
		allColumns[i] = i
	}

	return &ParquetReader{
		file:        f,
		reader:      reader,
		arrowReader: arrowReader,
		schema:      schema,
		columnNames: columnNames,
		columnTypes: columnTypes,
		allColumns:  allColumns,
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

// ReadBatch reads the next batch of rows using row-group streaming
// to avoid loading the entire file into memory
func (r *ParquetReader) ReadBatch(batchSize int) ([]map[string]any, error) {
	ctx := context.Background()

	// Initialize numRowGroups on first call
	if r.numRowGroups == 0 {
		r.numRowGroups = r.reader.NumRowGroups()
	}

	// Check if we've exhausted all data
	if r.exhausted {
		return nil, io.EOF
	}

	rows := make([]map[string]any, 0, batchSize)

	for len(rows) < batchSize {
		// If we don't have a current batch or we've consumed it, get the next one
		if r.currentBatch == nil || r.batchIdx >= int(r.currentBatch.NumRows()) {
			// Release previous batch if any
			if r.currentBatch != nil {
				r.currentBatch.Release()
				r.currentBatch = nil
			}

			// Move to next row group if needed
			if r.rowGroupIdx >= r.numRowGroups {
				r.exhausted = true
				break
			}

			// Read the next row group
			table, err := r.arrowReader.ReadRowGroups(ctx, r.allColumns, []int{r.rowGroupIdx})
			if err != nil {
				return nil, fmt.Errorf("failed to read row group %d: %w", r.rowGroupIdx, err)
			}
			r.rowGroupIdx++

			if table.NumRows() == 0 {
				table.Release()
				continue // Try next row group
			}

			// Create a record reader from the table
			chunkSize := int64(batchSize)
			if chunkSize <= 0 || chunkSize > table.NumRows() {
				chunkSize = table.NumRows()
			}
			tableReader := array.NewTableReader(table, chunkSize)

			if tableReader.Next() {
				r.currentBatch = tableReader.RecordBatch()
				r.currentBatch.Retain() // Keep the record alive after releasing tableReader
				r.batchIdx = 0
			}
			tableReader.Release()
			table.Release()

			if r.currentBatch == nil {
				continue // Try next row group
			}
		}

		// Extract rows from the current batch
		for len(rows) < batchSize && r.batchIdx < int(r.currentBatch.NumRows()) {
			row := make(map[string]any)

			for colIdx, colName := range r.columnNames {
				col := r.currentBatch.Column(colIdx)
				value := extractValue(col, r.batchIdx)
				row[colName] = value
			}

			rows = append(rows, row)
			r.batchIdx++
		}
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
