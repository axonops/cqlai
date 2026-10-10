package batch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// outputJSON outputs data in JSON format
func (e *Executor) outputJSON(data [][]string) error {
	if len(data) == 0 {
		fmt.Fprintln(e.writer, "[]")
		return nil
	}

	headers := data[0]
	var results []map[string]string

	for i := 1; i < len(data); i++ {
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(data[i]) {
				row[header] = data[i][j]
			}
		}
		results = append(results, row)
	}

	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// outputJSONWithRawData outputs JSON using the raw data map for better type preservation
func (e *Executor) outputJSONWithRawData(result db.QueryResult) error {
	if len(result.RawData) == 0 {
		fmt.Fprintln(e.writer, "[]")
		return nil
	}

	// Use the raw data directly for JSON output
	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result.RawData)
}

// outputStreamingJSON outputs streaming data in JSON format
func (e *Executor) outputStreamingJSON(ctx context.Context, result db.StreamingQueryResult) error {
	fmt.Fprint(e.writer, "[")
	first := true

	// Get column information from the iterator
	cols := result.Iterator.Columns()

	// Debug: log the column types we received
	logger.DebugfToFile("batch", "ColumnTypes from result: %v", result.ColumnTypes)
	logger.DebugfToFile("batch", "Number of columns: %d, Number of column types: %d", len(cols), len(result.ColumnTypes))

	// Get the current keyspace for UDT lookups - prefer from result, then session manager
	currentKeyspace := result.Keyspace
	if currentKeyspace == "" && e.sessionManager != nil {
		currentKeyspace = e.sessionManager.CurrentKeyspace()
	}
	decoder := e.session.NewRowDecoder(cols, result.ColumnTypes, currentKeyspace)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(e.writer, "\n]")
			return nil
		default:
			if !result.Iterator.Scan(decoder.Dest()...) {
				fmt.Fprintln(e.writer, "\n]")
				return result.Iterator.Close()
			}

			rowMap := decoder.Row()

			// Write comma if not first row
			if !first {
				fmt.Fprint(e.writer, ",")
			}
			first = false

			// Encode and write row
			fmt.Fprint(e.writer, "\n  ")
			encoder := json.NewEncoder(e.writer)
			encoder.SetIndent("  ", "  ")
			// Each value as JSON can hold it: a duration as 1mo2d3ns, a
			// decimal as its digits, a blob as 0x hex.
			if err := encoder.Encode(db.JSONValue(rowMap)); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		}
	}
}
