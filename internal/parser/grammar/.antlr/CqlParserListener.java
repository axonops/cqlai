// Generated from /home/hayato/git/cqlai/internal/parser/grammar/CqlParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link CqlParser}.
 */
public interface CqlParserListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link CqlParser#root}.
	 * @param ctx the parse tree
	 */
	void enterRoot(CqlParser.RootContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#root}.
	 * @param ctx the parse tree
	 */
	void exitRoot(CqlParser.RootContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#cqls}.
	 * @param ctx the parse tree
	 */
	void enterCqls(CqlParser.CqlsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#cqls}.
	 * @param ctx the parse tree
	 */
	void exitCqls(CqlParser.CqlsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#statementSeparator}.
	 * @param ctx the parse tree
	 */
	void enterStatementSeparator(CqlParser.StatementSeparatorContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#statementSeparator}.
	 * @param ctx the parse tree
	 */
	void exitStatementSeparator(CqlParser.StatementSeparatorContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#empty_}.
	 * @param ctx the parse tree
	 */
	void enterEmpty_(CqlParser.Empty_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#empty_}.
	 * @param ctx the parse tree
	 */
	void exitEmpty_(CqlParser.Empty_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#cql}.
	 * @param ctx the parse tree
	 */
	void enterCql(CqlParser.CqlContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#cql}.
	 * @param ctx the parse tree
	 */
	void exitCql(CqlParser.CqlContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#describeCommand}.
	 * @param ctx the parse tree
	 */
	void enterDescribeCommand(CqlParser.DescribeCommandContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#describeCommand}.
	 * @param ctx the parse tree
	 */
	void exitDescribeCommand(CqlParser.DescribeCommandContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#consistencyCommand}.
	 * @param ctx the parse tree
	 */
	void enterConsistencyCommand(CqlParser.ConsistencyCommandContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#consistencyCommand}.
	 * @param ctx the parse tree
	 */
	void exitConsistencyCommand(CqlParser.ConsistencyCommandContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#outputFormatCommand}.
	 * @param ctx the parse tree
	 */
	void enterOutputFormatCommand(CqlParser.OutputFormatCommandContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#outputFormatCommand}.
	 * @param ctx the parse tree
	 */
	void exitOutputFormatCommand(CqlParser.OutputFormatCommandContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#revoke}.
	 * @param ctx the parse tree
	 */
	void enterRevoke(CqlParser.RevokeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#revoke}.
	 * @param ctx the parse tree
	 */
	void exitRevoke(CqlParser.RevokeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#listRoles}.
	 * @param ctx the parse tree
	 */
	void enterListRoles(CqlParser.ListRolesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#listRoles}.
	 * @param ctx the parse tree
	 */
	void exitListRoles(CqlParser.ListRolesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#listPermissions}.
	 * @param ctx the parse tree
	 */
	void enterListPermissions(CqlParser.ListPermissionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#listPermissions}.
	 * @param ctx the parse tree
	 */
	void exitListPermissions(CqlParser.ListPermissionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#grant}.
	 * @param ctx the parse tree
	 */
	void enterGrant(CqlParser.GrantContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#grant}.
	 * @param ctx the parse tree
	 */
	void exitGrant(CqlParser.GrantContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#priviledge}.
	 * @param ctx the parse tree
	 */
	void enterPriviledge(CqlParser.PriviledgeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#priviledge}.
	 * @param ctx the parse tree
	 */
	void exitPriviledge(CqlParser.PriviledgeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#resource}.
	 * @param ctx the parse tree
	 */
	void enterResource(CqlParser.ResourceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#resource}.
	 * @param ctx the parse tree
	 */
	void exitResource(CqlParser.ResourceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createUser}.
	 * @param ctx the parse tree
	 */
	void enterCreateUser(CqlParser.CreateUserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createUser}.
	 * @param ctx the parse tree
	 */
	void exitCreateUser(CqlParser.CreateUserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createRole}.
	 * @param ctx the parse tree
	 */
	void enterCreateRole(CqlParser.CreateRoleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createRole}.
	 * @param ctx the parse tree
	 */
	void exitCreateRole(CqlParser.CreateRoleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createType}.
	 * @param ctx the parse tree
	 */
	void enterCreateType(CqlParser.CreateTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createType}.
	 * @param ctx the parse tree
	 */
	void exitCreateType(CqlParser.CreateTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#typeMemberColumnList}.
	 * @param ctx the parse tree
	 */
	void enterTypeMemberColumnList(CqlParser.TypeMemberColumnListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#typeMemberColumnList}.
	 * @param ctx the parse tree
	 */
	void exitTypeMemberColumnList(CqlParser.TypeMemberColumnListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createTrigger}.
	 * @param ctx the parse tree
	 */
	void enterCreateTrigger(CqlParser.CreateTriggerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createTrigger}.
	 * @param ctx the parse tree
	 */
	void exitCreateTrigger(CqlParser.CreateTriggerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createMaterializedView}.
	 * @param ctx the parse tree
	 */
	void enterCreateMaterializedView(CqlParser.CreateMaterializedViewContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createMaterializedView}.
	 * @param ctx the parse tree
	 */
	void exitCreateMaterializedView(CqlParser.CreateMaterializedViewContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#materializedViewWhere}.
	 * @param ctx the parse tree
	 */
	void enterMaterializedViewWhere(CqlParser.MaterializedViewWhereContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#materializedViewWhere}.
	 * @param ctx the parse tree
	 */
	void exitMaterializedViewWhere(CqlParser.MaterializedViewWhereContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#columnNotNullList}.
	 * @param ctx the parse tree
	 */
	void enterColumnNotNullList(CqlParser.ColumnNotNullListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#columnNotNullList}.
	 * @param ctx the parse tree
	 */
	void exitColumnNotNullList(CqlParser.ColumnNotNullListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#columnNotNull}.
	 * @param ctx the parse tree
	 */
	void enterColumnNotNull(CqlParser.ColumnNotNullContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#columnNotNull}.
	 * @param ctx the parse tree
	 */
	void exitColumnNotNull(CqlParser.ColumnNotNullContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#materializedViewOptions}.
	 * @param ctx the parse tree
	 */
	void enterMaterializedViewOptions(CqlParser.MaterializedViewOptionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#materializedViewOptions}.
	 * @param ctx the parse tree
	 */
	void exitMaterializedViewOptions(CqlParser.MaterializedViewOptionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createKeyspace}.
	 * @param ctx the parse tree
	 */
	void enterCreateKeyspace(CqlParser.CreateKeyspaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createKeyspace}.
	 * @param ctx the parse tree
	 */
	void exitCreateKeyspace(CqlParser.CreateKeyspaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createFunction}.
	 * @param ctx the parse tree
	 */
	void enterCreateFunction(CqlParser.CreateFunctionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createFunction}.
	 * @param ctx the parse tree
	 */
	void exitCreateFunction(CqlParser.CreateFunctionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#codeBlock}.
	 * @param ctx the parse tree
	 */
	void enterCodeBlock(CqlParser.CodeBlockContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#codeBlock}.
	 * @param ctx the parse tree
	 */
	void exitCodeBlock(CqlParser.CodeBlockContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#paramList}.
	 * @param ctx the parse tree
	 */
	void enterParamList(CqlParser.ParamListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#paramList}.
	 * @param ctx the parse tree
	 */
	void exitParamList(CqlParser.ParamListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#returnMode}.
	 * @param ctx the parse tree
	 */
	void enterReturnMode(CqlParser.ReturnModeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#returnMode}.
	 * @param ctx the parse tree
	 */
	void exitReturnMode(CqlParser.ReturnModeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createAggregate}.
	 * @param ctx the parse tree
	 */
	void enterCreateAggregate(CqlParser.CreateAggregateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createAggregate}.
	 * @param ctx the parse tree
	 */
	void exitCreateAggregate(CqlParser.CreateAggregateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#initCondDefinition}.
	 * @param ctx the parse tree
	 */
	void enterInitCondDefinition(CqlParser.InitCondDefinitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#initCondDefinition}.
	 * @param ctx the parse tree
	 */
	void exitInitCondDefinition(CqlParser.InitCondDefinitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#initCondHash}.
	 * @param ctx the parse tree
	 */
	void enterInitCondHash(CqlParser.InitCondHashContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#initCondHash}.
	 * @param ctx the parse tree
	 */
	void exitInitCondHash(CqlParser.InitCondHashContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#initCondHashItem}.
	 * @param ctx the parse tree
	 */
	void enterInitCondHashItem(CqlParser.InitCondHashItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#initCondHashItem}.
	 * @param ctx the parse tree
	 */
	void exitInitCondHashItem(CqlParser.InitCondHashItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#initCondListNested}.
	 * @param ctx the parse tree
	 */
	void enterInitCondListNested(CqlParser.InitCondListNestedContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#initCondListNested}.
	 * @param ctx the parse tree
	 */
	void exitInitCondListNested(CqlParser.InitCondListNestedContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#initCondList}.
	 * @param ctx the parse tree
	 */
	void enterInitCondList(CqlParser.InitCondListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#initCondList}.
	 * @param ctx the parse tree
	 */
	void exitInitCondList(CqlParser.InitCondListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#orReplace}.
	 * @param ctx the parse tree
	 */
	void enterOrReplace(CqlParser.OrReplaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#orReplace}.
	 * @param ctx the parse tree
	 */
	void exitOrReplace(CqlParser.OrReplaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterUser}.
	 * @param ctx the parse tree
	 */
	void enterAlterUser(CqlParser.AlterUserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterUser}.
	 * @param ctx the parse tree
	 */
	void exitAlterUser(CqlParser.AlterUserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#userPassword}.
	 * @param ctx the parse tree
	 */
	void enterUserPassword(CqlParser.UserPasswordContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#userPassword}.
	 * @param ctx the parse tree
	 */
	void exitUserPassword(CqlParser.UserPasswordContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#userSuperUser}.
	 * @param ctx the parse tree
	 */
	void enterUserSuperUser(CqlParser.UserSuperUserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#userSuperUser}.
	 * @param ctx the parse tree
	 */
	void exitUserSuperUser(CqlParser.UserSuperUserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterType}.
	 * @param ctx the parse tree
	 */
	void enterAlterType(CqlParser.AlterTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterType}.
	 * @param ctx the parse tree
	 */
	void exitAlterType(CqlParser.AlterTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeOperation}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeOperation(CqlParser.AlterTypeOperationContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeOperation}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeOperation(CqlParser.AlterTypeOperationContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeRename}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeRename(CqlParser.AlterTypeRenameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeRename}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeRename(CqlParser.AlterTypeRenameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeRenameList}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeRenameList(CqlParser.AlterTypeRenameListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeRenameList}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeRenameList(CqlParser.AlterTypeRenameListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeRenameItem}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeRenameItem(CqlParser.AlterTypeRenameItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeRenameItem}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeRenameItem(CqlParser.AlterTypeRenameItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeAdd}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeAdd(CqlParser.AlterTypeAddContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeAdd}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeAdd(CqlParser.AlterTypeAddContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTypeAlterType}.
	 * @param ctx the parse tree
	 */
	void enterAlterTypeAlterType(CqlParser.AlterTypeAlterTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTypeAlterType}.
	 * @param ctx the parse tree
	 */
	void exitAlterTypeAlterType(CqlParser.AlterTypeAlterTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTable}.
	 * @param ctx the parse tree
	 */
	void enterAlterTable(CqlParser.AlterTableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTable}.
	 * @param ctx the parse tree
	 */
	void exitAlterTable(CqlParser.AlterTableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableOperation}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableOperation(CqlParser.AlterTableOperationContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableOperation}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableOperation(CqlParser.AlterTableOperationContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableWith}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableWith(CqlParser.AlterTableWithContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableWith}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableWith(CqlParser.AlterTableWithContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableRename}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableRename(CqlParser.AlterTableRenameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableRename}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableRename(CqlParser.AlterTableRenameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableDropCompactStorage}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableDropCompactStorage(CqlParser.AlterTableDropCompactStorageContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableDropCompactStorage}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableDropCompactStorage(CqlParser.AlterTableDropCompactStorageContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableDropColumns}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableDropColumns(CqlParser.AlterTableDropColumnsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableDropColumns}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableDropColumns(CqlParser.AlterTableDropColumnsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableDropColumnList}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableDropColumnList(CqlParser.AlterTableDropColumnListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableDropColumnList}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableDropColumnList(CqlParser.AlterTableDropColumnListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableAdd}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableAdd(CqlParser.AlterTableAddContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableAdd}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableAdd(CqlParser.AlterTableAddContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterTableColumnDefinition}.
	 * @param ctx the parse tree
	 */
	void enterAlterTableColumnDefinition(CqlParser.AlterTableColumnDefinitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterTableColumnDefinition}.
	 * @param ctx the parse tree
	 */
	void exitAlterTableColumnDefinition(CqlParser.AlterTableColumnDefinitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterRole}.
	 * @param ctx the parse tree
	 */
	void enterAlterRole(CqlParser.AlterRoleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterRole}.
	 * @param ctx the parse tree
	 */
	void exitAlterRole(CqlParser.AlterRoleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#roleWith}.
	 * @param ctx the parse tree
	 */
	void enterRoleWith(CqlParser.RoleWithContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#roleWith}.
	 * @param ctx the parse tree
	 */
	void exitRoleWith(CqlParser.RoleWithContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#roleWithOptions}.
	 * @param ctx the parse tree
	 */
	void enterRoleWithOptions(CqlParser.RoleWithOptionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#roleWithOptions}.
	 * @param ctx the parse tree
	 */
	void exitRoleWithOptions(CqlParser.RoleWithOptionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterMaterializedView}.
	 * @param ctx the parse tree
	 */
	void enterAlterMaterializedView(CqlParser.AlterMaterializedViewContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterMaterializedView}.
	 * @param ctx the parse tree
	 */
	void exitAlterMaterializedView(CqlParser.AlterMaterializedViewContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropUser}.
	 * @param ctx the parse tree
	 */
	void enterDropUser(CqlParser.DropUserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropUser}.
	 * @param ctx the parse tree
	 */
	void exitDropUser(CqlParser.DropUserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropType}.
	 * @param ctx the parse tree
	 */
	void enterDropType(CqlParser.DropTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropType}.
	 * @param ctx the parse tree
	 */
	void exitDropType(CqlParser.DropTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropMaterializedView}.
	 * @param ctx the parse tree
	 */
	void enterDropMaterializedView(CqlParser.DropMaterializedViewContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropMaterializedView}.
	 * @param ctx the parse tree
	 */
	void exitDropMaterializedView(CqlParser.DropMaterializedViewContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropAggregate}.
	 * @param ctx the parse tree
	 */
	void enterDropAggregate(CqlParser.DropAggregateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropAggregate}.
	 * @param ctx the parse tree
	 */
	void exitDropAggregate(CqlParser.DropAggregateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropFunction}.
	 * @param ctx the parse tree
	 */
	void enterDropFunction(CqlParser.DropFunctionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropFunction}.
	 * @param ctx the parse tree
	 */
	void exitDropFunction(CqlParser.DropFunctionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropTrigger}.
	 * @param ctx the parse tree
	 */
	void enterDropTrigger(CqlParser.DropTriggerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropTrigger}.
	 * @param ctx the parse tree
	 */
	void exitDropTrigger(CqlParser.DropTriggerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropRole}.
	 * @param ctx the parse tree
	 */
	void enterDropRole(CqlParser.DropRoleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropRole}.
	 * @param ctx the parse tree
	 */
	void exitDropRole(CqlParser.DropRoleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropTable}.
	 * @param ctx the parse tree
	 */
	void enterDropTable(CqlParser.DropTableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropTable}.
	 * @param ctx the parse tree
	 */
	void exitDropTable(CqlParser.DropTableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropKeyspace}.
	 * @param ctx the parse tree
	 */
	void enterDropKeyspace(CqlParser.DropKeyspaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropKeyspace}.
	 * @param ctx the parse tree
	 */
	void exitDropKeyspace(CqlParser.DropKeyspaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dropIndex}.
	 * @param ctx the parse tree
	 */
	void enterDropIndex(CqlParser.DropIndexContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dropIndex}.
	 * @param ctx the parse tree
	 */
	void exitDropIndex(CqlParser.DropIndexContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createTable}.
	 * @param ctx the parse tree
	 */
	void enterCreateTable(CqlParser.CreateTableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createTable}.
	 * @param ctx the parse tree
	 */
	void exitCreateTable(CqlParser.CreateTableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#withElement}.
	 * @param ctx the parse tree
	 */
	void enterWithElement(CqlParser.WithElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#withElement}.
	 * @param ctx the parse tree
	 */
	void exitWithElement(CqlParser.WithElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#tableOptions}.
	 * @param ctx the parse tree
	 */
	void enterTableOptions(CqlParser.TableOptionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#tableOptions}.
	 * @param ctx the parse tree
	 */
	void exitTableOptions(CqlParser.TableOptionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#clusteringOrder}.
	 * @param ctx the parse tree
	 */
	void enterClusteringOrder(CqlParser.ClusteringOrderContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#clusteringOrder}.
	 * @param ctx the parse tree
	 */
	void exitClusteringOrder(CqlParser.ClusteringOrderContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#tableOptionItem}.
	 * @param ctx the parse tree
	 */
	void enterTableOptionItem(CqlParser.TableOptionItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#tableOptionItem}.
	 * @param ctx the parse tree
	 */
	void exitTableOptionItem(CqlParser.TableOptionItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#tableOptionName}.
	 * @param ctx the parse tree
	 */
	void enterTableOptionName(CqlParser.TableOptionNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#tableOptionName}.
	 * @param ctx the parse tree
	 */
	void exitTableOptionName(CqlParser.TableOptionNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#tableOptionValue}.
	 * @param ctx the parse tree
	 */
	void enterTableOptionValue(CqlParser.TableOptionValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#tableOptionValue}.
	 * @param ctx the parse tree
	 */
	void exitTableOptionValue(CqlParser.TableOptionValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#optionHash}.
	 * @param ctx the parse tree
	 */
	void enterOptionHash(CqlParser.OptionHashContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#optionHash}.
	 * @param ctx the parse tree
	 */
	void exitOptionHash(CqlParser.OptionHashContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#optionHashItem}.
	 * @param ctx the parse tree
	 */
	void enterOptionHashItem(CqlParser.OptionHashItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#optionHashItem}.
	 * @param ctx the parse tree
	 */
	void exitOptionHashItem(CqlParser.OptionHashItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#optionHashKey}.
	 * @param ctx the parse tree
	 */
	void enterOptionHashKey(CqlParser.OptionHashKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#optionHashKey}.
	 * @param ctx the parse tree
	 */
	void exitOptionHashKey(CqlParser.OptionHashKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#optionHashValue}.
	 * @param ctx the parse tree
	 */
	void enterOptionHashValue(CqlParser.OptionHashValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#optionHashValue}.
	 * @param ctx the parse tree
	 */
	void exitOptionHashValue(CqlParser.OptionHashValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#columnDefinitionList}.
	 * @param ctx the parse tree
	 */
	void enterColumnDefinitionList(CqlParser.ColumnDefinitionListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#columnDefinitionList}.
	 * @param ctx the parse tree
	 */
	void exitColumnDefinitionList(CqlParser.ColumnDefinitionListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#columnDefinition}.
	 * @param ctx the parse tree
	 */
	void enterColumnDefinition(CqlParser.ColumnDefinitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#columnDefinition}.
	 * @param ctx the parse tree
	 */
	void exitColumnDefinition(CqlParser.ColumnDefinitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#primaryKeyColumn}.
	 * @param ctx the parse tree
	 */
	void enterPrimaryKeyColumn(CqlParser.PrimaryKeyColumnContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#primaryKeyColumn}.
	 * @param ctx the parse tree
	 */
	void exitPrimaryKeyColumn(CqlParser.PrimaryKeyColumnContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#primaryKeyElement}.
	 * @param ctx the parse tree
	 */
	void enterPrimaryKeyElement(CqlParser.PrimaryKeyElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#primaryKeyElement}.
	 * @param ctx the parse tree
	 */
	void exitPrimaryKeyElement(CqlParser.PrimaryKeyElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#primaryKeyDefinition}.
	 * @param ctx the parse tree
	 */
	void enterPrimaryKeyDefinition(CqlParser.PrimaryKeyDefinitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#primaryKeyDefinition}.
	 * @param ctx the parse tree
	 */
	void exitPrimaryKeyDefinition(CqlParser.PrimaryKeyDefinitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#singlePrimaryKey}.
	 * @param ctx the parse tree
	 */
	void enterSinglePrimaryKey(CqlParser.SinglePrimaryKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#singlePrimaryKey}.
	 * @param ctx the parse tree
	 */
	void exitSinglePrimaryKey(CqlParser.SinglePrimaryKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#compoundKey}.
	 * @param ctx the parse tree
	 */
	void enterCompoundKey(CqlParser.CompoundKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#compoundKey}.
	 * @param ctx the parse tree
	 */
	void exitCompoundKey(CqlParser.CompoundKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#compositeKey}.
	 * @param ctx the parse tree
	 */
	void enterCompositeKey(CqlParser.CompositeKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#compositeKey}.
	 * @param ctx the parse tree
	 */
	void exitCompositeKey(CqlParser.CompositeKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#partitionKeyList}.
	 * @param ctx the parse tree
	 */
	void enterPartitionKeyList(CqlParser.PartitionKeyListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#partitionKeyList}.
	 * @param ctx the parse tree
	 */
	void exitPartitionKeyList(CqlParser.PartitionKeyListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#clusteringKeyList}.
	 * @param ctx the parse tree
	 */
	void enterClusteringKeyList(CqlParser.ClusteringKeyListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#clusteringKeyList}.
	 * @param ctx the parse tree
	 */
	void exitClusteringKeyList(CqlParser.ClusteringKeyListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#partitionKey}.
	 * @param ctx the parse tree
	 */
	void enterPartitionKey(CqlParser.PartitionKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#partitionKey}.
	 * @param ctx the parse tree
	 */
	void exitPartitionKey(CqlParser.PartitionKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#clusteringKey}.
	 * @param ctx the parse tree
	 */
	void enterClusteringKey(CqlParser.ClusteringKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#clusteringKey}.
	 * @param ctx the parse tree
	 */
	void exitClusteringKey(CqlParser.ClusteringKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#applyBatch}.
	 * @param ctx the parse tree
	 */
	void enterApplyBatch(CqlParser.ApplyBatchContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#applyBatch}.
	 * @param ctx the parse tree
	 */
	void exitApplyBatch(CqlParser.ApplyBatchContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#beginBatch}.
	 * @param ctx the parse tree
	 */
	void enterBeginBatch(CqlParser.BeginBatchContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#beginBatch}.
	 * @param ctx the parse tree
	 */
	void exitBeginBatch(CqlParser.BeginBatchContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#batchType}.
	 * @param ctx the parse tree
	 */
	void enterBatchType(CqlParser.BatchTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#batchType}.
	 * @param ctx the parse tree
	 */
	void exitBatchType(CqlParser.BatchTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#alterKeyspace}.
	 * @param ctx the parse tree
	 */
	void enterAlterKeyspace(CqlParser.AlterKeyspaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#alterKeyspace}.
	 * @param ctx the parse tree
	 */
	void exitAlterKeyspace(CqlParser.AlterKeyspaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#replicationList}.
	 * @param ctx the parse tree
	 */
	void enterReplicationList(CqlParser.ReplicationListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#replicationList}.
	 * @param ctx the parse tree
	 */
	void exitReplicationList(CqlParser.ReplicationListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#replicationListItem}.
	 * @param ctx the parse tree
	 */
	void enterReplicationListItem(CqlParser.ReplicationListItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#replicationListItem}.
	 * @param ctx the parse tree
	 */
	void exitReplicationListItem(CqlParser.ReplicationListItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#durableWrites}.
	 * @param ctx the parse tree
	 */
	void enterDurableWrites(CqlParser.DurableWritesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#durableWrites}.
	 * @param ctx the parse tree
	 */
	void exitDurableWrites(CqlParser.DurableWritesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#use_}.
	 * @param ctx the parse tree
	 */
	void enterUse_(CqlParser.Use_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#use_}.
	 * @param ctx the parse tree
	 */
	void exitUse_(CqlParser.Use_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#truncate}.
	 * @param ctx the parse tree
	 */
	void enterTruncate(CqlParser.TruncateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#truncate}.
	 * @param ctx the parse tree
	 */
	void exitTruncate(CqlParser.TruncateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#createIndex}.
	 * @param ctx the parse tree
	 */
	void enterCreateIndex(CqlParser.CreateIndexContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#createIndex}.
	 * @param ctx the parse tree
	 */
	void exitCreateIndex(CqlParser.CreateIndexContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#indexName}.
	 * @param ctx the parse tree
	 */
	void enterIndexName(CqlParser.IndexNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#indexName}.
	 * @param ctx the parse tree
	 */
	void exitIndexName(CqlParser.IndexNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#indexColumnSpec}.
	 * @param ctx the parse tree
	 */
	void enterIndexColumnSpec(CqlParser.IndexColumnSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#indexColumnSpec}.
	 * @param ctx the parse tree
	 */
	void exitIndexColumnSpec(CqlParser.IndexColumnSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#indexKeysSpec}.
	 * @param ctx the parse tree
	 */
	void enterIndexKeysSpec(CqlParser.IndexKeysSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#indexKeysSpec}.
	 * @param ctx the parse tree
	 */
	void exitIndexKeysSpec(CqlParser.IndexKeysSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#indexEntriesSSpec}.
	 * @param ctx the parse tree
	 */
	void enterIndexEntriesSSpec(CqlParser.IndexEntriesSSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#indexEntriesSSpec}.
	 * @param ctx the parse tree
	 */
	void exitIndexEntriesSSpec(CqlParser.IndexEntriesSSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#indexFullSpec}.
	 * @param ctx the parse tree
	 */
	void enterIndexFullSpec(CqlParser.IndexFullSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#indexFullSpec}.
	 * @param ctx the parse tree
	 */
	void exitIndexFullSpec(CqlParser.IndexFullSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#delete_}.
	 * @param ctx the parse tree
	 */
	void enterDelete_(CqlParser.Delete_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#delete_}.
	 * @param ctx the parse tree
	 */
	void exitDelete_(CqlParser.Delete_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#deleteColumnList}.
	 * @param ctx the parse tree
	 */
	void enterDeleteColumnList(CqlParser.DeleteColumnListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#deleteColumnList}.
	 * @param ctx the parse tree
	 */
	void exitDeleteColumnList(CqlParser.DeleteColumnListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#deleteColumnItem}.
	 * @param ctx the parse tree
	 */
	void enterDeleteColumnItem(CqlParser.DeleteColumnItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#deleteColumnItem}.
	 * @param ctx the parse tree
	 */
	void exitDeleteColumnItem(CqlParser.DeleteColumnItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#update}.
	 * @param ctx the parse tree
	 */
	void enterUpdate(CqlParser.UpdateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#update}.
	 * @param ctx the parse tree
	 */
	void exitUpdate(CqlParser.UpdateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ifSpec}.
	 * @param ctx the parse tree
	 */
	void enterIfSpec(CqlParser.IfSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ifSpec}.
	 * @param ctx the parse tree
	 */
	void exitIfSpec(CqlParser.IfSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ifConditionList}.
	 * @param ctx the parse tree
	 */
	void enterIfConditionList(CqlParser.IfConditionListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ifConditionList}.
	 * @param ctx the parse tree
	 */
	void exitIfConditionList(CqlParser.IfConditionListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ifCondition}.
	 * @param ctx the parse tree
	 */
	void enterIfCondition(CqlParser.IfConditionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ifCondition}.
	 * @param ctx the parse tree
	 */
	void exitIfCondition(CqlParser.IfConditionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignments}.
	 * @param ctx the parse tree
	 */
	void enterAssignments(CqlParser.AssignmentsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignments}.
	 * @param ctx the parse tree
	 */
	void exitAssignments(CqlParser.AssignmentsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignmentElement}.
	 * @param ctx the parse tree
	 */
	void enterAssignmentElement(CqlParser.AssignmentElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignmentElement}.
	 * @param ctx the parse tree
	 */
	void exitAssignmentElement(CqlParser.AssignmentElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignmentSet}.
	 * @param ctx the parse tree
	 */
	void enterAssignmentSet(CqlParser.AssignmentSetContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignmentSet}.
	 * @param ctx the parse tree
	 */
	void exitAssignmentSet(CqlParser.AssignmentSetContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignmentMap}.
	 * @param ctx the parse tree
	 */
	void enterAssignmentMap(CqlParser.AssignmentMapContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignmentMap}.
	 * @param ctx the parse tree
	 */
	void exitAssignmentMap(CqlParser.AssignmentMapContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignmentList}.
	 * @param ctx the parse tree
	 */
	void enterAssignmentList(CqlParser.AssignmentListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignmentList}.
	 * @param ctx the parse tree
	 */
	void exitAssignmentList(CqlParser.AssignmentListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#assignmentTuple}.
	 * @param ctx the parse tree
	 */
	void enterAssignmentTuple(CqlParser.AssignmentTupleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#assignmentTuple}.
	 * @param ctx the parse tree
	 */
	void exitAssignmentTuple(CqlParser.AssignmentTupleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#insert}.
	 * @param ctx the parse tree
	 */
	void enterInsert(CqlParser.InsertContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#insert}.
	 * @param ctx the parse tree
	 */
	void exitInsert(CqlParser.InsertContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#usingTtlTimestamp}.
	 * @param ctx the parse tree
	 */
	void enterUsingTtlTimestamp(CqlParser.UsingTtlTimestampContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#usingTtlTimestamp}.
	 * @param ctx the parse tree
	 */
	void exitUsingTtlTimestamp(CqlParser.UsingTtlTimestampContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#timestamp}.
	 * @param ctx the parse tree
	 */
	void enterTimestamp(CqlParser.TimestampContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#timestamp}.
	 * @param ctx the parse tree
	 */
	void exitTimestamp(CqlParser.TimestampContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ttl}.
	 * @param ctx the parse tree
	 */
	void enterTtl(CqlParser.TtlContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ttl}.
	 * @param ctx the parse tree
	 */
	void exitTtl(CqlParser.TtlContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#usingTimestampSpec}.
	 * @param ctx the parse tree
	 */
	void enterUsingTimestampSpec(CqlParser.UsingTimestampSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#usingTimestampSpec}.
	 * @param ctx the parse tree
	 */
	void exitUsingTimestampSpec(CqlParser.UsingTimestampSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ifNotExist}.
	 * @param ctx the parse tree
	 */
	void enterIfNotExist(CqlParser.IfNotExistContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ifNotExist}.
	 * @param ctx the parse tree
	 */
	void exitIfNotExist(CqlParser.IfNotExistContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#ifExist}.
	 * @param ctx the parse tree
	 */
	void enterIfExist(CqlParser.IfExistContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#ifExist}.
	 * @param ctx the parse tree
	 */
	void exitIfExist(CqlParser.IfExistContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#insertValuesSpec}.
	 * @param ctx the parse tree
	 */
	void enterInsertValuesSpec(CqlParser.InsertValuesSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#insertValuesSpec}.
	 * @param ctx the parse tree
	 */
	void exitInsertValuesSpec(CqlParser.InsertValuesSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#insertColumnSpec}.
	 * @param ctx the parse tree
	 */
	void enterInsertColumnSpec(CqlParser.InsertColumnSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#insertColumnSpec}.
	 * @param ctx the parse tree
	 */
	void exitInsertColumnSpec(CqlParser.InsertColumnSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#columnList}.
	 * @param ctx the parse tree
	 */
	void enterColumnList(CqlParser.ColumnListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#columnList}.
	 * @param ctx the parse tree
	 */
	void exitColumnList(CqlParser.ColumnListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#expressionList}.
	 * @param ctx the parse tree
	 */
	void enterExpressionList(CqlParser.ExpressionListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#expressionList}.
	 * @param ctx the parse tree
	 */
	void exitExpressionList(CqlParser.ExpressionListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#expression}.
	 * @param ctx the parse tree
	 */
	void enterExpression(CqlParser.ExpressionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#expression}.
	 * @param ctx the parse tree
	 */
	void exitExpression(CqlParser.ExpressionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#select_}.
	 * @param ctx the parse tree
	 */
	void enterSelect_(CqlParser.Select_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#select_}.
	 * @param ctx the parse tree
	 */
	void exitSelect_(CqlParser.Select_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#allowFilteringSpec}.
	 * @param ctx the parse tree
	 */
	void enterAllowFilteringSpec(CqlParser.AllowFilteringSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#allowFilteringSpec}.
	 * @param ctx the parse tree
	 */
	void exitAllowFilteringSpec(CqlParser.AllowFilteringSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#limitSpec}.
	 * @param ctx the parse tree
	 */
	void enterLimitSpec(CqlParser.LimitSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#limitSpec}.
	 * @param ctx the parse tree
	 */
	void exitLimitSpec(CqlParser.LimitSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#fromSpec}.
	 * @param ctx the parse tree
	 */
	void enterFromSpec(CqlParser.FromSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#fromSpec}.
	 * @param ctx the parse tree
	 */
	void exitFromSpec(CqlParser.FromSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#fromSpecElement}.
	 * @param ctx the parse tree
	 */
	void enterFromSpecElement(CqlParser.FromSpecElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#fromSpecElement}.
	 * @param ctx the parse tree
	 */
	void exitFromSpecElement(CqlParser.FromSpecElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#orderSpec}.
	 * @param ctx the parse tree
	 */
	void enterOrderSpec(CqlParser.OrderSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#orderSpec}.
	 * @param ctx the parse tree
	 */
	void exitOrderSpec(CqlParser.OrderSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#orderSpecElement}.
	 * @param ctx the parse tree
	 */
	void enterOrderSpecElement(CqlParser.OrderSpecElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#orderSpecElement}.
	 * @param ctx the parse tree
	 */
	void exitOrderSpecElement(CqlParser.OrderSpecElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#whereSpec}.
	 * @param ctx the parse tree
	 */
	void enterWhereSpec(CqlParser.WhereSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#whereSpec}.
	 * @param ctx the parse tree
	 */
	void exitWhereSpec(CqlParser.WhereSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#distinctSpec}.
	 * @param ctx the parse tree
	 */
	void enterDistinctSpec(CqlParser.DistinctSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#distinctSpec}.
	 * @param ctx the parse tree
	 */
	void exitDistinctSpec(CqlParser.DistinctSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#selectElements}.
	 * @param ctx the parse tree
	 */
	void enterSelectElements(CqlParser.SelectElementsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#selectElements}.
	 * @param ctx the parse tree
	 */
	void exitSelectElements(CqlParser.SelectElementsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#selectElement}.
	 * @param ctx the parse tree
	 */
	void enterSelectElement(CqlParser.SelectElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#selectElement}.
	 * @param ctx the parse tree
	 */
	void exitSelectElement(CqlParser.SelectElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#relationElements}.
	 * @param ctx the parse tree
	 */
	void enterRelationElements(CqlParser.RelationElementsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#relationElements}.
	 * @param ctx the parse tree
	 */
	void exitRelationElements(CqlParser.RelationElementsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#relationElement}.
	 * @param ctx the parse tree
	 */
	void enterRelationElement(CqlParser.RelationElementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#relationElement}.
	 * @param ctx the parse tree
	 */
	void exitRelationElement(CqlParser.RelationElementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#relalationContains}.
	 * @param ctx the parse tree
	 */
	void enterRelalationContains(CqlParser.RelalationContainsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#relalationContains}.
	 * @param ctx the parse tree
	 */
	void exitRelalationContains(CqlParser.RelalationContainsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#relalationContainsKey}.
	 * @param ctx the parse tree
	 */
	void enterRelalationContainsKey(CqlParser.RelalationContainsKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#relalationContainsKey}.
	 * @param ctx the parse tree
	 */
	void exitRelalationContainsKey(CqlParser.RelalationContainsKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#functionCall}.
	 * @param ctx the parse tree
	 */
	void enterFunctionCall(CqlParser.FunctionCallContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#functionCall}.
	 * @param ctx the parse tree
	 */
	void exitFunctionCall(CqlParser.FunctionCallContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#functionArgs}.
	 * @param ctx the parse tree
	 */
	void enterFunctionArgs(CqlParser.FunctionArgsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#functionArgs}.
	 * @param ctx the parse tree
	 */
	void exitFunctionArgs(CqlParser.FunctionArgsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#constant}.
	 * @param ctx the parse tree
	 */
	void enterConstant(CqlParser.ConstantContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#constant}.
	 * @param ctx the parse tree
	 */
	void exitConstant(CqlParser.ConstantContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#decimalLiteral}.
	 * @param ctx the parse tree
	 */
	void enterDecimalLiteral(CqlParser.DecimalLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#decimalLiteral}.
	 * @param ctx the parse tree
	 */
	void exitDecimalLiteral(CqlParser.DecimalLiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#floatLiteral}.
	 * @param ctx the parse tree
	 */
	void enterFloatLiteral(CqlParser.FloatLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#floatLiteral}.
	 * @param ctx the parse tree
	 */
	void exitFloatLiteral(CqlParser.FloatLiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#stringLiteral}.
	 * @param ctx the parse tree
	 */
	void enterStringLiteral(CqlParser.StringLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#stringLiteral}.
	 * @param ctx the parse tree
	 */
	void exitStringLiteral(CqlParser.StringLiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#booleanLiteral}.
	 * @param ctx the parse tree
	 */
	void enterBooleanLiteral(CqlParser.BooleanLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#booleanLiteral}.
	 * @param ctx the parse tree
	 */
	void exitBooleanLiteral(CqlParser.BooleanLiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#hexadecimalLiteral}.
	 * @param ctx the parse tree
	 */
	void enterHexadecimalLiteral(CqlParser.HexadecimalLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#hexadecimalLiteral}.
	 * @param ctx the parse tree
	 */
	void exitHexadecimalLiteral(CqlParser.HexadecimalLiteralContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#keyspace}.
	 * @param ctx the parse tree
	 */
	void enterKeyspace(CqlParser.KeyspaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#keyspace}.
	 * @param ctx the parse tree
	 */
	void exitKeyspace(CqlParser.KeyspaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#table}.
	 * @param ctx the parse tree
	 */
	void enterTable(CqlParser.TableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#table}.
	 * @param ctx the parse tree
	 */
	void exitTable(CqlParser.TableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#column}.
	 * @param ctx the parse tree
	 */
	void enterColumn(CqlParser.ColumnContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#column}.
	 * @param ctx the parse tree
	 */
	void exitColumn(CqlParser.ColumnContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dataType}.
	 * @param ctx the parse tree
	 */
	void enterDataType(CqlParser.DataTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dataType}.
	 * @param ctx the parse tree
	 */
	void exitDataType(CqlParser.DataTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dataTypeName}.
	 * @param ctx the parse tree
	 */
	void enterDataTypeName(CqlParser.DataTypeNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dataTypeName}.
	 * @param ctx the parse tree
	 */
	void exitDataTypeName(CqlParser.DataTypeNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#dataTypeDefinition}.
	 * @param ctx the parse tree
	 */
	void enterDataTypeDefinition(CqlParser.DataTypeDefinitionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#dataTypeDefinition}.
	 * @param ctx the parse tree
	 */
	void exitDataTypeDefinition(CqlParser.DataTypeDefinitionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#orderDirection}.
	 * @param ctx the parse tree
	 */
	void enterOrderDirection(CqlParser.OrderDirectionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#orderDirection}.
	 * @param ctx the parse tree
	 */
	void exitOrderDirection(CqlParser.OrderDirectionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#role}.
	 * @param ctx the parse tree
	 */
	void enterRole(CqlParser.RoleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#role}.
	 * @param ctx the parse tree
	 */
	void exitRole(CqlParser.RoleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#trigger}.
	 * @param ctx the parse tree
	 */
	void enterTrigger(CqlParser.TriggerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#trigger}.
	 * @param ctx the parse tree
	 */
	void exitTrigger(CqlParser.TriggerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#triggerClass}.
	 * @param ctx the parse tree
	 */
	void enterTriggerClass(CqlParser.TriggerClassContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#triggerClass}.
	 * @param ctx the parse tree
	 */
	void exitTriggerClass(CqlParser.TriggerClassContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#materializedView}.
	 * @param ctx the parse tree
	 */
	void enterMaterializedView(CqlParser.MaterializedViewContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#materializedView}.
	 * @param ctx the parse tree
	 */
	void exitMaterializedView(CqlParser.MaterializedViewContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#type_}.
	 * @param ctx the parse tree
	 */
	void enterType_(CqlParser.Type_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#type_}.
	 * @param ctx the parse tree
	 */
	void exitType_(CqlParser.Type_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#aggregate}.
	 * @param ctx the parse tree
	 */
	void enterAggregate(CqlParser.AggregateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#aggregate}.
	 * @param ctx the parse tree
	 */
	void exitAggregate(CqlParser.AggregateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#function_}.
	 * @param ctx the parse tree
	 */
	void enterFunction_(CqlParser.Function_Context ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#function_}.
	 * @param ctx the parse tree
	 */
	void exitFunction_(CqlParser.Function_Context ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#language}.
	 * @param ctx the parse tree
	 */
	void enterLanguage(CqlParser.LanguageContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#language}.
	 * @param ctx the parse tree
	 */
	void exitLanguage(CqlParser.LanguageContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#user}.
	 * @param ctx the parse tree
	 */
	void enterUser(CqlParser.UserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#user}.
	 * @param ctx the parse tree
	 */
	void exitUser(CqlParser.UserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#password}.
	 * @param ctx the parse tree
	 */
	void enterPassword(CqlParser.PasswordContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#password}.
	 * @param ctx the parse tree
	 */
	void exitPassword(CqlParser.PasswordContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#hashKey}.
	 * @param ctx the parse tree
	 */
	void enterHashKey(CqlParser.HashKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#hashKey}.
	 * @param ctx the parse tree
	 */
	void exitHashKey(CqlParser.HashKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#param}.
	 * @param ctx the parse tree
	 */
	void enterParam(CqlParser.ParamContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#param}.
	 * @param ctx the parse tree
	 */
	void exitParam(CqlParser.ParamContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#paramName}.
	 * @param ctx the parse tree
	 */
	void enterParamName(CqlParser.ParamNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#paramName}.
	 * @param ctx the parse tree
	 */
	void exitParamName(CqlParser.ParamNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAdd}.
	 * @param ctx the parse tree
	 */
	void enterKwAdd(CqlParser.KwAddContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAdd}.
	 * @param ctx the parse tree
	 */
	void exitKwAdd(CqlParser.KwAddContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAggregate}.
	 * @param ctx the parse tree
	 */
	void enterKwAggregate(CqlParser.KwAggregateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAggregate}.
	 * @param ctx the parse tree
	 */
	void exitKwAggregate(CqlParser.KwAggregateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAggregates}.
	 * @param ctx the parse tree
	 */
	void enterKwAggregates(CqlParser.KwAggregatesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAggregates}.
	 * @param ctx the parse tree
	 */
	void exitKwAggregates(CqlParser.KwAggregatesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAll}.
	 * @param ctx the parse tree
	 */
	void enterKwAll(CqlParser.KwAllContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAll}.
	 * @param ctx the parse tree
	 */
	void exitKwAll(CqlParser.KwAllContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAllPermissions}.
	 * @param ctx the parse tree
	 */
	void enterKwAllPermissions(CqlParser.KwAllPermissionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAllPermissions}.
	 * @param ctx the parse tree
	 */
	void exitKwAllPermissions(CqlParser.KwAllPermissionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAllow}.
	 * @param ctx the parse tree
	 */
	void enterKwAllow(CqlParser.KwAllowContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAllow}.
	 * @param ctx the parse tree
	 */
	void exitKwAllow(CqlParser.KwAllowContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAlter}.
	 * @param ctx the parse tree
	 */
	void enterKwAlter(CqlParser.KwAlterContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAlter}.
	 * @param ctx the parse tree
	 */
	void exitKwAlter(CqlParser.KwAlterContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAnd}.
	 * @param ctx the parse tree
	 */
	void enterKwAnd(CqlParser.KwAndContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAnd}.
	 * @param ctx the parse tree
	 */
	void exitKwAnd(CqlParser.KwAndContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwApply}.
	 * @param ctx the parse tree
	 */
	void enterKwApply(CqlParser.KwApplyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwApply}.
	 * @param ctx the parse tree
	 */
	void exitKwApply(CqlParser.KwApplyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAs}.
	 * @param ctx the parse tree
	 */
	void enterKwAs(CqlParser.KwAsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAs}.
	 * @param ctx the parse tree
	 */
	void exitKwAs(CqlParser.KwAsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAsc}.
	 * @param ctx the parse tree
	 */
	void enterKwAsc(CqlParser.KwAscContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAsc}.
	 * @param ctx the parse tree
	 */
	void exitKwAsc(CqlParser.KwAscContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwAuthorize}.
	 * @param ctx the parse tree
	 */
	void enterKwAuthorize(CqlParser.KwAuthorizeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwAuthorize}.
	 * @param ctx the parse tree
	 */
	void exitKwAuthorize(CqlParser.KwAuthorizeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwBatch}.
	 * @param ctx the parse tree
	 */
	void enterKwBatch(CqlParser.KwBatchContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwBatch}.
	 * @param ctx the parse tree
	 */
	void exitKwBatch(CqlParser.KwBatchContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwBegin}.
	 * @param ctx the parse tree
	 */
	void enterKwBegin(CqlParser.KwBeginContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwBegin}.
	 * @param ctx the parse tree
	 */
	void exitKwBegin(CqlParser.KwBeginContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwBy}.
	 * @param ctx the parse tree
	 */
	void enterKwBy(CqlParser.KwByContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwBy}.
	 * @param ctx the parse tree
	 */
	void exitKwBy(CqlParser.KwByContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwCalled}.
	 * @param ctx the parse tree
	 */
	void enterKwCalled(CqlParser.KwCalledContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwCalled}.
	 * @param ctx the parse tree
	 */
	void exitKwCalled(CqlParser.KwCalledContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwCluster}.
	 * @param ctx the parse tree
	 */
	void enterKwCluster(CqlParser.KwClusterContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwCluster}.
	 * @param ctx the parse tree
	 */
	void exitKwCluster(CqlParser.KwClusterContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwClustering}.
	 * @param ctx the parse tree
	 */
	void enterKwClustering(CqlParser.KwClusteringContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwClustering}.
	 * @param ctx the parse tree
	 */
	void exitKwClustering(CqlParser.KwClusteringContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwCompact}.
	 * @param ctx the parse tree
	 */
	void enterKwCompact(CqlParser.KwCompactContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwCompact}.
	 * @param ctx the parse tree
	 */
	void exitKwCompact(CqlParser.KwCompactContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwConnection}.
	 * @param ctx the parse tree
	 */
	void enterKwConnection(CqlParser.KwConnectionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwConnection}.
	 * @param ctx the parse tree
	 */
	void exitKwConnection(CqlParser.KwConnectionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwConsistency}.
	 * @param ctx the parse tree
	 */
	void enterKwConsistency(CqlParser.KwConsistencyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwConsistency}.
	 * @param ctx the parse tree
	 */
	void exitKwConsistency(CqlParser.KwConsistencyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwConsistencyLevel}.
	 * @param ctx the parse tree
	 */
	void enterKwConsistencyLevel(CqlParser.KwConsistencyLevelContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwConsistencyLevel}.
	 * @param ctx the parse tree
	 */
	void exitKwConsistencyLevel(CqlParser.KwConsistencyLevelContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOutput}.
	 * @param ctx the parse tree
	 */
	void enterKwOutput(CqlParser.KwOutputContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOutput}.
	 * @param ctx the parse tree
	 */
	void exitKwOutput(CqlParser.KwOutputContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOutputFormatType}.
	 * @param ctx the parse tree
	 */
	void enterKwOutputFormatType(CqlParser.KwOutputFormatTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOutputFormatType}.
	 * @param ctx the parse tree
	 */
	void exitKwOutputFormatType(CqlParser.KwOutputFormatTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwContains}.
	 * @param ctx the parse tree
	 */
	void enterKwContains(CqlParser.KwContainsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwContains}.
	 * @param ctx the parse tree
	 */
	void exitKwContains(CqlParser.KwContainsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwCreate}.
	 * @param ctx the parse tree
	 */
	void enterKwCreate(CqlParser.KwCreateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwCreate}.
	 * @param ctx the parse tree
	 */
	void exitKwCreate(CqlParser.KwCreateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDelete}.
	 * @param ctx the parse tree
	 */
	void enterKwDelete(CqlParser.KwDeleteContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDelete}.
	 * @param ctx the parse tree
	 */
	void exitKwDelete(CqlParser.KwDeleteContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDesc}.
	 * @param ctx the parse tree
	 */
	void enterKwDesc(CqlParser.KwDescContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDesc}.
	 * @param ctx the parse tree
	 */
	void exitKwDesc(CqlParser.KwDescContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDescribe}.
	 * @param ctx the parse tree
	 */
	void enterKwDescribe(CqlParser.KwDescribeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDescribe}.
	 * @param ctx the parse tree
	 */
	void exitKwDescribe(CqlParser.KwDescribeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDistinct}.
	 * @param ctx the parse tree
	 */
	void enterKwDistinct(CqlParser.KwDistinctContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDistinct}.
	 * @param ctx the parse tree
	 */
	void exitKwDistinct(CqlParser.KwDistinctContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDrop}.
	 * @param ctx the parse tree
	 */
	void enterKwDrop(CqlParser.KwDropContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDrop}.
	 * @param ctx the parse tree
	 */
	void exitKwDrop(CqlParser.KwDropContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwDurableWrites}.
	 * @param ctx the parse tree
	 */
	void enterKwDurableWrites(CqlParser.KwDurableWritesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwDurableWrites}.
	 * @param ctx the parse tree
	 */
	void exitKwDurableWrites(CqlParser.KwDurableWritesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwEntries}.
	 * @param ctx the parse tree
	 */
	void enterKwEntries(CqlParser.KwEntriesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwEntries}.
	 * @param ctx the parse tree
	 */
	void exitKwEntries(CqlParser.KwEntriesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwExecute}.
	 * @param ctx the parse tree
	 */
	void enterKwExecute(CqlParser.KwExecuteContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwExecute}.
	 * @param ctx the parse tree
	 */
	void exitKwExecute(CqlParser.KwExecuteContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwExists}.
	 * @param ctx the parse tree
	 */
	void enterKwExists(CqlParser.KwExistsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwExists}.
	 * @param ctx the parse tree
	 */
	void exitKwExists(CqlParser.KwExistsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFiltering}.
	 * @param ctx the parse tree
	 */
	void enterKwFiltering(CqlParser.KwFilteringContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFiltering}.
	 * @param ctx the parse tree
	 */
	void exitKwFiltering(CqlParser.KwFilteringContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFinalfunc}.
	 * @param ctx the parse tree
	 */
	void enterKwFinalfunc(CqlParser.KwFinalfuncContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFinalfunc}.
	 * @param ctx the parse tree
	 */
	void exitKwFinalfunc(CqlParser.KwFinalfuncContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFrom}.
	 * @param ctx the parse tree
	 */
	void enterKwFrom(CqlParser.KwFromContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFrom}.
	 * @param ctx the parse tree
	 */
	void exitKwFrom(CqlParser.KwFromContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFull}.
	 * @param ctx the parse tree
	 */
	void enterKwFull(CqlParser.KwFullContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFull}.
	 * @param ctx the parse tree
	 */
	void exitKwFull(CqlParser.KwFullContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFunction}.
	 * @param ctx the parse tree
	 */
	void enterKwFunction(CqlParser.KwFunctionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFunction}.
	 * @param ctx the parse tree
	 */
	void exitKwFunction(CqlParser.KwFunctionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwFunctions}.
	 * @param ctx the parse tree
	 */
	void enterKwFunctions(CqlParser.KwFunctionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwFunctions}.
	 * @param ctx the parse tree
	 */
	void exitKwFunctions(CqlParser.KwFunctionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwGrant}.
	 * @param ctx the parse tree
	 */
	void enterKwGrant(CqlParser.KwGrantContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwGrant}.
	 * @param ctx the parse tree
	 */
	void exitKwGrant(CqlParser.KwGrantContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwIf}.
	 * @param ctx the parse tree
	 */
	void enterKwIf(CqlParser.KwIfContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwIf}.
	 * @param ctx the parse tree
	 */
	void exitKwIf(CqlParser.KwIfContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwIn}.
	 * @param ctx the parse tree
	 */
	void enterKwIn(CqlParser.KwInContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwIn}.
	 * @param ctx the parse tree
	 */
	void exitKwIn(CqlParser.KwInContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwIndex}.
	 * @param ctx the parse tree
	 */
	void enterKwIndex(CqlParser.KwIndexContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwIndex}.
	 * @param ctx the parse tree
	 */
	void exitKwIndex(CqlParser.KwIndexContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwInitcond}.
	 * @param ctx the parse tree
	 */
	void enterKwInitcond(CqlParser.KwInitcondContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwInitcond}.
	 * @param ctx the parse tree
	 */
	void exitKwInitcond(CqlParser.KwInitcondContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwInput}.
	 * @param ctx the parse tree
	 */
	void enterKwInput(CqlParser.KwInputContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwInput}.
	 * @param ctx the parse tree
	 */
	void exitKwInput(CqlParser.KwInputContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwInsert}.
	 * @param ctx the parse tree
	 */
	void enterKwInsert(CqlParser.KwInsertContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwInsert}.
	 * @param ctx the parse tree
	 */
	void exitKwInsert(CqlParser.KwInsertContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwInto}.
	 * @param ctx the parse tree
	 */
	void enterKwInto(CqlParser.KwIntoContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwInto}.
	 * @param ctx the parse tree
	 */
	void exitKwInto(CqlParser.KwIntoContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwIs}.
	 * @param ctx the parse tree
	 */
	void enterKwIs(CqlParser.KwIsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwIs}.
	 * @param ctx the parse tree
	 */
	void exitKwIs(CqlParser.KwIsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwJson}.
	 * @param ctx the parse tree
	 */
	void enterKwJson(CqlParser.KwJsonContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwJson}.
	 * @param ctx the parse tree
	 */
	void exitKwJson(CqlParser.KwJsonContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwKey}.
	 * @param ctx the parse tree
	 */
	void enterKwKey(CqlParser.KwKeyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwKey}.
	 * @param ctx the parse tree
	 */
	void exitKwKey(CqlParser.KwKeyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwKeys}.
	 * @param ctx the parse tree
	 */
	void enterKwKeys(CqlParser.KwKeysContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwKeys}.
	 * @param ctx the parse tree
	 */
	void exitKwKeys(CqlParser.KwKeysContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwKeyspace}.
	 * @param ctx the parse tree
	 */
	void enterKwKeyspace(CqlParser.KwKeyspaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwKeyspace}.
	 * @param ctx the parse tree
	 */
	void exitKwKeyspace(CqlParser.KwKeyspaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwKeyspaces}.
	 * @param ctx the parse tree
	 */
	void enterKwKeyspaces(CqlParser.KwKeyspacesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwKeyspaces}.
	 * @param ctx the parse tree
	 */
	void exitKwKeyspaces(CqlParser.KwKeyspacesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwLanguage}.
	 * @param ctx the parse tree
	 */
	void enterKwLanguage(CqlParser.KwLanguageContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwLanguage}.
	 * @param ctx the parse tree
	 */
	void exitKwLanguage(CqlParser.KwLanguageContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwLimit}.
	 * @param ctx the parse tree
	 */
	void enterKwLimit(CqlParser.KwLimitContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwLimit}.
	 * @param ctx the parse tree
	 */
	void exitKwLimit(CqlParser.KwLimitContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwList}.
	 * @param ctx the parse tree
	 */
	void enterKwList(CqlParser.KwListContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwList}.
	 * @param ctx the parse tree
	 */
	void exitKwList(CqlParser.KwListContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwLogged}.
	 * @param ctx the parse tree
	 */
	void enterKwLogged(CqlParser.KwLoggedContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwLogged}.
	 * @param ctx the parse tree
	 */
	void exitKwLogged(CqlParser.KwLoggedContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwLocalSerial}.
	 * @param ctx the parse tree
	 */
	void enterKwLocalSerial(CqlParser.KwLocalSerialContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwLocalSerial}.
	 * @param ctx the parse tree
	 */
	void exitKwLocalSerial(CqlParser.KwLocalSerialContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwLogin}.
	 * @param ctx the parse tree
	 */
	void enterKwLogin(CqlParser.KwLoginContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwLogin}.
	 * @param ctx the parse tree
	 */
	void exitKwLogin(CqlParser.KwLoginContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwMaterialized}.
	 * @param ctx the parse tree
	 */
	void enterKwMaterialized(CqlParser.KwMaterializedContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwMaterialized}.
	 * @param ctx the parse tree
	 */
	void exitKwMaterialized(CqlParser.KwMaterializedContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwModify}.
	 * @param ctx the parse tree
	 */
	void enterKwModify(CqlParser.KwModifyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwModify}.
	 * @param ctx the parse tree
	 */
	void exitKwModify(CqlParser.KwModifyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwNosuperuser}.
	 * @param ctx the parse tree
	 */
	void enterKwNosuperuser(CqlParser.KwNosuperuserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwNosuperuser}.
	 * @param ctx the parse tree
	 */
	void exitKwNosuperuser(CqlParser.KwNosuperuserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwNorecursive}.
	 * @param ctx the parse tree
	 */
	void enterKwNorecursive(CqlParser.KwNorecursiveContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwNorecursive}.
	 * @param ctx the parse tree
	 */
	void exitKwNorecursive(CqlParser.KwNorecursiveContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwNot}.
	 * @param ctx the parse tree
	 */
	void enterKwNot(CqlParser.KwNotContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwNot}.
	 * @param ctx the parse tree
	 */
	void exitKwNot(CqlParser.KwNotContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwNull}.
	 * @param ctx the parse tree
	 */
	void enterKwNull(CqlParser.KwNullContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwNull}.
	 * @param ctx the parse tree
	 */
	void exitKwNull(CqlParser.KwNullContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOf}.
	 * @param ctx the parse tree
	 */
	void enterKwOf(CqlParser.KwOfContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOf}.
	 * @param ctx the parse tree
	 */
	void exitKwOf(CqlParser.KwOfContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOn}.
	 * @param ctx the parse tree
	 */
	void enterKwOn(CqlParser.KwOnContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOn}.
	 * @param ctx the parse tree
	 */
	void exitKwOn(CqlParser.KwOnContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOptions}.
	 * @param ctx the parse tree
	 */
	void enterKwOptions(CqlParser.KwOptionsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOptions}.
	 * @param ctx the parse tree
	 */
	void exitKwOptions(CqlParser.KwOptionsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOr}.
	 * @param ctx the parse tree
	 */
	void enterKwOr(CqlParser.KwOrContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOr}.
	 * @param ctx the parse tree
	 */
	void exitKwOr(CqlParser.KwOrContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwOrder}.
	 * @param ctx the parse tree
	 */
	void enterKwOrder(CqlParser.KwOrderContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwOrder}.
	 * @param ctx the parse tree
	 */
	void exitKwOrder(CqlParser.KwOrderContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwPassword}.
	 * @param ctx the parse tree
	 */
	void enterKwPassword(CqlParser.KwPasswordContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwPassword}.
	 * @param ctx the parse tree
	 */
	void exitKwPassword(CqlParser.KwPasswordContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwPrimary}.
	 * @param ctx the parse tree
	 */
	void enterKwPrimary(CqlParser.KwPrimaryContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwPrimary}.
	 * @param ctx the parse tree
	 */
	void exitKwPrimary(CqlParser.KwPrimaryContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwRename}.
	 * @param ctx the parse tree
	 */
	void enterKwRename(CqlParser.KwRenameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwRename}.
	 * @param ctx the parse tree
	 */
	void exitKwRename(CqlParser.KwRenameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwReplace}.
	 * @param ctx the parse tree
	 */
	void enterKwReplace(CqlParser.KwReplaceContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwReplace}.
	 * @param ctx the parse tree
	 */
	void exitKwReplace(CqlParser.KwReplaceContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwReplication}.
	 * @param ctx the parse tree
	 */
	void enterKwReplication(CqlParser.KwReplicationContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwReplication}.
	 * @param ctx the parse tree
	 */
	void exitKwReplication(CqlParser.KwReplicationContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwReturns}.
	 * @param ctx the parse tree
	 */
	void enterKwReturns(CqlParser.KwReturnsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwReturns}.
	 * @param ctx the parse tree
	 */
	void exitKwReturns(CqlParser.KwReturnsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwRole}.
	 * @param ctx the parse tree
	 */
	void enterKwRole(CqlParser.KwRoleContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwRole}.
	 * @param ctx the parse tree
	 */
	void exitKwRole(CqlParser.KwRoleContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwRoles}.
	 * @param ctx the parse tree
	 */
	void enterKwRoles(CqlParser.KwRolesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwRoles}.
	 * @param ctx the parse tree
	 */
	void exitKwRoles(CqlParser.KwRolesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwSelect}.
	 * @param ctx the parse tree
	 */
	void enterKwSelect(CqlParser.KwSelectContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwSelect}.
	 * @param ctx the parse tree
	 */
	void exitKwSelect(CqlParser.KwSelectContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwSerial}.
	 * @param ctx the parse tree
	 */
	void enterKwSerial(CqlParser.KwSerialContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwSerial}.
	 * @param ctx the parse tree
	 */
	void exitKwSerial(CqlParser.KwSerialContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwSet}.
	 * @param ctx the parse tree
	 */
	void enterKwSet(CqlParser.KwSetContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwSet}.
	 * @param ctx the parse tree
	 */
	void exitKwSet(CqlParser.KwSetContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwSfunc}.
	 * @param ctx the parse tree
	 */
	void enterKwSfunc(CqlParser.KwSfuncContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwSfunc}.
	 * @param ctx the parse tree
	 */
	void exitKwSfunc(CqlParser.KwSfuncContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwStorage}.
	 * @param ctx the parse tree
	 */
	void enterKwStorage(CqlParser.KwStorageContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwStorage}.
	 * @param ctx the parse tree
	 */
	void exitKwStorage(CqlParser.KwStorageContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwStype}.
	 * @param ctx the parse tree
	 */
	void enterKwStype(CqlParser.KwStypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwStype}.
	 * @param ctx the parse tree
	 */
	void exitKwStype(CqlParser.KwStypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwSuperuser}.
	 * @param ctx the parse tree
	 */
	void enterKwSuperuser(CqlParser.KwSuperuserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwSuperuser}.
	 * @param ctx the parse tree
	 */
	void exitKwSuperuser(CqlParser.KwSuperuserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTable}.
	 * @param ctx the parse tree
	 */
	void enterKwTable(CqlParser.KwTableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTable}.
	 * @param ctx the parse tree
	 */
	void exitKwTable(CqlParser.KwTableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTables}.
	 * @param ctx the parse tree
	 */
	void enterKwTables(CqlParser.KwTablesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTables}.
	 * @param ctx the parse tree
	 */
	void exitKwTables(CqlParser.KwTablesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTimestamp}.
	 * @param ctx the parse tree
	 */
	void enterKwTimestamp(CqlParser.KwTimestampContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTimestamp}.
	 * @param ctx the parse tree
	 */
	void exitKwTimestamp(CqlParser.KwTimestampContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTo}.
	 * @param ctx the parse tree
	 */
	void enterKwTo(CqlParser.KwToContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTo}.
	 * @param ctx the parse tree
	 */
	void exitKwTo(CqlParser.KwToContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTrigger}.
	 * @param ctx the parse tree
	 */
	void enterKwTrigger(CqlParser.KwTriggerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTrigger}.
	 * @param ctx the parse tree
	 */
	void exitKwTrigger(CqlParser.KwTriggerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTruncate}.
	 * @param ctx the parse tree
	 */
	void enterKwTruncate(CqlParser.KwTruncateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTruncate}.
	 * @param ctx the parse tree
	 */
	void exitKwTruncate(CqlParser.KwTruncateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTtl}.
	 * @param ctx the parse tree
	 */
	void enterKwTtl(CqlParser.KwTtlContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTtl}.
	 * @param ctx the parse tree
	 */
	void exitKwTtl(CqlParser.KwTtlContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwType}.
	 * @param ctx the parse tree
	 */
	void enterKwType(CqlParser.KwTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwType}.
	 * @param ctx the parse tree
	 */
	void exitKwType(CqlParser.KwTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwTypes}.
	 * @param ctx the parse tree
	 */
	void enterKwTypes(CqlParser.KwTypesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwTypes}.
	 * @param ctx the parse tree
	 */
	void exitKwTypes(CqlParser.KwTypesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwUnlogged}.
	 * @param ctx the parse tree
	 */
	void enterKwUnlogged(CqlParser.KwUnloggedContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwUnlogged}.
	 * @param ctx the parse tree
	 */
	void exitKwUnlogged(CqlParser.KwUnloggedContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwUpdate}.
	 * @param ctx the parse tree
	 */
	void enterKwUpdate(CqlParser.KwUpdateContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwUpdate}.
	 * @param ctx the parse tree
	 */
	void exitKwUpdate(CqlParser.KwUpdateContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwUse}.
	 * @param ctx the parse tree
	 */
	void enterKwUse(CqlParser.KwUseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwUse}.
	 * @param ctx the parse tree
	 */
	void exitKwUse(CqlParser.KwUseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwUser}.
	 * @param ctx the parse tree
	 */
	void enterKwUser(CqlParser.KwUserContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwUser}.
	 * @param ctx the parse tree
	 */
	void exitKwUser(CqlParser.KwUserContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwUsing}.
	 * @param ctx the parse tree
	 */
	void enterKwUsing(CqlParser.KwUsingContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwUsing}.
	 * @param ctx the parse tree
	 */
	void exitKwUsing(CqlParser.KwUsingContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwValues}.
	 * @param ctx the parse tree
	 */
	void enterKwValues(CqlParser.KwValuesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwValues}.
	 * @param ctx the parse tree
	 */
	void exitKwValues(CqlParser.KwValuesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwView}.
	 * @param ctx the parse tree
	 */
	void enterKwView(CqlParser.KwViewContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwView}.
	 * @param ctx the parse tree
	 */
	void exitKwView(CqlParser.KwViewContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwWhere}.
	 * @param ctx the parse tree
	 */
	void enterKwWhere(CqlParser.KwWhereContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwWhere}.
	 * @param ctx the parse tree
	 */
	void exitKwWhere(CqlParser.KwWhereContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwWith}.
	 * @param ctx the parse tree
	 */
	void enterKwWith(CqlParser.KwWithContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwWith}.
	 * @param ctx the parse tree
	 */
	void exitKwWith(CqlParser.KwWithContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#kwRevoke}.
	 * @param ctx the parse tree
	 */
	void enterKwRevoke(CqlParser.KwRevokeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#kwRevoke}.
	 * @param ctx the parse tree
	 */
	void exitKwRevoke(CqlParser.KwRevokeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketLr}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketLr(CqlParser.SyntaxBracketLrContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketLr}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketLr(CqlParser.SyntaxBracketLrContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketRr}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketRr(CqlParser.SyntaxBracketRrContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketRr}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketRr(CqlParser.SyntaxBracketRrContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketLc}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketLc(CqlParser.SyntaxBracketLcContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketLc}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketLc(CqlParser.SyntaxBracketLcContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketRc}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketRc(CqlParser.SyntaxBracketRcContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketRc}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketRc(CqlParser.SyntaxBracketRcContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketLa}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketLa(CqlParser.SyntaxBracketLaContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketLa}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketLa(CqlParser.SyntaxBracketLaContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketRa}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketRa(CqlParser.SyntaxBracketRaContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketRa}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketRa(CqlParser.SyntaxBracketRaContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketLs}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketLs(CqlParser.SyntaxBracketLsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketLs}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketLs(CqlParser.SyntaxBracketLsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxBracketRs}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxBracketRs(CqlParser.SyntaxBracketRsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxBracketRs}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxBracketRs(CqlParser.SyntaxBracketRsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxComma}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxComma(CqlParser.SyntaxCommaContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxComma}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxComma(CqlParser.SyntaxCommaContext ctx);
	/**
	 * Enter a parse tree produced by {@link CqlParser#syntaxColon}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxColon(CqlParser.SyntaxColonContext ctx);
	/**
	 * Exit a parse tree produced by {@link CqlParser#syntaxColon}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxColon(CqlParser.SyntaxColonContext ctx);
}