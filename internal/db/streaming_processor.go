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
func (sp *StreamingProcessor) LoadResults(ctx context.Context, maxRows int) (Page, error) {
	if sp.iterator == nil {
		return Page{}, fmt.Errorf("iterator is nil")
	}

	page := Page{
		Rows: make([][]string, 0, maxRows),
		Raw:  make([]map[string]interface{}, 0, maxRows),
	}

	for len(page.Rows) < maxRows {
		select {
		case <-ctx.Done():
			return page, ctx.Err()
		default:
			// Use MapScan to handle NULLs properly
			rowMap := make(map[string]interface{})
			if !sp.iterator.MapScan(rowMap) {
				// No more rows or error occurred
				if err := sp.iterator.Close(); err != nil {
					return page, fmt.Errorf("iterator error: %w", err)
				}
				return page, nil
			}

			// Both the values as they are drawn and the values themselves. The
			// second set is what JSON is built from: a timestamp formatted for
			// a table cell, or a collection written the way cqlsh writes one,
			// cannot be turned back into what it came from.
			page.Rows = append(page.Rows, sp.formatRow(rowMap))
			page.Raw = append(page.Raw, rowMap)
		}
	}

	// We loaded maxRows, there might be more
	page.HasMore = true
	return page, nil
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

// Page is a batch of rows as they came back: the values as they are drawn, and
// the values themselves.
//
// Both, because they answer different questions and neither can be got from the
// other. The strings are what a table cell holds. The raw values are what JSON
// is built from, and what the OUTPUT format could not be changed without: a
// result fetched as JSON is one column of documents, and no amount of redrawing
// turns that back into columns.
type Page struct {
	Rows    [][]string
	Raw     []map[string]interface{}
	HasMore bool
}

// StreamingResult represents a complete result that can be loaded progressively
type StreamingResult struct {
	Headers     []string
	Rows        [][]string
	HasMore     bool
	LoadMore    func(ctx context.Context, count int) (Page, error)
	Close       func() error
	ElapsedTime time.Duration
}

// ProcessStreamingQuery creates a StreamingResult that loads data on demand
func (s *Session) ProcessStreamingQuery(result StreamingQueryResult) *StreamingResult {
	processor := NewStreamingProcessor(result, s)

	return &StreamingResult{
		Headers: processor.GetHeaders(),
		Rows:    [][]string{},
		HasMore: true,
		LoadMore: func(ctx context.Context, count int) (Page, error) {
			return processor.LoadResults(ctx, count)
		},
		Close: func() error {
			return processor.Close()
		},
		ElapsedTime: time.Since(result.StartTime),
	}
}
