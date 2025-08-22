package router

import (
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// CqlCommandVisitor is an interface for visiting CQL commands.
type CqlCommandVisitor interface {
	VisitDescribeCommand(ctx *grammar.DescribeCommandContext) interface{}
	VisitKwDescribe(ctx *grammar.KwDescribeContext) interface{}
	VisitConsistencyCommand(ctx *grammar.ConsistencyCommandContext) interface{}
	VisitRevoke(ctx *grammar.RevokeContext) interface{}
	VisitListRoles(ctx *grammar.ListRolesContext) interface{}
	VisitListPermissions(ctx *grammar.ListPermissionsContext) interface{}
	VisitGrant(ctx *grammar.GrantContext) interface{}
	VisitCreateUser(ctx *grammar.CreateUserContext) interface{}
	VisitCreateRole(ctx *grammar.CreateRoleContext) interface{}
	VisitCreateType(ctx *grammar.CreateTypeContext) interface{}
	VisitCreateTrigger(ctx *grammar.CreateTriggerContext) interface{}
	VisitCreateMaterializedView(ctx *grammar.CreateMaterializedViewContext) interface{}
	VisitCreateKeyspace(ctx *grammar.CreateKeyspaceContext) interface{}
	VisitCreateFunction(ctx *grammar.CreateFunctionContext) interface{}
	VisitCreateAggregate(ctx *grammar.CreateAggregateContext) interface{}
	VisitAlterUser(ctx *grammar.AlterUserContext) interface{}
	VisitAlterType(ctx *grammar.AlterTypeContext) interface{}
	VisitAlterTable(ctx *grammar.AlterTableContext) interface{}
	VisitAlterRole(ctx *grammar.AlterRoleContext) interface{}
	VisitAlterMaterializedView(ctx *grammar.AlterMaterializedViewContext) interface{}
	VisitDropUser(ctx *grammar.DropUserContext) interface{}
	VisitDropType(ctx *grammar.DropTypeContext) interface{}
	VisitDropMaterializedView(ctx *grammar.DropMaterializedViewContext) interface{}
	VisitDropAggregate(ctx *grammar.DropAggregateContext) interface{}
	VisitDropFunction(ctx *grammar.DropFunctionContext) interface{}
	VisitDropTrigger(ctx *grammar.DropTriggerContext) interface{}
	VisitDropRole(ctx *grammar.DropRoleContext) interface{}
	VisitDropTable(ctx *grammar.DropTableContext) interface{}
	VisitDropKeyspace(ctx *grammar.DropKeyspaceContext) interface{}
	VisitDropIndex(ctx *grammar.DropIndexContext) interface{}
	VisitCreateTable(ctx *grammar.CreateTableContext) interface{}
	VisitApplyBatch(ctx *grammar.ApplyBatchContext) interface{}
	VisitBeginBatch(ctx *grammar.BeginBatchContext) interface{}
	VisitAlterKeyspace(ctx *grammar.AlterKeyspaceContext) interface{}
	VisitUse_(ctx *grammar.Use_Context) interface{}
	VisitTruncate(ctx *grammar.TruncateContext) interface{}
	VisitCreateIndex(ctx *grammar.CreateIndexContext) interface{}
	VisitDelete_(ctx *grammar.Delete_Context) interface{}
	VisitUpdate(ctx *grammar.UpdateContext) interface{}
	VisitInsert(ctx *grammar.InsertContext) interface{}
	VisitSelect_(ctx *grammar.Select_Context) interface{}
}
