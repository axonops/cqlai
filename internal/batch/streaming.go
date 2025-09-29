package batch

import (
	"context"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/ui"
)

// handleStreamingResult handles streaming query results with automatic pagination
func (e *Executor) handleStreamingResult(ctx context.Context, result db.StreamingQueryResult) error {
	defer result.Iterator.Close()

	// For CSV and JSON, we need to handle differently
	switch e.options.Format {
	case OutputFormatCSV:
		return e.outputStreamingCSV(ctx, result)
	case OutputFormatJSON:
		return e.outputStreamingJSON(ctx, result)
	}

	// For ASCII and Table formats, stream with pagination
	rowCount := 0
	var rows [][]string
	isFirstBatch := true
	var columnWidths []int // Store column widths from first batch


	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(e.writer, "\n\nQuery interrupted by user")
			return nil
		default:
			// Use MapScan to handle NULLs properly - gocql can panic on NULL values with Scan()
			rowMap := make(map[string]interface{})
			if !result.Iterator.MapScan(rowMap) {
				// Check for errors
				if err := result.Iterator.Close(); err != nil {
					return fmt.Errorf("iterator error: %w", err)
				}

				// Output final batch if we have rows
				switch {
				case len(rows) > 0 && isFirstBatch:
					// If this is the only batch, include headers
					allData := append([][]string{result.Headers}, rows...)
					output := ui.FormatASCIITable(allData)
					// Remove the row count from FormatASCIITable output as we'll add our own
					output = strings.TrimSuffix(output, fmt.Sprintf("\n(%d rows)\n", len(rows)))
					output = strings.TrimSuffix(output, fmt.Sprintf("\n(%d row)\n", len(rows)))
					fmt.Fprint(e.writer, output)
				case len(rows) > 0:
					// For final batch of multi-batch output, just output rows with stored widths
					allData := append([][]string{result.Headers}, rows...)
					output := ui.FormatASCIITableRowsOnlyWithWidths(allData, columnWidths)
					fmt.Fprint(e.writer, output)
					// Add bottom border using stored widths
					bottomBorder := ui.FormatASCIITableBottomWithWidths(allData, columnWidths)
					fmt.Fprint(e.writer, bottomBorder)
				case isFirstBatch:
					// No rows at all, just print headers
					headerData := [][]string{result.Headers}
					output := ui.FormatASCIITable(headerData)
					// Remove the row count as we'll add our own
					output = strings.TrimSuffix(output, "\n(0 rows)\n")
					fmt.Fprint(e.writer, output)
				default:
					// No rows in final batch but had previous batches - just add bottom border with stored widths
					bottomBorder := ui.FormatASCIITableBottomWithWidths([][]string{result.Headers}, columnWidths)
					fmt.Fprint(e.writer, bottomBorder)
				}

				fmt.Fprintf(e.writer, "\n(%d rows)\n", rowCount)
				return nil
			}

			// Convert row to string array using MapScan results
			row := make([]string, len(result.ColumnNames))
			for i, colName := range result.ColumnNames {
				if val, exists := rowMap[colName]; exists {
					row[i] = db.FormatValue(val)
				} else {
					row[i] = "null"
				}
			}

			rows = append(rows, row)
			rowCount++

			// Output batch based on configured page size
			batchSize := e.options.PageSize
			if batchSize <= 0 {
				batchSize = 100 // Default to 100 if not set
			}
			if len(rows) >= batchSize {
				// For the first batch, include headers and calculate column widths
				if isFirstBatch {
					// Combine headers with rows to calculate proper widths
					allData := append([][]string{result.Headers}, rows...)
					// Store column widths for subsequent batches
					columnWidths = ui.CalculateColumnWidths(allData)

					output := ui.FormatASCIITable(allData)
					// Remove the row count from FormatASCIITable output
					output = strings.TrimSuffix(output, fmt.Sprintf("\n(%d rows)\n", len(rows)))
					output = strings.TrimSuffix(output, fmt.Sprintf("\n(%d row)\n", len(rows)))

					// Remove the bottom border - we'll add it at the very end
					lines := strings.Split(output, "\n")
					// Find and remove the last border (line starting with +) and any trailing empty lines
					done := false
					for i := len(lines) - 1; i >= 0 && !done; i-- {
						switch {
						case lines[i] == "":
							// Remove empty line
							lines = lines[:i]
						case strings.HasPrefix(lines[i], "+") && strings.Contains(lines[i], "-"):
							// Found the bottom border, remove it
							lines = lines[:i]
							done = true
						default:
							// Found a data line, stop
							done = true
						}
					}
					output = strings.Join(lines, "\n") + "\n"
					fmt.Fprint(e.writer, output)
					isFirstBatch = false
				} else {
					// For subsequent batches, use the stored column widths for consistent formatting
					allData := append([][]string{result.Headers}, rows...)
					output := ui.FormatASCIITableRowsOnlyWithWidths(allData, columnWidths)
					fmt.Fprint(e.writer, output)
				}
				// Clear rows for next batch
				rows = [][]string{}
			}
		}
	}
}

// handleQueryResult handles non-streaming query results
func (e *Executor) handleQueryResult(result db.QueryResult) error {
	switch e.options.Format {
	case OutputFormatJSON:
		// For JSON, use raw data if available
		if len(result.RawData) > 0 {
			return e.outputJSONWithRawData(result)
		}
		return e.outputJSON(result.Data)
	case OutputFormatCSV:
		return e.outputCSV(result.Data)
	default:
		// Default to table format
		// Note: outputTable already includes row count
		return e.outputTable(result.Data)
	}
}