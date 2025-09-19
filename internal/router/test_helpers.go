// +build !integration

package router

import (
	"fmt"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

// MockDBSession wraps MockSession to provide a db.Session compatible type
type MockDBSession struct {
	// Do not embed db.Session to avoid calling through to nil methods
	Session *db.Session
	mock    *MockSession
}

// MockSession implements a minimal mock for testing
type MockSession struct {
	queryResult interface{}
	rowCount    int
}

// NewMockSession creates a new mock session wrapped in db.Session compatible type
func NewMockSession(queryResult interface{}) *MockDBSession {
	return &MockDBSession{
		Session: &db.Session{},
		mock: &MockSession{
			queryResult: queryResult,
		},
	}
}

// NewMockSessionWithRowCount creates a new mock session with a specific row count
func NewMockSessionWithRowCount(rowCount int) *MockDBSession {
	return &MockDBSession{
		Session: &db.Session{},
		mock: &MockSession{
			rowCount: rowCount,
		},
	}
}

// IsConnected returns true for mock
func (m *MockDBSession) IsConnected() bool {
	return true
}

// ExecuteCQLQuery returns the mock query result
func (m *MockDBSession) ExecuteCQLQuery(query string) interface{} {
	if m.mock.queryResult != nil {
		return m.mock.queryResult
	}

	// Generate test data based on rowCount
	headers := []string{"id", "name", "value", "timestamp", "status"}
	columnTypes := []string{"int", "text", "double", "timestamp", "text"}

	data := make([][]string, m.mock.rowCount)
	rawData := make([]map[string]interface{}, m.mock.rowCount)

	for i := 0; i < m.mock.rowCount; i++ {
		data[i] = []string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("Name_%d", i),
			fmt.Sprintf("%.2f", float64(i)*1.5),
			time.Now().Format(time.RFC3339),
			"active",
		}
		// Use int32 directly to avoid overflow issues
		var id int32
		if i <= 2147483647 {
			id = int32(i) // #nosec G115 - bounds check ensures no overflow
		} else {
			id = int32(i % 2147483647) // #nosec G115 - modulo ensures value is in range
		}
		rawData[i] = map[string]interface{}{
			"id":        id,
			"name":      fmt.Sprintf("Name_%d", i),
			"value":     float64(i) * 1.5,
			"timestamp": time.Now(),
			"status":    "active",
		}
	}

	return db.QueryResult{
		Headers:     headers,
		ColumnTypes: columnTypes,
		Data:        data,
		RawData:     rawData,
	}
}

// ExecuteStreamingQuery returns the mock query result bypassing the Query method
func (m *MockDBSession) ExecuteStreamingQuery(query string) interface{} {
	// Directly return the result without calling through to Query
	return m.ExecuteCQLQuery(query)
}

// Query overrides to prevent nil pointer when Query is called
func (m *MockDBSession) Query(stmt string, values ...interface{}) interface{} {
	// Return a stub to prevent nil pointer - this shouldn't be called in tests
	return nil
}