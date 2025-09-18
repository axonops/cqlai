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
	currentBatch arrow.Record
	batchIdx     int
	totalRows    int64
}

// NewParquetReader creates a new Parquet reader
func NewParquetReader(filename string) (*ParquetReader, error) {
	// Open the file
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Create Parquet file reader
	reader, err := file.NewParquetReader(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}

	// Create Arrow reader for easier data access
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		reader.Close()
		f.Close()
		return nil, fmt.Errorf("failed to create arrow reader: %w", err)
	}

	// Get schema
	schema, err := arrowReader.Schema()
	if err != nil {
		reader.Close()
		f.Close()
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

	// If we don't have a current batch or we've consumed it, get the next row group
	if r.currentBatch == nil || r.batchIdx >= int(r.currentBatch.NumRows()) {
		if r.rowGroupIdx >= r.reader.NumRowGroups() {
			return nil, io.EOF
		}

		// Read the next row group
		rowGroupReader := r.arrowReader.RowGroup(r.rowGroupIdx)
		r.rowGroupIdx++

		// Read all columns from this row group
		// Try passing explicit column indices instead of nil
		columnIndices := make([]int, len(r.columnNames))
		for i := range columnIndices {
			columnIndices[i] = i
		}
		table, err := rowGroupReader.ReadTable(ctx, columnIndices)
		if err != nil {
			return nil, fmt.Errorf("failed to read row group: %w", err)
		}

		// Get the first chunk as a record batch
		if table.NumRows() > 0 {
			// Create a record batch from the table
			// Use a reasonable chunk size - if batchSize is 0, use table size
			chunkSize := int64(batchSize)
			if chunkSize <= 0 {
				chunkSize = table.NumRows()
			}
			reader := array.NewTableReader(table, chunkSize)

			if reader.Next() {
				r.currentBatch = reader.RecordBatch()
				r.currentBatch.Retain() // Keep the record alive
				r.batchIdx = 0
				// Don't return here - continue to extract rows below
			} else {
				table.Release()
				reader.Release()
				return nil, fmt.Errorf("failed to read record batch from table with %d rows", table.NumRows())
			}
			reader.Release()
			table.Release()
		} else {
			table.Release()
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

// extractValue extracts a value from an Arrow array at the given index
func extractValue(col arrow.Array, idx int) any {
	if col.IsNull(idx) {
		return nil
	}

	switch c := col.(type) {
	case *array.Boolean:
		return c.Value(idx)
	case *array.Int8:
		return c.Value(idx)
	case *array.Int16:
		return c.Value(idx)
	case *array.Int32:
		return c.Value(idx)
	case *array.Int64:
		return c.Value(idx)
	case *array.Uint8:
		return c.Value(idx)
	case *array.Uint16:
		return c.Value(idx)
	case *array.Uint32:
		return c.Value(idx)
	case *array.Uint64:
		return c.Value(idx)
	case *array.Float32:
		return c.Value(idx)
	case *array.Float64:
		return c.Value(idx)
	case *array.String:
		return c.Value(idx)
	case *array.LargeString:
		return c.Value(idx)
	case *array.Binary:
		return c.Value(idx)
	case *array.LargeBinary:
		return c.Value(idx)
	case *array.FixedSizeBinary:
		return c.Value(idx)
	case *array.Date32:
		return c.Value(idx).ToTime()
	case *array.Date64:
		return c.Value(idx).ToTime()
	case *array.Timestamp:
		return c.Value(idx).ToTime(c.DataType().(*arrow.TimestampType).Unit)
	default:
		// For complex types, try to convert to string
		return fmt.Sprintf("%v", c.ValueStr(idx))
	}
}

// Close closes the Parquet reader
func (r *ParquetReader) Close() error {
	if r.currentBatch != nil {
		r.currentBatch.Release()
	}
	if r.reader != nil {
		r.reader.Close()
	}
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}