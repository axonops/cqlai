// Code generated from CqlParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package grammar // CqlParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by CqlParser.
type CqlParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by CqlParser#root.
	VisitRoot(ctx *RootContext) interface{}

	// Visit a parse tree produced by CqlParser#cqls.
	VisitCqls(ctx *CqlsContext) interface{}

	// Visit a parse tree produced by CqlParser#statementSeparator.
	VisitStatementSeparator(ctx *StatementSeparatorContext) interface{}

	// Visit a parse tree produced by CqlParser#empty_.
	VisitEmpty_(ctx *Empty_Context) interface{}

	// Visit a parse tree produced by CqlParser#cql.
	VisitCql(ctx *CqlContext) interface{}

	// Visit a parse tree produced by CqlParser#describeCommand.
	VisitDescribeCommand(ctx *DescribeCommandContext) interface{}

	// Visit a parse tree produced by CqlParser#consistencyCommand.
	VisitConsistencyCommand(ctx *ConsistencyCommandContext) interface{}

	// Visit a parse tree produced by CqlParser#revoke.
	VisitRevoke(ctx *RevokeContext) interface{}

	// Visit a parse tree produced by CqlParser#listRoles.
	VisitListRoles(ctx *ListRolesContext) interface{}

	// Visit a parse tree produced by CqlParser#listPermissions.
	VisitListPermissions(ctx *ListPermissionsContext) interface{}

	// Visit a parse tree produced by CqlParser#grant.
	VisitGrant(ctx *GrantContext) interface{}

	// Visit a parse tree produced by CqlParser#priviledge.
	VisitPriviledge(ctx *PriviledgeContext) interface{}

	// Visit a parse tree produced by CqlParser#resource.
	VisitResource(ctx *ResourceContext) interface{}

	// Visit a parse tree produced by CqlParser#createUser.
	VisitCreateUser(ctx *CreateUserContext) interface{}

	// Visit a parse tree produced by CqlParser#createRole.
	VisitCreateRole(ctx *CreateRoleContext) interface{}

	// Visit a parse tree produced by CqlParser#createType.
	VisitCreateType(ctx *CreateTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#typeMemberColumnList.
	VisitTypeMemberColumnList(ctx *TypeMemberColumnListContext) interface{}

	// Visit a parse tree produced by CqlParser#createTrigger.
	VisitCreateTrigger(ctx *CreateTriggerContext) interface{}

	// Visit a parse tree produced by CqlParser#createMaterializedView.
	VisitCreateMaterializedView(ctx *CreateMaterializedViewContext) interface{}

	// Visit a parse tree produced by CqlParser#materializedViewWhere.
	VisitMaterializedViewWhere(ctx *MaterializedViewWhereContext) interface{}

	// Visit a parse tree produced by CqlParser#columnNotNullList.
	VisitColumnNotNullList(ctx *ColumnNotNullListContext) interface{}

	// Visit a parse tree produced by CqlParser#columnNotNull.
	VisitColumnNotNull(ctx *ColumnNotNullContext) interface{}

	// Visit a parse tree produced by CqlParser#materializedViewOptions.
	VisitMaterializedViewOptions(ctx *MaterializedViewOptionsContext) interface{}

	// Visit a parse tree produced by CqlParser#createKeyspace.
	VisitCreateKeyspace(ctx *CreateKeyspaceContext) interface{}

	// Visit a parse tree produced by CqlParser#createFunction.
	VisitCreateFunction(ctx *CreateFunctionContext) interface{}

	// Visit a parse tree produced by CqlParser#codeBlock.
	VisitCodeBlock(ctx *CodeBlockContext) interface{}

	// Visit a parse tree produced by CqlParser#paramList.
	VisitParamList(ctx *ParamListContext) interface{}

	// Visit a parse tree produced by CqlParser#returnMode.
	VisitReturnMode(ctx *ReturnModeContext) interface{}

	// Visit a parse tree produced by CqlParser#createAggregate.
	VisitCreateAggregate(ctx *CreateAggregateContext) interface{}

	// Visit a parse tree produced by CqlParser#initCondDefinition.
	VisitInitCondDefinition(ctx *InitCondDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#initCondHash.
	VisitInitCondHash(ctx *InitCondHashContext) interface{}

	// Visit a parse tree produced by CqlParser#initCondHashItem.
	VisitInitCondHashItem(ctx *InitCondHashItemContext) interface{}

	// Visit a parse tree produced by CqlParser#initCondListNested.
	VisitInitCondListNested(ctx *InitCondListNestedContext) interface{}

	// Visit a parse tree produced by CqlParser#initCondList.
	VisitInitCondList(ctx *InitCondListContext) interface{}

	// Visit a parse tree produced by CqlParser#orReplace.
	VisitOrReplace(ctx *OrReplaceContext) interface{}

	// Visit a parse tree produced by CqlParser#alterUser.
	VisitAlterUser(ctx *AlterUserContext) interface{}

	// Visit a parse tree produced by CqlParser#userPassword.
	VisitUserPassword(ctx *UserPasswordContext) interface{}

	// Visit a parse tree produced by CqlParser#userSuperUser.
	VisitUserSuperUser(ctx *UserSuperUserContext) interface{}

	// Visit a parse tree produced by CqlParser#alterType.
	VisitAlterType(ctx *AlterTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeOperation.
	VisitAlterTypeOperation(ctx *AlterTypeOperationContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeRename.
	VisitAlterTypeRename(ctx *AlterTypeRenameContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeRenameList.
	VisitAlterTypeRenameList(ctx *AlterTypeRenameListContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeRenameItem.
	VisitAlterTypeRenameItem(ctx *AlterTypeRenameItemContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeAdd.
	VisitAlterTypeAdd(ctx *AlterTypeAddContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTypeAlterType.
	VisitAlterTypeAlterType(ctx *AlterTypeAlterTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTable.
	VisitAlterTable(ctx *AlterTableContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableOperation.
	VisitAlterTableOperation(ctx *AlterTableOperationContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableWith.
	VisitAlterTableWith(ctx *AlterTableWithContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableRename.
	VisitAlterTableRename(ctx *AlterTableRenameContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableDropCompactStorage.
	VisitAlterTableDropCompactStorage(ctx *AlterTableDropCompactStorageContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableDropColumns.
	VisitAlterTableDropColumns(ctx *AlterTableDropColumnsContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableDropColumnList.
	VisitAlterTableDropColumnList(ctx *AlterTableDropColumnListContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableAdd.
	VisitAlterTableAdd(ctx *AlterTableAddContext) interface{}

	// Visit a parse tree produced by CqlParser#alterTableColumnDefinition.
	VisitAlterTableColumnDefinition(ctx *AlterTableColumnDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#alterRole.
	VisitAlterRole(ctx *AlterRoleContext) interface{}

	// Visit a parse tree produced by CqlParser#roleWith.
	VisitRoleWith(ctx *RoleWithContext) interface{}

	// Visit a parse tree produced by CqlParser#roleWithOptions.
	VisitRoleWithOptions(ctx *RoleWithOptionsContext) interface{}

	// Visit a parse tree produced by CqlParser#alterMaterializedView.
	VisitAlterMaterializedView(ctx *AlterMaterializedViewContext) interface{}

	// Visit a parse tree produced by CqlParser#dropUser.
	VisitDropUser(ctx *DropUserContext) interface{}

	// Visit a parse tree produced by CqlParser#dropType.
	VisitDropType(ctx *DropTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#dropMaterializedView.
	VisitDropMaterializedView(ctx *DropMaterializedViewContext) interface{}

	// Visit a parse tree produced by CqlParser#dropAggregate.
	VisitDropAggregate(ctx *DropAggregateContext) interface{}

	// Visit a parse tree produced by CqlParser#dropFunction.
	VisitDropFunction(ctx *DropFunctionContext) interface{}

	// Visit a parse tree produced by CqlParser#dropTrigger.
	VisitDropTrigger(ctx *DropTriggerContext) interface{}

	// Visit a parse tree produced by CqlParser#dropRole.
	VisitDropRole(ctx *DropRoleContext) interface{}

	// Visit a parse tree produced by CqlParser#dropTable.
	VisitDropTable(ctx *DropTableContext) interface{}

	// Visit a parse tree produced by CqlParser#dropKeyspace.
	VisitDropKeyspace(ctx *DropKeyspaceContext) interface{}

	// Visit a parse tree produced by CqlParser#dropIndex.
	VisitDropIndex(ctx *DropIndexContext) interface{}

	// Visit a parse tree produced by CqlParser#createTable.
	VisitCreateTable(ctx *CreateTableContext) interface{}

	// Visit a parse tree produced by CqlParser#withElement.
	VisitWithElement(ctx *WithElementContext) interface{}

	// Visit a parse tree produced by CqlParser#tableOptions.
	VisitTableOptions(ctx *TableOptionsContext) interface{}

	// Visit a parse tree produced by CqlParser#clusteringOrder.
	VisitClusteringOrder(ctx *ClusteringOrderContext) interface{}

	// Visit a parse tree produced by CqlParser#tableOptionItem.
	VisitTableOptionItem(ctx *TableOptionItemContext) interface{}

	// Visit a parse tree produced by CqlParser#tableOptionName.
	VisitTableOptionName(ctx *TableOptionNameContext) interface{}

	// Visit a parse tree produced by CqlParser#tableOptionValue.
	VisitTableOptionValue(ctx *TableOptionValueContext) interface{}

	// Visit a parse tree produced by CqlParser#optionHash.
	VisitOptionHash(ctx *OptionHashContext) interface{}

	// Visit a parse tree produced by CqlParser#optionHashItem.
	VisitOptionHashItem(ctx *OptionHashItemContext) interface{}

	// Visit a parse tree produced by CqlParser#optionHashKey.
	VisitOptionHashKey(ctx *OptionHashKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#optionHashValue.
	VisitOptionHashValue(ctx *OptionHashValueContext) interface{}

	// Visit a parse tree produced by CqlParser#columnDefinitionList.
	VisitColumnDefinitionList(ctx *ColumnDefinitionListContext) interface{}

	// Visit a parse tree produced by CqlParser#columnDefinition.
	VisitColumnDefinition(ctx *ColumnDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#primaryKeyColumn.
	VisitPrimaryKeyColumn(ctx *PrimaryKeyColumnContext) interface{}

	// Visit a parse tree produced by CqlParser#primaryKeyElement.
	VisitPrimaryKeyElement(ctx *PrimaryKeyElementContext) interface{}

	// Visit a parse tree produced by CqlParser#primaryKeyDefinition.
	VisitPrimaryKeyDefinition(ctx *PrimaryKeyDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#singlePrimaryKey.
	VisitSinglePrimaryKey(ctx *SinglePrimaryKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#compoundKey.
	VisitCompoundKey(ctx *CompoundKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#compositeKey.
	VisitCompositeKey(ctx *CompositeKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#partitionKeyList.
	VisitPartitionKeyList(ctx *PartitionKeyListContext) interface{}

	// Visit a parse tree produced by CqlParser#clusteringKeyList.
	VisitClusteringKeyList(ctx *ClusteringKeyListContext) interface{}

	// Visit a parse tree produced by CqlParser#partitionKey.
	VisitPartitionKey(ctx *PartitionKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#clusteringKey.
	VisitClusteringKey(ctx *ClusteringKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#applyBatch.
	VisitApplyBatch(ctx *ApplyBatchContext) interface{}

	// Visit a parse tree produced by CqlParser#beginBatch.
	VisitBeginBatch(ctx *BeginBatchContext) interface{}

	// Visit a parse tree produced by CqlParser#batchType.
	VisitBatchType(ctx *BatchTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#alterKeyspace.
	VisitAlterKeyspace(ctx *AlterKeyspaceContext) interface{}

	// Visit a parse tree produced by CqlParser#replicationList.
	VisitReplicationList(ctx *ReplicationListContext) interface{}

	// Visit a parse tree produced by CqlParser#replicationListItem.
	VisitReplicationListItem(ctx *ReplicationListItemContext) interface{}

	// Visit a parse tree produced by CqlParser#durableWrites.
	VisitDurableWrites(ctx *DurableWritesContext) interface{}

	// Visit a parse tree produced by CqlParser#use_.
	VisitUse_(ctx *Use_Context) interface{}

	// Visit a parse tree produced by CqlParser#truncate.
	VisitTruncate(ctx *TruncateContext) interface{}

	// Visit a parse tree produced by CqlParser#createIndex.
	VisitCreateIndex(ctx *CreateIndexContext) interface{}

	// Visit a parse tree produced by CqlParser#indexName.
	VisitIndexName(ctx *IndexNameContext) interface{}

	// Visit a parse tree produced by CqlParser#indexColumnSpec.
	VisitIndexColumnSpec(ctx *IndexColumnSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#indexKeysSpec.
	VisitIndexKeysSpec(ctx *IndexKeysSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#indexEntriesSSpec.
	VisitIndexEntriesSSpec(ctx *IndexEntriesSSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#indexFullSpec.
	VisitIndexFullSpec(ctx *IndexFullSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#delete_.
	VisitDelete_(ctx *Delete_Context) interface{}

	// Visit a parse tree produced by CqlParser#deleteColumnList.
	VisitDeleteColumnList(ctx *DeleteColumnListContext) interface{}

	// Visit a parse tree produced by CqlParser#deleteColumnItem.
	VisitDeleteColumnItem(ctx *DeleteColumnItemContext) interface{}

	// Visit a parse tree produced by CqlParser#update.
	VisitUpdate(ctx *UpdateContext) interface{}

	// Visit a parse tree produced by CqlParser#ifSpec.
	VisitIfSpec(ctx *IfSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#ifConditionList.
	VisitIfConditionList(ctx *IfConditionListContext) interface{}

	// Visit a parse tree produced by CqlParser#ifCondition.
	VisitIfCondition(ctx *IfConditionContext) interface{}

	// Visit a parse tree produced by CqlParser#assignments.
	VisitAssignments(ctx *AssignmentsContext) interface{}

	// Visit a parse tree produced by CqlParser#assignmentElement.
	VisitAssignmentElement(ctx *AssignmentElementContext) interface{}

	// Visit a parse tree produced by CqlParser#assignmentSet.
	VisitAssignmentSet(ctx *AssignmentSetContext) interface{}

	// Visit a parse tree produced by CqlParser#assignmentMap.
	VisitAssignmentMap(ctx *AssignmentMapContext) interface{}

	// Visit a parse tree produced by CqlParser#assignmentList.
	VisitAssignmentList(ctx *AssignmentListContext) interface{}

	// Visit a parse tree produced by CqlParser#assignmentTuple.
	VisitAssignmentTuple(ctx *AssignmentTupleContext) interface{}

	// Visit a parse tree produced by CqlParser#insert.
	VisitInsert(ctx *InsertContext) interface{}

	// Visit a parse tree produced by CqlParser#usingTtlTimestamp.
	VisitUsingTtlTimestamp(ctx *UsingTtlTimestampContext) interface{}

	// Visit a parse tree produced by CqlParser#timestamp.
	VisitTimestamp(ctx *TimestampContext) interface{}

	// Visit a parse tree produced by CqlParser#ttl.
	VisitTtl(ctx *TtlContext) interface{}

	// Visit a parse tree produced by CqlParser#usingTimestampSpec.
	VisitUsingTimestampSpec(ctx *UsingTimestampSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#ifNotExist.
	VisitIfNotExist(ctx *IfNotExistContext) interface{}

	// Visit a parse tree produced by CqlParser#ifExist.
	VisitIfExist(ctx *IfExistContext) interface{}

	// Visit a parse tree produced by CqlParser#insertValuesSpec.
	VisitInsertValuesSpec(ctx *InsertValuesSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#insertColumnSpec.
	VisitInsertColumnSpec(ctx *InsertColumnSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#columnList.
	VisitColumnList(ctx *ColumnListContext) interface{}

	// Visit a parse tree produced by CqlParser#expressionList.
	VisitExpressionList(ctx *ExpressionListContext) interface{}

	// Visit a parse tree produced by CqlParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by CqlParser#select_.
	VisitSelect_(ctx *Select_Context) interface{}

	// Visit a parse tree produced by CqlParser#allowFilteringSpec.
	VisitAllowFilteringSpec(ctx *AllowFilteringSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#limitSpec.
	VisitLimitSpec(ctx *LimitSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#fromSpec.
	VisitFromSpec(ctx *FromSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#fromSpecElement.
	VisitFromSpecElement(ctx *FromSpecElementContext) interface{}

	// Visit a parse tree produced by CqlParser#orderSpec.
	VisitOrderSpec(ctx *OrderSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#orderSpecElement.
	VisitOrderSpecElement(ctx *OrderSpecElementContext) interface{}

	// Visit a parse tree produced by CqlParser#whereSpec.
	VisitWhereSpec(ctx *WhereSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#distinctSpec.
	VisitDistinctSpec(ctx *DistinctSpecContext) interface{}

	// Visit a parse tree produced by CqlParser#selectElements.
	VisitSelectElements(ctx *SelectElementsContext) interface{}

	// Visit a parse tree produced by CqlParser#selectElement.
	VisitSelectElement(ctx *SelectElementContext) interface{}

	// Visit a parse tree produced by CqlParser#relationElements.
	VisitRelationElements(ctx *RelationElementsContext) interface{}

	// Visit a parse tree produced by CqlParser#relationElement.
	VisitRelationElement(ctx *RelationElementContext) interface{}

	// Visit a parse tree produced by CqlParser#relalationContains.
	VisitRelalationContains(ctx *RelalationContainsContext) interface{}

	// Visit a parse tree produced by CqlParser#relalationContainsKey.
	VisitRelalationContainsKey(ctx *RelalationContainsKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) interface{}

	// Visit a parse tree produced by CqlParser#functionArgs.
	VisitFunctionArgs(ctx *FunctionArgsContext) interface{}

	// Visit a parse tree produced by CqlParser#constant.
	VisitConstant(ctx *ConstantContext) interface{}

	// Visit a parse tree produced by CqlParser#decimalLiteral.
	VisitDecimalLiteral(ctx *DecimalLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#floatLiteral.
	VisitFloatLiteral(ctx *FloatLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#stringLiteral.
	VisitStringLiteral(ctx *StringLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#booleanLiteral.
	VisitBooleanLiteral(ctx *BooleanLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#hexadecimalLiteral.
	VisitHexadecimalLiteral(ctx *HexadecimalLiteralContext) interface{}

	// Visit a parse tree produced by CqlParser#keyspace.
	VisitKeyspace(ctx *KeyspaceContext) interface{}

	// Visit a parse tree produced by CqlParser#table.
	VisitTable(ctx *TableContext) interface{}

	// Visit a parse tree produced by CqlParser#column.
	VisitColumn(ctx *ColumnContext) interface{}

	// Visit a parse tree produced by CqlParser#dataType.
	VisitDataType(ctx *DataTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#dataTypeName.
	VisitDataTypeName(ctx *DataTypeNameContext) interface{}

	// Visit a parse tree produced by CqlParser#dataTypeDefinition.
	VisitDataTypeDefinition(ctx *DataTypeDefinitionContext) interface{}

	// Visit a parse tree produced by CqlParser#orderDirection.
	VisitOrderDirection(ctx *OrderDirectionContext) interface{}

	// Visit a parse tree produced by CqlParser#role.
	VisitRole(ctx *RoleContext) interface{}

	// Visit a parse tree produced by CqlParser#trigger.
	VisitTrigger(ctx *TriggerContext) interface{}

	// Visit a parse tree produced by CqlParser#triggerClass.
	VisitTriggerClass(ctx *TriggerClassContext) interface{}

	// Visit a parse tree produced by CqlParser#materializedView.
	VisitMaterializedView(ctx *MaterializedViewContext) interface{}

	// Visit a parse tree produced by CqlParser#type_.
	VisitType_(ctx *Type_Context) interface{}

	// Visit a parse tree produced by CqlParser#aggregate.
	VisitAggregate(ctx *AggregateContext) interface{}

	// Visit a parse tree produced by CqlParser#function_.
	VisitFunction_(ctx *Function_Context) interface{}

	// Visit a parse tree produced by CqlParser#language.
	VisitLanguage(ctx *LanguageContext) interface{}

	// Visit a parse tree produced by CqlParser#user.
	VisitUser(ctx *UserContext) interface{}

	// Visit a parse tree produced by CqlParser#password.
	VisitPassword(ctx *PasswordContext) interface{}

	// Visit a parse tree produced by CqlParser#hashKey.
	VisitHashKey(ctx *HashKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#param.
	VisitParam(ctx *ParamContext) interface{}

	// Visit a parse tree produced by CqlParser#paramName.
	VisitParamName(ctx *ParamNameContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAdd.
	VisitKwAdd(ctx *KwAddContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAggregate.
	VisitKwAggregate(ctx *KwAggregateContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAggregates.
	VisitKwAggregates(ctx *KwAggregatesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAll.
	VisitKwAll(ctx *KwAllContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAllPermissions.
	VisitKwAllPermissions(ctx *KwAllPermissionsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAllow.
	VisitKwAllow(ctx *KwAllowContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAlter.
	VisitKwAlter(ctx *KwAlterContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAnd.
	VisitKwAnd(ctx *KwAndContext) interface{}

	// Visit a parse tree produced by CqlParser#kwApply.
	VisitKwApply(ctx *KwApplyContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAs.
	VisitKwAs(ctx *KwAsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAsc.
	VisitKwAsc(ctx *KwAscContext) interface{}

	// Visit a parse tree produced by CqlParser#kwAuthorize.
	VisitKwAuthorize(ctx *KwAuthorizeContext) interface{}

	// Visit a parse tree produced by CqlParser#kwBatch.
	VisitKwBatch(ctx *KwBatchContext) interface{}

	// Visit a parse tree produced by CqlParser#kwBegin.
	VisitKwBegin(ctx *KwBeginContext) interface{}

	// Visit a parse tree produced by CqlParser#kwBy.
	VisitKwBy(ctx *KwByContext) interface{}

	// Visit a parse tree produced by CqlParser#kwCalled.
	VisitKwCalled(ctx *KwCalledContext) interface{}

	// Visit a parse tree produced by CqlParser#kwCluster.
	VisitKwCluster(ctx *KwClusterContext) interface{}

	// Visit a parse tree produced by CqlParser#kwClustering.
	VisitKwClustering(ctx *KwClusteringContext) interface{}

	// Visit a parse tree produced by CqlParser#kwCompact.
	VisitKwCompact(ctx *KwCompactContext) interface{}

	// Visit a parse tree produced by CqlParser#kwConnection.
	VisitKwConnection(ctx *KwConnectionContext) interface{}

	// Visit a parse tree produced by CqlParser#kwConsistency.
	VisitKwConsistency(ctx *KwConsistencyContext) interface{}

	// Visit a parse tree produced by CqlParser#kwConsistencyLevel.
	VisitKwConsistencyLevel(ctx *KwConsistencyLevelContext) interface{}

	// Visit a parse tree produced by CqlParser#kwContains.
	VisitKwContains(ctx *KwContainsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwCreate.
	VisitKwCreate(ctx *KwCreateContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDelete.
	VisitKwDelete(ctx *KwDeleteContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDesc.
	VisitKwDesc(ctx *KwDescContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDescribe.
	VisitKwDescribe(ctx *KwDescribeContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDistinct.
	VisitKwDistinct(ctx *KwDistinctContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDrop.
	VisitKwDrop(ctx *KwDropContext) interface{}

	// Visit a parse tree produced by CqlParser#kwDurableWrites.
	VisitKwDurableWrites(ctx *KwDurableWritesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwEntries.
	VisitKwEntries(ctx *KwEntriesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwExecute.
	VisitKwExecute(ctx *KwExecuteContext) interface{}

	// Visit a parse tree produced by CqlParser#kwExists.
	VisitKwExists(ctx *KwExistsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFiltering.
	VisitKwFiltering(ctx *KwFilteringContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFinalfunc.
	VisitKwFinalfunc(ctx *KwFinalfuncContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFrom.
	VisitKwFrom(ctx *KwFromContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFull.
	VisitKwFull(ctx *KwFullContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFunction.
	VisitKwFunction(ctx *KwFunctionContext) interface{}

	// Visit a parse tree produced by CqlParser#kwFunctions.
	VisitKwFunctions(ctx *KwFunctionsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwGrant.
	VisitKwGrant(ctx *KwGrantContext) interface{}

	// Visit a parse tree produced by CqlParser#kwIf.
	VisitKwIf(ctx *KwIfContext) interface{}

	// Visit a parse tree produced by CqlParser#kwIn.
	VisitKwIn(ctx *KwInContext) interface{}

	// Visit a parse tree produced by CqlParser#kwIndex.
	VisitKwIndex(ctx *KwIndexContext) interface{}

	// Visit a parse tree produced by CqlParser#kwInitcond.
	VisitKwInitcond(ctx *KwInitcondContext) interface{}

	// Visit a parse tree produced by CqlParser#kwInput.
	VisitKwInput(ctx *KwInputContext) interface{}

	// Visit a parse tree produced by CqlParser#kwInsert.
	VisitKwInsert(ctx *KwInsertContext) interface{}

	// Visit a parse tree produced by CqlParser#kwInto.
	VisitKwInto(ctx *KwIntoContext) interface{}

	// Visit a parse tree produced by CqlParser#kwIs.
	VisitKwIs(ctx *KwIsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwJson.
	VisitKwJson(ctx *KwJsonContext) interface{}

	// Visit a parse tree produced by CqlParser#kwKey.
	VisitKwKey(ctx *KwKeyContext) interface{}

	// Visit a parse tree produced by CqlParser#kwKeys.
	VisitKwKeys(ctx *KwKeysContext) interface{}

	// Visit a parse tree produced by CqlParser#kwKeyspace.
	VisitKwKeyspace(ctx *KwKeyspaceContext) interface{}

	// Visit a parse tree produced by CqlParser#kwKeyspaces.
	VisitKwKeyspaces(ctx *KwKeyspacesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwLanguage.
	VisitKwLanguage(ctx *KwLanguageContext) interface{}

	// Visit a parse tree produced by CqlParser#kwLimit.
	VisitKwLimit(ctx *KwLimitContext) interface{}

	// Visit a parse tree produced by CqlParser#kwList.
	VisitKwList(ctx *KwListContext) interface{}

	// Visit a parse tree produced by CqlParser#kwLogged.
	VisitKwLogged(ctx *KwLoggedContext) interface{}

	// Visit a parse tree produced by CqlParser#kwLocalSerial.
	VisitKwLocalSerial(ctx *KwLocalSerialContext) interface{}

	// Visit a parse tree produced by CqlParser#kwLogin.
	VisitKwLogin(ctx *KwLoginContext) interface{}

	// Visit a parse tree produced by CqlParser#kwMaterialized.
	VisitKwMaterialized(ctx *KwMaterializedContext) interface{}

	// Visit a parse tree produced by CqlParser#kwModify.
	VisitKwModify(ctx *KwModifyContext) interface{}

	// Visit a parse tree produced by CqlParser#kwNosuperuser.
	VisitKwNosuperuser(ctx *KwNosuperuserContext) interface{}

	// Visit a parse tree produced by CqlParser#kwNorecursive.
	VisitKwNorecursive(ctx *KwNorecursiveContext) interface{}

	// Visit a parse tree produced by CqlParser#kwNot.
	VisitKwNot(ctx *KwNotContext) interface{}

	// Visit a parse tree produced by CqlParser#kwNull.
	VisitKwNull(ctx *KwNullContext) interface{}

	// Visit a parse tree produced by CqlParser#kwOf.
	VisitKwOf(ctx *KwOfContext) interface{}

	// Visit a parse tree produced by CqlParser#kwOn.
	VisitKwOn(ctx *KwOnContext) interface{}

	// Visit a parse tree produced by CqlParser#kwOptions.
	VisitKwOptions(ctx *KwOptionsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwOr.
	VisitKwOr(ctx *KwOrContext) interface{}

	// Visit a parse tree produced by CqlParser#kwOrder.
	VisitKwOrder(ctx *KwOrderContext) interface{}

	// Visit a parse tree produced by CqlParser#kwPassword.
	VisitKwPassword(ctx *KwPasswordContext) interface{}

	// Visit a parse tree produced by CqlParser#kwPrimary.
	VisitKwPrimary(ctx *KwPrimaryContext) interface{}

	// Visit a parse tree produced by CqlParser#kwRename.
	VisitKwRename(ctx *KwRenameContext) interface{}

	// Visit a parse tree produced by CqlParser#kwReplace.
	VisitKwReplace(ctx *KwReplaceContext) interface{}

	// Visit a parse tree produced by CqlParser#kwReplication.
	VisitKwReplication(ctx *KwReplicationContext) interface{}

	// Visit a parse tree produced by CqlParser#kwReturns.
	VisitKwReturns(ctx *KwReturnsContext) interface{}

	// Visit a parse tree produced by CqlParser#kwRole.
	VisitKwRole(ctx *KwRoleContext) interface{}

	// Visit a parse tree produced by CqlParser#kwRoles.
	VisitKwRoles(ctx *KwRolesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwSelect.
	VisitKwSelect(ctx *KwSelectContext) interface{}

	// Visit a parse tree produced by CqlParser#kwSerial.
	VisitKwSerial(ctx *KwSerialContext) interface{}

	// Visit a parse tree produced by CqlParser#kwSet.
	VisitKwSet(ctx *KwSetContext) interface{}

	// Visit a parse tree produced by CqlParser#kwSfunc.
	VisitKwSfunc(ctx *KwSfuncContext) interface{}

	// Visit a parse tree produced by CqlParser#kwStorage.
	VisitKwStorage(ctx *KwStorageContext) interface{}

	// Visit a parse tree produced by CqlParser#kwStype.
	VisitKwStype(ctx *KwStypeContext) interface{}

	// Visit a parse tree produced by CqlParser#kwSuperuser.
	VisitKwSuperuser(ctx *KwSuperuserContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTable.
	VisitKwTable(ctx *KwTableContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTables.
	VisitKwTables(ctx *KwTablesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTimestamp.
	VisitKwTimestamp(ctx *KwTimestampContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTo.
	VisitKwTo(ctx *KwToContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTrigger.
	VisitKwTrigger(ctx *KwTriggerContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTruncate.
	VisitKwTruncate(ctx *KwTruncateContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTtl.
	VisitKwTtl(ctx *KwTtlContext) interface{}

	// Visit a parse tree produced by CqlParser#kwType.
	VisitKwType(ctx *KwTypeContext) interface{}

	// Visit a parse tree produced by CqlParser#kwTypes.
	VisitKwTypes(ctx *KwTypesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwUnlogged.
	VisitKwUnlogged(ctx *KwUnloggedContext) interface{}

	// Visit a parse tree produced by CqlParser#kwUpdate.
	VisitKwUpdate(ctx *KwUpdateContext) interface{}

	// Visit a parse tree produced by CqlParser#kwUse.
	VisitKwUse(ctx *KwUseContext) interface{}

	// Visit a parse tree produced by CqlParser#kwUser.
	VisitKwUser(ctx *KwUserContext) interface{}

	// Visit a parse tree produced by CqlParser#kwUsing.
	VisitKwUsing(ctx *KwUsingContext) interface{}

	// Visit a parse tree produced by CqlParser#kwValues.
	VisitKwValues(ctx *KwValuesContext) interface{}

	// Visit a parse tree produced by CqlParser#kwView.
	VisitKwView(ctx *KwViewContext) interface{}

	// Visit a parse tree produced by CqlParser#kwWhere.
	VisitKwWhere(ctx *KwWhereContext) interface{}

	// Visit a parse tree produced by CqlParser#kwWith.
	VisitKwWith(ctx *KwWithContext) interface{}

	// Visit a parse tree produced by CqlParser#kwRevoke.
	VisitKwRevoke(ctx *KwRevokeContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketLr.
	VisitSyntaxBracketLr(ctx *SyntaxBracketLrContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketRr.
	VisitSyntaxBracketRr(ctx *SyntaxBracketRrContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketLc.
	VisitSyntaxBracketLc(ctx *SyntaxBracketLcContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketRc.
	VisitSyntaxBracketRc(ctx *SyntaxBracketRcContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketLa.
	VisitSyntaxBracketLa(ctx *SyntaxBracketLaContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketRa.
	VisitSyntaxBracketRa(ctx *SyntaxBracketRaContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketLs.
	VisitSyntaxBracketLs(ctx *SyntaxBracketLsContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxBracketRs.
	VisitSyntaxBracketRs(ctx *SyntaxBracketRsContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxComma.
	VisitSyntaxComma(ctx *SyntaxCommaContext) interface{}

	// Visit a parse tree produced by CqlParser#syntaxColon.
	VisitSyntaxColon(ctx *SyntaxColonContext) interface{}
}
