package router

import (
	"fmt"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// BatchState maintains the state of batch operations
type BatchState struct {
	batch      *gocql.Batch
	isActive   bool
	batchType  string   // LOGGED, UNLOGGED, or COUNTER
	statements []string // Keep track of statements added to batch
}

// batchState is a session-level batch state
// In a real implementation, this might be stored in the session or model
var batchState = &BatchState{}

// VisitBeginBatch handles BEGIN BATCH statements
func (v *CqlCommandVisitorImpl) VisitBeginBatch(ctx *grammar.BeginBatchContext) interface{} {
	// Check if a batch is already active
	if batchState.isActive {
		return fmt.Errorf("batch already in progress - use APPLY BATCH to execute or discard the current batch")
	}

	// Parse the batch type
	text := strings.ToUpper(ctx.GetText())
	batchType := gocql.LoggedBatch // default
	batchTypeStr := "LOGGED"

	if strings.Contains(text, "UNLOGGED") {
		batchType = gocql.UnloggedBatch
		batchTypeStr = "UNLOGGED"
	} else if strings.Contains(text, "COUNTER") {
		batchType = gocql.CounterBatch
		batchTypeStr = "COUNTER"
	}

	// Create a new batch
	batchState.batch = v.session.CreateBatch(batchType)
	batchState.isActive = true
	batchState.batchType = batchTypeStr
	batchState.statements = []string{}

	return fmt.Sprintf("started %s batch - add statements and use APPLY BATCH when ready", batchTypeStr)
}

// VisitApplyBatch handles APPLY BATCH statements
func (v *CqlCommandVisitorImpl) VisitApplyBatch(ctx *grammar.ApplyBatchContext) interface{} {
	// Check if a batch is active
	if !batchState.isActive {
		return fmt.Errorf("no batch in progress - use BEGIN BATCH to start a batch")
	}

	// Execute the batch
	if err := v.session.ExecuteBatch(batchState.batch); err != nil {
		// Reset batch state even on error
		batchState.isActive = false
		batchState.batch = nil
		batchState.statements = []string{}
		return fmt.Errorf("execution of APPLY BATCH failed: %v", err)
	}

	// Get the number of statements that were executed
	stmtCount := len(batchState.statements)

	// Reset batch state
	batchState.isActive = false
	batchState.batch = nil
	batchState.statements = []string{}

	return fmt.Sprintf("batch applied successfully (%d statements executed)", stmtCount)
}

// AddToBatch adds a statement to the current batch if one is active
func (v *CqlCommandVisitorImpl) AddToBatch(query string) (bool, error) {
	if !batchState.isActive {
		return false, nil // No batch active, statement should be executed normally
	}

	// Add the query to the batch
	v.session.AddToBatch(batchState.batch, query)
	batchState.statements = append(batchState.statements, query)

	return true, nil // Statement was added to batch
}

// Modified INSERT to support batch operations
func (v *CqlCommandVisitorImpl) VisitInsertBatch(ctx *grammar.InsertContext) interface{} {
	query := ctx.GetText()

	// Check if we should add to batch
	if added, err := v.AddToBatch(query); err != nil {
		return fmt.Errorf("failed to add INSERT to batch: %v", err)
	} else if added {
		return fmt.Sprintf("INSERT added to batch (statement %d)", len(batchState.statements))
	}

	// Execute normally if not in batch mode
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("exectution of INSERT failed: %v", err)
	}

	return "INSERT successful"
}

// Modified UPDATE to support batch operations
func (v *CqlCommandVisitorImpl) VisitUpdateBatch(ctx *grammar.UpdateContext) interface{} {
	query := ctx.GetText()

	// Check if we should add to batch
	if added, err := v.AddToBatch(query); err != nil {
		return fmt.Errorf("failed to add UPDATE to batch: %v", err)
	} else if added {
		return fmt.Sprintf("UPDATE added to batch (statement %d)", len(batchState.statements))
	}

	// Execute normally if not in batch mode
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("execution of UPDATE failed: %v", err)
	}

	return "UPDATE successful"
}

// Modified DELETE to support batch operations
func (v *CqlCommandVisitorImpl) VisitDelete_Batch(ctx *grammar.Delete_Context) interface{} {
	query := ctx.GetText()

	// Check if we should add to batch
	if added, err := v.AddToBatch(query); err != nil {
		return fmt.Errorf("failed to add DELETE to batch: %v", err)
	} else if added {
		return fmt.Sprintf("DELETE added to batch (statement %d)", len(batchState.statements))
	}

	// Execute normally if not in batch mode
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("execution of DELETE failed: %v", err)
	}

	return "DELETE successful"
}

// GetBatchStatus returns the current batch status
func (v *CqlCommandVisitorImpl) GetBatchStatus() string {
	if !batchState.isActive {
		return "No batch in progress"
	}

	return fmt.Sprintf("%s batch in progress with %d statement(s)",
		batchState.batchType, len(batchState.statements))
}

// DiscardBatch discards the current batch without executing
func (v *CqlCommandVisitorImpl) DiscardBatch() string {
	if !batchState.isActive {
		return "No batch to discard"
	}

	stmtCount := len(batchState.statements)
	batchState.isActive = false
	batchState.batch = nil
	batchState.statements = []string{}

	return fmt.Sprintf("Batch discarded (%d statements discarded)", stmtCount)
}
