package batch

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/axonops/cqlai/internal/db"
)

// outputCSV outputs data in CSV format
func (e *Executor) outputCSV(data [][]string) error {
	csvWriter := csv.NewWriter(e.writer)
	if e.options.FieldSep != "" && len(e.options.FieldSep) == 1 {
		csvWriter.Comma = rune(e.options.FieldSep[0])
	}

	for i, row := range data {
		// Skip header if NoHeader is set and this is the first row
		if i == 0 && e.options.NoHeader {
			continue
		}
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV: %w", err)
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

// outputStreamingCSV outputs streaming data in CSV format
func (e *Executor) outputStreamingCSV(ctx context.Context, result db.StreamingQueryResult) error {
	csvWriter := csv.NewWriter(e.writer)
	if e.options.FieldSep != "" && len(e.options.FieldSep) == 1 {
		csvWriter.Comma = rune(e.options.FieldSep[0])
	}

	// Write headers unless NoHeader is set
	if !e.options.NoHeader {
		if err := csvWriter.Write(result.Headers); err != nil {
			return fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	// Get column information from the iterator
	cols := result.Iterator.Columns()

	// Get the current keyspace for UDT lookups
	currentKeyspace := result.Keyspace
	if currentKeyspace == "" && e.sessionManager != nil {
		currentKeyspace = e.sessionManager.CurrentKeyspace()
	}

	// The same reading of a row as JSON output's: a NULL is nil, a tuple is
	// one value read from a destination for each of its elements, and a UDT
	// is decoded from its bytes.
	decoder := e.session.NewRowDecoder(cols, result.ColumnTypes, currentKeyspace)
	handler := db.NewCQLTypeHandler()

	// Precompute column name to index map to avoid O(cols^2) lookup per row
	columnIndexMap := make(map[string]int, len(result.ColumnNames))
	for j, name := range result.ColumnNames {
		columnIndexMap[name] = j
	}

	rowCount := 0
	for {
		select {
		case <-ctx.Done():
			csvWriter.Flush()
			return nil
		default:
			if !result.Iterator.Scan(decoder.Dest()...) {
				csvWriter.Flush()
				return result.Iterator.Close()
			}

			// Convert row to string array; a NULL is an empty field.
			values := decoder.Row()
			row := make([]string, len(result.ColumnNames))
			for _, col := range cols {
				colIdx, found := columnIndexMap[col.Name]
				if !found {
					continue
				}
				switch val := values[col.Name]; {
				case val == nil:
					row[colIdx] = ""
				case col.TypeInfo != nil:
					row[colIdx] = handler.FormatValue(val, col.TypeInfo)
				default:
					row[colIdx] = db.FormatValue(val)
				}
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
			rowCount++

			// Flush periodically - every 1000 rows instead of every row
			if rowCount%1000 == 0 {
				csvWriter.Flush()
			}
		}
	}
}
