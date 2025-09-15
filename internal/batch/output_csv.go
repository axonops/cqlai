package batch

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
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

	// Create scan destinations - use RawBytes for UDT columns
	scanDest := make([]interface{}, len(cols))
	udtColumns := make(map[int]*db.CQLTypeInfo)

	// Get UDT decoder and registry if we might have UDT columns
	var decoder *db.BinaryDecoder
	var registry *db.UDTRegistry
	if e.session != nil {
		registry = e.session.GetUDTRegistry()
		decoder = db.NewBinaryDecoder(registry)
	}

	// Get the current keyspace for UDT lookups
	currentKeyspace := result.Keyspace
	if currentKeyspace == "" && e.sessionManager != nil {
		currentKeyspace = e.sessionManager.CurrentKeyspace()
	}

	for i, col := range cols {
		if col.TypeInfo.Type() == gocql.TypeUDT {
			scanDest[i] = new(db.RawBytes)

			// Try to get the full type string from column types if available
			if i < len(result.ColumnTypes) {
				typeStr := result.ColumnTypes[i]
				if typeStr != "" && typeStr != "udt" {
					// Parse the type string to get the UDT info
					parsedType, err := db.ParseCQLType(typeStr)
					if err == nil && parsedType != nil {
						udtColumns[i] = parsedType
					}
				}
			}
		} else {
			scanDest[i] = new(interface{})
		}
	}

	for {
		select {
		case <-ctx.Done():
			csvWriter.Flush()
			return nil
		default:
			if !result.Iterator.Scan(scanDest...) {
				csvWriter.Flush()
				return result.Iterator.Close()
			}

			// Convert row to string array
			row := make([]string, len(result.ColumnNames))
			for i, col := range cols {
				colName := col.Name
				colIdx := -1
				for j, name := range result.ColumnNames {
					if name == colName {
						colIdx = j
						break
					}
				}

				if colIdx >= 0 {
					if udtTypeInfo, hasUDT := udtColumns[i]; hasUDT {
						// Handle UDT column
						rawBytes := scanDest[i].(*db.RawBytes)
						if rawBytes != nil && *rawBytes != nil && decoder != nil {
							// Determine keyspace
							keyspace := udtTypeInfo.Keyspace
							if keyspace == "" {
								keyspace = currentKeyspace
							}

							decodedValue, err := decoder.Decode([]byte(*rawBytes), udtTypeInfo, keyspace)
							if err == nil {
								row[colIdx] = db.FormatValue(decodedValue)
							} else {
								logger.DebugfToFile("batch", "Failed to decode UDT for %s: %v", colName, err)
								row[colIdx] = db.FormatValue(*rawBytes)
							}
						} else {
							row[colIdx] = ""
						}
					} else {
						// Regular column
						val := *(scanDest[i].(*interface{}))
						if val == nil {
							row[colIdx] = ""
						} else {
							row[colIdx] = db.FormatValue(val)
						}
					}
				}
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}

			// Flush periodically for large result sets
			csvWriter.Flush()
		}
	}
}