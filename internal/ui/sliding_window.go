package ui

import (
	"fmt"
	"time"
	
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/logger"
)

// SlidingWindowTable manages a sliding window of table data with memory limits
type SlidingWindowTable struct {
	// Configuration
	MaxRows        int  // Maximum rows to keep in memory (e.g., 10000)
	MaxMemoryBytes int64 // Maximum memory usage in bytes (e.g., 10MB)
	
	// Data storage
	Headers      []string   // Column headers with (PK)/(C) indicators (always kept)
	ColumnNames  []string   // Original column names without indicators (for data lookup)
	Rows         [][]string // Current window of rows
	ColumnTypes  []string   // Column types (always kept)
	
	// Window tracking
	FirstRowIndex int64 // Global index of first row in window
	TotalRowsSeen int64 // Total rows processed (may be more than in memory)
	CurrentMemory int64 // Approximate current memory usage
	
	// Iterator state for loading more data
	iterator     interface{} // Store the gocql iterator if still available
	hasMoreData  bool       // Whether more data can be fetched
	
	// Indicators for UI
	DataDroppedAtStart bool // True if we've dropped rows from the beginning
	DataAvailableAtEnd bool // True if more data can be loaded
	
	// Capture tracking
	LastCapturedRow int64 // Index of the last row written to capture file
}

// NewSlidingWindowTable creates a new sliding window table
func NewSlidingWindowTable(maxRows int, maxMemoryMB int) *SlidingWindowTable {
	return &SlidingWindowTable{
		MaxRows:        maxRows,
		MaxMemoryBytes: int64(maxMemoryMB * 1024 * 1024),
		Rows:           make([][]string, 0),
		FirstRowIndex:  0,
		TotalRowsSeen:  0,
		CurrentMemory:  0,
		hasMoreData:    false,
	}
}

// AddRow adds a row to the sliding window, potentially evicting old rows
func (swt *SlidingWindowTable) AddRow(row []string) {
	// Calculate approximate memory for this row
	rowMemory := swt.calculateRowMemory(row)
	
	// Check if adding this row would exceed limits
	if swt.needsEviction(rowMemory) {
		swt.evictOldestRows(rowMemory)
	}
	
	// Add the new row
	swt.Rows = append(swt.Rows, row)
	swt.CurrentMemory += rowMemory
	swt.TotalRowsSeen++
	
	logger.DebugfToFile("SlidingWindowTable", "Added row %d, window size: %d, memory: %d bytes", 
		swt.TotalRowsSeen, len(swt.Rows), swt.CurrentMemory)
}

// needsEviction checks if we need to evict rows before adding a new one
func (swt *SlidingWindowTable) needsEviction(newRowMemory int64) bool {
	// Check row count limit
	if len(swt.Rows) >= swt.MaxRows {
		return true
	}
	
	// Check memory limit
	if swt.CurrentMemory + newRowMemory > swt.MaxMemoryBytes {
		return true
	}
	
	return false
}

// evictOldestRows removes rows from the beginning to make space
func (swt *SlidingWindowTable) evictOldestRows(neededMemory int64) {
	rowsToEvict := 0
	freedMemory := int64(0)
	
	// Calculate how many rows to evict
	// Evict at least 10% of rows or enough to free needed memory
	minEvict := len(swt.Rows) / 10
	if minEvict < 1 {
		minEvict = 1
	}
	
	for i := 0; i < len(swt.Rows); i++ {
		freedMemory += swt.calculateRowMemory(swt.Rows[i])
		rowsToEvict++
		
		// Stop if we've freed enough memory and evicted minimum
		if rowsToEvict >= minEvict && 
		   (len(swt.Rows) - rowsToEvict < swt.MaxRows) &&
		   (swt.CurrentMemory - freedMemory + neededMemory <= swt.MaxMemoryBytes) {
			break
		}
	}
	
	// Evict the rows
	if rowsToEvict > 0 && rowsToEvict < len(swt.Rows) {
		logger.DebugfToFile("SlidingWindowTable", "Evicting %d rows, freeing ~%d bytes", 
			rowsToEvict, freedMemory)
		
		swt.Rows = swt.Rows[rowsToEvict:]
		swt.FirstRowIndex += int64(rowsToEvict)
		swt.CurrentMemory -= freedMemory
		swt.DataDroppedAtStart = true
	}
}

// calculateRowMemory estimates memory usage for a row
func (swt *SlidingWindowTable) calculateRowMemory(row []string) int64 {
	memory := int64(24) // Slice overhead
	for _, cell := range row {
		memory += int64(len(cell)) + 24 // String content + string header
	}
	return memory
}

// GetVisibleRows returns rows for display within the specified range
func (swt *SlidingWindowTable) GetVisibleRows(startIdx, endIdx int) [][]string {
	// Adjust indices relative to our window
	windowStart := startIdx - int(swt.FirstRowIndex)
	windowEnd := endIdx - int(swt.FirstRowIndex)
	
	// Clamp to available data
	if windowStart < 0 {
		windowStart = 0
	}
	if windowEnd > len(swt.Rows) {
		windowEnd = len(swt.Rows)
	}
	
	if windowStart >= len(swt.Rows) {
		return [][]string{}
	}
	
	return swt.Rows[windowStart:windowEnd]
}

// CanScrollUp returns true if there's data before the current window
func (swt *SlidingWindowTable) CanScrollUp() bool {
	return swt.FirstRowIndex > 0
}

// CanScrollDown returns true if there's data after the current window
func (swt *SlidingWindowTable) CanScrollDown() bool {
	return swt.hasMoreData
}

// GetStatusInfo returns information about the sliding window state
func (swt *SlidingWindowTable) GetStatusInfo() string {
	if swt.DataDroppedAtStart {
		startRow := swt.FirstRowIndex + 1
		endRow := swt.FirstRowIndex + int64(len(swt.Rows))
		return fmt.Sprintf("Showing rows %d-%d (earlier rows dropped due to memory limit)", 
			startRow, endRow)
	}
	return ""
}

// LoadMoreRows loads more rows from the iterator if available
func (swt *SlidingWindowTable) LoadMoreRows(maxRows int) int {
	if swt.iterator == nil || !swt.hasMoreData {
		return 0
	}
	
	// Cast iterator to *gocql.Iter
	iter, ok := swt.iterator.(*gocql.Iter)
	if !ok {
		logger.DebugToFile("SlidingWindowTable", "Iterator is not *gocql.Iter")
		return 0
	}
	
	loadedRows := 0
	for loadedRows < maxRows {
		rowMap := make(map[string]interface{})
		if !iter.MapScan(rowMap) {
			// No more data or error
			swt.hasMoreData = false
			if err := iter.Close(); err != nil {
				logger.DebugfToFile("SlidingWindowTable", "Iterator close error: %v", err)
			}
			swt.iterator = nil
			break
		}
		
		// Convert row to string array using column names
		row := make([]string, len(swt.ColumnNames))
		for i, colName := range swt.ColumnNames {
			if val, ok := rowMap[colName]; ok {
				if val == nil {
					row[i] = "null"
				} else {
					// Handle different types appropriately
					switch typed := val.(type) {
					case gocql.UUID:
						row[i] = typed.String()
					case []byte:
						row[i] = fmt.Sprintf("0x%x", typed)
					case time.Time:
						row[i] = typed.Format(time.RFC3339)
					default:
						row[i] = fmt.Sprintf("%v", val)
					}
				}
			} else {
				row[i] = "null"
			}
		}
		
		swt.AddRow(row)
		loadedRows++
	}
	
	logger.DebugfToFile("SlidingWindowTable", "Loaded %d more rows", loadedRows)
	return loadedRows
}

// GetUncapturedRows returns rows that haven't been written to capture file yet
func (swt *SlidingWindowTable) GetUncapturedRows() [][]string {
	if swt.LastCapturedRow >= swt.TotalRowsSeen {
		// All rows have been captured
		return nil
	}
	
	// Calculate how many uncaptured rows we have
	startIdx := swt.LastCapturedRow - swt.FirstRowIndex
	if startIdx < 0 {
		// All current rows are uncaptured
		return swt.Rows
	}
	
	if startIdx >= int64(len(swt.Rows)) {
		// No uncaptured rows in current window
		return nil
	}
	
	// Return the uncaptured portion
	return swt.Rows[startIdx:]
}

// MarkRowsAsCaptured updates the last captured row index
func (swt *SlidingWindowTable) MarkRowsAsCaptured(count int) {
	swt.LastCapturedRow += int64(count)
	if swt.LastCapturedRow > swt.TotalRowsSeen {
		swt.LastCapturedRow = swt.TotalRowsSeen
	}
}

// Reset clears the sliding window
func (swt *SlidingWindowTable) Reset() {
	swt.Rows = make([][]string, 0)
	swt.FirstRowIndex = 0
	swt.TotalRowsSeen = 0
	swt.CurrentMemory = 0
	swt.DataDroppedAtStart = false
	swt.DataAvailableAtEnd = false
	swt.iterator = nil
	swt.hasMoreData = false
	swt.LastCapturedRow = 0
}