package router

import (
	"fmt"
	"github.com/axonops/cqlai/internal/logger"
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// VisitDescribeCommand visits a describe command.
// Note: DESCRIBE commands are client-side meta-commands, not part of the CQL protocol.
// They work by querying the system_schema tables, which is what cqlsh does internally.
func (v *CqlCommandVisitorImpl) VisitDescribeCommand(ctx *grammar.DescribeCommandContext) interface{} {
	logger.DebugToFile("VisitDescribeCommand", "Called")
	if ctx.KwKeyspaces() != nil {
		return v.describeKeyspaces()
	}

	if ctx.KwTables() != nil {
		return v.describeTables()
	}

	if ctx.KwCluster() != nil {
		return v.describeCluster()
	}

	if ctx.KwTypes() != nil {
		return v.describeTypes()
	}

	if ctx.KwFunctions() != nil {
		return v.describeFunctions()
	}

	if ctx.KwAggregates() != nil {
		return v.describeAggregates()
	}

	// DESCRIBE SCHEMA
	if ctx.KwSchema() != nil {
		return v.describeSchema()
	}

	// DESCRIBE TABLE [keyspace.]table
	if ctx.KwTable() != nil && ctx.Table() != nil {
		tableName := ctx.Table().GetText()
		logger.DebugToFile("VisitDescribeCommand", fmt.Sprintf("Table token text: '%s'", tableName))
		// Check if there's also a keyspace token
		if ctx.Keyspace() != nil {
			keyspaceName := ctx.Keyspace().GetText()
			logger.DebugToFile("VisitDescribeCommand", fmt.Sprintf("Keyspace token text: '%s'", keyspaceName))
			tableName = keyspaceName + "." + tableName
		}
		logger.DebugToFile("VisitDescribeCommand", fmt.Sprintf("Calling describeTable with: '%s'", tableName))
		// The describeTable function will handle splitting it if needed
		return v.describeTable(tableName)
	}

	// DESCRIBE KEYSPACE keyspace
	if ctx.KwKeyspace() != nil && ctx.Keyspace() != nil {
		return v.describeKeyspace(ctx.Keyspace().GetText())
	}

	// DESCRIBE INDEX index_name
	if ctx.IndexName() != nil {
		return v.describeIndex(ctx.IndexName().GetText())
	}

	// DESCRIBE MATERIALIZED VIEW view_name
	if ctx.KwMaterialized() != nil && ctx.KwView() != nil {
		// Need to get the view name - this might need adjustment based on grammar
		return v.describeMaterializedView("")
	}

	// DESCRIBE TYPE [keyspace.]type
	if ctx.KwType() != nil && ctx.Type_() != nil {
		typeName := ctx.Type_().GetText()
		if ctx.Keyspace() != nil {
			typeName = ctx.Keyspace().GetText() + "." + typeName
		}
		return v.describeType(typeName)
	}

	// DESCRIBE FUNCTION [keyspace.]function
	if ctx.KwFunction() != nil && ctx.Function_() != nil {
		functionName := ctx.Function_().GetText()
		if ctx.Keyspace() != nil {
			functionName = ctx.Keyspace().GetText() + "." + functionName
		}
		return v.describeFunction(functionName)
	}

	// DESCRIBE AGGREGATE [keyspace.]aggregate
	if ctx.KwAggregate() != nil && ctx.Aggregate() != nil {
		aggregateName := ctx.Aggregate().GetText()
		if ctx.Keyspace() != nil {
			aggregateName = ctx.Keyspace().GetText() + "." + aggregateName
		}
		return v.describeAggregate(aggregateName)
	}

	return "DESCRIBE command not yet implemented for this object. Supported: KEYSPACES, TABLES, CLUSTER, SCHEMA, TABLE <name>, KEYSPACE <name>, TYPES, FUNCTIONS, INDEX <name>, VIEW <name>, TYPE <name>, FUNCTION <name>, AGGREGATES, AGGREGATE <name>"
}
