// Code generated from CqlParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package grammar // CqlParser
import "github.com/antlr4-go/antlr/v4"

type BaseCqlParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseCqlParserVisitor) VisitRoot(ctx *RootContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCqls(ctx *CqlsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitStatementSeparator(ctx *StatementSeparatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitEmpty_(ctx *Empty_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCql(ctx *CqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDescribeCommand(ctx *DescribeCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitConsistencyCommand(ctx *ConsistencyCommandContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRevoke(ctx *RevokeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitListRoles(ctx *ListRolesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitListPermissions(ctx *ListPermissionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitGrant(ctx *GrantContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPriviledge(ctx *PriviledgeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitResource(ctx *ResourceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateUser(ctx *CreateUserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateRole(ctx *CreateRoleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateType(ctx *CreateTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTypeMemberColumnList(ctx *TypeMemberColumnListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateTrigger(ctx *CreateTriggerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateMaterializedView(ctx *CreateMaterializedViewContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitMaterializedViewWhere(ctx *MaterializedViewWhereContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumnNotNullList(ctx *ColumnNotNullListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumnNotNull(ctx *ColumnNotNullContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitMaterializedViewOptions(ctx *MaterializedViewOptionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateKeyspace(ctx *CreateKeyspaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateFunction(ctx *CreateFunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCodeBlock(ctx *CodeBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitParamList(ctx *ParamListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitReturnMode(ctx *ReturnModeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateAggregate(ctx *CreateAggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInitCondDefinition(ctx *InitCondDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInitCondHash(ctx *InitCondHashContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInitCondHashItem(ctx *InitCondHashItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInitCondListNested(ctx *InitCondListNestedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInitCondList(ctx *InitCondListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOrReplace(ctx *OrReplaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterUser(ctx *AlterUserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUserPassword(ctx *UserPasswordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUserSuperUser(ctx *UserSuperUserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterType(ctx *AlterTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeOperation(ctx *AlterTypeOperationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeRename(ctx *AlterTypeRenameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeRenameList(ctx *AlterTypeRenameListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeRenameItem(ctx *AlterTypeRenameItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeAdd(ctx *AlterTypeAddContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTypeAlterType(ctx *AlterTypeAlterTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTable(ctx *AlterTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableOperation(ctx *AlterTableOperationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableWith(ctx *AlterTableWithContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableRename(ctx *AlterTableRenameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableDropCompactStorage(ctx *AlterTableDropCompactStorageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableDropColumns(ctx *AlterTableDropColumnsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableDropColumnList(ctx *AlterTableDropColumnListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableAdd(ctx *AlterTableAddContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterTableColumnDefinition(ctx *AlterTableColumnDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterRole(ctx *AlterRoleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRoleWith(ctx *RoleWithContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRoleWithOptions(ctx *RoleWithOptionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterMaterializedView(ctx *AlterMaterializedViewContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropUser(ctx *DropUserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropType(ctx *DropTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropMaterializedView(ctx *DropMaterializedViewContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropAggregate(ctx *DropAggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropFunction(ctx *DropFunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropTrigger(ctx *DropTriggerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropRole(ctx *DropRoleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropTable(ctx *DropTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropKeyspace(ctx *DropKeyspaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDropIndex(ctx *DropIndexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateTable(ctx *CreateTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitWithElement(ctx *WithElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTableOptions(ctx *TableOptionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitClusteringOrder(ctx *ClusteringOrderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTableOptionItem(ctx *TableOptionItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTableOptionName(ctx *TableOptionNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTableOptionValue(ctx *TableOptionValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOptionHash(ctx *OptionHashContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOptionHashItem(ctx *OptionHashItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOptionHashKey(ctx *OptionHashKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOptionHashValue(ctx *OptionHashValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumnDefinitionList(ctx *ColumnDefinitionListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumnDefinition(ctx *ColumnDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPrimaryKeyColumn(ctx *PrimaryKeyColumnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPrimaryKeyElement(ctx *PrimaryKeyElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPrimaryKeyDefinition(ctx *PrimaryKeyDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSinglePrimaryKey(ctx *SinglePrimaryKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCompoundKey(ctx *CompoundKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCompositeKey(ctx *CompositeKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPartitionKeyList(ctx *PartitionKeyListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitClusteringKeyList(ctx *ClusteringKeyListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPartitionKey(ctx *PartitionKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitClusteringKey(ctx *ClusteringKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitApplyBatch(ctx *ApplyBatchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitBeginBatch(ctx *BeginBatchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitBatchType(ctx *BatchTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAlterKeyspace(ctx *AlterKeyspaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitReplicationList(ctx *ReplicationListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitReplicationListItem(ctx *ReplicationListItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDurableWrites(ctx *DurableWritesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUse_(ctx *Use_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTruncate(ctx *TruncateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitCreateIndex(ctx *CreateIndexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIndexName(ctx *IndexNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIndexColumnSpec(ctx *IndexColumnSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIndexKeysSpec(ctx *IndexKeysSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIndexEntriesSSpec(ctx *IndexEntriesSSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIndexFullSpec(ctx *IndexFullSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDelete_(ctx *Delete_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDeleteColumnList(ctx *DeleteColumnListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDeleteColumnItem(ctx *DeleteColumnItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUpdate(ctx *UpdateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIfSpec(ctx *IfSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIfConditionList(ctx *IfConditionListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIfCondition(ctx *IfConditionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignments(ctx *AssignmentsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignmentElement(ctx *AssignmentElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignmentSet(ctx *AssignmentSetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignmentMap(ctx *AssignmentMapContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignmentList(ctx *AssignmentListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAssignmentTuple(ctx *AssignmentTupleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInsert(ctx *InsertContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUsingTtlTimestamp(ctx *UsingTtlTimestampContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTimestamp(ctx *TimestampContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTtl(ctx *TtlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUsingTimestampSpec(ctx *UsingTimestampSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIfNotExist(ctx *IfNotExistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitIfExist(ctx *IfExistContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInsertValuesSpec(ctx *InsertValuesSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitInsertColumnSpec(ctx *InsertColumnSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumnList(ctx *ColumnListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitExpressionList(ctx *ExpressionListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSelect_(ctx *Select_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAllowFilteringSpec(ctx *AllowFilteringSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitLimitSpec(ctx *LimitSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFromSpec(ctx *FromSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFromSpecElement(ctx *FromSpecElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOrderSpec(ctx *OrderSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOrderSpecElement(ctx *OrderSpecElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitWhereSpec(ctx *WhereSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDistinctSpec(ctx *DistinctSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSelectElements(ctx *SelectElementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSelectElement(ctx *SelectElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRelationElements(ctx *RelationElementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRelationElement(ctx *RelationElementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRelalationContains(ctx *RelalationContainsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRelalationContainsKey(ctx *RelalationContainsKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFunctionArgs(ctx *FunctionArgsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitConstant(ctx *ConstantContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDecimalLiteral(ctx *DecimalLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFloatLiteral(ctx *FloatLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitStringLiteral(ctx *StringLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitHexadecimalLiteral(ctx *HexadecimalLiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKeyspace(ctx *KeyspaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTable(ctx *TableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitColumn(ctx *ColumnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDataType(ctx *DataTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDataTypeName(ctx *DataTypeNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitDataTypeDefinition(ctx *DataTypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitOrderDirection(ctx *OrderDirectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitRole(ctx *RoleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTrigger(ctx *TriggerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitTriggerClass(ctx *TriggerClassContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitMaterializedView(ctx *MaterializedViewContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitType_(ctx *Type_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitAggregate(ctx *AggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitFunction_(ctx *Function_Context) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitLanguage(ctx *LanguageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitUser(ctx *UserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitPassword(ctx *PasswordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitHashKey(ctx *HashKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitParam(ctx *ParamContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitParamName(ctx *ParamNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAdd(ctx *KwAddContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAggregate(ctx *KwAggregateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAggregates(ctx *KwAggregatesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAll(ctx *KwAllContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAllPermissions(ctx *KwAllPermissionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAllow(ctx *KwAllowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAlter(ctx *KwAlterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAnd(ctx *KwAndContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwApply(ctx *KwApplyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAs(ctx *KwAsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAsc(ctx *KwAscContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwAuthorize(ctx *KwAuthorizeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwBatch(ctx *KwBatchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwBegin(ctx *KwBeginContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwBy(ctx *KwByContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwCalled(ctx *KwCalledContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwCluster(ctx *KwClusterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwClustering(ctx *KwClusteringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwCompact(ctx *KwCompactContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwConnection(ctx *KwConnectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwConsistency(ctx *KwConsistencyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwConsistencyLevel(ctx *KwConsistencyLevelContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwContains(ctx *KwContainsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwCreate(ctx *KwCreateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDelete(ctx *KwDeleteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDesc(ctx *KwDescContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDescribe(ctx *KwDescribeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDistinct(ctx *KwDistinctContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDrop(ctx *KwDropContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwDurableWrites(ctx *KwDurableWritesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwEntries(ctx *KwEntriesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwExecute(ctx *KwExecuteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwExists(ctx *KwExistsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFiltering(ctx *KwFilteringContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFinalfunc(ctx *KwFinalfuncContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFrom(ctx *KwFromContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFull(ctx *KwFullContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFunction(ctx *KwFunctionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwFunctions(ctx *KwFunctionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwGrant(ctx *KwGrantContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwIf(ctx *KwIfContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwIn(ctx *KwInContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwIndex(ctx *KwIndexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwInitcond(ctx *KwInitcondContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwInput(ctx *KwInputContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwInsert(ctx *KwInsertContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwInto(ctx *KwIntoContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwIs(ctx *KwIsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwJson(ctx *KwJsonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwKey(ctx *KwKeyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwKeys(ctx *KwKeysContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwKeyspace(ctx *KwKeyspaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwKeyspaces(ctx *KwKeyspacesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwLanguage(ctx *KwLanguageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwLimit(ctx *KwLimitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwList(ctx *KwListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwLogged(ctx *KwLoggedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwLocalSerial(ctx *KwLocalSerialContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwLogin(ctx *KwLoginContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwMaterialized(ctx *KwMaterializedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwModify(ctx *KwModifyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwNosuperuser(ctx *KwNosuperuserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwNorecursive(ctx *KwNorecursiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwNot(ctx *KwNotContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwNull(ctx *KwNullContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwOf(ctx *KwOfContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwOn(ctx *KwOnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwOptions(ctx *KwOptionsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwOr(ctx *KwOrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwOrder(ctx *KwOrderContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwPassword(ctx *KwPasswordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwPrimary(ctx *KwPrimaryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwRename(ctx *KwRenameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwReplace(ctx *KwReplaceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwReplication(ctx *KwReplicationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwReturns(ctx *KwReturnsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwRole(ctx *KwRoleContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwRoles(ctx *KwRolesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSchema(ctx *KwSchemaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSelect(ctx *KwSelectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSerial(ctx *KwSerialContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSet(ctx *KwSetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSfunc(ctx *KwSfuncContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwStorage(ctx *KwStorageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwStype(ctx *KwStypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwSuperuser(ctx *KwSuperuserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTable(ctx *KwTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTables(ctx *KwTablesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTimestamp(ctx *KwTimestampContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTo(ctx *KwToContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTrigger(ctx *KwTriggerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTruncate(ctx *KwTruncateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTtl(ctx *KwTtlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwType(ctx *KwTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwTypes(ctx *KwTypesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwUnlogged(ctx *KwUnloggedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwUpdate(ctx *KwUpdateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwUse(ctx *KwUseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwUser(ctx *KwUserContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwUsing(ctx *KwUsingContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwValues(ctx *KwValuesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwView(ctx *KwViewContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwWhere(ctx *KwWhereContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwWith(ctx *KwWithContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitKwRevoke(ctx *KwRevokeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketLr(ctx *SyntaxBracketLrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketRr(ctx *SyntaxBracketRrContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketLc(ctx *SyntaxBracketLcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketRc(ctx *SyntaxBracketRcContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketLa(ctx *SyntaxBracketLaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketRa(ctx *SyntaxBracketRaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketLs(ctx *SyntaxBracketLsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxBracketRs(ctx *SyntaxBracketRsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxComma(ctx *SyntaxCommaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCqlParserVisitor) VisitSyntaxColon(ctx *SyntaxColonContext) interface{} {
	return v.VisitChildren(ctx)
}
