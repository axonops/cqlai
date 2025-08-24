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
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		v.session.AddToBatch(batchState.batch, query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("INSERT added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the INSERT query
	if err := v.session.ExecuteDMLCommand(query); err != nil {
		return fmt.Errorf("INSERT failed: %v", err)
	}
	
	return "INSERT successful"
}

// VisitUpdate handles UPDATE statements
func (v *CqlCommandVisitorImpl) VisitUpdate(ctx *grammar.UpdateContext) interface{} {
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		v.session.AddToBatch(batchState.batch, query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("UPDATE added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the UPDATE query
	if err := v.session.ExecuteDMLCommand(query); err != nil {
		return fmt.Errorf("UPDATE failed: %v", err)
	}
	
	return "UPDATE successful"
}

// VisitDelete_ handles DELETE statements
func (v *CqlCommandVisitorImpl) VisitDelete_(ctx *grammar.Delete_Context) interface{} {
	query := ctx.GetText()
	
	// Check if we're in batch mode
	if batchState != nil && batchState.isActive {
		v.session.AddToBatch(batchState.batch, query)
		batchState.statements = append(batchState.statements, query)
		return fmt.Sprintf("DELETE added to batch (statement %d)", len(batchState.statements))
	}
	
	// Execute the DELETE query
	if err := v.session.ExecuteDMLCommand(query); err != nil {
		return fmt.Errorf("DELETE failed: %v", err)
	}
	
	return "DELETE successful"
}

// VisitTruncate handles TRUNCATE statements
func (v *CqlCommandVisitorImpl) VisitTruncate(ctx *grammar.TruncateContext) interface{} {
	query := ctx.GetText()
	
	// TRUNCATE is a dangerous operation - in a real shell, you might want to add confirmation
	// For now, we'll execute it directly
	if err := v.session.ExecuteDMLCommand(query); err != nil {
		return fmt.Errorf("TRUNCATE failed: %v", err)
	}
	
	// Extract table name for the success message
	parts := strings.Fields(strings.ToUpper(query))
	tableName := ""
	if len(parts) >= 2 {
		tableName = parts[1]
	}
	
	if tableName != "" {
		return fmt.Sprintf("TRUNCATE successful - all data removed from %s", tableName)
	}
	return "TRUNCATE successful"
}

