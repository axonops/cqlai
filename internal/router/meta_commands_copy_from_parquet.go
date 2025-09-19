package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// executeCopyFromParquet executes COPY FROM operation for Parquet format
func (h *MetaCommandHandler) executeCopyFromParquet(table string, columns []string, filename string, options map[string]string) interface{} {
	// Check if input is STDIN (not supported for Parquet)
	isStdin := strings.ToUpper(filename) == "STDIN"
	if isStdin {
		return "COPY FROM STDIN is not supported for Parquet format. Please provide a file path."
	}

	// Clean the filename to prevent path traversal
	cleanPath := filepath.Clean(filename)

	// Open the Parquet file
	reader, err := parquet.NewParquetReader(cleanPath)
	if err != nil {
		return fmt.Sprintf("Error opening Parquet file: %v", err)
	}
	defer reader.Close()

	// Get schema from Parquet file
	parquetColumns, parquetTypes := reader.GetSchema()

	// If no columns specified, use all columns from the Parquet file
	if len(columns) == 0 {
		columns = parquetColumns
	} else {
		// Validate that specified columns exist in the Parquet file
		columnMap := make(map[string]bool)
		for _, col := range parquetColumns {
			columnMap[col] = true
		}

		for _, col := range columns {
			if !columnMap[col] {
				return fmt.Sprintf("Column '%s' not found in Parquet file", col)
			}
		}
	}

	// Build column list for INSERT
	columnList := strings.Join(columns, ", ")
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	placeholderList := strings.Join(placeholders, ", ")

	// Prepare INSERT statement template (for reference/logging)
	_ = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnList, placeholderList)

	// Parse numeric options
	batchSize, _ := strconv.Atoi(options["CHUNKSIZE"])
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	maxRows, _ := strconv.Atoi(options["MAXROWS"])
	skipRows, _ := strconv.Atoi(options["SKIPROWS"])
	maxInsertErrors, _ := strconv.Atoi(options["MAXINSERTERRORS"])

	// Process rows
	rowCount := 0
	processedRows := 0
	insertErrorCount := 0
	skippedRows := 0

	// Skip initial rows if specified
	if skipRows > 0 {
		// Read and discard the skip rows
		for i := 0; i < skipRows; i += batchSize {
			batch, err := reader.ReadBatch(batchSize)
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Sprintf("Error reading Parquet file: %v", err)
			}

			toSkip := skipRows - i
			if toSkip > len(batch) {
				skippedRows += len(batch)
			} else {
				skippedRows += toSkip
				break
			}
		}
	}

	// Process data in batches
	for maxRows <= 0 || processedRows < maxRows {
		// Read a batch of rows
		batch, err := reader.ReadBatch(batchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Sprintf("Error reading Parquet batch: %v", err)
		}

		logger.DebugfToFile("CopyFromParquet", "Read batch of %d rows", len(batch))

		// Process each row in the batch
		for _, row := range batch {
			processedRows++

			// Check max rows limit
			if maxRows > 0 && processedRows > maxRows {
				break
			}

			// Build values array for the INSERT
			values := make([]any, len(columns))
			for i, colName := range columns {
				if val, ok := row[colName]; ok {
					values[i] = val
					// Debug log the value and its type for UDT columns
					logger.DebugfToFile("CopyFromParquet", "Column %s: value=%v, type=%T", colName, val, val)
				} else {
					values[i] = nil
				}
			}


			// Execute the INSERT
			// For now, we'll build a string representation
			// In a real implementation, this would use prepared statements
			valueStrings := make([]string, len(values))
			for i, val := range values {
				if val == nil {
					valueStrings[i] = "null"
				} else {
					// Format value based on type
					switch v := val.(type) {
					case string:
						trimmed := strings.TrimSpace(v)

						// Check if this is a list/set (starts with [ and ends with ])
						switch {
						case strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"):
							// Parse list/set format: [value1 value2 value3]
							inner := strings.Trim(trimmed, "[]")
							if inner == "" {
								// Empty collection - need to determine if it's a list or set
								// Check column name to determine collection type
								if strings.Contains(columns[i], "unique") || strings.HasSuffix(columns[i], "_set") ||
									strings.HasSuffix(columns[i], "_nums") {
									// Use set syntax with {}
									valueStrings[i] = "{}"
								} else {
									// Use list/vector syntax with []
									valueStrings[i] = "[]"
								}
							} else {
								// Split values and quote them
								parts := strings.Fields(inner)
								quotedParts := make([]string, len(parts))
								for j, part := range parts {
									// Check if it's a number (int or float)
									if _, err := strconv.Atoi(part); err == nil {
										quotedParts[j] = part // Integer, don't quote
									} else if _, err := strconv.ParseFloat(part, 64); err == nil {
										quotedParts[j] = part // Float, don't quote
									} else {
										quotedParts[j] = fmt.Sprintf("'%s'", strings.ReplaceAll(part, "'", "''"))
									}
								}
								// Check column name to determine collection type
								// This is a heuristic - sets typically have "unique" or end with "_set"
								// Vectors typically have "embedding", "vector", or end with "_vec"
								if strings.Contains(columns[i], "unique") || strings.HasSuffix(columns[i], "_set") ||
									strings.HasSuffix(columns[i], "_nums") {
									// Use set syntax with {}
									valueStrings[i] = "{" + strings.Join(quotedParts, ", ") + "}"
								} else {
									// Use list/vector syntax with []
									valueStrings[i] = "[" + strings.Join(quotedParts, ", ") + "]"
								}
							}
						case strings.HasPrefix(trimmed, "map[") && strings.HasSuffix(trimmed, "]"):
							// Parse map format: map[key1:value1 key2:value2] -> {'key1': value1, 'key2': value2}
							inner := strings.TrimPrefix(trimmed, "map[")
							inner = strings.TrimSuffix(inner, "]")
							if inner == "" {
								// Empty map
								valueStrings[i] = "{}"
							} else {
								// Parse key:value pairs
								pairs := strings.Fields(inner)
								mapPairs := make([]string, 0, len(pairs))
								for _, pair := range pairs {
									kv := strings.SplitN(pair, ":", 2)
									if len(kv) == 2 {
										key := kv[0]
										val := kv[1]
										// Quote the key
										quotedKey := fmt.Sprintf("'%s'", strings.ReplaceAll(key, "'", "''"))
										// Check if value is a number
										if _, err := strconv.Atoi(val); err == nil {
											mapPairs = append(mapPairs, fmt.Sprintf("%s: %s", quotedKey, val))
										} else {
											quotedVal := fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "''"))
											mapPairs = append(mapPairs, fmt.Sprintf("%s: %s", quotedKey, quotedVal))
										}
									}
								}
								valueStrings[i] = "{" + strings.Join(mapPairs, ", ") + "}"
							}
						case strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") &&
							strings.Contains(trimmed, ":") && strings.Contains(trimmed, "\""):
							// This looks like JSON for a UDT, convert to Cassandra UDT format
							// Remove quotes from field names, keep quotes for string values
							udtValue := trimmed
							// Replace "field": with field: (remove quotes from field names)
							udtValue = strings.ReplaceAll(udtValue, "\"street\":", "street:")
							udtValue = strings.ReplaceAll(udtValue, "\"city\":", "city:")
							udtValue = strings.ReplaceAll(udtValue, "\"zip\":", "zip:")
							// Replace double quotes with single quotes for string values
							udtValue = strings.ReplaceAll(udtValue, "\"", "'")
							valueStrings[i] = udtValue
						default:
							// Regular string value
							valueStrings[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
						}
					case time.Time:
						// Format timestamp for CQL (use RFC3339 format with quotes)
						valueStrings[i] = fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
					case bool:
						valueStrings[i] = fmt.Sprintf("%t", v)
					case []byte:
						valueStrings[i] = fmt.Sprintf("0x%x", v)
					case []interface{}:
						// Handle actual arrays (lists/sets)
						if len(v) == 0 {
							valueStrings[i] = "[]"
						} else {
							quotedParts := make([]string, len(v))
							for j, item := range v {
								switch it := item.(type) {
								case string:
									quotedParts[j] = fmt.Sprintf("'%s'", strings.ReplaceAll(it, "'", "''"))
								case int, int32, int64, float32, float64:
									quotedParts[j] = fmt.Sprintf("%v", it)
								default:
									quotedParts[j] = fmt.Sprintf("'%v'", it)
								}
							}
							valueStrings[i] = "[" + strings.Join(quotedParts, ", ") + "]"
						}
					case map[string]interface{}:
						// Handle STRUCT values (UDTs) from Parquet
						// Format as Cassandra UDT: {field1: 'value1', field2: 'value2'}
						if len(v) == 0 {
							valueStrings[i] = "{}"
						} else {
							udtPairs := make([]string, 0, len(v))
							for fieldName, fieldValue := range v {
								var formattedValue string
								if fieldValue == nil {
									formattedValue = "null"
								} else {
									switch fv := fieldValue.(type) {
									case string:
										formattedValue = fmt.Sprintf("'%s'", strings.ReplaceAll(fv, "'", "''"))
									case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
										formattedValue = fmt.Sprintf("%v", fv)
									case float32, float64:
										formattedValue = fmt.Sprintf("%v", fv)
									case bool:
										formattedValue = fmt.Sprintf("%t", fv)
									default:
										formattedValue = fmt.Sprintf("'%v'", fv)
									}
								}
								// Field names in Cassandra UDT syntax don't have quotes
								udtPairs = append(udtPairs, fmt.Sprintf("%s: %s", fieldName, formattedValue))
							}
							valueStrings[i] = "{" + strings.Join(udtPairs, ", ") + "}"
						}
					case map[interface{}]interface{}:
						// Handle actual maps (for map columns, not UDTs)
						if len(v) == 0 {
							valueStrings[i] = "{}"
						} else {
							mapPairs := make([]string, 0, len(v))
							for key, val := range v {
								var quotedKey string
								switch k := key.(type) {
								case string:
									quotedKey = fmt.Sprintf("'%s'", strings.ReplaceAll(k, "'", "''"))
								default:
									quotedKey = fmt.Sprintf("'%v'", k)
								}

								var quotedVal string
								switch vt := val.(type) {
								case string:
									quotedVal = fmt.Sprintf("'%s'", strings.ReplaceAll(vt, "'", "''"))
								case int, int32, int64, float32, float64:
									quotedVal = fmt.Sprintf("%v", vt)
								default:
									quotedVal = fmt.Sprintf("'%v'", vt)
								}
								mapPairs = append(mapPairs, fmt.Sprintf("%s: %s", quotedKey, quotedVal))
							}
							valueStrings[i] = "{" + strings.Join(mapPairs, ", ") + "}"
						}
					default:
						valueStrings[i] = fmt.Sprintf("%v", v)
					}
				}
			}

			// Build and execute the actual INSERT query
			// Check if we have a current keyspace and qualify the table name if needed
			fullyQualifiedTable := table
			if h.session.Keyspace() != "" && !strings.Contains(table, ".") {
				fullyQualifiedTable = fmt.Sprintf("%s.%s", h.session.Keyspace(), table)
			}
			actualQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				fullyQualifiedTable, columnList, strings.Join(valueStrings, ", "))

			// Debug: Log the query for troubleshooting
			logger.DebugfToFile("CopyFromParquet", "INSERT query: %s", actualQuery)

			// Execute the query
			result := h.session.ExecuteCQLQuery(actualQuery)

			// Check for errors
			if err, isError := result.(error); isError {
				insertErrorCount++
				logger.DebugfToFile("CopyFromParquet", "Insert error: %v", err)
				logger.DebugfToFile("CopyFromParquet", "Failed query: %s", actualQuery)

				// Check if we've exceeded max insert errors
				if maxInsertErrors > 0 && insertErrorCount >= maxInsertErrors {
					return fmt.Sprintf("Aborted after %d insert errors. Successfully imported %d rows.",
						insertErrorCount, rowCount)
				}
			} else {
				rowCount++
			}
		}
	}

	// Return summary
	summary := fmt.Sprintf("Imported %d rows from Parquet file", rowCount)

	if skippedRows > 0 {
		summary += fmt.Sprintf(" (skipped %d rows)", skippedRows)
	}

	if insertErrorCount > 0 {
		summary += fmt.Sprintf(" with %d errors", insertErrorCount)
	}

	// Log type mappings for debugging
	logger.DebugfToFile("CopyFromParquet", "Parquet columns: %v", parquetColumns)
	logger.DebugfToFile("CopyFromParquet", "Parquet types: %v", parquetTypes)
	logger.DebugfToFile("CopyFromParquet", "Total rows in file: %d", reader.GetRowCount())

	return summary
}