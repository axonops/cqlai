package batch

import (
	"context"
	"fmt"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/ui"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
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
	// First, print the table header
	headerData := [][]string{result.Headers}
	headerOutput := ui.FormatASCIITableHeader(headerData)
	fmt.Fprint(e.writer, headerOutput)

	rowCount := 0
	var rows [][]string

	// Get column information from the iterator
	cols := result.Iterator.Columns()

	// For tables with tuples, gocql expects us to provide scan destinations
	// for each tuple field, not just the tuple column itself.
	// Calculate the total number of scan destinations needed.
	scanCount := 0
	tupleColumns := make(map[int]int) // Map column index to number of tuple elements

	for i, col := range cols {
		switch {
		case col.TypeInfo == nil:
			// Handle nil TypeInfo (virtual tables)
			scanCount++
		case col.TypeInfo.Type() == gocql.TypeTuple:
			// Get the tuple element count from TypeInfo - use safe type assertion
			if tupleInfo, ok := col.TypeInfo.(gocql.TupleTypeInfo); ok {
				tupleElements := len(tupleInfo.Elems)
				tupleColumns[i] = tupleElements
				scanCount += tupleElements
			} else {
				// If we can't get tuple info, treat as single column
				scanCount++
			}
		default:
			scanCount++
		}
	}

	// Create scan destinations
	scanDest := make([]interface{}, scanCount)
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

	// Set up scan destinations - handle tuples specially
	destIdx := 0
	for i, col := range cols {
		switch {
		case tupleColumns[i] > 0:
			// For tuple columns, create destinations for each element
			tupleCount := tupleColumns[i]
			for j := 0; j < tupleCount; j++ {
				scanDest[destIdx] = new(interface{})
				destIdx++
			}
		case col.TypeInfo == nil:
			// Handle nil TypeInfo (virtual tables)
			scanDest[destIdx] = new(interface{})
			destIdx++
		case col.TypeInfo.Type() == gocql.TypeUDT:
			scanDest[destIdx] = new(db.RawBytes)
			udtColumns[destIdx] = nil

			// Try to get the full type string from column types if available
			if i < len(result.ColumnTypes) {
				typeStr := result.ColumnTypes[i]
				if typeStr != "" && typeStr != "udt" {
					// Parse the type string to get the UDT info
					parsedType, err := db.ParseCQLType(typeStr)
					if err == nil && parsedType != nil {
						udtColumns[destIdx] = parsedType
					}
				} else if typeStr == "udt" || typeStr == "" {
					// If we just have "udt" without details, try to extract from gocql TypeInfo
					if udtInfo, ok := col.TypeInfo.(gocql.UDTTypeInfo); ok {
						// Create a minimal CQLTypeInfo with the UDT name
						udtColumns[destIdx] = &db.CQLTypeInfo{
							BaseType: "udt",
							UDTName:  udtInfo.Name,
							Keyspace: udtInfo.Keyspace,
						}
					}
				}
			}
			destIdx++
		default:
			scanDest[destIdx] = new(interface{})
			destIdx++
		}
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(e.writer, "\n\nQuery interrupted by user")
			return nil
		default:
			if !result.Iterator.Scan(scanDest...) {
				// Check for errors
				if err := result.Iterator.Close(); err != nil {
					return fmt.Errorf("iterator error: %w", err)
				}

				// Output final batch if we have rows
				if len(rows) > 0 {
					if err := e.outputStreamingRows(rows, result.Headers); err != nil {
						return err
					}
				}

				// Print bottom border and row count
				e.printTableBottom(result.Headers)
				fmt.Fprintf(e.writer, "\n(%d rows)\n", rowCount)
				return nil
			}

			// Convert row to string array
			row := make([]string, len(result.ColumnNames))

			// Process scanned values - reconstruct tuples
			destIdx := 0
			for i := range cols {
				if i >= len(row) {
					break
				}

				if tupleCount, isTuple := tupleColumns[i]; isTuple {
					// Reconstruct tuple from its elements
					tupleElements := make([]interface{}, tupleCount)
					for j := 0; j < tupleCount && destIdx < len(scanDest); j++ {
						if scanDest[destIdx] != nil {
							tupleElements[j] = *(scanDest[destIdx].(*interface{}))
						}
						destIdx++
					}
					// Format tuple as (elem1, elem2, ...)
					row[i] = db.FormatValue(tupleElements)
				} else if destIdx < len(scanDest) {
					if udtTypeInfo, hasUDT := udtColumns[destIdx]; hasUDT {
						// Handle UDT column
						rawBytes := scanDest[destIdx].(*db.RawBytes)
						if rawBytes != nil && *rawBytes != nil && decoder != nil {
							// Determine keyspace
							keyspace := udtTypeInfo.Keyspace
							if keyspace == "" {
								keyspace = currentKeyspace
							}

							decodedValue, err := decoder.Decode([]byte(*rawBytes), udtTypeInfo, keyspace)
							if err == nil {
								row[i] = db.FormatValue(decodedValue)
							} else {
								row[i] = db.FormatValue(*rawBytes)
							}
						} else {
							row[i] = "null"
						}
					} else {
						// Regular column
						val := *(scanDest[destIdx].(*interface{}))
						row[i] = db.FormatValue(val)
					}
					destIdx++
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
				if err := e.outputStreamingRows(rows, result.Headers); err != nil {
					return err
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
		err := e.outputTable(result.Data)
		if err == nil && len(result.Data) > 1 {
			fmt.Fprintf(e.writer, "\n(%d rows)\n", len(result.Data)-1)
		}
		return err
	}
}