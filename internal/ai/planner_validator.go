package ai

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/cluster"
)

// ValidateInsertPlan validates an INSERT plan against table schema
// Returns error if plan is invalid (missing partition keys, invalid columns, etc.)
func ValidateInsertPlan(plan *AIResult, metadata cluster.MetadataManager) error {
	if metadata == nil {
		// If no metadata available, skip validation
		return nil
	}

	if plan.Table == "" {
		return fmt.Errorf("table is required for INSERT")
	}

	if len(plan.Values) == 0 && !plan.InsertJSON {
		return fmt.Errorf("values are required for INSERT")
	}

	// Skip validation for INSERT JSON (different rules)
	if plan.InsertJSON {
		return nil
	}

	// Get table metadata
	tableMeta, err := metadata.GetTable(plan.Keyspace, plan.Table)
	if err != nil {
		// If we can't get metadata, skip validation (table might not exist yet for CREATE TABLE scenarios)
		return nil
	}

	if tableMeta == nil {
		// Table doesn't exist - might be valid for some scenarios
		return nil
	}

	// Validate all partition keys are present
	pkNames := tableMeta.GetPartitionKeyNames()
	for _, pkName := range pkNames {
		if _, exists := plan.Values[pkName]; !exists {
			return fmt.Errorf("missing partition key column: %s (required for INSERT)", pkName)
		}
	}

	// Validate all clustering keys are present
	ckNames := tableMeta.GetClusteringKeyNames()
	for _, ckName := range ckNames {
		if _, exists := plan.Values[ckName]; !exists {
			return fmt.Errorf("missing clustering key column: %s (required for INSERT)", ckName)
		}
	}

	return nil
}

// ValidateUpdatePlan validates an UPDATE plan against table schema
func ValidateUpdatePlan(plan *AIResult, metadata cluster.MetadataManager) error {
	if metadata == nil {
		return nil
	}

	if plan.Table == "" {
		return fmt.Errorf("table is required for UPDATE")
	}

	if len(plan.Values) == 0 && len(plan.CounterOps) == 0 && len(plan.CollectionOps) == 0 {
		return fmt.Errorf("no values to update")
	}

	if len(plan.Where) == 0 {
		return fmt.Errorf("WHERE clause is required for UPDATE")
	}

	// Get table metadata
	tableMeta, err := metadata.GetTable(plan.Keyspace, plan.Table)
	if err != nil {
		return nil // Skip validation if metadata unavailable
	}

	if tableMeta == nil {
		return nil // Table doesn't exist
	}

	// Check if update is modifying static columns only
	updatingOnlyStatic := true
	staticColumns := make(map[string]bool)
	for _, staticCol := range tableMeta.GetStaticColumns() {
		staticColumns[staticCol.Name] = true
	}

	// Check what columns are being updated
	for colName := range plan.Values {
		if !staticColumns[colName] {
			updatingOnlyStatic = false
			break
		}
	}

	// If updating regular columns, need full primary key in WHERE
	if !updatingOnlyStatic {
		// Extract WHERE columns
		whereColumns := make(map[string]bool)
		for _, w := range plan.Where {
			if w.Column != "" {
				whereColumns[w.Column] = true
			}
			// Handle tuple notation
			for _, col := range w.Columns {
				whereColumns[col] = true
			}
		}

		// Validate all partition keys present in WHERE
		pkNames := tableMeta.GetPartitionKeyNames()
		for _, pkName := range pkNames {
			if !whereColumns[pkName] {
				return fmt.Errorf("missing partition key in WHERE clause: %s (required for UPDATE)", pkName)
			}
		}

		// Validate all clustering keys present in WHERE
		ckNames := tableMeta.GetClusteringKeyNames()
		for _, ckName := range ckNames {
			if !whereColumns[ckName] {
				return fmt.Errorf("missing clustering key in WHERE clause: %s (required for UPDATE of regular columns)", ckName)
			}
		}
	} else {
		// Updating static columns only - just need partition key
		whereColumns := make(map[string]bool)
		for _, w := range plan.Where {
			if w.Column != "" {
				whereColumns[w.Column] = true
			}
			for _, col := range w.Columns {
				whereColumns[col] = true
			}
		}

		pkNames := tableMeta.GetPartitionKeyNames()
		for _, pkName := range pkNames {
			if !whereColumns[pkName] {
				return fmt.Errorf("missing partition key in WHERE clause: %s (required for UPDATE)", pkName)
			}
		}
		// Clustering keys NOT required for static column updates
	}

	return nil
}

// ValidateDeletePlan validates a DELETE plan against table schema
func ValidateDeletePlan(plan *AIResult, metadata cluster.MetadataManager) error {
	if metadata == nil {
		return nil
	}

	if plan.Table == "" {
		return fmt.Errorf("table is required for DELETE")
	}

	if len(plan.Where) == 0 {
		return fmt.Errorf("WHERE clause is required for DELETE (DELETE without WHERE is not allowed)")
	}

	// Get table metadata
	tableMeta, err := metadata.GetTable(plan.Keyspace, plan.Table)
	if err != nil {
		return nil
	}

	if tableMeta == nil {
		return nil
	}

	// Extract WHERE columns
	whereColumns := make(map[string]bool)
	for _, w := range plan.Where {
		if w.Column != "" {
			whereColumns[w.Column] = true
		}
		for _, col := range w.Columns {
			whereColumns[col] = true
		}
	}

	// DELETE requires at least partition key
	pkNames := tableMeta.GetPartitionKeyNames()
	hasPartitionKey := false
	for _, pkName := range pkNames {
		if whereColumns[pkName] {
			hasPartitionKey = true
		}
	}

	if !hasPartitionKey {
		return fmt.Errorf("WHERE clause must include at least one partition key column for DELETE")
	}

	// Clustering keys are optional for DELETE (allows range deletes)

	return nil
}

// ValidateBatchPlan validates a BATCH plan for cross-partition issues
func ValidateBatchPlan(plan *AIResult, metadata cluster.MetadataManager) error {
	if metadata == nil {
		return nil
	}

	if len(plan.BatchStatements) == 0 {
		return fmt.Errorf("BATCH must contain at least one statement")
	}

	// Check for counter/non-counter mixing
	hasCounter := false
	hasNonCounter := false
	batchType := strings.ToUpper(plan.BatchType)

	for i, stmt := range plan.BatchStatements {
		// Validate each statement
		switch strings.ToUpper(stmt.Operation) {
		case "INSERT":
			if err := ValidateInsertPlan(&stmt, metadata); err != nil {
				return fmt.Errorf("BATCH statement %d (INSERT): %w", i, err)
			}
			hasNonCounter = true

		case "UPDATE":
			if err := ValidateUpdatePlan(&stmt, metadata); err != nil {
				return fmt.Errorf("BATCH statement %d (UPDATE): %w", i, err)
			}

			// Check if updating counters
			if len(stmt.CounterOps) > 0 {
				hasCounter = true
			} else {
				hasNonCounter = true
			}

		case "DELETE":
			if err := ValidateDeletePlan(&stmt, metadata); err != nil {
				return fmt.Errorf("BATCH statement %d (DELETE): %w", i, err)
			}
			hasNonCounter = true

		default:
			return fmt.Errorf("BATCH statement %d: invalid operation %s (only INSERT, UPDATE, DELETE allowed)", i, stmt.Operation)
		}
	}

	// Validate counter mixing
	if hasCounter && hasNonCounter {
		return fmt.Errorf("cannot mix counter and non-counter operations in BATCH")
	}

	// If counter batch, must be COUNTER type
	if hasCounter && batchType != "COUNTER" {
		return fmt.Errorf("counter operations require BATCH type COUNTER")
	}

	// TODO: Add cross-partition detection
	// This requires extracting partition key values from each statement and comparing
	// For now, we'll skip this and add it in a follow-up commit

	return nil
}

// ValidatePlanWithMetadata validates a query plan using cluster metadata
// This is called BEFORE RenderCQL to catch errors early
func ValidatePlanWithMetadata(plan *AIResult, metadata cluster.MetadataManager) error {
	if plan == nil {
		return fmt.Errorf("plan cannot be nil")
	}

	if metadata == nil {
		// No metadata available - skip validation
		return nil
	}

	operation := strings.ToUpper(plan.Operation)

	switch operation {
	case "INSERT":
		return ValidateInsertPlan(plan, metadata)

	case "UPDATE":
		return ValidateUpdatePlan(plan, metadata)

	case "DELETE":
		return ValidateDeletePlan(plan, metadata)

	case "BATCH":
		return ValidateBatchPlan(plan, metadata)

	default:
		// Other operations don't need schema validation yet
		return nil
	}
}
