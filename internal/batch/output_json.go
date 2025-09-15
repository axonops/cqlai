package batch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
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

	// Get the current keyspace for UDT lookups - prefer from result, then session manager
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
				logger.DebugfToFile("batch", "Column %s: type from ColumnTypes = %s", col.Name, typeStr)
				if typeStr != "" && typeStr != "udt" {
					// Parse the type string to get the UDT info
					parsedType, err := db.ParseCQLType(typeStr)
					if err == nil && parsedType != nil {
						udtColumns[i] = parsedType
						logger.DebugfToFile("batch", "Parsed UDT type for %s: %+v", col.Name, parsedType)
					} else if err != nil {
						logger.DebugfToFile("batch", "Failed to parse type for %s: %v", col.Name, err)
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
			fmt.Fprintln(e.writer, "\n]")
			return nil
		default:
			if !result.Iterator.Scan(scanDest...) {
				fmt.Fprintln(e.writer, "\n]")
				return result.Iterator.Close()
			}

			// Build row map
			rowMap := make(map[string]interface{})
			for i, col := range cols {
				colName := col.Name
				if udtTypeInfo, hasUDT := udtColumns[i]; hasUDT {
					// Handle UDT column
					rawBytes := scanDest[i].(*db.RawBytes)
					if rawBytes != nil && *rawBytes != nil && decoder != nil {
						// Use the pre-parsed type info if available
						if udtTypeInfo != nil {
							// Determine keyspace - use from type info or current keyspace
							keyspace := udtTypeInfo.Keyspace
							if keyspace == "" {
								keyspace = currentKeyspace
							}

							logger.DebugfToFile("batch", "Decoding UDT %s with keyspace=%s, name=%s",
								colName, keyspace, udtTypeInfo.UDTName)

							decodedValue, err := decoder.Decode([]byte(*rawBytes), udtTypeInfo, keyspace)
							if err == nil {
								rowMap[colName] = decodedValue
								logger.DebugfToFile("batch", "Successfully decoded UDT %s: %+v", colName, decodedValue)
							} else {
								logger.DebugfToFile("batch", "Failed to decode UDT %s: %v", colName, err)
								// On error, try to return the raw bytes as hex string
								rowMap[colName] = fmt.Sprintf("0x%x", *rawBytes)
							}
						} else {
							logger.DebugfToFile("batch", "No UDT type info for %s", colName)
							rowMap[colName] = fmt.Sprintf("0x%x", *rawBytes)
						}
					} else {
						rowMap[colName] = nil
					}
				} else {
					// Regular column
					val := *(scanDest[i].(*interface{}))
					rowMap[colName] = val
				}
			}

			// Write comma if not first row
			if !first {
				fmt.Fprint(e.writer, ",")
			}
			first = false

			// Encode and write row
			fmt.Fprint(e.writer, "\n  ")
			encoder := json.NewEncoder(e.writer)
			encoder.SetIndent("  ", "  ")
			if err := encoder.Encode(rowMap); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		}
	}
}