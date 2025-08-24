package router

import (
	"fmt"
	"strings"

	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// VisitCreateKeyspace handles CREATE KEYSPACE statements
func (v *CqlCommandVisitorImpl) VisitCreateKeyspace(ctx *grammar.CreateKeyspaceContext) interface{} {

	// Execute the CREATE KEYSPACE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE KEYSPACE failed: %v", err)
	}

	// Extract keyspace name for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	keyspaceName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "KEYSPACE" && i+1 < len(parts) {
			keyspaceName = parts[i+1]
			break
		}
	}

	if keyspaceName != "" {
		return fmt.Sprintf("Keyspace '%s' created successfully", keyspaceName)
	}
	return "CREATE KEYSPACE successful"
}

// VisitCreateTable handles CREATE TABLE statements
func (v *CqlCommandVisitorImpl) VisitCreateTable(ctx *grammar.CreateTableContext) interface{} {

	// Execute the CREATE TABLE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE TABLE failed: %v", err)
	}

	// Extract table name for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	tableName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "TABLE" && i+1 < len(parts) {
			tableName = parts[i+1]
			// Remove parenthesis if present
			if idx := strings.Index(tableName, "("); idx > 0 {
				tableName = tableName[:idx]
			}
			break
		}
	}

	if tableName != "" {
		return fmt.Sprintf("Table '%s' created successfully", tableName)
	}
	return "CREATE TABLE successful"
}

// VisitCreateIndex handles CREATE INDEX statements
func (v *CqlCommandVisitorImpl) VisitCreateIndex(ctx *grammar.CreateIndexContext) interface{} {

	// Execute the CREATE INDEX query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE INDEX failed: %v", err)
	}

	// Extract index name if provided
	text := ctx.GetText()
	parts := strings.Fields(text)
	indexName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "INDEX" && i+1 < len(parts) {
			nextPart := strings.ToUpper(parts[i+1])
			if nextPart != "ON" && nextPart != "IF" {
				indexName = parts[i+1]
			}
			break
		}
	}

	if indexName != "" {
		return fmt.Sprintf("Index '%s' created successfully", indexName)
	}
	return "CREATE INDEX successful"
}

// VisitCreateType handles CREATE TYPE statements
func (v *CqlCommandVisitorImpl) VisitCreateType(ctx *grammar.CreateTypeContext) interface{} {

	// Execute the CREATE TYPE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE TYPE failed: %v", err)
	}

	return "CREATE TYPE successful"
}

// VisitCreateFunction handles CREATE FUNCTION statements
func (v *CqlCommandVisitorImpl) VisitCreateFunction(ctx *grammar.CreateFunctionContext) interface{} {

	// Execute the CREATE FUNCTION query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE FUNCTION failed: %v", err)
	}

	return "CREATE FUNCTION successful"
}

// VisitCreateAggregate handles CREATE AGGREGATE statements
func (v *CqlCommandVisitorImpl) VisitCreateAggregate(ctx *grammar.CreateAggregateContext) interface{} {

	// Execute the CREATE AGGREGATE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE AGGREGATE failed: %v", err)
	}

	return "CREATE AGGREGATE successful"
}

// VisitCreateMaterializedView handles CREATE MATERIALIZED VIEW statements
func (v *CqlCommandVisitorImpl) VisitCreateMaterializedView(ctx *grammar.CreateMaterializedViewContext) interface{} {

	// Execute the CREATE MATERIALIZED VIEW query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE MATERIALIZED VIEW failed: %v", err)
	}

	return "CREATE MATERIALIZED VIEW successful"
}

// VisitCreateTrigger handles CREATE TRIGGER statements
func (v *CqlCommandVisitorImpl) VisitCreateTrigger(ctx *grammar.CreateTriggerContext) interface{} {

	// Execute the CREATE TRIGGER query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("CREATE TRIGGER failed: %v", err)
	}

	return "CREATE TRIGGER successful"
}

// VisitDropKeyspace handles DROP KEYSPACE statements
func (v *CqlCommandVisitorImpl) VisitDropKeyspace(ctx *grammar.DropKeyspaceContext) interface{} {

	// Execute the DROP KEYSPACE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP KEYSPACE failed: %v", err)
	}

	// Extract keyspace name for success message
	parts := strings.Fields(ctx.GetText())
	keyspaceName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "KEYSPACE" && i+1 < len(parts) {
			keyspaceName = parts[i+1]
			break
		}
	}

	if keyspaceName != "" {
		// Clear current keyspace if we just dropped it
		if v.session.CurrentKeyspace() == keyspaceName {
			v.session.SetKeyspace("")
		}
		return fmt.Sprintf("Keyspace '%s' dropped successfully", keyspaceName)
	}
	return "DROP KEYSPACE successful"
}

// VisitDropTable handles DROP TABLE statements
func (v *CqlCommandVisitorImpl) VisitDropTable(ctx *grammar.DropTableContext) interface{} {

	// Execute the DROP TABLE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP TABLE failed: %v", err)
	}

	parts := strings.Fields(ctx.GetText())
	tableName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "TABLE" && i+1 < len(parts) {
			tableName = parts[i+1]
			break
		}
	}

	if tableName != "" {
		return fmt.Sprintf("Table '%s' dropped successfully", tableName)
	}
	return "DROP TABLE successful"
}

// VisitDropIndex handles DROP INDEX statements
func (v *CqlCommandVisitorImpl) VisitDropIndex(ctx *grammar.DropIndexContext) interface{} {

	// Execute the DROP INDEX query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP INDEX failed: %v", err)
	}

	return "DROP INDEX successful"
}

// VisitDropType handles DROP TYPE statements
func (v *CqlCommandVisitorImpl) VisitDropType(ctx *grammar.DropTypeContext) interface{} {

	// Execute the DROP TYPE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP TYPE failed: %v", err)
	}

	return "DROP TYPE successful"
}

// VisitDropFunction handles DROP FUNCTION statements
func (v *CqlCommandVisitorImpl) VisitDropFunction(ctx *grammar.DropFunctionContext) interface{} {

	// Execute the DROP FUNCTION query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP FUNCTION failed: %v", err)
	}

	return "DROP FUNCTION successful"
}

// VisitDropAggregate handles DROP AGGREGATE statements
func (v *CqlCommandVisitorImpl) VisitDropAggregate(ctx *grammar.DropAggregateContext) interface{} {

	// Execute the DROP AGGREGATE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP AGGREGATE failed: %v", err)
	}

	return "DROP AGGREGATE successful"
}

// VisitDropMaterializedView handles DROP MATERIALIZED VIEW statements
func (v *CqlCommandVisitorImpl) VisitDropMaterializedView(ctx *grammar.DropMaterializedViewContext) interface{} {

	// Execute the DROP MATERIALIZED VIEW query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP MATERIALIZED VIEW failed: %v", err)
	}

	return "DROP MATERIALIZED VIEW successful"
}

// VisitDropTrigger handles DROP TRIGGER statements
func (v *CqlCommandVisitorImpl) VisitDropTrigger(ctx *grammar.DropTriggerContext) interface{} {

	// Execute the DROP TRIGGER query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("DROP TRIGGER failed: %v", err)
	}

	return "DROP TRIGGER successful"
}

// VisitAlterKeyspace handles ALTER KEYSPACE statements
func (v *CqlCommandVisitorImpl) VisitAlterKeyspace(ctx *grammar.AlterKeyspaceContext) interface{} {

	// Execute the ALTER KEYSPACE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("ALTER KEYSPACE failed: %v", err)
	}

	return "ALTER KEYSPACE successful"
}

// VisitAlterTable handles ALTER TABLE statements
func (v *CqlCommandVisitorImpl) VisitAlterTable(ctx *grammar.AlterTableContext) interface{} {

	// Execute the ALTER TABLE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("ALTER TABLE failed: %v", err)
	}

	return "ALTER TABLE successful"
}

// VisitAlterType handles ALTER TYPE statements
func (v *CqlCommandVisitorImpl) VisitAlterType(ctx *grammar.AlterTypeContext) interface{} {

	// Execute the ALTER TYPE query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("ALTER TYPE failed: %v", err)
	}

	return "ALTER TYPE successful"
}

// VisitAlterMaterializedView handles ALTER MATERIALIZED VIEW statements
func (v *CqlCommandVisitorImpl) VisitAlterMaterializedView(ctx *grammar.AlterMaterializedViewContext) interface{} {

	// Execute the ALTER MATERIALIZED VIEW query
	if err := v.session.Query(ctx.GetText()).Exec(); err != nil {
		return fmt.Errorf("ALTER MATERIALIZED VIEW failed: %v", err)
	}

	return "ALTER MATERIALIZED VIEW successful"
}
