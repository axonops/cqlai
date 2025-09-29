package db

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/cassandra-gocql-driver/v2"
)

// StreamingProcessor handles progressive loading and formatting of query results
type StreamingProcessor struct {
	iterator        *gocql.Iter
	headers         []string
	columnNames     []string
	columnTypes     []string
	currentKeyspace string
	tableName       string
	session         *Session
	typeHandler     *CQLTypeHandler
	decoder         *BinaryDecoder
}

// NewStreamingProcessor creates a new streaming processor for progressive result loading
func NewStreamingProcessor(result StreamingQueryResult, session *Session) *StreamingProcessor {
	var decoder *BinaryDecoder
	if session != nil {
		registry := session.GetUDTRegistry()
		if registry != nil {
			decoder = NewBinaryDecoder(registry)
		}
	}

	// Extract table name from the query result if possible
	tableName := ""
	// TODO: Could parse from query or pass explicitly

	return &StreamingProcessor{
		iterator:        result.Iterator,
		headers:         result.Headers,
		columnNames:     result.ColumnNames,
		columnTypes:     result.ColumnTypes,
		currentKeyspace: result.Keyspace,
		tableName:       tableName,
		session:         session,
		typeHandler:     NewCQLTypeHandler(),
		decoder:         decoder,
	}
}

// LoadResults loads a batch of results from the iterator
// Returns the formatted rows, whether more results exist, and any error
func (sp *StreamingProcessor) LoadResults(ctx context.Context, maxRows int) ([][]string, bool, error) {
	if sp.iterator == nil {
		return nil, false, fmt.Errorf("iterator is nil")
	}

	rows := make([][]string, 0, maxRows)
	rowCount := 0

	for rowCount < maxRows {
		select {
		case <-ctx.Done():
			return rows, false, ctx.Err()
		default:
			// Use MapScan to handle NULLs properly
			rowMap := make(map[string]interface{})
			if !sp.iterator.MapScan(rowMap) {
				// No more rows or error occurred
				if err := sp.iterator.Close(); err != nil {
					return rows, false, fmt.Errorf("iterator error: %w", err)
				}
				return rows, false, nil
			}

			// Convert row to string array
			row := sp.formatRow(rowMap)
			rows = append(rows, row)
			rowCount++
		}
	}

	// We loaded maxRows, there might be more
	return rows, true, nil
}

// formatRow formats a single row from MapScan results
func (sp *StreamingProcessor) formatRow(rowMap map[string]interface{}) []string {
	row := make([]string, len(sp.columnNames))

	// Get column information for type-aware formatting
	cols := sp.iterator.Columns()

	for i, colName := range sp.columnNames {
		val, exists := rowMap[colName]
		if !exists || val == nil {
			row[i] = sp.typeHandler.NullString
			continue
		}

		// Find column info for this column
		var col *gocql.ColumnInfo
		for _, c := range cols {
			if c.Name == colName {
				col = &c
				break
			}
		}

		// Check if it's a UDT that needs special handling
		if col != nil && sp.isUDT(col) && sp.decoder != nil {
			// Try to decode UDT if we have bytes
			if bytes, ok := val.([]byte); ok && len(bytes) > 0 {
				if sp.currentKeyspace != "" && sp.tableName != "" && sp.session != nil {
					// Get full type definition from system tables
					fullType := sp.session.GetColumnTypeFromSystemTable(sp.currentKeyspace, sp.tableName, colName)
					if fullType != "" {
						if typeInfo, err := ParseCQLType(fullType); err == nil {
							if decoded, err := sp.decoder.Decode(bytes, typeInfo, sp.currentKeyspace); err == nil {
								val = decoded
							}
						}
					}
				}
			}
		}

		// Format the value
		if col != nil && col.TypeInfo != nil {
			row[i] = sp.typeHandler.FormatValue(val, col.TypeInfo)
		} else {
			row[i] = FormatValue(val)
		}
	}

	return row
}

// isUDT safely checks if a column is a UDT type
func (sp *StreamingProcessor) isUDT(col *gocql.ColumnInfo) bool {
	if col == nil || col.TypeInfo == nil {
		return false
	}

	// Use defer/recover to catch any panic from Type() call
	isUDT := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				// TypeInfo.Type() panicked, not a UDT
				isUDT = false
			}
		}()
		isUDT = col.TypeInfo.Type() == gocql.TypeUDT
	}()

	return isUDT
}

// Close closes the iterator if it's still open
func (sp *StreamingProcessor) Close() error {
	if sp.iterator != nil {
		return sp.iterator.Close()
	}
	return nil
}

// GetHeaders returns the column headers
func (sp *StreamingProcessor) GetHeaders() []string {
	return sp.headers
}

// GetColumnNames returns the raw column names (without PK/C indicators)
func (sp *StreamingProcessor) GetColumnNames() []string {
	return sp.columnNames
}

// StreamingResult represents a complete result that can be loaded progressively
type StreamingResult struct {
	Headers     []string
	Rows        [][]string
	HasMore     bool
	LoadMore    func(ctx context.Context, count int) ([][]string, bool, error)
	Close       func() error
	ElapsedTime time.Duration
}

// ProcessStreamingQuery creates a StreamingResult that loads data on demand
func (s *Session) ProcessStreamingQuery(result StreamingQueryResult) *StreamingResult {
	processor := NewStreamingProcessor(result, s)

	return &StreamingResult{
		Headers:     processor.GetHeaders(),
		Rows:        [][]string{},
		HasMore:     true,
		LoadMore: func(ctx context.Context, count int) ([][]string, bool, error) {
			rows, hasMore, err := processor.LoadResults(ctx, count)
			if err != nil {
				return nil, false, err
			}
			return rows, hasMore, nil
		},
		Close: func() error {
			return processor.Close()
		},
		ElapsedTime: time.Since(result.StartTime),
	}
}