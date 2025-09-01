package db

import (
	"fmt"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// ExecuteDMLCommand executes a DML command (INSERT, UPDATE, DELETE, TRUNCATE)
func (s *Session) ExecuteDMLCommand(query string) error {
	return s.Query(query).Exec()
}

// AddToBatch adds a query to the current batch
func (s *Session) AddToBatch(batch *gocql.Batch, query string) {
	if batch != nil {
		batch.Query(query)
	}
}

// CreateBatch creates a new batch with the specified type
func (s *Session) CreateBatch(batchType gocql.BatchType) *gocql.Batch {
	return s.Batch(batchType)
}

// ExecuteBatch executes a batch of statements
func (s *Session) ExecuteBatch(batch *gocql.Batch) error {
	if batch == nil {
		return fmt.Errorf("no batch to execute")
	}
	return batch.Exec()
}