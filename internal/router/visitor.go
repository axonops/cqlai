package router

import (
	"github.com/axonops/cqlai/internal/logger"
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// Visitor is a visitor for the CqlParser.
type Visitor struct {
	*grammar.BaseCqlParserVisitor
	commands CqlCommandVisitor
}

// NewVisitor creates a new Visitor.
func NewVisitor(commands CqlCommandVisitor) *Visitor {
	return &Visitor{
		BaseCqlParserVisitor: &grammar.BaseCqlParserVisitor{},
		commands:             commands,
	}
}

// VisitRoot visits the root of the parse tree.
func (v *Visitor) VisitRoot(ctx *grammar.RootContext) interface{} {
	if ctx.Cqls() != nil {
		return ctx.Cqls().Accept(v)
	}
	return ""
}

// VisitCqls visits a list of cql statements.
func (v *Visitor) VisitCqls(ctx *grammar.CqlsContext) interface{} {
	var result interface{}
	for _, cql := range ctx.AllCql() {
		result = cql.Accept(v)
	}
	return result
}

// VisitCql visits a cql statement.
func (v *Visitor) VisitCql(ctx *grammar.CqlContext) interface{} {
	// Add debug logging
	if ctx.DescribeCommand() != nil {
		logger.DebugfToFile("VisitCql", "Found DescribeCommand: %v", ctx.DescribeCommand().GetText())
	}
	if ctx.Revoke() != nil {
		return v.commands.VisitRevoke(ctx.Revoke().(*grammar.RevokeContext))
	}
	if ctx.ListRoles() != nil {
		return v.commands.VisitListRoles(ctx.ListRoles().(*grammar.ListRolesContext))
	}
	if ctx.ListPermissions() != nil {
		return v.commands.VisitListPermissions(ctx.ListPermissions().(*grammar.ListPermissionsContext))
	}
	if ctx.Grant() != nil {
		return v.commands.VisitGrant(ctx.Grant().(*grammar.GrantContext))
	}
	if ctx.CreateUser() != nil {
		return v.commands.VisitCreateUser(ctx.CreateUser().(*grammar.CreateUserContext))
	}
	if ctx.CreateRole() != nil {
		return v.commands.VisitCreateRole(ctx.CreateRole().(*grammar.CreateRoleContext))
	}
	if ctx.CreateType() != nil {
		return v.commands.VisitCreateType(ctx.CreateType().(*grammar.CreateTypeContext))
	}
	if ctx.CreateTrigger() != nil {
		return v.commands.VisitCreateTrigger(ctx.CreateTrigger().(*grammar.CreateTriggerContext))
	}
	if ctx.CreateMaterializedView() != nil {
		return v.commands.VisitCreateMaterializedView(ctx.CreateMaterializedView().(*grammar.CreateMaterializedViewContext))
	}
	if ctx.CreateKeyspace() != nil {
		return v.commands.VisitCreateKeyspace(ctx.CreateKeyspace().(*grammar.CreateKeyspaceContext))
	}
	if ctx.CreateFunction() != nil {
		return v.commands.VisitCreateFunction(ctx.CreateFunction().(*grammar.CreateFunctionContext))
	}
	if ctx.CreateAggregate() != nil {
		return v.commands.VisitCreateAggregate(ctx.CreateAggregate().(*grammar.CreateAggregateContext))
	}
	if ctx.AlterUser() != nil {
		return v.commands.VisitAlterUser(ctx.AlterUser().(*grammar.AlterUserContext))
	}
	if ctx.AlterType() != nil {
		return v.commands.VisitAlterType(ctx.AlterType().(*grammar.AlterTypeContext))
	}
	if ctx.AlterTable() != nil {
		return v.commands.VisitAlterTable(ctx.AlterTable().(*grammar.AlterTableContext))
	}
	if ctx.AlterRole() != nil {
		return v.commands.VisitAlterRole(ctx.AlterRole().(*grammar.AlterRoleContext))
	}
	if ctx.AlterMaterializedView() != nil {
		return v.commands.VisitAlterMaterializedView(ctx.AlterMaterializedView().(*grammar.AlterMaterializedViewContext))
	}
	if ctx.DropUser() != nil {
		return v.commands.VisitDropUser(ctx.DropUser().(*grammar.DropUserContext))
	}
	if ctx.DropType() != nil {
		return v.commands.VisitDropType(ctx.DropType().(*grammar.DropTypeContext))
	}
	if ctx.DropMaterializedView() != nil {
		return v.commands.VisitDropMaterializedView(ctx.DropMaterializedView().(*grammar.DropMaterializedViewContext))
	}
	if ctx.DropAggregate() != nil {
		return v.commands.VisitDropAggregate(ctx.DropAggregate().(*grammar.DropAggregateContext))
	}
	if ctx.DropFunction() != nil {
		return v.commands.VisitDropFunction(ctx.DropFunction().(*grammar.DropFunctionContext))
	}
	if ctx.DropTrigger() != nil {
		return v.commands.VisitDropTrigger(ctx.DropTrigger().(*grammar.DropTriggerContext))
	}
	if ctx.DropRole() != nil {
		return v.commands.VisitDropRole(ctx.DropRole().(*grammar.DropRoleContext))
	}
	if ctx.DropTable() != nil {
		return v.commands.VisitDropTable(ctx.DropTable().(*grammar.DropTableContext))
	}
	if ctx.DropKeyspace() != nil {
		return v.commands.VisitDropKeyspace(ctx.DropKeyspace().(*grammar.DropKeyspaceContext))
	}
	if ctx.DropIndex() != nil {
		return v.commands.VisitDropIndex(ctx.DropIndex().(*grammar.DropIndexContext))
	}
	if ctx.CreateTable() != nil {
		return v.commands.VisitCreateTable(ctx.CreateTable().(*grammar.CreateTableContext))
	}
	if ctx.ApplyBatch() != nil {
		return v.commands.VisitApplyBatch(ctx.ApplyBatch().(*grammar.ApplyBatchContext))
	}
	if ctx.DescribeCommand() != nil {
		return v.commands.VisitDescribeCommand(ctx.DescribeCommand().(*grammar.DescribeCommandContext))
	}
	if ctx.AlterKeyspace() != nil {
		return v.commands.VisitAlterKeyspace(ctx.AlterKeyspace().(*grammar.AlterKeyspaceContext))
	}
	if ctx.Use_() != nil {
		return v.commands.VisitUse_(ctx.Use_().(*grammar.Use_Context))
	}
	if ctx.Truncate() != nil {
		return v.commands.VisitTruncate(ctx.Truncate().(*grammar.TruncateContext))
	}
	if ctx.CreateIndex() != nil {
		return v.commands.VisitCreateIndex(ctx.CreateIndex().(*grammar.CreateIndexContext))
	}
	if ctx.Delete_() != nil {
		return v.commands.VisitDelete_(ctx.Delete_().(*grammar.Delete_Context))
	}
	if ctx.Update() != nil {
		return v.commands.VisitUpdate(ctx.Update().(*grammar.UpdateContext))
	}
	if ctx.Insert() != nil {
		return v.commands.VisitInsert(ctx.Insert().(*grammar.InsertContext))
	}
	if ctx.Select_() != nil {
		return v.commands.VisitSelect_(ctx.Select_().(*grammar.Select_Context))
	}
	// Handle the consistency command
	if ctx.ConsistencyCommand() != nil {
		return v.commands.VisitConsistencyCommand(ctx.ConsistencyCommand().(*grammar.ConsistencyCommandContext))
	}

	return v.VisitChildren(ctx)
}
