package router

import (
	"fmt"
	"strings"
	
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// batchState is defined in batch_commands.go and shared across DML commands
// It maintains the state of batch operations

// VisitInsert handles INSERT statements
func (v *CqlCommandVisitorImpl) VisitInsert(ctx *grammar.InsertContext) interface{} {
	// Get the full text of the INSERT statement
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		batchState.batch.Query(query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("INSERT added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the INSERT query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("INSERT failed: %v", err)
	}
	
	// Return success message
	return "INSERT successful"
}

// VisitUpdate handles UPDATE statements
func (v *CqlCommandVisitorImpl) VisitUpdate(ctx *grammar.UpdateContext) interface{} {
	// Get the full text of the UPDATE statement
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		batchState.batch.Query(query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("UPDATE added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the UPDATE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("UPDATE failed: %v", err)
	}
	
	// Get the number of rows that might have been affected
	// Note: Cassandra doesn't return affected row count like traditional SQL databases
	return "UPDATE successful"
}

// VisitDelete_ handles DELETE statements
func (v *CqlCommandVisitorImpl) VisitDelete_(ctx *grammar.Delete_Context) interface{} {
	// Get the full text of the DELETE statement
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		batchState.batch.Query(query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("DELETE added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the DELETE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("DELETE failed: %v", err)
	}
	
	return "DELETE successful"
}

// VisitTruncate handles TRUNCATE statements
func (v *CqlCommandVisitorImpl) VisitTruncate(ctx *grammar.TruncateContext) interface{} {
	// Get the full text of the TRUNCATE statement
	query := ctx.GetText()
	
	// TRUNCATE is a dangerous operation - in a real shell, you might want to add confirmation
	// For now, we'll execute it directly
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("TRUNCATE failed: %v", err)
	}
	
	// Extract table name for the success message
	text := ctx.GetText()
	parts := strings.Fields(strings.ToUpper(text))
	tableName := ""
	if len(parts) >= 2 {
		tableName = parts[1]
	}
	
	if tableName != "" {
		return fmt.Sprintf("TRUNCATE successful - all data removed from %s", tableName)
	}
	return "TRUNCATE successful"
}

