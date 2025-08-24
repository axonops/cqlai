// Generated from /home/hayato/git/cqlai/internal/parser/grammar/CqlParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class CqlParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		LR_BRACKET=1, RR_BRACKET=2, LC_BRACKET=3, RC_BRACKET=4, LS_BRACKET=5, 
		RS_BRACKET=6, COMMA=7, SEMI=8, COLON=9, DOT=10, STAR=11, DIVIDE=12, MODULE=13, 
		PLUS=14, MINUSMINUS=15, MINUS=16, DQUOTE=17, SQUOTE=18, OPERATOR_EQ=19, 
		OPERATOR_LT=20, OPERATOR_GT=21, OPERATOR_LTE=22, OPERATOR_GTE=23, K_ADD=24, 
		K_AGGREGATE=25, K_AGGREGATES=26, K_ALL=27, K_ALLOW=28, K_ALTER=29, K_AND=30, 
		K_ANY=31, K_APPLY=32, K_AS=33, K_ASC=34, K_AUTHORIZE=35, K_BATCH=36, K_BEGIN=37, 
		K_BY=38, K_CALLED=39, K_CLUSTER=40, K_CLUSTERING=41, K_COLUMNFAMILY=42, 
		K_COMPACT=43, K_CONNECTION=44, K_CONSISTENCY=45, K_CONTAINS=46, K_CREATE=47, 
		K_CUSTOM=48, K_DELETE=49, K_DESC=50, K_DESCRIBE=51, K_DISTINCT=52, K_DROP=53, 
		K_DURABLE_WRITES=54, K_EACH_QUORUM=55, K_ENTRIES=56, K_EXECUTE=57, K_EXISTS=58, 
		K_FALSE=59, K_FILTERING=60, K_FINALFUNC=61, K_FROM=62, K_FULL=63, K_FUNCTION=64, 
		K_FUNCTIONS=65, K_GRANT=66, K_IF=67, K_IN=68, K_INDEX=69, K_INFINITY=70, 
		K_INITCOND=71, K_INPUT=72, K_INSERT=73, K_INTO=74, K_IS=75, K_JSON=76, 
		K_KEY=77, K_KEYS=78, K_KEYSPACE=79, K_KEYSPACES=80, K_LANGUAGE=81, K_LEVEL=82, 
		K_LIMIT=83, K_LOCAL_ONE=84, K_LOCAL_QUORUM=85, K_LOCAL_SERIAL=86, K_LOGGED=87, 
		K_LOGIN=88, K_MATERIALIZED=89, K_MODIFY=90, K_NAN=91, K_NORECURSIVE=92, 
		K_NOSUPERUSER=93, K_NOT=94, K_NULL=95, K_OF=96, K_ON=97, K_ONE=98, K_OPTIONS=99, 
		K_OR=100, K_ORDER=101, K_OUTPUT=102, K_PARTITION=103, K_PASSWORD=104, 
		K_PER=105, K_PERMISSION=106, K_PERMISSIONS=107, K_PRIMARY=108, K_QUORUM=109, 
		K_RENAME=110, K_REPLACE=111, K_REPLICATION=112, K_RETURNS=113, K_REVOKE=114, 
		K_ROLE=115, K_ROLES=116, K_SCHEMA=117, K_SELECT=118, K_SERIAL=119, K_SET=120, 
		K_SFUNC=121, K_STATIC=122, K_STORAGE=123, K_STYPE=124, K_SUPERUSER=125, 
		K_TABLE=126, K_TABLES=127, K_THREE=128, K_TIMESTAMP=129, K_TO=130, K_TOKEN=131, 
		K_TRIGGER=132, K_TRUE=133, K_TRUNCATE=134, K_TTL=135, K_TWO=136, K_TYPE=137, 
		K_TYPES=138, K_UNLOGGED=139, K_UPDATE=140, K_USE=141, K_USER=142, K_USING=143, 
		K_UUID=144, K_VALUES=145, K_VIEW=146, K_WHERE=147, K_WITH=148, K_WRITETIME=149, 
		K_ASCII=150, K_BIGINT=151, K_BLOB=152, K_BOOLEAN=153, K_COUNTER=154, K_DATE=155, 
		K_DECIMAL=156, K_DOUBLE=157, K_FLOAT=158, K_FROZEN=159, K_INET=160, K_INT=161, 
		K_LIST=162, K_MAP=163, K_SMALLINT=164, K_TEXT=165, K_TIMEUUID=166, K_TIME=167, 
		K_TINYINT=168, K_TUPLE=169, K_VARCHAR=170, K_VARINT=171, CODE_BLOCK=172, 
		STRING_LITERAL=173, DECIMAL_LITERAL=174, FLOAT_LITERAL=175, HEXADECIMAL_LITERAL=176, 
		REAL_LITERAL=177, OBJECT_NAME=178, UUID=179, SPACE=180, SPEC_MYSQL_COMMENT=181, 
		COMMENT_INPUT=182, LINE_COMMENT=183;
	public static final int
		RULE_root = 0, RULE_cqls = 1, RULE_statementSeparator = 2, RULE_empty_ = 3, 
		RULE_cql = 4, RULE_describeCommand = 5, RULE_consistencyCommand = 6, RULE_outputFormatCommand = 7, 
		RULE_revoke = 8, RULE_listRoles = 9, RULE_listPermissions = 10, RULE_grant = 11, 
		RULE_priviledge = 12, RULE_resource = 13, RULE_createUser = 14, RULE_createRole = 15, 
		RULE_createType = 16, RULE_typeMemberColumnList = 17, RULE_createTrigger = 18, 
		RULE_createMaterializedView = 19, RULE_materializedViewWhere = 20, RULE_columnNotNullList = 21, 
		RULE_columnNotNull = 22, RULE_materializedViewOptions = 23, RULE_createKeyspace = 24, 
		RULE_createFunction = 25, RULE_codeBlock = 26, RULE_paramList = 27, RULE_returnMode = 28, 
		RULE_createAggregate = 29, RULE_initCondDefinition = 30, RULE_initCondHash = 31, 
		RULE_initCondHashItem = 32, RULE_initCondListNested = 33, RULE_initCondList = 34, 
		RULE_orReplace = 35, RULE_alterUser = 36, RULE_userPassword = 37, RULE_userSuperUser = 38, 
		RULE_alterType = 39, RULE_alterTypeOperation = 40, RULE_alterTypeRename = 41, 
		RULE_alterTypeRenameList = 42, RULE_alterTypeRenameItem = 43, RULE_alterTypeAdd = 44, 
		RULE_alterTypeAlterType = 45, RULE_alterTable = 46, RULE_alterTableOperation = 47, 
		RULE_alterTableWith = 48, RULE_alterTableRename = 49, RULE_alterTableDropCompactStorage = 50, 
		RULE_alterTableDropColumns = 51, RULE_alterTableDropColumnList = 52, RULE_alterTableAdd = 53, 
		RULE_alterTableColumnDefinition = 54, RULE_alterRole = 55, RULE_roleWith = 56, 
		RULE_roleWithOptions = 57, RULE_alterMaterializedView = 58, RULE_dropUser = 59, 
		RULE_dropType = 60, RULE_dropMaterializedView = 61, RULE_dropAggregate = 62, 
		RULE_dropFunction = 63, RULE_dropTrigger = 64, RULE_dropRole = 65, RULE_dropTable = 66, 
		RULE_dropKeyspace = 67, RULE_dropIndex = 68, RULE_createTable = 69, RULE_withElement = 70, 
		RULE_tableOptions = 71, RULE_clusteringOrder = 72, RULE_tableOptionItem = 73, 
		RULE_tableOptionName = 74, RULE_tableOptionValue = 75, RULE_optionHash = 76, 
		RULE_optionHashItem = 77, RULE_optionHashKey = 78, RULE_optionHashValue = 79, 
		RULE_columnDefinitionList = 80, RULE_columnDefinition = 81, RULE_primaryKeyColumn = 82, 
		RULE_primaryKeyElement = 83, RULE_primaryKeyDefinition = 84, RULE_singlePrimaryKey = 85, 
		RULE_compoundKey = 86, RULE_compositeKey = 87, RULE_partitionKeyList = 88, 
		RULE_clusteringKeyList = 89, RULE_partitionKey = 90, RULE_clusteringKey = 91, 
		RULE_applyBatch = 92, RULE_beginBatch = 93, RULE_batchType = 94, RULE_alterKeyspace = 95, 
		RULE_replicationList = 96, RULE_replicationListItem = 97, RULE_durableWrites = 98, 
		RULE_use_ = 99, RULE_truncate = 100, RULE_createIndex = 101, RULE_indexName = 102, 
		RULE_indexColumnSpec = 103, RULE_indexKeysSpec = 104, RULE_indexEntriesSSpec = 105, 
		RULE_indexFullSpec = 106, RULE_delete_ = 107, RULE_deleteColumnList = 108, 
		RULE_deleteColumnItem = 109, RULE_update = 110, RULE_ifSpec = 111, RULE_ifConditionList = 112, 
		RULE_ifCondition = 113, RULE_assignments = 114, RULE_assignmentElement = 115, 
		RULE_assignmentSet = 116, RULE_assignmentMap = 117, RULE_assignmentList = 118, 
		RULE_assignmentTuple = 119, RULE_insert = 120, RULE_usingTtlTimestamp = 121, 
		RULE_timestamp = 122, RULE_ttl = 123, RULE_usingTimestampSpec = 124, RULE_ifNotExist = 125, 
		RULE_ifExist = 126, RULE_insertValuesSpec = 127, RULE_insertColumnSpec = 128, 
		RULE_columnList = 129, RULE_expressionList = 130, RULE_expression = 131, 
		RULE_select_ = 132, RULE_allowFilteringSpec = 133, RULE_limitSpec = 134, 
		RULE_fromSpec = 135, RULE_fromSpecElement = 136, RULE_orderSpec = 137, 
		RULE_orderSpecElement = 138, RULE_whereSpec = 139, RULE_distinctSpec = 140, 
		RULE_selectElements = 141, RULE_selectElement = 142, RULE_relationElements = 143, 
		RULE_relationElement = 144, RULE_relalationContains = 145, RULE_relalationContainsKey = 146, 
		RULE_functionCall = 147, RULE_functionArgs = 148, RULE_constant = 149, 
		RULE_decimalLiteral = 150, RULE_floatLiteral = 151, RULE_stringLiteral = 152, 
		RULE_booleanLiteral = 153, RULE_hexadecimalLiteral = 154, RULE_keyspace = 155, 
		RULE_table = 156, RULE_column = 157, RULE_dataType = 158, RULE_dataTypeName = 159, 
		RULE_dataTypeDefinition = 160, RULE_orderDirection = 161, RULE_role = 162, 
		RULE_trigger = 163, RULE_triggerClass = 164, RULE_materializedView = 165, 
		RULE_type_ = 166, RULE_aggregate = 167, RULE_function_ = 168, RULE_language = 169, 
		RULE_user = 170, RULE_password = 171, RULE_hashKey = 172, RULE_param = 173, 
		RULE_paramName = 174, RULE_kwAdd = 175, RULE_kwAggregate = 176, RULE_kwAggregates = 177, 
		RULE_kwAll = 178, RULE_kwAllPermissions = 179, RULE_kwAllow = 180, RULE_kwAlter = 181, 
		RULE_kwAnd = 182, RULE_kwApply = 183, RULE_kwAs = 184, RULE_kwAsc = 185, 
		RULE_kwAuthorize = 186, RULE_kwBatch = 187, RULE_kwBegin = 188, RULE_kwBy = 189, 
		RULE_kwCalled = 190, RULE_kwCluster = 191, RULE_kwClustering = 192, RULE_kwCompact = 193, 
		RULE_kwConnection = 194, RULE_kwConsistency = 195, RULE_kwConsistencyLevel = 196, 
		RULE_kwOutput = 197, RULE_kwOutputFormatType = 198, RULE_kwContains = 199, 
		RULE_kwCreate = 200, RULE_kwDelete = 201, RULE_kwDesc = 202, RULE_kwDescribe = 203, 
		RULE_kwDistinct = 204, RULE_kwDrop = 205, RULE_kwDurableWrites = 206, 
		RULE_kwEntries = 207, RULE_kwExecute = 208, RULE_kwExists = 209, RULE_kwFiltering = 210, 
		RULE_kwFinalfunc = 211, RULE_kwFrom = 212, RULE_kwFull = 213, RULE_kwFunction = 214, 
		RULE_kwFunctions = 215, RULE_kwGrant = 216, RULE_kwIf = 217, RULE_kwIn = 218, 
		RULE_kwIndex = 219, RULE_kwInitcond = 220, RULE_kwInput = 221, RULE_kwInsert = 222, 
		RULE_kwInto = 223, RULE_kwIs = 224, RULE_kwJson = 225, RULE_kwKey = 226, 
		RULE_kwKeys = 227, RULE_kwKeyspace = 228, RULE_kwKeyspaces = 229, RULE_kwLanguage = 230, 
		RULE_kwLimit = 231, RULE_kwList = 232, RULE_kwLogged = 233, RULE_kwLocalSerial = 234, 
		RULE_kwLogin = 235, RULE_kwMaterialized = 236, RULE_kwModify = 237, RULE_kwNosuperuser = 238, 
		RULE_kwNorecursive = 239, RULE_kwNot = 240, RULE_kwNull = 241, RULE_kwOf = 242, 
		RULE_kwOn = 243, RULE_kwOptions = 244, RULE_kwOr = 245, RULE_kwOrder = 246, 
		RULE_kwPassword = 247, RULE_kwPrimary = 248, RULE_kwRename = 249, RULE_kwReplace = 250, 
		RULE_kwReplication = 251, RULE_kwReturns = 252, RULE_kwRole = 253, RULE_kwRoles = 254, 
		RULE_kwSelect = 255, RULE_kwSerial = 256, RULE_kwSet = 257, RULE_kwSfunc = 258, 
		RULE_kwStorage = 259, RULE_kwStype = 260, RULE_kwSuperuser = 261, RULE_kwTable = 262, 
		RULE_kwTables = 263, RULE_kwTimestamp = 264, RULE_kwTo = 265, RULE_kwTrigger = 266, 
		RULE_kwTruncate = 267, RULE_kwTtl = 268, RULE_kwType = 269, RULE_kwTypes = 270, 
		RULE_kwUnlogged = 271, RULE_kwUpdate = 272, RULE_kwUse = 273, RULE_kwUser = 274, 
		RULE_kwUsing = 275, RULE_kwValues = 276, RULE_kwView = 277, RULE_kwWhere = 278, 
		RULE_kwWith = 279, RULE_kwRevoke = 280, RULE_syntaxBracketLr = 281, RULE_syntaxBracketRr = 282, 
		RULE_syntaxBracketLc = 283, RULE_syntaxBracketRc = 284, RULE_syntaxBracketLa = 285, 
		RULE_syntaxBracketRa = 286, RULE_syntaxBracketLs = 287, RULE_syntaxBracketRs = 288, 
		RULE_syntaxComma = 289, RULE_syntaxColon = 290;
	private static String[] makeRuleNames() {
		return new String[] {
			"root", "cqls", "statementSeparator", "empty_", "cql", "describeCommand", 
			"consistencyCommand", "outputFormatCommand", "revoke", "listRoles", "listPermissions", 
			"grant", "priviledge", "resource", "createUser", "createRole", "createType", 
			"typeMemberColumnList", "createTrigger", "createMaterializedView", "materializedViewWhere", 
			"columnNotNullList", "columnNotNull", "materializedViewOptions", "createKeyspace", 
			"createFunction", "codeBlock", "paramList", "returnMode", "createAggregate", 
			"initCondDefinition", "initCondHash", "initCondHashItem", "initCondListNested", 
			"initCondList", "orReplace", "alterUser", "userPassword", "userSuperUser", 
			"alterType", "alterTypeOperation", "alterTypeRename", "alterTypeRenameList", 
			"alterTypeRenameItem", "alterTypeAdd", "alterTypeAlterType", "alterTable", 
			"alterTableOperation", "alterTableWith", "alterTableRename", "alterTableDropCompactStorage", 
			"alterTableDropColumns", "alterTableDropColumnList", "alterTableAdd", 
			"alterTableColumnDefinition", "alterRole", "roleWith", "roleWithOptions", 
			"alterMaterializedView", "dropUser", "dropType", "dropMaterializedView", 
			"dropAggregate", "dropFunction", "dropTrigger", "dropRole", "dropTable", 
			"dropKeyspace", "dropIndex", "createTable", "withElement", "tableOptions", 
			"clusteringOrder", "tableOptionItem", "tableOptionName", "tableOptionValue", 
			"optionHash", "optionHashItem", "optionHashKey", "optionHashValue", "columnDefinitionList", 
			"columnDefinition", "primaryKeyColumn", "primaryKeyElement", "primaryKeyDefinition", 
			"singlePrimaryKey", "compoundKey", "compositeKey", "partitionKeyList", 
			"clusteringKeyList", "partitionKey", "clusteringKey", "applyBatch", "beginBatch", 
			"batchType", "alterKeyspace", "replicationList", "replicationListItem", 
			"durableWrites", "use_", "truncate", "createIndex", "indexName", "indexColumnSpec", 
			"indexKeysSpec", "indexEntriesSSpec", "indexFullSpec", "delete_", "deleteColumnList", 
			"deleteColumnItem", "update", "ifSpec", "ifConditionList", "ifCondition", 
			"assignments", "assignmentElement", "assignmentSet", "assignmentMap", 
			"assignmentList", "assignmentTuple", "insert", "usingTtlTimestamp", "timestamp", 
			"ttl", "usingTimestampSpec", "ifNotExist", "ifExist", "insertValuesSpec", 
			"insertColumnSpec", "columnList", "expressionList", "expression", "select_", 
			"allowFilteringSpec", "limitSpec", "fromSpec", "fromSpecElement", "orderSpec", 
			"orderSpecElement", "whereSpec", "distinctSpec", "selectElements", "selectElement", 
			"relationElements", "relationElement", "relalationContains", "relalationContainsKey", 
			"functionCall", "functionArgs", "constant", "decimalLiteral", "floatLiteral", 
			"stringLiteral", "booleanLiteral", "hexadecimalLiteral", "keyspace", 
			"table", "column", "dataType", "dataTypeName", "dataTypeDefinition", 
			"orderDirection", "role", "trigger", "triggerClass", "materializedView", 
			"type_", "aggregate", "function_", "language", "user", "password", "hashKey", 
			"param", "paramName", "kwAdd", "kwAggregate", "kwAggregates", "kwAll", 
			"kwAllPermissions", "kwAllow", "kwAlter", "kwAnd", "kwApply", "kwAs", 
			"kwAsc", "kwAuthorize", "kwBatch", "kwBegin", "kwBy", "kwCalled", "kwCluster", 
			"kwClustering", "kwCompact", "kwConnection", "kwConsistency", "kwConsistencyLevel", 
			"kwOutput", "kwOutputFormatType", "kwContains", "kwCreate", "kwDelete", 
			"kwDesc", "kwDescribe", "kwDistinct", "kwDrop", "kwDurableWrites", "kwEntries", 
			"kwExecute", "kwExists", "kwFiltering", "kwFinalfunc", "kwFrom", "kwFull", 
			"kwFunction", "kwFunctions", "kwGrant", "kwIf", "kwIn", "kwIndex", "kwInitcond", 
			"kwInput", "kwInsert", "kwInto", "kwIs", "kwJson", "kwKey", "kwKeys", 
			"kwKeyspace", "kwKeyspaces", "kwLanguage", "kwLimit", "kwList", "kwLogged", 
			"kwLocalSerial", "kwLogin", "kwMaterialized", "kwModify", "kwNosuperuser", 
			"kwNorecursive", "kwNot", "kwNull", "kwOf", "kwOn", "kwOptions", "kwOr", 
			"kwOrder", "kwPassword", "kwPrimary", "kwRename", "kwReplace", "kwReplication", 
			"kwReturns", "kwRole", "kwRoles", "kwSelect", "kwSerial", "kwSet", "kwSfunc", 
			"kwStorage", "kwStype", "kwSuperuser", "kwTable", "kwTables", "kwTimestamp", 
			"kwTo", "kwTrigger", "kwTruncate", "kwTtl", "kwType", "kwTypes", "kwUnlogged", 
			"kwUpdate", "kwUse", "kwUser", "kwUsing", "kwValues", "kwView", "kwWhere", 
			"kwWith", "kwRevoke", "syntaxBracketLr", "syntaxBracketRr", "syntaxBracketLc", 
			"syntaxBracketRc", "syntaxBracketLa", "syntaxBracketRa", "syntaxBracketLs", 
			"syntaxBracketRs", "syntaxComma", "syntaxColon"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'('", "')'", "'{'", "'}'", "'['", "']'", "','", "';'", "':'", 
			"'.'", "'*'", "'/'", "'%'", "'+'", "'--'", "'-'", "'\"'", "'''", "'='", 
			"'<'", "'>'", "'<='", "'>='", "'ADD'", "'AGGREGATE'", "'AGGREGATES'", 
			"'ALL'", "'ALLOW'", "'ALTER'", "'AND'", "'ANY'", "'APPLY'", "'AS'", "'ASC'", 
			"'AUTHORIZE'", "'BATCH'", "'BEGIN'", "'BY'", "'CALLED'", "'CLUSTER'", 
			"'CLUSTERING'", "'COLUMNFAMILY'", "'COMPACT'", "'CONNECTION'", "'CONSISTENCY'", 
			"'CONTAINS'", "'CREATE'", "'CUSTOM'", "'DELETE'", "'DESC'", "'DESCRIBE'", 
			"'DISTINCT'", "'DROP'", "'DURABLE_WRITES'", "'EACH_QUORUM'", "'ENTRIES'", 
			"'EXECUTE'", "'EXISTS'", "'FALSE'", "'FILTERING'", "'FINALFUNC'", "'FROM'", 
			"'FULL'", "'FUNCTION'", "'FUNCTIONS'", "'GRANT'", "'IF'", "'IN'", "'INDEX'", 
			"'INFINITY'", "'INITCOND'", "'INPUT'", "'INSERT'", "'INTO'", "'IS'", 
			"'JSON'", "'KEY'", "'KEYS'", "'KEYSPACE'", "'KEYSPACES'", "'LANGUAGE'", 
			"'LEVEL'", "'LIMIT'", "'LOCAL_ONE'", "'LOCAL_QUORUM'", "'LOCAL_SERIAL'", 
			"'LOGGED'", "'LOGIN'", "'MATERIALIZED'", "'MODIFY'", "'NAN'", "'NORECURSIVE'", 
			"'NOSUPERUSER'", "'NOT'", "'NULL'", "'OF'", "'ON'", "'ONE'", "'OPTIONS'", 
			"'OR'", "'ORDER'", "'OUTPUT'", "'PARTITION'", "'PASSWORD'", "'PER'", 
			"'PERMISSION'", "'PERMISSIONS'", "'PRIMARY'", "'QUORUM'", "'RENAME'", 
			"'REPLACE'", "'REPLICATION'", "'RETURNS'", "'REVOKE'", "'ROLE'", "'ROLES'", 
			"'SCHEMA'", "'SELECT'", "'SERIAL'", "'SET'", "'SFUNC'", "'STATIC'", "'STORAGE'", 
			"'STYPE'", "'SUPERUSER'", "'TABLE'", "'TABLES'", "'THREE'", "'TIMESTAMP'", 
			"'TO'", "'TOKEN'", "'TRIGGER'", "'TRUE'", "'TRUNCATE'", "'TTL'", "'TWO'", 
			"'TYPE'", "'TYPES'", "'UNLOGGED'", "'UPDATE'", "'USE'", "'USER'", "'USING'", 
			"'UUID'", "'VALUES'", "'VIEW'", "'WHERE'", "'WITH'", "'WRITETIME'", "'ASCII'", 
			"'BIGINT'", "'BLOB'", "'BOOLEAN'", "'COUNTER'", "'DATE'", "'DECIMAL'", 
			"'DOUBLE'", "'FLOAT'", "'FROZEN'", "'INET'", "'INT'", "'LIST'", "'MAP'", 
			"'SMALLINT'", "'TEXT'", "'TIMEUUID'", "'TIME'", "'TINYINT'", "'TUPLE'", 
			"'VARCHAR'", "'VARINT'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, "LR_BRACKET", "RR_BRACKET", "LC_BRACKET", "RC_BRACKET", "LS_BRACKET", 
			"RS_BRACKET", "COMMA", "SEMI", "COLON", "DOT", "STAR", "DIVIDE", "MODULE", 
			"PLUS", "MINUSMINUS", "MINUS", "DQUOTE", "SQUOTE", "OPERATOR_EQ", "OPERATOR_LT", 
			"OPERATOR_GT", "OPERATOR_LTE", "OPERATOR_GTE", "K_ADD", "K_AGGREGATE", 
			"K_AGGREGATES", "K_ALL", "K_ALLOW", "K_ALTER", "K_AND", "K_ANY", "K_APPLY", 
			"K_AS", "K_ASC", "K_AUTHORIZE", "K_BATCH", "K_BEGIN", "K_BY", "K_CALLED", 
			"K_CLUSTER", "K_CLUSTERING", "K_COLUMNFAMILY", "K_COMPACT", "K_CONNECTION", 
			"K_CONSISTENCY", "K_CONTAINS", "K_CREATE", "K_CUSTOM", "K_DELETE", "K_DESC", 
			"K_DESCRIBE", "K_DISTINCT", "K_DROP", "K_DURABLE_WRITES", "K_EACH_QUORUM", 
			"K_ENTRIES", "K_EXECUTE", "K_EXISTS", "K_FALSE", "K_FILTERING", "K_FINALFUNC", 
			"K_FROM", "K_FULL", "K_FUNCTION", "K_FUNCTIONS", "K_GRANT", "K_IF", "K_IN", 
			"K_INDEX", "K_INFINITY", "K_INITCOND", "K_INPUT", "K_INSERT", "K_INTO", 
			"K_IS", "K_JSON", "K_KEY", "K_KEYS", "K_KEYSPACE", "K_KEYSPACES", "K_LANGUAGE", 
			"K_LEVEL", "K_LIMIT", "K_LOCAL_ONE", "K_LOCAL_QUORUM", "K_LOCAL_SERIAL", 
			"K_LOGGED", "K_LOGIN", "K_MATERIALIZED", "K_MODIFY", "K_NAN", "K_NORECURSIVE", 
			"K_NOSUPERUSER", "K_NOT", "K_NULL", "K_OF", "K_ON", "K_ONE", "K_OPTIONS", 
			"K_OR", "K_ORDER", "K_OUTPUT", "K_PARTITION", "K_PASSWORD", "K_PER", 
			"K_PERMISSION", "K_PERMISSIONS", "K_PRIMARY", "K_QUORUM", "K_RENAME", 
			"K_REPLACE", "K_REPLICATION", "K_RETURNS", "K_REVOKE", "K_ROLE", "K_ROLES", 
			"K_SCHEMA", "K_SELECT", "K_SERIAL", "K_SET", "K_SFUNC", "K_STATIC", "K_STORAGE", 
			"K_STYPE", "K_SUPERUSER", "K_TABLE", "K_TABLES", "K_THREE", "K_TIMESTAMP", 
			"K_TO", "K_TOKEN", "K_TRIGGER", "K_TRUE", "K_TRUNCATE", "K_TTL", "K_TWO", 
			"K_TYPE", "K_TYPES", "K_UNLOGGED", "K_UPDATE", "K_USE", "K_USER", "K_USING", 
			"K_UUID", "K_VALUES", "K_VIEW", "K_WHERE", "K_WITH", "K_WRITETIME", "K_ASCII", 
			"K_BIGINT", "K_BLOB", "K_BOOLEAN", "K_COUNTER", "K_DATE", "K_DECIMAL", 
			"K_DOUBLE", "K_FLOAT", "K_FROZEN", "K_INET", "K_INT", "K_LIST", "K_MAP", 
			"K_SMALLINT", "K_TEXT", "K_TIMEUUID", "K_TIME", "K_TINYINT", "K_TUPLE", 
			"K_VARCHAR", "K_VARINT", "CODE_BLOCK", "STRING_LITERAL", "DECIMAL_LITERAL", 
			"FLOAT_LITERAL", "HEXADECIMAL_LITERAL", "REAL_LITERAL", "OBJECT_NAME", 
			"UUID", "SPACE", "SPEC_MYSQL_COMMENT", "COMMENT_INPUT", "LINE_COMMENT"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "CqlParser.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public CqlParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RootContext extends ParserRuleContext {
		public TerminalNode EOF() { return getToken(CqlParser.EOF, 0); }
		public CqlsContext cqls() {
			return getRuleContext(CqlsContext.class,0);
		}
		public TerminalNode MINUSMINUS() { return getToken(CqlParser.MINUSMINUS, 0); }
		public RootContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_root; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRoot(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRoot(this);
		}
	}

	public final RootContext root() throws RecognitionException {
		RootContext _localctx = new RootContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_root);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(583);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if ((((_la) & ~0x3f) == 0 && ((1L << _la) & 13123913059926272L) != 0) || ((((_la - 66)) & ~0x3f) == 0 && ((1L << (_la - 66)) & 4785143323558017L) != 0) || ((((_la - 134)) & ~0x3f) == 0 && ((1L << (_la - 134)) & 268435649L) != 0)) {
				{
				setState(582);
				cqls();
				}
			}

			setState(586);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==MINUSMINUS) {
				{
				setState(585);
				match(MINUSMINUS);
				}
			}

			setState(588);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CqlsContext extends ParserRuleContext {
		public List<CqlContext> cql() {
			return getRuleContexts(CqlContext.class);
		}
		public CqlContext cql(int i) {
			return getRuleContext(CqlContext.class,i);
		}
		public List<Empty_Context> empty_() {
			return getRuleContexts(Empty_Context.class);
		}
		public Empty_Context empty_(int i) {
			return getRuleContext(Empty_Context.class,i);
		}
		public List<StatementSeparatorContext> statementSeparator() {
			return getRuleContexts(StatementSeparatorContext.class);
		}
		public StatementSeparatorContext statementSeparator(int i) {
			return getRuleContext(StatementSeparatorContext.class,i);
		}
		public List<TerminalNode> MINUSMINUS() { return getTokens(CqlParser.MINUSMINUS); }
		public TerminalNode MINUSMINUS(int i) {
			return getToken(CqlParser.MINUSMINUS, i);
		}
		public CqlsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_cqls; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCqls(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCqls(this);
		}
	}

	public final CqlsContext cqls() throws RecognitionException {
		CqlsContext _localctx = new CqlsContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_cqls);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(599);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					{
					setState(597);
					_errHandler.sync(this);
					switch (_input.LA(1)) {
					case K_ALTER:
					case K_APPLY:
					case K_BEGIN:
					case K_CONSISTENCY:
					case K_CREATE:
					case K_DELETE:
					case K_DESC:
					case K_DESCRIBE:
					case K_DROP:
					case K_GRANT:
					case K_INSERT:
					case K_OUTPUT:
					case K_REVOKE:
					case K_SELECT:
					case K_TRUNCATE:
					case K_UPDATE:
					case K_USE:
					case K_LIST:
						{
						setState(590);
						cql();
						setState(592);
						_errHandler.sync(this);
						_la = _input.LA(1);
						if (_la==MINUSMINUS) {
							{
							setState(591);
							match(MINUSMINUS);
							}
						}

						setState(594);
						statementSeparator();
						}
						break;
					case SEMI:
						{
						setState(596);
						empty_();
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					} 
				}
				setState(601);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,4,_ctx);
			}
			setState(610);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_ALTER:
			case K_APPLY:
			case K_BEGIN:
			case K_CONSISTENCY:
			case K_CREATE:
			case K_DELETE:
			case K_DESC:
			case K_DESCRIBE:
			case K_DROP:
			case K_GRANT:
			case K_INSERT:
			case K_OUTPUT:
			case K_REVOKE:
			case K_SELECT:
			case K_TRUNCATE:
			case K_UPDATE:
			case K_USE:
			case K_LIST:
				{
				setState(602);
				cql();
				setState(607);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,6,_ctx) ) {
				case 1:
					{
					setState(604);
					_errHandler.sync(this);
					_la = _input.LA(1);
					if (_la==MINUSMINUS) {
						{
						setState(603);
						match(MINUSMINUS);
						}
					}

					setState(606);
					statementSeparator();
					}
					break;
				}
				}
				break;
			case SEMI:
				{
				setState(609);
				empty_();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StatementSeparatorContext extends ParserRuleContext {
		public TerminalNode SEMI() { return getToken(CqlParser.SEMI, 0); }
		public StatementSeparatorContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statementSeparator; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterStatementSeparator(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitStatementSeparator(this);
		}
	}

	public final StatementSeparatorContext statementSeparator() throws RecognitionException {
		StatementSeparatorContext _localctx = new StatementSeparatorContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_statementSeparator);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(612);
			match(SEMI);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Empty_Context extends ParserRuleContext {
		public StatementSeparatorContext statementSeparator() {
			return getRuleContext(StatementSeparatorContext.class,0);
		}
		public Empty_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_empty_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterEmpty_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitEmpty_(this);
		}
	}

	public final Empty_Context empty_() throws RecognitionException {
		Empty_Context _localctx = new Empty_Context(_ctx, getState());
		enterRule(_localctx, 6, RULE_empty_);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(614);
			statementSeparator();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CqlContext extends ParserRuleContext {
		public AlterKeyspaceContext alterKeyspace() {
			return getRuleContext(AlterKeyspaceContext.class,0);
		}
		public AlterMaterializedViewContext alterMaterializedView() {
			return getRuleContext(AlterMaterializedViewContext.class,0);
		}
		public AlterRoleContext alterRole() {
			return getRuleContext(AlterRoleContext.class,0);
		}
		public AlterTableContext alterTable() {
			return getRuleContext(AlterTableContext.class,0);
		}
		public AlterTypeContext alterType() {
			return getRuleContext(AlterTypeContext.class,0);
		}
		public AlterUserContext alterUser() {
			return getRuleContext(AlterUserContext.class,0);
		}
		public ApplyBatchContext applyBatch() {
			return getRuleContext(ApplyBatchContext.class,0);
		}
		public ConsistencyCommandContext consistencyCommand() {
			return getRuleContext(ConsistencyCommandContext.class,0);
		}
		public OutputFormatCommandContext outputFormatCommand() {
			return getRuleContext(OutputFormatCommandContext.class,0);
		}
		public CreateAggregateContext createAggregate() {
			return getRuleContext(CreateAggregateContext.class,0);
		}
		public CreateFunctionContext createFunction() {
			return getRuleContext(CreateFunctionContext.class,0);
		}
		public CreateIndexContext createIndex() {
			return getRuleContext(CreateIndexContext.class,0);
		}
		public CreateKeyspaceContext createKeyspace() {
			return getRuleContext(CreateKeyspaceContext.class,0);
		}
		public CreateMaterializedViewContext createMaterializedView() {
			return getRuleContext(CreateMaterializedViewContext.class,0);
		}
		public CreateRoleContext createRole() {
			return getRuleContext(CreateRoleContext.class,0);
		}
		public CreateTableContext createTable() {
			return getRuleContext(CreateTableContext.class,0);
		}
		public CreateTriggerContext createTrigger() {
			return getRuleContext(CreateTriggerContext.class,0);
		}
		public CreateTypeContext createType() {
			return getRuleContext(CreateTypeContext.class,0);
		}
		public CreateUserContext createUser() {
			return getRuleContext(CreateUserContext.class,0);
		}
		public Delete_Context delete_() {
			return getRuleContext(Delete_Context.class,0);
		}
		public DescribeCommandContext describeCommand() {
			return getRuleContext(DescribeCommandContext.class,0);
		}
		public DropAggregateContext dropAggregate() {
			return getRuleContext(DropAggregateContext.class,0);
		}
		public DropFunctionContext dropFunction() {
			return getRuleContext(DropFunctionContext.class,0);
		}
		public DropIndexContext dropIndex() {
			return getRuleContext(DropIndexContext.class,0);
		}
		public DropKeyspaceContext dropKeyspace() {
			return getRuleContext(DropKeyspaceContext.class,0);
		}
		public DropMaterializedViewContext dropMaterializedView() {
			return getRuleContext(DropMaterializedViewContext.class,0);
		}
		public DropRoleContext dropRole() {
			return getRuleContext(DropRoleContext.class,0);
		}
		public DropTableContext dropTable() {
			return getRuleContext(DropTableContext.class,0);
		}
		public DropTriggerContext dropTrigger() {
			return getRuleContext(DropTriggerContext.class,0);
		}
		public DropTypeContext dropType() {
			return getRuleContext(DropTypeContext.class,0);
		}
		public DropUserContext dropUser() {
			return getRuleContext(DropUserContext.class,0);
		}
		public GrantContext grant() {
			return getRuleContext(GrantContext.class,0);
		}
		public InsertContext insert() {
			return getRuleContext(InsertContext.class,0);
		}
		public ListPermissionsContext listPermissions() {
			return getRuleContext(ListPermissionsContext.class,0);
		}
		public ListRolesContext listRoles() {
			return getRuleContext(ListRolesContext.class,0);
		}
		public RevokeContext revoke() {
			return getRuleContext(RevokeContext.class,0);
		}
		public Select_Context select_() {
			return getRuleContext(Select_Context.class,0);
		}
		public TruncateContext truncate() {
			return getRuleContext(TruncateContext.class,0);
		}
		public UpdateContext update() {
			return getRuleContext(UpdateContext.class,0);
		}
		public Use_Context use_() {
			return getRuleContext(Use_Context.class,0);
		}
		public CqlContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_cql; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCql(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCql(this);
		}
	}

	public final CqlContext cql() throws RecognitionException {
		CqlContext _localctx = new CqlContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_cql);
		try {
			setState(656);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,8,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(616);
				alterKeyspace();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(617);
				alterMaterializedView();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(618);
				alterRole();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(619);
				alterTable();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(620);
				alterType();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(621);
				alterUser();
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(622);
				applyBatch();
				}
				break;
			case 8:
				enterOuterAlt(_localctx, 8);
				{
				setState(623);
				consistencyCommand();
				}
				break;
			case 9:
				enterOuterAlt(_localctx, 9);
				{
				setState(624);
				outputFormatCommand();
				}
				break;
			case 10:
				enterOuterAlt(_localctx, 10);
				{
				setState(625);
				createAggregate();
				}
				break;
			case 11:
				enterOuterAlt(_localctx, 11);
				{
				setState(626);
				createFunction();
				}
				break;
			case 12:
				enterOuterAlt(_localctx, 12);
				{
				setState(627);
				createIndex();
				}
				break;
			case 13:
				enterOuterAlt(_localctx, 13);
				{
				setState(628);
				createKeyspace();
				}
				break;
			case 14:
				enterOuterAlt(_localctx, 14);
				{
				setState(629);
				createMaterializedView();
				}
				break;
			case 15:
				enterOuterAlt(_localctx, 15);
				{
				setState(630);
				createRole();
				}
				break;
			case 16:
				enterOuterAlt(_localctx, 16);
				{
				setState(631);
				createTable();
				}
				break;
			case 17:
				enterOuterAlt(_localctx, 17);
				{
				setState(632);
				createTrigger();
				}
				break;
			case 18:
				enterOuterAlt(_localctx, 18);
				{
				setState(633);
				createType();
				}
				break;
			case 19:
				enterOuterAlt(_localctx, 19);
				{
				setState(634);
				createUser();
				}
				break;
			case 20:
				enterOuterAlt(_localctx, 20);
				{
				setState(635);
				delete_();
				}
				break;
			case 21:
				enterOuterAlt(_localctx, 21);
				{
				setState(636);
				describeCommand();
				}
				break;
			case 22:
				enterOuterAlt(_localctx, 22);
				{
				setState(637);
				dropAggregate();
				}
				break;
			case 23:
				enterOuterAlt(_localctx, 23);
				{
				setState(638);
				dropFunction();
				}
				break;
			case 24:
				enterOuterAlt(_localctx, 24);
				{
				setState(639);
				dropIndex();
				}
				break;
			case 25:
				enterOuterAlt(_localctx, 25);
				{
				setState(640);
				dropKeyspace();
				}
				break;
			case 26:
				enterOuterAlt(_localctx, 26);
				{
				setState(641);
				dropMaterializedView();
				}
				break;
			case 27:
				enterOuterAlt(_localctx, 27);
				{
				setState(642);
				dropRole();
				}
				break;
			case 28:
				enterOuterAlt(_localctx, 28);
				{
				setState(643);
				dropTable();
				}
				break;
			case 29:
				enterOuterAlt(_localctx, 29);
				{
				setState(644);
				dropTrigger();
				}
				break;
			case 30:
				enterOuterAlt(_localctx, 30);
				{
				setState(645);
				dropType();
				}
				break;
			case 31:
				enterOuterAlt(_localctx, 31);
				{
				setState(646);
				dropUser();
				}
				break;
			case 32:
				enterOuterAlt(_localctx, 32);
				{
				setState(647);
				grant();
				}
				break;
			case 33:
				enterOuterAlt(_localctx, 33);
				{
				setState(648);
				insert();
				}
				break;
			case 34:
				enterOuterAlt(_localctx, 34);
				{
				setState(649);
				listPermissions();
				}
				break;
			case 35:
				enterOuterAlt(_localctx, 35);
				{
				setState(650);
				listRoles();
				}
				break;
			case 36:
				enterOuterAlt(_localctx, 36);
				{
				setState(651);
				revoke();
				}
				break;
			case 37:
				enterOuterAlt(_localctx, 37);
				{
				setState(652);
				select_();
				}
				break;
			case 38:
				enterOuterAlt(_localctx, 38);
				{
				setState(653);
				truncate();
				}
				break;
			case 39:
				enterOuterAlt(_localctx, 39);
				{
				setState(654);
				update();
				}
				break;
			case 40:
				enterOuterAlt(_localctx, 40);
				{
				setState(655);
				use_();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DescribeCommandContext extends ParserRuleContext {
		public KwDescribeContext kwDescribe() {
			return getRuleContext(KwDescribeContext.class,0);
		}
		public KwKeyspacesContext kwKeyspaces() {
			return getRuleContext(KwKeyspacesContext.class,0);
		}
		public KwKeyspaceContext kwKeyspace() {
			return getRuleContext(KwKeyspaceContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public KwTablesContext kwTables() {
			return getRuleContext(KwTablesContext.class,0);
		}
		public KwTypeContext kwType() {
			return getRuleContext(KwTypeContext.class,0);
		}
		public Type_Context type_() {
			return getRuleContext(Type_Context.class,0);
		}
		public KwTypesContext kwTypes() {
			return getRuleContext(KwTypesContext.class,0);
		}
		public KwFunctionContext kwFunction() {
			return getRuleContext(KwFunctionContext.class,0);
		}
		public Function_Context function_() {
			return getRuleContext(Function_Context.class,0);
		}
		public KwFunctionsContext kwFunctions() {
			return getRuleContext(KwFunctionsContext.class,0);
		}
		public KwAggregatesContext kwAggregates() {
			return getRuleContext(KwAggregatesContext.class,0);
		}
		public KwAggregateContext kwAggregate() {
			return getRuleContext(KwAggregateContext.class,0);
		}
		public AggregateContext aggregate() {
			return getRuleContext(AggregateContext.class,0);
		}
		public KwClusterContext kwCluster() {
			return getRuleContext(KwClusterContext.class,0);
		}
		public KwConnectionContext kwConnection() {
			return getRuleContext(KwConnectionContext.class,0);
		}
		public KwIndexContext kwIndex() {
			return getRuleContext(KwIndexContext.class,0);
		}
		public IndexNameContext indexName() {
			return getRuleContext(IndexNameContext.class,0);
		}
		public KwMaterializedContext kwMaterialized() {
			return getRuleContext(KwMaterializedContext.class,0);
		}
		public KwViewContext kwView() {
			return getRuleContext(KwViewContext.class,0);
		}
		public MaterializedViewContext materializedView() {
			return getRuleContext(MaterializedViewContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DescribeCommandContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_describeCommand; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDescribeCommand(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDescribeCommand(this);
		}
	}

	public final DescribeCommandContext describeCommand() throws RecognitionException {
		DescribeCommandContext _localctx = new DescribeCommandContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_describeCommand);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(658);
			kwDescribe();
			setState(718);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_KEYSPACES:
				{
				setState(659);
				kwKeyspaces();
				}
				break;
			case K_KEYSPACE:
				{
				setState(660);
				kwKeyspace();
				setState(661);
				keyspace();
				}
				break;
			case K_TABLE:
				{
				setState(663);
				kwTable();
				setState(667);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,9,_ctx) ) {
				case 1:
					{
					setState(664);
					keyspace();
					setState(665);
					match(DOT);
					}
					break;
				}
				setState(669);
				table();
				}
				break;
			case K_TABLES:
				{
				setState(671);
				kwTables();
				}
				break;
			case K_TYPE:
				{
				setState(672);
				kwType();
				setState(676);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,10,_ctx) ) {
				case 1:
					{
					setState(673);
					keyspace();
					setState(674);
					match(DOT);
					}
					break;
				}
				setState(678);
				type_();
				}
				break;
			case K_TYPES:
				{
				setState(680);
				kwTypes();
				}
				break;
			case K_FUNCTION:
				{
				setState(681);
				kwFunction();
				setState(685);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,11,_ctx) ) {
				case 1:
					{
					setState(682);
					keyspace();
					setState(683);
					match(DOT);
					}
					break;
				}
				setState(687);
				function_();
				}
				break;
			case K_FUNCTIONS:
				{
				setState(689);
				kwFunctions();
				}
				break;
			case K_AGGREGATES:
				{
				setState(690);
				kwAggregates();
				}
				break;
			case K_AGGREGATE:
				{
				setState(691);
				kwAggregate();
				setState(695);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,12,_ctx) ) {
				case 1:
					{
					setState(692);
					keyspace();
					setState(693);
					match(DOT);
					}
					break;
				}
				setState(697);
				aggregate();
				}
				break;
			case K_CLUSTER:
				{
				setState(699);
				kwCluster();
				}
				break;
			case K_CONNECTION:
				{
				setState(700);
				kwConnection();
				}
				break;
			case K_INDEX:
				{
				setState(701);
				kwIndex();
				setState(705);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,13,_ctx) ) {
				case 1:
					{
					setState(702);
					keyspace();
					setState(703);
					match(DOT);
					}
					break;
				}
				setState(707);
				indexName();
				}
				break;
			case K_MATERIALIZED:
				{
				setState(709);
				kwMaterialized();
				setState(710);
				kwView();
				setState(714);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,14,_ctx) ) {
				case 1:
					{
					setState(711);
					keyspace();
					setState(712);
					match(DOT);
					}
					break;
				}
				setState(716);
				materializedView();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ConsistencyCommandContext extends ParserRuleContext {
		public KwConsistencyContext kwConsistency() {
			return getRuleContext(KwConsistencyContext.class,0);
		}
		public KwConsistencyLevelContext kwConsistencyLevel() {
			return getRuleContext(KwConsistencyLevelContext.class,0);
		}
		public ConsistencyCommandContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_consistencyCommand; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterConsistencyCommand(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitConsistencyCommand(this);
		}
	}

	public final ConsistencyCommandContext consistencyCommand() throws RecognitionException {
		ConsistencyCommandContext _localctx = new ConsistencyCommandContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_consistencyCommand);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(720);
			kwConsistency();
			setState(721);
			kwConsistencyLevel();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OutputFormatCommandContext extends ParserRuleContext {
		public KwOutputContext kwOutput() {
			return getRuleContext(KwOutputContext.class,0);
		}
		public KwOutputFormatTypeContext kwOutputFormatType() {
			return getRuleContext(KwOutputFormatTypeContext.class,0);
		}
		public OutputFormatCommandContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_outputFormatCommand; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOutputFormatCommand(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOutputFormatCommand(this);
		}
	}

	public final OutputFormatCommandContext outputFormatCommand() throws RecognitionException {
		OutputFormatCommandContext _localctx = new OutputFormatCommandContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_outputFormatCommand);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(723);
			kwOutput();
			setState(725);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_TABLE || _la==K_ASCII) {
				{
				setState(724);
				kwOutputFormatType();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RevokeContext extends ParserRuleContext {
		public KwRevokeContext kwRevoke() {
			return getRuleContext(KwRevokeContext.class,0);
		}
		public PriviledgeContext priviledge() {
			return getRuleContext(PriviledgeContext.class,0);
		}
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public ResourceContext resource() {
			return getRuleContext(ResourceContext.class,0);
		}
		public KwFromContext kwFrom() {
			return getRuleContext(KwFromContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public RevokeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_revoke; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRevoke(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRevoke(this);
		}
	}

	public final RevokeContext revoke() throws RecognitionException {
		RevokeContext _localctx = new RevokeContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_revoke);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(727);
			kwRevoke();
			setState(728);
			priviledge();
			setState(729);
			kwOn();
			setState(730);
			resource();
			setState(731);
			kwFrom();
			setState(732);
			role();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ListRolesContext extends ParserRuleContext {
		public KwListContext kwList() {
			return getRuleContext(KwListContext.class,0);
		}
		public KwRolesContext kwRoles() {
			return getRuleContext(KwRolesContext.class,0);
		}
		public KwOfContext kwOf() {
			return getRuleContext(KwOfContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public KwNorecursiveContext kwNorecursive() {
			return getRuleContext(KwNorecursiveContext.class,0);
		}
		public ListRolesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_listRoles; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterListRoles(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitListRoles(this);
		}
	}

	public final ListRolesContext listRoles() throws RecognitionException {
		ListRolesContext _localctx = new ListRolesContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_listRoles);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(734);
			kwList();
			setState(735);
			kwRoles();
			setState(739);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_OF) {
				{
				setState(736);
				kwOf();
				setState(737);
				role();
				}
			}

			setState(742);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_NORECURSIVE) {
				{
				setState(741);
				kwNorecursive();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ListPermissionsContext extends ParserRuleContext {
		public KwListContext kwList() {
			return getRuleContext(KwListContext.class,0);
		}
		public PriviledgeContext priviledge() {
			return getRuleContext(PriviledgeContext.class,0);
		}
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public ResourceContext resource() {
			return getRuleContext(ResourceContext.class,0);
		}
		public KwOfContext kwOf() {
			return getRuleContext(KwOfContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public ListPermissionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_listPermissions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterListPermissions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitListPermissions(this);
		}
	}

	public final ListPermissionsContext listPermissions() throws RecognitionException {
		ListPermissionsContext _localctx = new ListPermissionsContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_listPermissions);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(744);
			kwList();
			setState(745);
			priviledge();
			setState(749);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_ON) {
				{
				setState(746);
				kwOn();
				setState(747);
				resource();
				}
			}

			setState(754);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_OF) {
				{
				setState(751);
				kwOf();
				setState(752);
				role();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class GrantContext extends ParserRuleContext {
		public KwGrantContext kwGrant() {
			return getRuleContext(KwGrantContext.class,0);
		}
		public PriviledgeContext priviledge() {
			return getRuleContext(PriviledgeContext.class,0);
		}
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public ResourceContext resource() {
			return getRuleContext(ResourceContext.class,0);
		}
		public KwToContext kwTo() {
			return getRuleContext(KwToContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public GrantContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_grant; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterGrant(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitGrant(this);
		}
	}

	public final GrantContext grant() throws RecognitionException {
		GrantContext _localctx = new GrantContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_grant);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(756);
			kwGrant();
			setState(757);
			priviledge();
			setState(758);
			kwOn();
			setState(759);
			resource();
			setState(760);
			kwTo();
			setState(761);
			role();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PriviledgeContext extends ParserRuleContext {
		public KwAllContext kwAll() {
			return getRuleContext(KwAllContext.class,0);
		}
		public KwAllPermissionsContext kwAllPermissions() {
			return getRuleContext(KwAllPermissionsContext.class,0);
		}
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwAuthorizeContext kwAuthorize() {
			return getRuleContext(KwAuthorizeContext.class,0);
		}
		public KwDescribeContext kwDescribe() {
			return getRuleContext(KwDescribeContext.class,0);
		}
		public KwExecuteContext kwExecute() {
			return getRuleContext(KwExecuteContext.class,0);
		}
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwModifyContext kwModify() {
			return getRuleContext(KwModifyContext.class,0);
		}
		public KwSelectContext kwSelect() {
			return getRuleContext(KwSelectContext.class,0);
		}
		public PriviledgeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_priviledge; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPriviledge(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPriviledge(this);
		}
	}

	public final PriviledgeContext priviledge() throws RecognitionException {
		PriviledgeContext _localctx = new PriviledgeContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_priviledge);
		try {
			setState(775);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_ALL:
				enterOuterAlt(_localctx, 1);
				{
				setState(765);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,21,_ctx) ) {
				case 1:
					{
					setState(763);
					kwAll();
					}
					break;
				case 2:
					{
					setState(764);
					kwAllPermissions();
					}
					break;
				}
				}
				break;
			case K_ALTER:
				enterOuterAlt(_localctx, 2);
				{
				setState(767);
				kwAlter();
				}
				break;
			case K_AUTHORIZE:
				enterOuterAlt(_localctx, 3);
				{
				setState(768);
				kwAuthorize();
				}
				break;
			case K_DESC:
			case K_DESCRIBE:
				enterOuterAlt(_localctx, 4);
				{
				setState(769);
				kwDescribe();
				}
				break;
			case K_EXECUTE:
				enterOuterAlt(_localctx, 5);
				{
				setState(770);
				kwExecute();
				}
				break;
			case K_CREATE:
				enterOuterAlt(_localctx, 6);
				{
				setState(771);
				kwCreate();
				}
				break;
			case K_DROP:
				enterOuterAlt(_localctx, 7);
				{
				setState(772);
				kwDrop();
				}
				break;
			case K_MODIFY:
				enterOuterAlt(_localctx, 8);
				{
				setState(773);
				kwModify();
				}
				break;
			case K_SELECT:
				enterOuterAlt(_localctx, 9);
				{
				setState(774);
				kwSelect();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ResourceContext extends ParserRuleContext {
		public KwAllContext kwAll() {
			return getRuleContext(KwAllContext.class,0);
		}
		public KwFunctionsContext kwFunctions() {
			return getRuleContext(KwFunctionsContext.class,0);
		}
		public KwInContext kwIn() {
			return getRuleContext(KwInContext.class,0);
		}
		public KwKeyspaceContext kwKeyspace() {
			return getRuleContext(KwKeyspaceContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public KwFunctionContext kwFunction() {
			return getRuleContext(KwFunctionContext.class,0);
		}
		public Function_Context function_() {
			return getRuleContext(Function_Context.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public KwKeyspacesContext kwKeyspaces() {
			return getRuleContext(KwKeyspacesContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public KwRolesContext kwRoles() {
			return getRuleContext(KwRolesContext.class,0);
		}
		public KwRoleContext kwRole() {
			return getRuleContext(KwRoleContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public ResourceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_resource; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterResource(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitResource(this);
		}
	}

	public final ResourceContext resource() throws RecognitionException {
		ResourceContext _localctx = new ResourceContext(_ctx, getState());
		enterRule(_localctx, 26, RULE_resource);
		int _la;
		try {
			setState(815);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,26,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(777);
				kwAll();
				setState(778);
				kwFunctions();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(780);
				kwAll();
				setState(781);
				kwFunctions();
				setState(782);
				kwIn();
				setState(783);
				kwKeyspace();
				setState(784);
				keyspace();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(786);
				kwFunction();
				setState(790);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,23,_ctx) ) {
				case 1:
					{
					setState(787);
					keyspace();
					setState(788);
					match(DOT);
					}
					break;
				}
				setState(792);
				function_();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(794);
				kwAll();
				setState(795);
				kwKeyspaces();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(797);
				kwKeyspace();
				setState(798);
				keyspace();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(801);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_TABLE) {
					{
					setState(800);
					kwTable();
					}
				}

				setState(806);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,25,_ctx) ) {
				case 1:
					{
					setState(803);
					keyspace();
					setState(804);
					match(DOT);
					}
					break;
				}
				setState(808);
				table();
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(809);
				kwAll();
				setState(810);
				kwRoles();
				}
				break;
			case 8:
				enterOuterAlt(_localctx, 8);
				{
				setState(812);
				kwRole();
				setState(813);
				role();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateUserContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwUserContext kwUser() {
			return getRuleContext(KwUserContext.class,0);
		}
		public UserContext user() {
			return getRuleContext(UserContext.class,0);
		}
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public KwPasswordContext kwPassword() {
			return getRuleContext(KwPasswordContext.class,0);
		}
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KwSuperuserContext kwSuperuser() {
			return getRuleContext(KwSuperuserContext.class,0);
		}
		public KwNosuperuserContext kwNosuperuser() {
			return getRuleContext(KwNosuperuserContext.class,0);
		}
		public CreateUserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createUser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateUser(this);
		}
	}

	public final CreateUserContext createUser() throws RecognitionException {
		CreateUserContext _localctx = new CreateUserContext(_ctx, getState());
		enterRule(_localctx, 28, RULE_createUser);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(817);
			kwCreate();
			setState(818);
			kwUser();
			setState(820);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(819);
				ifNotExist();
				}
			}

			setState(822);
			user();
			setState(823);
			kwWith();
			setState(824);
			kwPassword();
			setState(825);
			stringLiteral();
			setState(828);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_SUPERUSER:
				{
				setState(826);
				kwSuperuser();
				}
				break;
			case K_NOSUPERUSER:
				{
				setState(827);
				kwNosuperuser();
				}
				break;
			case EOF:
			case SEMI:
			case MINUSMINUS:
				break;
			default:
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateRoleContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwRoleContext kwRole() {
			return getRuleContext(KwRoleContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public RoleWithContext roleWith() {
			return getRuleContext(RoleWithContext.class,0);
		}
		public CreateRoleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createRole; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateRole(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateRole(this);
		}
	}

	public final CreateRoleContext createRole() throws RecognitionException {
		CreateRoleContext _localctx = new CreateRoleContext(_ctx, getState());
		enterRule(_localctx, 30, RULE_createRole);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(830);
			kwCreate();
			setState(831);
			kwRole();
			setState(833);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(832);
				ifNotExist();
				}
			}

			setState(835);
			role();
			setState(837);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WITH) {
				{
				setState(836);
				roleWith();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateTypeContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwTypeContext kwType() {
			return getRuleContext(KwTypeContext.class,0);
		}
		public Type_Context type_() {
			return getRuleContext(Type_Context.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public TypeMemberColumnListContext typeMemberColumnList() {
			return getRuleContext(TypeMemberColumnListContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public CreateTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateType(this);
		}
	}

	public final CreateTypeContext createType() throws RecognitionException {
		CreateTypeContext _localctx = new CreateTypeContext(_ctx, getState());
		enterRule(_localctx, 32, RULE_createType);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(839);
			kwCreate();
			setState(840);
			kwType();
			setState(842);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(841);
				ifNotExist();
				}
			}

			setState(847);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,32,_ctx) ) {
			case 1:
				{
				setState(844);
				keyspace();
				setState(845);
				match(DOT);
				}
				break;
			}
			setState(849);
			type_();
			setState(850);
			syntaxBracketLr();
			setState(851);
			typeMemberColumnList();
			setState(852);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TypeMemberColumnListContext extends ParserRuleContext {
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<DataTypeContext> dataType() {
			return getRuleContexts(DataTypeContext.class);
		}
		public DataTypeContext dataType(int i) {
			return getRuleContext(DataTypeContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public TypeMemberColumnListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_typeMemberColumnList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTypeMemberColumnList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTypeMemberColumnList(this);
		}
	}

	public final TypeMemberColumnListContext typeMemberColumnList() throws RecognitionException {
		TypeMemberColumnListContext _localctx = new TypeMemberColumnListContext(_ctx, getState());
		enterRule(_localctx, 34, RULE_typeMemberColumnList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(854);
			column();
			setState(855);
			dataType();
			setState(862);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(856);
				syntaxComma();
				setState(857);
				column();
				setState(858);
				dataType();
				}
				}
				setState(864);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateTriggerContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwTriggerContext kwTrigger() {
			return getRuleContext(KwTriggerContext.class,0);
		}
		public TriggerContext trigger() {
			return getRuleContext(TriggerContext.class,0);
		}
		public KwUsingContext kwUsing() {
			return getRuleContext(KwUsingContext.class,0);
		}
		public TriggerClassContext triggerClass() {
			return getRuleContext(TriggerClassContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public CreateTriggerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createTrigger; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateTrigger(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateTrigger(this);
		}
	}

	public final CreateTriggerContext createTrigger() throws RecognitionException {
		CreateTriggerContext _localctx = new CreateTriggerContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_createTrigger);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(865);
			kwCreate();
			setState(866);
			kwTrigger();
			setState(868);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(867);
				ifNotExist();
				}
			}

			setState(873);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,35,_ctx) ) {
			case 1:
				{
				setState(870);
				keyspace();
				setState(871);
				match(DOT);
				}
				break;
			}
			setState(875);
			trigger();
			setState(876);
			kwUsing();
			setState(877);
			triggerClass();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateMaterializedViewContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwMaterializedContext kwMaterialized() {
			return getRuleContext(KwMaterializedContext.class,0);
		}
		public KwViewContext kwView() {
			return getRuleContext(KwViewContext.class,0);
		}
		public MaterializedViewContext materializedView() {
			return getRuleContext(MaterializedViewContext.class,0);
		}
		public KwAsContext kwAs() {
			return getRuleContext(KwAsContext.class,0);
		}
		public KwSelectContext kwSelect() {
			return getRuleContext(KwSelectContext.class,0);
		}
		public List<ColumnListContext> columnList() {
			return getRuleContexts(ColumnListContext.class);
		}
		public ColumnListContext columnList(int i) {
			return getRuleContext(ColumnListContext.class,i);
		}
		public KwFromContext kwFrom() {
			return getRuleContext(KwFromContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public MaterializedViewWhereContext materializedViewWhere() {
			return getRuleContext(MaterializedViewWhereContext.class,0);
		}
		public KwPrimaryContext kwPrimary() {
			return getRuleContext(KwPrimaryContext.class,0);
		}
		public KwKeyContext kwKey() {
			return getRuleContext(KwKeyContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public List<KeyspaceContext> keyspace() {
			return getRuleContexts(KeyspaceContext.class);
		}
		public KeyspaceContext keyspace(int i) {
			return getRuleContext(KeyspaceContext.class,i);
		}
		public List<TerminalNode> DOT() { return getTokens(CqlParser.DOT); }
		public TerminalNode DOT(int i) {
			return getToken(CqlParser.DOT, i);
		}
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public MaterializedViewOptionsContext materializedViewOptions() {
			return getRuleContext(MaterializedViewOptionsContext.class,0);
		}
		public CreateMaterializedViewContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createMaterializedView; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateMaterializedView(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateMaterializedView(this);
		}
	}

	public final CreateMaterializedViewContext createMaterializedView() throws RecognitionException {
		CreateMaterializedViewContext _localctx = new CreateMaterializedViewContext(_ctx, getState());
		enterRule(_localctx, 38, RULE_createMaterializedView);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(879);
			kwCreate();
			setState(880);
			kwMaterialized();
			setState(881);
			kwView();
			setState(883);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(882);
				ifNotExist();
				}
			}

			setState(888);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,37,_ctx) ) {
			case 1:
				{
				setState(885);
				keyspace();
				setState(886);
				match(DOT);
				}
				break;
			}
			setState(890);
			materializedView();
			setState(891);
			kwAs();
			setState(892);
			kwSelect();
			setState(893);
			columnList();
			setState(894);
			kwFrom();
			setState(898);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,38,_ctx) ) {
			case 1:
				{
				setState(895);
				keyspace();
				setState(896);
				match(DOT);
				}
				break;
			}
			setState(900);
			table();
			setState(901);
			materializedViewWhere();
			setState(902);
			kwPrimary();
			setState(903);
			kwKey();
			setState(904);
			syntaxBracketLr();
			setState(905);
			columnList();
			setState(906);
			syntaxBracketRr();
			setState(910);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WITH) {
				{
				setState(907);
				kwWith();
				setState(908);
				materializedViewOptions();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class MaterializedViewWhereContext extends ParserRuleContext {
		public KwWhereContext kwWhere() {
			return getRuleContext(KwWhereContext.class,0);
		}
		public ColumnNotNullListContext columnNotNullList() {
			return getRuleContext(ColumnNotNullListContext.class,0);
		}
		public KwAndContext kwAnd() {
			return getRuleContext(KwAndContext.class,0);
		}
		public RelationElementsContext relationElements() {
			return getRuleContext(RelationElementsContext.class,0);
		}
		public MaterializedViewWhereContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_materializedViewWhere; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterMaterializedViewWhere(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitMaterializedViewWhere(this);
		}
	}

	public final MaterializedViewWhereContext materializedViewWhere() throws RecognitionException {
		MaterializedViewWhereContext _localctx = new MaterializedViewWhereContext(_ctx, getState());
		enterRule(_localctx, 40, RULE_materializedViewWhere);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(912);
			kwWhere();
			setState(913);
			columnNotNullList();
			setState(917);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_AND) {
				{
				setState(914);
				kwAnd();
				setState(915);
				relationElements();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnNotNullListContext extends ParserRuleContext {
		public List<ColumnNotNullContext> columnNotNull() {
			return getRuleContexts(ColumnNotNullContext.class);
		}
		public ColumnNotNullContext columnNotNull(int i) {
			return getRuleContext(ColumnNotNullContext.class,i);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public ColumnNotNullListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_columnNotNullList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumnNotNullList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumnNotNullList(this);
		}
	}

	public final ColumnNotNullListContext columnNotNullList() throws RecognitionException {
		ColumnNotNullListContext _localctx = new ColumnNotNullListContext(_ctx, getState());
		enterRule(_localctx, 42, RULE_columnNotNullList);
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(919);
			columnNotNull();
			setState(925);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,41,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					{
					{
					setState(920);
					kwAnd();
					setState(921);
					columnNotNull();
					}
					} 
				}
				setState(927);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,41,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnNotNullContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public KwIsContext kwIs() {
			return getRuleContext(KwIsContext.class,0);
		}
		public KwNotContext kwNot() {
			return getRuleContext(KwNotContext.class,0);
		}
		public KwNullContext kwNull() {
			return getRuleContext(KwNullContext.class,0);
		}
		public ColumnNotNullContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_columnNotNull; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumnNotNull(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumnNotNull(this);
		}
	}

	public final ColumnNotNullContext columnNotNull() throws RecognitionException {
		ColumnNotNullContext _localctx = new ColumnNotNullContext(_ctx, getState());
		enterRule(_localctx, 44, RULE_columnNotNull);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(928);
			column();
			setState(929);
			kwIs();
			setState(930);
			kwNot();
			setState(931);
			kwNull();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class MaterializedViewOptionsContext extends ParserRuleContext {
		public TableOptionsContext tableOptions() {
			return getRuleContext(TableOptionsContext.class,0);
		}
		public KwAndContext kwAnd() {
			return getRuleContext(KwAndContext.class,0);
		}
		public ClusteringOrderContext clusteringOrder() {
			return getRuleContext(ClusteringOrderContext.class,0);
		}
		public MaterializedViewOptionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_materializedViewOptions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterMaterializedViewOptions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitMaterializedViewOptions(this);
		}
	}

	public final MaterializedViewOptionsContext materializedViewOptions() throws RecognitionException {
		MaterializedViewOptionsContext _localctx = new MaterializedViewOptionsContext(_ctx, getState());
		enterRule(_localctx, 46, RULE_materializedViewOptions);
		try {
			setState(939);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,42,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(933);
				tableOptions();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(934);
				tableOptions();
				setState(935);
				kwAnd();
				setState(936);
				clusteringOrder();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(938);
				clusteringOrder();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateKeyspaceContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwKeyspaceContext kwKeyspace() {
			return getRuleContext(KwKeyspaceContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public KwReplicationContext kwReplication() {
			return getRuleContext(KwReplicationContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public ReplicationListContext replicationList() {
			return getRuleContext(ReplicationListContext.class,0);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KwAndContext kwAnd() {
			return getRuleContext(KwAndContext.class,0);
		}
		public DurableWritesContext durableWrites() {
			return getRuleContext(DurableWritesContext.class,0);
		}
		public CreateKeyspaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createKeyspace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateKeyspace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateKeyspace(this);
		}
	}

	public final CreateKeyspaceContext createKeyspace() throws RecognitionException {
		CreateKeyspaceContext _localctx = new CreateKeyspaceContext(_ctx, getState());
		enterRule(_localctx, 48, RULE_createKeyspace);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(941);
			kwCreate();
			setState(942);
			kwKeyspace();
			setState(944);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(943);
				ifNotExist();
				}
			}

			setState(946);
			keyspace();
			setState(947);
			kwWith();
			setState(948);
			kwReplication();
			setState(949);
			match(OPERATOR_EQ);
			setState(950);
			syntaxBracketLc();
			setState(951);
			replicationList();
			setState(952);
			syntaxBracketRc();
			setState(956);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_AND) {
				{
				setState(953);
				kwAnd();
				setState(954);
				durableWrites();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateFunctionContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwFunctionContext kwFunction() {
			return getRuleContext(KwFunctionContext.class,0);
		}
		public Function_Context function_() {
			return getRuleContext(Function_Context.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public ReturnModeContext returnMode() {
			return getRuleContext(ReturnModeContext.class,0);
		}
		public KwReturnsContext kwReturns() {
			return getRuleContext(KwReturnsContext.class,0);
		}
		public DataTypeContext dataType() {
			return getRuleContext(DataTypeContext.class,0);
		}
		public KwLanguageContext kwLanguage() {
			return getRuleContext(KwLanguageContext.class,0);
		}
		public LanguageContext language() {
			return getRuleContext(LanguageContext.class,0);
		}
		public KwAsContext kwAs() {
			return getRuleContext(KwAsContext.class,0);
		}
		public CodeBlockContext codeBlock() {
			return getRuleContext(CodeBlockContext.class,0);
		}
		public OrReplaceContext orReplace() {
			return getRuleContext(OrReplaceContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public ParamListContext paramList() {
			return getRuleContext(ParamListContext.class,0);
		}
		public CreateFunctionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createFunction; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateFunction(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateFunction(this);
		}
	}

	public final CreateFunctionContext createFunction() throws RecognitionException {
		CreateFunctionContext _localctx = new CreateFunctionContext(_ctx, getState());
		enterRule(_localctx, 50, RULE_createFunction);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(958);
			kwCreate();
			setState(960);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_OR) {
				{
				setState(959);
				orReplace();
				}
			}

			setState(962);
			kwFunction();
			setState(964);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(963);
				ifNotExist();
				}
			}

			setState(969);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,47,_ctx) ) {
			case 1:
				{
				setState(966);
				keyspace();
				setState(967);
				match(DOT);
				}
				break;
			}
			setState(971);
			function_();
			setState(972);
			syntaxBracketLr();
			setState(974);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_INPUT || _la==OBJECT_NAME) {
				{
				setState(973);
				paramList();
				}
			}

			setState(976);
			syntaxBracketRr();
			setState(977);
			returnMode();
			setState(978);
			kwReturns();
			setState(979);
			dataType();
			setState(980);
			kwLanguage();
			setState(981);
			language();
			setState(982);
			kwAs();
			setState(983);
			codeBlock();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CodeBlockContext extends ParserRuleContext {
		public TerminalNode CODE_BLOCK() { return getToken(CqlParser.CODE_BLOCK, 0); }
		public TerminalNode STRING_LITERAL() { return getToken(CqlParser.STRING_LITERAL, 0); }
		public CodeBlockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_codeBlock; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCodeBlock(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCodeBlock(this);
		}
	}

	public final CodeBlockContext codeBlock() throws RecognitionException {
		CodeBlockContext _localctx = new CodeBlockContext(_ctx, getState());
		enterRule(_localctx, 52, RULE_codeBlock);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(985);
			_la = _input.LA(1);
			if ( !(_la==CODE_BLOCK || _la==STRING_LITERAL) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ParamListContext extends ParserRuleContext {
		public List<ParamContext> param() {
			return getRuleContexts(ParamContext.class);
		}
		public ParamContext param(int i) {
			return getRuleContext(ParamContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public ParamListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_paramList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterParamList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitParamList(this);
		}
	}

	public final ParamListContext paramList() throws RecognitionException {
		ParamListContext _localctx = new ParamListContext(_ctx, getState());
		enterRule(_localctx, 54, RULE_paramList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(987);
			param();
			setState(993);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(988);
				syntaxComma();
				setState(989);
				param();
				}
				}
				setState(995);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReturnModeContext extends ParserRuleContext {
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public List<KwNullContext> kwNull() {
			return getRuleContexts(KwNullContext.class);
		}
		public KwNullContext kwNull(int i) {
			return getRuleContext(KwNullContext.class,i);
		}
		public KwInputContext kwInput() {
			return getRuleContext(KwInputContext.class,0);
		}
		public KwCalledContext kwCalled() {
			return getRuleContext(KwCalledContext.class,0);
		}
		public KwReturnsContext kwReturns() {
			return getRuleContext(KwReturnsContext.class,0);
		}
		public ReturnModeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returnMode; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterReturnMode(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitReturnMode(this);
		}
	}

	public final ReturnModeContext returnMode() throws RecognitionException {
		ReturnModeContext _localctx = new ReturnModeContext(_ctx, getState());
		enterRule(_localctx, 56, RULE_returnMode);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1000);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_CALLED:
				{
				setState(996);
				kwCalled();
				}
				break;
			case K_RETURNS:
				{
				setState(997);
				kwReturns();
				setState(998);
				kwNull();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(1002);
			kwOn();
			setState(1003);
			kwNull();
			setState(1004);
			kwInput();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateAggregateContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwAggregateContext kwAggregate() {
			return getRuleContext(KwAggregateContext.class,0);
		}
		public AggregateContext aggregate() {
			return getRuleContext(AggregateContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public List<DataTypeContext> dataType() {
			return getRuleContexts(DataTypeContext.class);
		}
		public DataTypeContext dataType(int i) {
			return getRuleContext(DataTypeContext.class,i);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public KwSfuncContext kwSfunc() {
			return getRuleContext(KwSfuncContext.class,0);
		}
		public List<Function_Context> function_() {
			return getRuleContexts(Function_Context.class);
		}
		public Function_Context function_(int i) {
			return getRuleContext(Function_Context.class,i);
		}
		public KwStypeContext kwStype() {
			return getRuleContext(KwStypeContext.class,0);
		}
		public KwFinalfuncContext kwFinalfunc() {
			return getRuleContext(KwFinalfuncContext.class,0);
		}
		public KwInitcondContext kwInitcond() {
			return getRuleContext(KwInitcondContext.class,0);
		}
		public InitCondDefinitionContext initCondDefinition() {
			return getRuleContext(InitCondDefinitionContext.class,0);
		}
		public OrReplaceContext orReplace() {
			return getRuleContext(OrReplaceContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public CreateAggregateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createAggregate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateAggregate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateAggregate(this);
		}
	}

	public final CreateAggregateContext createAggregate() throws RecognitionException {
		CreateAggregateContext _localctx = new CreateAggregateContext(_ctx, getState());
		enterRule(_localctx, 58, RULE_createAggregate);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1006);
			kwCreate();
			setState(1008);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_OR) {
				{
				setState(1007);
				orReplace();
				}
			}

			setState(1010);
			kwAggregate();
			setState(1012);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1011);
				ifNotExist();
				}
			}

			setState(1017);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,53,_ctx) ) {
			case 1:
				{
				setState(1014);
				keyspace();
				setState(1015);
				match(DOT);
				}
				break;
			}
			setState(1019);
			aggregate();
			setState(1020);
			syntaxBracketLr();
			setState(1021);
			dataType();
			setState(1022);
			syntaxBracketRr();
			setState(1023);
			kwSfunc();
			setState(1024);
			function_();
			setState(1025);
			kwStype();
			setState(1026);
			dataType();
			setState(1027);
			kwFinalfunc();
			setState(1028);
			function_();
			setState(1029);
			kwInitcond();
			setState(1030);
			initCondDefinition();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InitCondDefinitionContext extends ParserRuleContext {
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public InitCondListContext initCondList() {
			return getRuleContext(InitCondListContext.class,0);
		}
		public InitCondListNestedContext initCondListNested() {
			return getRuleContext(InitCondListNestedContext.class,0);
		}
		public InitCondHashContext initCondHash() {
			return getRuleContext(InitCondHashContext.class,0);
		}
		public InitCondDefinitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_initCondDefinition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInitCondDefinition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInitCondDefinition(this);
		}
	}

	public final InitCondDefinitionContext initCondDefinition() throws RecognitionException {
		InitCondDefinitionContext _localctx = new InitCondDefinitionContext(_ctx, getState());
		enterRule(_localctx, 60, RULE_initCondDefinition);
		try {
			setState(1036);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,54,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1032);
				constant();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1033);
				initCondList();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1034);
				initCondListNested();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(1035);
				initCondHash();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InitCondHashContext extends ParserRuleContext {
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public List<InitCondHashItemContext> initCondHashItem() {
			return getRuleContexts(InitCondHashItemContext.class);
		}
		public InitCondHashItemContext initCondHashItem(int i) {
			return getRuleContext(InitCondHashItemContext.class,i);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public InitCondHashContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_initCondHash; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInitCondHash(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInitCondHash(this);
		}
	}

	public final InitCondHashContext initCondHash() throws RecognitionException {
		InitCondHashContext _localctx = new InitCondHashContext(_ctx, getState());
		enterRule(_localctx, 62, RULE_initCondHash);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1038);
			syntaxBracketLc();
			setState(1039);
			initCondHashItem();
			setState(1045);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1040);
				syntaxComma();
				setState(1041);
				initCondHashItem();
				}
				}
				setState(1047);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1048);
			syntaxBracketRc();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InitCondHashItemContext extends ParserRuleContext {
		public HashKeyContext hashKey() {
			return getRuleContext(HashKeyContext.class,0);
		}
		public TerminalNode COLON() { return getToken(CqlParser.COLON, 0); }
		public InitCondDefinitionContext initCondDefinition() {
			return getRuleContext(InitCondDefinitionContext.class,0);
		}
		public InitCondHashItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_initCondHashItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInitCondHashItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInitCondHashItem(this);
		}
	}

	public final InitCondHashItemContext initCondHashItem() throws RecognitionException {
		InitCondHashItemContext _localctx = new InitCondHashItemContext(_ctx, getState());
		enterRule(_localctx, 64, RULE_initCondHashItem);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1050);
			hashKey();
			setState(1051);
			match(COLON);
			setState(1052);
			initCondDefinition();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InitCondListNestedContext extends ParserRuleContext {
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public List<InitCondListContext> initCondList() {
			return getRuleContexts(InitCondListContext.class);
		}
		public InitCondListContext initCondList(int i) {
			return getRuleContext(InitCondListContext.class,i);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public InitCondListNestedContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_initCondListNested; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInitCondListNested(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInitCondListNested(this);
		}
	}

	public final InitCondListNestedContext initCondListNested() throws RecognitionException {
		InitCondListNestedContext _localctx = new InitCondListNestedContext(_ctx, getState());
		enterRule(_localctx, 66, RULE_initCondListNested);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1054);
			syntaxBracketLr();
			setState(1055);
			initCondList();
			setState(1062);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LR_BRACKET || _la==COMMA) {
				{
				setState(1060);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case COMMA:
					{
					setState(1056);
					syntaxComma();
					setState(1057);
					constant();
					}
					break;
				case LR_BRACKET:
					{
					setState(1059);
					initCondList();
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(1064);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1065);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InitCondListContext extends ParserRuleContext {
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public InitCondListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_initCondList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInitCondList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInitCondList(this);
		}
	}

	public final InitCondListContext initCondList() throws RecognitionException {
		InitCondListContext _localctx = new InitCondListContext(_ctx, getState());
		enterRule(_localctx, 68, RULE_initCondList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1067);
			syntaxBracketLr();
			setState(1068);
			constant();
			setState(1074);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1069);
				syntaxComma();
				setState(1070);
				constant();
				}
				}
				setState(1076);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1077);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrReplaceContext extends ParserRuleContext {
		public KwOrContext kwOr() {
			return getRuleContext(KwOrContext.class,0);
		}
		public KwReplaceContext kwReplace() {
			return getRuleContext(KwReplaceContext.class,0);
		}
		public OrReplaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orReplace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOrReplace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOrReplace(this);
		}
	}

	public final OrReplaceContext orReplace() throws RecognitionException {
		OrReplaceContext _localctx = new OrReplaceContext(_ctx, getState());
		enterRule(_localctx, 70, RULE_orReplace);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1079);
			kwOr();
			setState(1080);
			kwReplace();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterUserContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwUserContext kwUser() {
			return getRuleContext(KwUserContext.class,0);
		}
		public UserContext user() {
			return getRuleContext(UserContext.class,0);
		}
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public UserPasswordContext userPassword() {
			return getRuleContext(UserPasswordContext.class,0);
		}
		public UserSuperUserContext userSuperUser() {
			return getRuleContext(UserSuperUserContext.class,0);
		}
		public AlterUserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterUser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterUser(this);
		}
	}

	public final AlterUserContext alterUser() throws RecognitionException {
		AlterUserContext _localctx = new AlterUserContext(_ctx, getState());
		enterRule(_localctx, 72, RULE_alterUser);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1082);
			kwAlter();
			setState(1083);
			kwUser();
			setState(1084);
			user();
			setState(1085);
			kwWith();
			setState(1086);
			userPassword();
			setState(1088);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_NOSUPERUSER || _la==K_SUPERUSER) {
				{
				setState(1087);
				userSuperUser();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UserPasswordContext extends ParserRuleContext {
		public KwPasswordContext kwPassword() {
			return getRuleContext(KwPasswordContext.class,0);
		}
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public UserPasswordContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_userPassword; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUserPassword(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUserPassword(this);
		}
	}

	public final UserPasswordContext userPassword() throws RecognitionException {
		UserPasswordContext _localctx = new UserPasswordContext(_ctx, getState());
		enterRule(_localctx, 74, RULE_userPassword);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1090);
			kwPassword();
			setState(1091);
			stringLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UserSuperUserContext extends ParserRuleContext {
		public KwSuperuserContext kwSuperuser() {
			return getRuleContext(KwSuperuserContext.class,0);
		}
		public KwNosuperuserContext kwNosuperuser() {
			return getRuleContext(KwNosuperuserContext.class,0);
		}
		public UserSuperUserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_userSuperUser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUserSuperUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUserSuperUser(this);
		}
	}

	public final UserSuperUserContext userSuperUser() throws RecognitionException {
		UserSuperUserContext _localctx = new UserSuperUserContext(_ctx, getState());
		enterRule(_localctx, 76, RULE_userSuperUser);
		try {
			setState(1095);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_SUPERUSER:
				enterOuterAlt(_localctx, 1);
				{
				setState(1093);
				kwSuperuser();
				}
				break;
			case K_NOSUPERUSER:
				enterOuterAlt(_localctx, 2);
				{
				setState(1094);
				kwNosuperuser();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwTypeContext kwType() {
			return getRuleContext(KwTypeContext.class,0);
		}
		public Type_Context type_() {
			return getRuleContext(Type_Context.class,0);
		}
		public AlterTypeOperationContext alterTypeOperation() {
			return getRuleContext(AlterTypeOperationContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public AlterTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterType(this);
		}
	}

	public final AlterTypeContext alterType() throws RecognitionException {
		AlterTypeContext _localctx = new AlterTypeContext(_ctx, getState());
		enterRule(_localctx, 78, RULE_alterType);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1097);
			kwAlter();
			setState(1098);
			kwType();
			setState(1102);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,61,_ctx) ) {
			case 1:
				{
				setState(1099);
				keyspace();
				setState(1100);
				match(DOT);
				}
				break;
			}
			setState(1104);
			type_();
			setState(1105);
			alterTypeOperation();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeOperationContext extends ParserRuleContext {
		public AlterTypeAlterTypeContext alterTypeAlterType() {
			return getRuleContext(AlterTypeAlterTypeContext.class,0);
		}
		public AlterTypeAddContext alterTypeAdd() {
			return getRuleContext(AlterTypeAddContext.class,0);
		}
		public AlterTypeRenameContext alterTypeRename() {
			return getRuleContext(AlterTypeRenameContext.class,0);
		}
		public AlterTypeOperationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeOperation; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeOperation(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeOperation(this);
		}
	}

	public final AlterTypeOperationContext alterTypeOperation() throws RecognitionException {
		AlterTypeOperationContext _localctx = new AlterTypeOperationContext(_ctx, getState());
		enterRule(_localctx, 80, RULE_alterTypeOperation);
		try {
			setState(1110);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_ALTER:
				enterOuterAlt(_localctx, 1);
				{
				setState(1107);
				alterTypeAlterType();
				}
				break;
			case K_ADD:
				enterOuterAlt(_localctx, 2);
				{
				setState(1108);
				alterTypeAdd();
				}
				break;
			case K_RENAME:
				enterOuterAlt(_localctx, 3);
				{
				setState(1109);
				alterTypeRename();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeRenameContext extends ParserRuleContext {
		public KwRenameContext kwRename() {
			return getRuleContext(KwRenameContext.class,0);
		}
		public AlterTypeRenameListContext alterTypeRenameList() {
			return getRuleContext(AlterTypeRenameListContext.class,0);
		}
		public AlterTypeRenameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeRename; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeRename(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeRename(this);
		}
	}

	public final AlterTypeRenameContext alterTypeRename() throws RecognitionException {
		AlterTypeRenameContext _localctx = new AlterTypeRenameContext(_ctx, getState());
		enterRule(_localctx, 82, RULE_alterTypeRename);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1112);
			kwRename();
			setState(1113);
			alterTypeRenameList();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeRenameListContext extends ParserRuleContext {
		public List<AlterTypeRenameItemContext> alterTypeRenameItem() {
			return getRuleContexts(AlterTypeRenameItemContext.class);
		}
		public AlterTypeRenameItemContext alterTypeRenameItem(int i) {
			return getRuleContext(AlterTypeRenameItemContext.class,i);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public AlterTypeRenameListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeRenameList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeRenameList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeRenameList(this);
		}
	}

	public final AlterTypeRenameListContext alterTypeRenameList() throws RecognitionException {
		AlterTypeRenameListContext _localctx = new AlterTypeRenameListContext(_ctx, getState());
		enterRule(_localctx, 84, RULE_alterTypeRenameList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1115);
			alterTypeRenameItem();
			setState(1121);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==K_AND) {
				{
				{
				setState(1116);
				kwAnd();
				setState(1117);
				alterTypeRenameItem();
				}
				}
				setState(1123);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeRenameItemContext extends ParserRuleContext {
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public KwToContext kwTo() {
			return getRuleContext(KwToContext.class,0);
		}
		public AlterTypeRenameItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeRenameItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeRenameItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeRenameItem(this);
		}
	}

	public final AlterTypeRenameItemContext alterTypeRenameItem() throws RecognitionException {
		AlterTypeRenameItemContext _localctx = new AlterTypeRenameItemContext(_ctx, getState());
		enterRule(_localctx, 86, RULE_alterTypeRenameItem);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1124);
			column();
			setState(1125);
			kwTo();
			setState(1126);
			column();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeAddContext extends ParserRuleContext {
		public KwAddContext kwAdd() {
			return getRuleContext(KwAddContext.class,0);
		}
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<DataTypeContext> dataType() {
			return getRuleContexts(DataTypeContext.class);
		}
		public DataTypeContext dataType(int i) {
			return getRuleContext(DataTypeContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AlterTypeAddContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeAdd; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeAdd(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeAdd(this);
		}
	}

	public final AlterTypeAddContext alterTypeAdd() throws RecognitionException {
		AlterTypeAddContext _localctx = new AlterTypeAddContext(_ctx, getState());
		enterRule(_localctx, 88, RULE_alterTypeAdd);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1128);
			kwAdd();
			setState(1129);
			column();
			setState(1130);
			dataType();
			setState(1137);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1131);
				syntaxComma();
				setState(1132);
				column();
				setState(1133);
				dataType();
				}
				}
				setState(1139);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTypeAlterTypeContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public KwTypeContext kwType() {
			return getRuleContext(KwTypeContext.class,0);
		}
		public DataTypeContext dataType() {
			return getRuleContext(DataTypeContext.class,0);
		}
		public AlterTypeAlterTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTypeAlterType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTypeAlterType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTypeAlterType(this);
		}
	}

	public final AlterTypeAlterTypeContext alterTypeAlterType() throws RecognitionException {
		AlterTypeAlterTypeContext _localctx = new AlterTypeAlterTypeContext(_ctx, getState());
		enterRule(_localctx, 90, RULE_alterTypeAlterType);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1140);
			kwAlter();
			setState(1141);
			column();
			setState(1142);
			kwType();
			setState(1143);
			dataType();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public AlterTableOperationContext alterTableOperation() {
			return getRuleContext(AlterTableOperationContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public AlterTableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTable; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTable(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTable(this);
		}
	}

	public final AlterTableContext alterTable() throws RecognitionException {
		AlterTableContext _localctx = new AlterTableContext(_ctx, getState());
		enterRule(_localctx, 92, RULE_alterTable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1145);
			kwAlter();
			setState(1146);
			kwTable();
			setState(1150);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,65,_ctx) ) {
			case 1:
				{
				setState(1147);
				keyspace();
				setState(1148);
				match(DOT);
				}
				break;
			}
			setState(1152);
			table();
			setState(1153);
			alterTableOperation();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableOperationContext extends ParserRuleContext {
		public AlterTableAddContext alterTableAdd() {
			return getRuleContext(AlterTableAddContext.class,0);
		}
		public AlterTableDropColumnsContext alterTableDropColumns() {
			return getRuleContext(AlterTableDropColumnsContext.class,0);
		}
		public AlterTableDropCompactStorageContext alterTableDropCompactStorage() {
			return getRuleContext(AlterTableDropCompactStorageContext.class,0);
		}
		public AlterTableRenameContext alterTableRename() {
			return getRuleContext(AlterTableRenameContext.class,0);
		}
		public AlterTableWithContext alterTableWith() {
			return getRuleContext(AlterTableWithContext.class,0);
		}
		public AlterTableOperationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableOperation; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableOperation(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableOperation(this);
		}
	}

	public final AlterTableOperationContext alterTableOperation() throws RecognitionException {
		AlterTableOperationContext _localctx = new AlterTableOperationContext(_ctx, getState());
		enterRule(_localctx, 94, RULE_alterTableOperation);
		try {
			setState(1160);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,66,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1155);
				alterTableAdd();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1156);
				alterTableDropColumns();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1157);
				alterTableDropCompactStorage();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(1158);
				alterTableRename();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(1159);
				alterTableWith();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableWithContext extends ParserRuleContext {
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public TableOptionsContext tableOptions() {
			return getRuleContext(TableOptionsContext.class,0);
		}
		public AlterTableWithContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableWith; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableWith(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableWith(this);
		}
	}

	public final AlterTableWithContext alterTableWith() throws RecognitionException {
		AlterTableWithContext _localctx = new AlterTableWithContext(_ctx, getState());
		enterRule(_localctx, 96, RULE_alterTableWith);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1162);
			kwWith();
			setState(1163);
			tableOptions();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableRenameContext extends ParserRuleContext {
		public KwRenameContext kwRename() {
			return getRuleContext(KwRenameContext.class,0);
		}
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public KwToContext kwTo() {
			return getRuleContext(KwToContext.class,0);
		}
		public AlterTableRenameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableRename; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableRename(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableRename(this);
		}
	}

	public final AlterTableRenameContext alterTableRename() throws RecognitionException {
		AlterTableRenameContext _localctx = new AlterTableRenameContext(_ctx, getState());
		enterRule(_localctx, 98, RULE_alterTableRename);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1165);
			kwRename();
			setState(1166);
			column();
			setState(1167);
			kwTo();
			setState(1168);
			column();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableDropCompactStorageContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwCompactContext kwCompact() {
			return getRuleContext(KwCompactContext.class,0);
		}
		public KwStorageContext kwStorage() {
			return getRuleContext(KwStorageContext.class,0);
		}
		public AlterTableDropCompactStorageContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableDropCompactStorage; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableDropCompactStorage(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableDropCompactStorage(this);
		}
	}

	public final AlterTableDropCompactStorageContext alterTableDropCompactStorage() throws RecognitionException {
		AlterTableDropCompactStorageContext _localctx = new AlterTableDropCompactStorageContext(_ctx, getState());
		enterRule(_localctx, 100, RULE_alterTableDropCompactStorage);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1170);
			kwDrop();
			setState(1171);
			kwCompact();
			setState(1172);
			kwStorage();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableDropColumnsContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public AlterTableDropColumnListContext alterTableDropColumnList() {
			return getRuleContext(AlterTableDropColumnListContext.class,0);
		}
		public AlterTableDropColumnsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableDropColumns; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableDropColumns(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableDropColumns(this);
		}
	}

	public final AlterTableDropColumnsContext alterTableDropColumns() throws RecognitionException {
		AlterTableDropColumnsContext _localctx = new AlterTableDropColumnsContext(_ctx, getState());
		enterRule(_localctx, 102, RULE_alterTableDropColumns);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1174);
			kwDrop();
			setState(1175);
			alterTableDropColumnList();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableDropColumnListContext extends ParserRuleContext {
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AlterTableDropColumnListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableDropColumnList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableDropColumnList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableDropColumnList(this);
		}
	}

	public final AlterTableDropColumnListContext alterTableDropColumnList() throws RecognitionException {
		AlterTableDropColumnListContext _localctx = new AlterTableDropColumnListContext(_ctx, getState());
		enterRule(_localctx, 104, RULE_alterTableDropColumnList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1177);
			column();
			setState(1183);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1178);
				syntaxComma();
				setState(1179);
				column();
				}
				}
				setState(1185);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableAddContext extends ParserRuleContext {
		public KwAddContext kwAdd() {
			return getRuleContext(KwAddContext.class,0);
		}
		public AlterTableColumnDefinitionContext alterTableColumnDefinition() {
			return getRuleContext(AlterTableColumnDefinitionContext.class,0);
		}
		public AlterTableAddContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableAdd; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableAdd(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableAdd(this);
		}
	}

	public final AlterTableAddContext alterTableAdd() throws RecognitionException {
		AlterTableAddContext _localctx = new AlterTableAddContext(_ctx, getState());
		enterRule(_localctx, 106, RULE_alterTableAdd);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1186);
			kwAdd();
			setState(1187);
			alterTableColumnDefinition();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterTableColumnDefinitionContext extends ParserRuleContext {
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<DataTypeContext> dataType() {
			return getRuleContexts(DataTypeContext.class);
		}
		public DataTypeContext dataType(int i) {
			return getRuleContext(DataTypeContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AlterTableColumnDefinitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterTableColumnDefinition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterTableColumnDefinition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterTableColumnDefinition(this);
		}
	}

	public final AlterTableColumnDefinitionContext alterTableColumnDefinition() throws RecognitionException {
		AlterTableColumnDefinitionContext _localctx = new AlterTableColumnDefinitionContext(_ctx, getState());
		enterRule(_localctx, 108, RULE_alterTableColumnDefinition);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1189);
			column();
			setState(1190);
			dataType();
			setState(1197);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1191);
				syntaxComma();
				setState(1192);
				column();
				setState(1193);
				dataType();
				}
				}
				setState(1199);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterRoleContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwRoleContext kwRole() {
			return getRuleContext(KwRoleContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public RoleWithContext roleWith() {
			return getRuleContext(RoleWithContext.class,0);
		}
		public AlterRoleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterRole; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterRole(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterRole(this);
		}
	}

	public final AlterRoleContext alterRole() throws RecognitionException {
		AlterRoleContext _localctx = new AlterRoleContext(_ctx, getState());
		enterRule(_localctx, 110, RULE_alterRole);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1200);
			kwAlter();
			setState(1201);
			kwRole();
			setState(1202);
			role();
			setState(1204);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WITH) {
				{
				setState(1203);
				roleWith();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RoleWithContext extends ParserRuleContext {
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public List<RoleWithOptionsContext> roleWithOptions() {
			return getRuleContexts(RoleWithOptionsContext.class);
		}
		public RoleWithOptionsContext roleWithOptions(int i) {
			return getRuleContext(RoleWithOptionsContext.class,i);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public RoleWithContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_roleWith; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRoleWith(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRoleWith(this);
		}
	}

	public final RoleWithContext roleWith() throws RecognitionException {
		RoleWithContext _localctx = new RoleWithContext(_ctx, getState());
		enterRule(_localctx, 112, RULE_roleWith);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1206);
			kwWith();
			{
			setState(1207);
			roleWithOptions();
			setState(1213);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==K_AND) {
				{
				{
				setState(1208);
				kwAnd();
				setState(1209);
				roleWithOptions();
				}
				}
				setState(1215);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RoleWithOptionsContext extends ParserRuleContext {
		public KwPasswordContext kwPassword() {
			return getRuleContext(KwPasswordContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public KwLoginContext kwLogin() {
			return getRuleContext(KwLoginContext.class,0);
		}
		public BooleanLiteralContext booleanLiteral() {
			return getRuleContext(BooleanLiteralContext.class,0);
		}
		public KwSuperuserContext kwSuperuser() {
			return getRuleContext(KwSuperuserContext.class,0);
		}
		public KwOptionsContext kwOptions() {
			return getRuleContext(KwOptionsContext.class,0);
		}
		public OptionHashContext optionHash() {
			return getRuleContext(OptionHashContext.class,0);
		}
		public RoleWithOptionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_roleWithOptions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRoleWithOptions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRoleWithOptions(this);
		}
	}

	public final RoleWithOptionsContext roleWithOptions() throws RecognitionException {
		RoleWithOptionsContext _localctx = new RoleWithOptionsContext(_ctx, getState());
		enterRule(_localctx, 114, RULE_roleWithOptions);
		try {
			setState(1232);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_PASSWORD:
				enterOuterAlt(_localctx, 1);
				{
				setState(1216);
				kwPassword();
				setState(1217);
				match(OPERATOR_EQ);
				setState(1218);
				stringLiteral();
				}
				break;
			case K_LOGIN:
				enterOuterAlt(_localctx, 2);
				{
				setState(1220);
				kwLogin();
				setState(1221);
				match(OPERATOR_EQ);
				setState(1222);
				booleanLiteral();
				}
				break;
			case K_SUPERUSER:
				enterOuterAlt(_localctx, 3);
				{
				setState(1224);
				kwSuperuser();
				setState(1225);
				match(OPERATOR_EQ);
				setState(1226);
				booleanLiteral();
				}
				break;
			case K_OPTIONS:
				enterOuterAlt(_localctx, 4);
				{
				setState(1228);
				kwOptions();
				setState(1229);
				match(OPERATOR_EQ);
				setState(1230);
				optionHash();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterMaterializedViewContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwMaterializedContext kwMaterialized() {
			return getRuleContext(KwMaterializedContext.class,0);
		}
		public KwViewContext kwView() {
			return getRuleContext(KwViewContext.class,0);
		}
		public MaterializedViewContext materializedView() {
			return getRuleContext(MaterializedViewContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public TableOptionsContext tableOptions() {
			return getRuleContext(TableOptionsContext.class,0);
		}
		public AlterMaterializedViewContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterMaterializedView; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterMaterializedView(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterMaterializedView(this);
		}
	}

	public final AlterMaterializedViewContext alterMaterializedView() throws RecognitionException {
		AlterMaterializedViewContext _localctx = new AlterMaterializedViewContext(_ctx, getState());
		enterRule(_localctx, 116, RULE_alterMaterializedView);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1234);
			kwAlter();
			setState(1235);
			kwMaterialized();
			setState(1236);
			kwView();
			setState(1240);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,72,_ctx) ) {
			case 1:
				{
				setState(1237);
				keyspace();
				setState(1238);
				match(DOT);
				}
				break;
			}
			setState(1242);
			materializedView();
			setState(1246);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WITH) {
				{
				setState(1243);
				kwWith();
				setState(1244);
				tableOptions();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropUserContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwUserContext kwUser() {
			return getRuleContext(KwUserContext.class,0);
		}
		public UserContext user() {
			return getRuleContext(UserContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public DropUserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropUser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropUser(this);
		}
	}

	public final DropUserContext dropUser() throws RecognitionException {
		DropUserContext _localctx = new DropUserContext(_ctx, getState());
		enterRule(_localctx, 118, RULE_dropUser);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1248);
			kwDrop();
			setState(1249);
			kwUser();
			setState(1251);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1250);
				ifExist();
				}
			}

			setState(1253);
			user();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropTypeContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwTypeContext kwType() {
			return getRuleContext(KwTypeContext.class,0);
		}
		public Type_Context type_() {
			return getRuleContext(Type_Context.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropType(this);
		}
	}

	public final DropTypeContext dropType() throws RecognitionException {
		DropTypeContext _localctx = new DropTypeContext(_ctx, getState());
		enterRule(_localctx, 120, RULE_dropType);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1255);
			kwDrop();
			setState(1256);
			kwType();
			setState(1258);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1257);
				ifExist();
				}
			}

			setState(1263);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,76,_ctx) ) {
			case 1:
				{
				setState(1260);
				keyspace();
				setState(1261);
				match(DOT);
				}
				break;
			}
			setState(1265);
			type_();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropMaterializedViewContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwMaterializedContext kwMaterialized() {
			return getRuleContext(KwMaterializedContext.class,0);
		}
		public KwViewContext kwView() {
			return getRuleContext(KwViewContext.class,0);
		}
		public MaterializedViewContext materializedView() {
			return getRuleContext(MaterializedViewContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropMaterializedViewContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropMaterializedView; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropMaterializedView(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropMaterializedView(this);
		}
	}

	public final DropMaterializedViewContext dropMaterializedView() throws RecognitionException {
		DropMaterializedViewContext _localctx = new DropMaterializedViewContext(_ctx, getState());
		enterRule(_localctx, 122, RULE_dropMaterializedView);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1267);
			kwDrop();
			setState(1268);
			kwMaterialized();
			setState(1269);
			kwView();
			setState(1271);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1270);
				ifExist();
				}
			}

			setState(1276);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,78,_ctx) ) {
			case 1:
				{
				setState(1273);
				keyspace();
				setState(1274);
				match(DOT);
				}
				break;
			}
			setState(1278);
			materializedView();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropAggregateContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwAggregateContext kwAggregate() {
			return getRuleContext(KwAggregateContext.class,0);
		}
		public AggregateContext aggregate() {
			return getRuleContext(AggregateContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropAggregateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropAggregate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropAggregate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropAggregate(this);
		}
	}

	public final DropAggregateContext dropAggregate() throws RecognitionException {
		DropAggregateContext _localctx = new DropAggregateContext(_ctx, getState());
		enterRule(_localctx, 124, RULE_dropAggregate);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1280);
			kwDrop();
			setState(1281);
			kwAggregate();
			setState(1283);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1282);
				ifExist();
				}
			}

			setState(1288);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,80,_ctx) ) {
			case 1:
				{
				setState(1285);
				keyspace();
				setState(1286);
				match(DOT);
				}
				break;
			}
			setState(1290);
			aggregate();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropFunctionContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwFunctionContext kwFunction() {
			return getRuleContext(KwFunctionContext.class,0);
		}
		public Function_Context function_() {
			return getRuleContext(Function_Context.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropFunctionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropFunction; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropFunction(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropFunction(this);
		}
	}

	public final DropFunctionContext dropFunction() throws RecognitionException {
		DropFunctionContext _localctx = new DropFunctionContext(_ctx, getState());
		enterRule(_localctx, 126, RULE_dropFunction);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1292);
			kwDrop();
			setState(1293);
			kwFunction();
			setState(1295);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1294);
				ifExist();
				}
			}

			setState(1300);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,82,_ctx) ) {
			case 1:
				{
				setState(1297);
				keyspace();
				setState(1298);
				match(DOT);
				}
				break;
			}
			setState(1302);
			function_();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropTriggerContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwTriggerContext kwTrigger() {
			return getRuleContext(KwTriggerContext.class,0);
		}
		public TriggerContext trigger() {
			return getRuleContext(TriggerContext.class,0);
		}
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropTriggerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropTrigger; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropTrigger(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropTrigger(this);
		}
	}

	public final DropTriggerContext dropTrigger() throws RecognitionException {
		DropTriggerContext _localctx = new DropTriggerContext(_ctx, getState());
		enterRule(_localctx, 128, RULE_dropTrigger);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1304);
			kwDrop();
			setState(1305);
			kwTrigger();
			setState(1307);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1306);
				ifExist();
				}
			}

			setState(1309);
			trigger();
			setState(1310);
			kwOn();
			setState(1314);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,84,_ctx) ) {
			case 1:
				{
				setState(1311);
				keyspace();
				setState(1312);
				match(DOT);
				}
				break;
			}
			setState(1316);
			table();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropRoleContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwRoleContext kwRole() {
			return getRuleContext(KwRoleContext.class,0);
		}
		public RoleContext role() {
			return getRuleContext(RoleContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public DropRoleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropRole; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropRole(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropRole(this);
		}
	}

	public final DropRoleContext dropRole() throws RecognitionException {
		DropRoleContext _localctx = new DropRoleContext(_ctx, getState());
		enterRule(_localctx, 130, RULE_dropRole);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1318);
			kwDrop();
			setState(1319);
			kwRole();
			setState(1321);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1320);
				ifExist();
				}
			}

			setState(1323);
			role();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropTableContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropTableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropTable; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropTable(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropTable(this);
		}
	}

	public final DropTableContext dropTable() throws RecognitionException {
		DropTableContext _localctx = new DropTableContext(_ctx, getState());
		enterRule(_localctx, 132, RULE_dropTable);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1325);
			kwDrop();
			setState(1326);
			kwTable();
			setState(1328);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1327);
				ifExist();
				}
			}

			setState(1333);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,87,_ctx) ) {
			case 1:
				{
				setState(1330);
				keyspace();
				setState(1331);
				match(DOT);
				}
				break;
			}
			setState(1335);
			table();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropKeyspaceContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwKeyspaceContext kwKeyspace() {
			return getRuleContext(KwKeyspaceContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public DropKeyspaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropKeyspace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropKeyspace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropKeyspace(this);
		}
	}

	public final DropKeyspaceContext dropKeyspace() throws RecognitionException {
		DropKeyspaceContext _localctx = new DropKeyspaceContext(_ctx, getState());
		enterRule(_localctx, 134, RULE_dropKeyspace);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1337);
			kwDrop();
			setState(1338);
			kwKeyspace();
			setState(1340);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1339);
				ifExist();
				}
			}

			setState(1342);
			keyspace();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DropIndexContext extends ParserRuleContext {
		public KwDropContext kwDrop() {
			return getRuleContext(KwDropContext.class,0);
		}
		public KwIndexContext kwIndex() {
			return getRuleContext(KwIndexContext.class,0);
		}
		public IndexNameContext indexName() {
			return getRuleContext(IndexNameContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public DropIndexContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dropIndex; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDropIndex(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDropIndex(this);
		}
	}

	public final DropIndexContext dropIndex() throws RecognitionException {
		DropIndexContext _localctx = new DropIndexContext(_ctx, getState());
		enterRule(_localctx, 136, RULE_dropIndex);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1344);
			kwDrop();
			setState(1345);
			kwIndex();
			setState(1347);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1346);
				ifExist();
				}
			}

			setState(1352);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,90,_ctx) ) {
			case 1:
				{
				setState(1349);
				keyspace();
				setState(1350);
				match(DOT);
				}
				break;
			}
			setState(1354);
			indexName();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateTableContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public ColumnDefinitionListContext columnDefinitionList() {
			return getRuleContext(ColumnDefinitionListContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public WithElementContext withElement() {
			return getRuleContext(WithElementContext.class,0);
		}
		public CreateTableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createTable; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateTable(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateTable(this);
		}
	}

	public final CreateTableContext createTable() throws RecognitionException {
		CreateTableContext _localctx = new CreateTableContext(_ctx, getState());
		enterRule(_localctx, 138, RULE_createTable);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1356);
			kwCreate();
			setState(1357);
			kwTable();
			setState(1359);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1358);
				ifNotExist();
				}
			}

			setState(1364);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,92,_ctx) ) {
			case 1:
				{
				setState(1361);
				keyspace();
				setState(1362);
				match(DOT);
				}
				break;
			}
			setState(1366);
			table();
			setState(1367);
			syntaxBracketLr();
			setState(1368);
			columnDefinitionList();
			setState(1369);
			syntaxBracketRr();
			setState(1371);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WITH) {
				{
				setState(1370);
				withElement();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class WithElementContext extends ParserRuleContext {
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public TableOptionsContext tableOptions() {
			return getRuleContext(TableOptionsContext.class,0);
		}
		public WithElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_withElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterWithElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitWithElement(this);
		}
	}

	public final WithElementContext withElement() throws RecognitionException {
		WithElementContext _localctx = new WithElementContext(_ctx, getState());
		enterRule(_localctx, 140, RULE_withElement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1373);
			kwWith();
			setState(1374);
			tableOptions();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TableOptionsContext extends ParserRuleContext {
		public KwCompactContext kwCompact() {
			return getRuleContext(KwCompactContext.class,0);
		}
		public KwStorageContext kwStorage() {
			return getRuleContext(KwStorageContext.class,0);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public TableOptionsContext tableOptions() {
			return getRuleContext(TableOptionsContext.class,0);
		}
		public ClusteringOrderContext clusteringOrder() {
			return getRuleContext(ClusteringOrderContext.class,0);
		}
		public List<TableOptionItemContext> tableOptionItem() {
			return getRuleContexts(TableOptionItemContext.class);
		}
		public TableOptionItemContext tableOptionItem(int i) {
			return getRuleContext(TableOptionItemContext.class,i);
		}
		public TableOptionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_tableOptions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTableOptions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTableOptions(this);
		}
	}

	public final TableOptionsContext tableOptions() throws RecognitionException {
		TableOptionsContext _localctx = new TableOptionsContext(_ctx, getState());
		enterRule(_localctx, 142, RULE_tableOptions);
		try {
			int _alt;
			setState(1398);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_COMPACT:
				enterOuterAlt(_localctx, 1);
				{
				setState(1376);
				kwCompact();
				setState(1377);
				kwStorage();
				setState(1381);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,94,_ctx) ) {
				case 1:
					{
					setState(1378);
					kwAnd();
					setState(1379);
					tableOptions();
					}
					break;
				}
				}
				break;
			case K_CLUSTERING:
				enterOuterAlt(_localctx, 2);
				{
				setState(1383);
				clusteringOrder();
				setState(1387);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,95,_ctx) ) {
				case 1:
					{
					setState(1384);
					kwAnd();
					setState(1385);
					tableOptions();
					}
					break;
				}
				}
				break;
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 3);
				{
				setState(1389);
				tableOptionItem();
				setState(1395);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,96,_ctx);
				while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
					if ( _alt==1 ) {
						{
						{
						setState(1390);
						kwAnd();
						setState(1391);
						tableOptionItem();
						}
						} 
					}
					setState(1397);
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,96,_ctx);
				}
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ClusteringOrderContext extends ParserRuleContext {
		public KwClusteringContext kwClustering() {
			return getRuleContext(KwClusteringContext.class,0);
		}
		public KwOrderContext kwOrder() {
			return getRuleContext(KwOrderContext.class,0);
		}
		public KwByContext kwBy() {
			return getRuleContext(KwByContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public List<OrderDirectionContext> orderDirection() {
			return getRuleContexts(OrderDirectionContext.class);
		}
		public OrderDirectionContext orderDirection(int i) {
			return getRuleContext(OrderDirectionContext.class,i);
		}
		public ClusteringOrderContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_clusteringOrder; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterClusteringOrder(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitClusteringOrder(this);
		}
	}

	public final ClusteringOrderContext clusteringOrder() throws RecognitionException {
		ClusteringOrderContext _localctx = new ClusteringOrderContext(_ctx, getState());
		enterRule(_localctx, 144, RULE_clusteringOrder);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1400);
			kwClustering();
			setState(1401);
			kwOrder();
			setState(1402);
			kwBy();
			setState(1403);
			syntaxBracketLr();
			{
			setState(1404);
			column();
			setState(1406);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_ASC || _la==K_DESC) {
				{
				setState(1405);
				orderDirection();
				}
			}

			}
			setState(1415);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1408);
				syntaxComma();
				setState(1409);
				column();
				setState(1411);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_ASC || _la==K_DESC) {
					{
					setState(1410);
					orderDirection();
					}
				}

				}
				}
				setState(1417);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1418);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TableOptionItemContext extends ParserRuleContext {
		public TableOptionNameContext tableOptionName() {
			return getRuleContext(TableOptionNameContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public TableOptionValueContext tableOptionValue() {
			return getRuleContext(TableOptionValueContext.class,0);
		}
		public OptionHashContext optionHash() {
			return getRuleContext(OptionHashContext.class,0);
		}
		public TableOptionItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_tableOptionItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTableOptionItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTableOptionItem(this);
		}
	}

	public final TableOptionItemContext tableOptionItem() throws RecognitionException {
		TableOptionItemContext _localctx = new TableOptionItemContext(_ctx, getState());
		enterRule(_localctx, 146, RULE_tableOptionItem);
		try {
			setState(1428);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,101,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1420);
				tableOptionName();
				setState(1421);
				match(OPERATOR_EQ);
				setState(1422);
				tableOptionValue();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1424);
				tableOptionName();
				setState(1425);
				match(OPERATOR_EQ);
				setState(1426);
				optionHash();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TableOptionNameContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TableOptionNameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_tableOptionName; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTableOptionName(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTableOptionName(this);
		}
	}

	public final TableOptionNameContext tableOptionName() throws RecognitionException {
		TableOptionNameContext _localctx = new TableOptionNameContext(_ctx, getState());
		enterRule(_localctx, 148, RULE_tableOptionName);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1430);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TableOptionValueContext extends ParserRuleContext {
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public FloatLiteralContext floatLiteral() {
			return getRuleContext(FloatLiteralContext.class,0);
		}
		public TableOptionValueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_tableOptionValue; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTableOptionValue(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTableOptionValue(this);
		}
	}

	public final TableOptionValueContext tableOptionValue() throws RecognitionException {
		TableOptionValueContext _localctx = new TableOptionValueContext(_ctx, getState());
		enterRule(_localctx, 150, RULE_tableOptionValue);
		try {
			setState(1434);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STRING_LITERAL:
				enterOuterAlt(_localctx, 1);
				{
				setState(1432);
				stringLiteral();
				}
				break;
			case DECIMAL_LITERAL:
			case FLOAT_LITERAL:
				enterOuterAlt(_localctx, 2);
				{
				setState(1433);
				floatLiteral();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OptionHashContext extends ParserRuleContext {
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public List<OptionHashItemContext> optionHashItem() {
			return getRuleContexts(OptionHashItemContext.class);
		}
		public OptionHashItemContext optionHashItem(int i) {
			return getRuleContext(OptionHashItemContext.class,i);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public OptionHashContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_optionHash; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOptionHash(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOptionHash(this);
		}
	}

	public final OptionHashContext optionHash() throws RecognitionException {
		OptionHashContext _localctx = new OptionHashContext(_ctx, getState());
		enterRule(_localctx, 152, RULE_optionHash);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1436);
			syntaxBracketLc();
			setState(1437);
			optionHashItem();
			setState(1443);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1438);
				syntaxComma();
				setState(1439);
				optionHashItem();
				}
				}
				setState(1445);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1446);
			syntaxBracketRc();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OptionHashItemContext extends ParserRuleContext {
		public OptionHashKeyContext optionHashKey() {
			return getRuleContext(OptionHashKeyContext.class,0);
		}
		public TerminalNode COLON() { return getToken(CqlParser.COLON, 0); }
		public OptionHashValueContext optionHashValue() {
			return getRuleContext(OptionHashValueContext.class,0);
		}
		public OptionHashItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_optionHashItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOptionHashItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOptionHashItem(this);
		}
	}

	public final OptionHashItemContext optionHashItem() throws RecognitionException {
		OptionHashItemContext _localctx = new OptionHashItemContext(_ctx, getState());
		enterRule(_localctx, 154, RULE_optionHashItem);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1448);
			optionHashKey();
			setState(1449);
			match(COLON);
			setState(1450);
			optionHashValue();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OptionHashKeyContext extends ParserRuleContext {
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public OptionHashKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_optionHashKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOptionHashKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOptionHashKey(this);
		}
	}

	public final OptionHashKeyContext optionHashKey() throws RecognitionException {
		OptionHashKeyContext _localctx = new OptionHashKeyContext(_ctx, getState());
		enterRule(_localctx, 156, RULE_optionHashKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1452);
			stringLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OptionHashValueContext extends ParserRuleContext {
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public FloatLiteralContext floatLiteral() {
			return getRuleContext(FloatLiteralContext.class,0);
		}
		public OptionHashValueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_optionHashValue; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOptionHashValue(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOptionHashValue(this);
		}
	}

	public final OptionHashValueContext optionHashValue() throws RecognitionException {
		OptionHashValueContext _localctx = new OptionHashValueContext(_ctx, getState());
		enterRule(_localctx, 158, RULE_optionHashValue);
		try {
			setState(1456);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STRING_LITERAL:
				enterOuterAlt(_localctx, 1);
				{
				setState(1454);
				stringLiteral();
				}
				break;
			case DECIMAL_LITERAL:
			case FLOAT_LITERAL:
				enterOuterAlt(_localctx, 2);
				{
				setState(1455);
				floatLiteral();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnDefinitionListContext extends ParserRuleContext {
		public List<ColumnDefinitionContext> columnDefinition() {
			return getRuleContexts(ColumnDefinitionContext.class);
		}
		public ColumnDefinitionContext columnDefinition(int i) {
			return getRuleContext(ColumnDefinitionContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public PrimaryKeyElementContext primaryKeyElement() {
			return getRuleContext(PrimaryKeyElementContext.class,0);
		}
		public ColumnDefinitionListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_columnDefinitionList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumnDefinitionList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumnDefinitionList(this);
		}
	}

	public final ColumnDefinitionListContext columnDefinitionList() throws RecognitionException {
		ColumnDefinitionListContext _localctx = new ColumnDefinitionListContext(_ctx, getState());
		enterRule(_localctx, 160, RULE_columnDefinitionList);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1458);
			columnDefinition();
			}
			setState(1464);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,105,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					{
					{
					setState(1459);
					syntaxComma();
					setState(1460);
					columnDefinition();
					}
					} 
				}
				setState(1466);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,105,_ctx);
			}
			setState(1470);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==COMMA) {
				{
				setState(1467);
				syntaxComma();
				setState(1468);
				primaryKeyElement();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnDefinitionContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public DataTypeContext dataType() {
			return getRuleContext(DataTypeContext.class,0);
		}
		public PrimaryKeyColumnContext primaryKeyColumn() {
			return getRuleContext(PrimaryKeyColumnContext.class,0);
		}
		public ColumnDefinitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_columnDefinition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumnDefinition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumnDefinition(this);
		}
	}

	public final ColumnDefinitionContext columnDefinition() throws RecognitionException {
		ColumnDefinitionContext _localctx = new ColumnDefinitionContext(_ctx, getState());
		enterRule(_localctx, 162, RULE_columnDefinition);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1472);
			column();
			setState(1473);
			dataType();
			setState(1475);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_PRIMARY) {
				{
				setState(1474);
				primaryKeyColumn();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PrimaryKeyColumnContext extends ParserRuleContext {
		public KwPrimaryContext kwPrimary() {
			return getRuleContext(KwPrimaryContext.class,0);
		}
		public KwKeyContext kwKey() {
			return getRuleContext(KwKeyContext.class,0);
		}
		public PrimaryKeyColumnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_primaryKeyColumn; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPrimaryKeyColumn(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPrimaryKeyColumn(this);
		}
	}

	public final PrimaryKeyColumnContext primaryKeyColumn() throws RecognitionException {
		PrimaryKeyColumnContext _localctx = new PrimaryKeyColumnContext(_ctx, getState());
		enterRule(_localctx, 164, RULE_primaryKeyColumn);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1477);
			kwPrimary();
			setState(1478);
			kwKey();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PrimaryKeyElementContext extends ParserRuleContext {
		public KwPrimaryContext kwPrimary() {
			return getRuleContext(KwPrimaryContext.class,0);
		}
		public KwKeyContext kwKey() {
			return getRuleContext(KwKeyContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public PrimaryKeyDefinitionContext primaryKeyDefinition() {
			return getRuleContext(PrimaryKeyDefinitionContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public PrimaryKeyElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_primaryKeyElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPrimaryKeyElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPrimaryKeyElement(this);
		}
	}

	public final PrimaryKeyElementContext primaryKeyElement() throws RecognitionException {
		PrimaryKeyElementContext _localctx = new PrimaryKeyElementContext(_ctx, getState());
		enterRule(_localctx, 166, RULE_primaryKeyElement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1480);
			kwPrimary();
			setState(1481);
			kwKey();
			setState(1482);
			syntaxBracketLr();
			setState(1483);
			primaryKeyDefinition();
			setState(1484);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PrimaryKeyDefinitionContext extends ParserRuleContext {
		public SinglePrimaryKeyContext singlePrimaryKey() {
			return getRuleContext(SinglePrimaryKeyContext.class,0);
		}
		public CompoundKeyContext compoundKey() {
			return getRuleContext(CompoundKeyContext.class,0);
		}
		public CompositeKeyContext compositeKey() {
			return getRuleContext(CompositeKeyContext.class,0);
		}
		public PrimaryKeyDefinitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_primaryKeyDefinition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPrimaryKeyDefinition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPrimaryKeyDefinition(this);
		}
	}

	public final PrimaryKeyDefinitionContext primaryKeyDefinition() throws RecognitionException {
		PrimaryKeyDefinitionContext _localctx = new PrimaryKeyDefinitionContext(_ctx, getState());
		enterRule(_localctx, 168, RULE_primaryKeyDefinition);
		try {
			setState(1489);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,108,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1486);
				singlePrimaryKey();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1487);
				compoundKey();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1488);
				compositeKey();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SinglePrimaryKeyContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public SinglePrimaryKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_singlePrimaryKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSinglePrimaryKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSinglePrimaryKey(this);
		}
	}

	public final SinglePrimaryKeyContext singlePrimaryKey() throws RecognitionException {
		SinglePrimaryKeyContext _localctx = new SinglePrimaryKeyContext(_ctx, getState());
		enterRule(_localctx, 170, RULE_singlePrimaryKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1491);
			column();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CompoundKeyContext extends ParserRuleContext {
		public PartitionKeyContext partitionKey() {
			return getRuleContext(PartitionKeyContext.class,0);
		}
		public SyntaxCommaContext syntaxComma() {
			return getRuleContext(SyntaxCommaContext.class,0);
		}
		public ClusteringKeyListContext clusteringKeyList() {
			return getRuleContext(ClusteringKeyListContext.class,0);
		}
		public CompoundKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_compoundKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCompoundKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCompoundKey(this);
		}
	}

	public final CompoundKeyContext compoundKey() throws RecognitionException {
		CompoundKeyContext _localctx = new CompoundKeyContext(_ctx, getState());
		enterRule(_localctx, 172, RULE_compoundKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1493);
			partitionKey();
			{
			setState(1494);
			syntaxComma();
			setState(1495);
			clusteringKeyList();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CompositeKeyContext extends ParserRuleContext {
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public PartitionKeyListContext partitionKeyList() {
			return getRuleContext(PartitionKeyListContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public SyntaxCommaContext syntaxComma() {
			return getRuleContext(SyntaxCommaContext.class,0);
		}
		public ClusteringKeyListContext clusteringKeyList() {
			return getRuleContext(ClusteringKeyListContext.class,0);
		}
		public CompositeKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_compositeKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCompositeKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCompositeKey(this);
		}
	}

	public final CompositeKeyContext compositeKey() throws RecognitionException {
		CompositeKeyContext _localctx = new CompositeKeyContext(_ctx, getState());
		enterRule(_localctx, 174, RULE_compositeKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1497);
			syntaxBracketLr();
			setState(1498);
			partitionKeyList();
			setState(1499);
			syntaxBracketRr();
			{
			setState(1500);
			syntaxComma();
			setState(1501);
			clusteringKeyList();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PartitionKeyListContext extends ParserRuleContext {
		public List<PartitionKeyContext> partitionKey() {
			return getRuleContexts(PartitionKeyContext.class);
		}
		public PartitionKeyContext partitionKey(int i) {
			return getRuleContext(PartitionKeyContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public PartitionKeyListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_partitionKeyList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPartitionKeyList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPartitionKeyList(this);
		}
	}

	public final PartitionKeyListContext partitionKeyList() throws RecognitionException {
		PartitionKeyListContext _localctx = new PartitionKeyListContext(_ctx, getState());
		enterRule(_localctx, 176, RULE_partitionKeyList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1503);
			partitionKey();
			}
			setState(1509);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1504);
				syntaxComma();
				setState(1505);
				partitionKey();
				}
				}
				setState(1511);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ClusteringKeyListContext extends ParserRuleContext {
		public List<ClusteringKeyContext> clusteringKey() {
			return getRuleContexts(ClusteringKeyContext.class);
		}
		public ClusteringKeyContext clusteringKey(int i) {
			return getRuleContext(ClusteringKeyContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public ClusteringKeyListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_clusteringKeyList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterClusteringKeyList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitClusteringKeyList(this);
		}
	}

	public final ClusteringKeyListContext clusteringKeyList() throws RecognitionException {
		ClusteringKeyListContext _localctx = new ClusteringKeyListContext(_ctx, getState());
		enterRule(_localctx, 178, RULE_clusteringKeyList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1512);
			clusteringKey();
			}
			setState(1518);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1513);
				syntaxComma();
				setState(1514);
				clusteringKey();
				}
				}
				setState(1520);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PartitionKeyContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public PartitionKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_partitionKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPartitionKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPartitionKey(this);
		}
	}

	public final PartitionKeyContext partitionKey() throws RecognitionException {
		PartitionKeyContext _localctx = new PartitionKeyContext(_ctx, getState());
		enterRule(_localctx, 180, RULE_partitionKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1521);
			column();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ClusteringKeyContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public ClusteringKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_clusteringKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterClusteringKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitClusteringKey(this);
		}
	}

	public final ClusteringKeyContext clusteringKey() throws RecognitionException {
		ClusteringKeyContext _localctx = new ClusteringKeyContext(_ctx, getState());
		enterRule(_localctx, 182, RULE_clusteringKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1523);
			column();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ApplyBatchContext extends ParserRuleContext {
		public KwApplyContext kwApply() {
			return getRuleContext(KwApplyContext.class,0);
		}
		public KwBatchContext kwBatch() {
			return getRuleContext(KwBatchContext.class,0);
		}
		public ApplyBatchContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_applyBatch; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterApplyBatch(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitApplyBatch(this);
		}
	}

	public final ApplyBatchContext applyBatch() throws RecognitionException {
		ApplyBatchContext _localctx = new ApplyBatchContext(_ctx, getState());
		enterRule(_localctx, 184, RULE_applyBatch);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1525);
			kwApply();
			setState(1526);
			kwBatch();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class BeginBatchContext extends ParserRuleContext {
		public KwBeginContext kwBegin() {
			return getRuleContext(KwBeginContext.class,0);
		}
		public KwBatchContext kwBatch() {
			return getRuleContext(KwBatchContext.class,0);
		}
		public BatchTypeContext batchType() {
			return getRuleContext(BatchTypeContext.class,0);
		}
		public UsingTimestampSpecContext usingTimestampSpec() {
			return getRuleContext(UsingTimestampSpecContext.class,0);
		}
		public BeginBatchContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_beginBatch; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterBeginBatch(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitBeginBatch(this);
		}
	}

	public final BeginBatchContext beginBatch() throws RecognitionException {
		BeginBatchContext _localctx = new BeginBatchContext(_ctx, getState());
		enterRule(_localctx, 186, RULE_beginBatch);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1528);
			kwBegin();
			setState(1530);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_LOGGED || _la==K_UNLOGGED) {
				{
				setState(1529);
				batchType();
				}
			}

			setState(1532);
			kwBatch();
			setState(1534);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_USING) {
				{
				setState(1533);
				usingTimestampSpec();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class BatchTypeContext extends ParserRuleContext {
		public KwLoggedContext kwLogged() {
			return getRuleContext(KwLoggedContext.class,0);
		}
		public KwUnloggedContext kwUnlogged() {
			return getRuleContext(KwUnloggedContext.class,0);
		}
		public BatchTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_batchType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterBatchType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitBatchType(this);
		}
	}

	public final BatchTypeContext batchType() throws RecognitionException {
		BatchTypeContext _localctx = new BatchTypeContext(_ctx, getState());
		enterRule(_localctx, 188, RULE_batchType);
		try {
			setState(1538);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_LOGGED:
				enterOuterAlt(_localctx, 1);
				{
				setState(1536);
				kwLogged();
				}
				break;
			case K_UNLOGGED:
				enterOuterAlt(_localctx, 2);
				{
				setState(1537);
				kwUnlogged();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AlterKeyspaceContext extends ParserRuleContext {
		public KwAlterContext kwAlter() {
			return getRuleContext(KwAlterContext.class,0);
		}
		public KwKeyspaceContext kwKeyspace() {
			return getRuleContext(KwKeyspaceContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public KwWithContext kwWith() {
			return getRuleContext(KwWithContext.class,0);
		}
		public KwReplicationContext kwReplication() {
			return getRuleContext(KwReplicationContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public ReplicationListContext replicationList() {
			return getRuleContext(ReplicationListContext.class,0);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public KwAndContext kwAnd() {
			return getRuleContext(KwAndContext.class,0);
		}
		public DurableWritesContext durableWrites() {
			return getRuleContext(DurableWritesContext.class,0);
		}
		public AlterKeyspaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_alterKeyspace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAlterKeyspace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAlterKeyspace(this);
		}
	}

	public final AlterKeyspaceContext alterKeyspace() throws RecognitionException {
		AlterKeyspaceContext _localctx = new AlterKeyspaceContext(_ctx, getState());
		enterRule(_localctx, 190, RULE_alterKeyspace);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1540);
			kwAlter();
			setState(1541);
			kwKeyspace();
			setState(1542);
			keyspace();
			setState(1543);
			kwWith();
			setState(1544);
			kwReplication();
			setState(1545);
			match(OPERATOR_EQ);
			setState(1546);
			syntaxBracketLc();
			setState(1547);
			replicationList();
			setState(1548);
			syntaxBracketRc();
			setState(1552);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_AND) {
				{
				setState(1549);
				kwAnd();
				setState(1550);
				durableWrites();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReplicationListContext extends ParserRuleContext {
		public List<ReplicationListItemContext> replicationListItem() {
			return getRuleContexts(ReplicationListItemContext.class);
		}
		public ReplicationListItemContext replicationListItem(int i) {
			return getRuleContext(ReplicationListItemContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public ReplicationListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_replicationList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterReplicationList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitReplicationList(this);
		}
	}

	public final ReplicationListContext replicationList() throws RecognitionException {
		ReplicationListContext _localctx = new ReplicationListContext(_ctx, getState());
		enterRule(_localctx, 192, RULE_replicationList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1554);
			replicationListItem();
			}
			setState(1560);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1555);
				syntaxComma();
				setState(1556);
				replicationListItem();
				}
				}
				setState(1562);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReplicationListItemContext extends ParserRuleContext {
		public List<TerminalNode> STRING_LITERAL() { return getTokens(CqlParser.STRING_LITERAL); }
		public TerminalNode STRING_LITERAL(int i) {
			return getToken(CqlParser.STRING_LITERAL, i);
		}
		public TerminalNode COLON() { return getToken(CqlParser.COLON, 0); }
		public TerminalNode DECIMAL_LITERAL() { return getToken(CqlParser.DECIMAL_LITERAL, 0); }
		public ReplicationListItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_replicationListItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterReplicationListItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitReplicationListItem(this);
		}
	}

	public final ReplicationListItemContext replicationListItem() throws RecognitionException {
		ReplicationListItemContext _localctx = new ReplicationListItemContext(_ctx, getState());
		enterRule(_localctx, 194, RULE_replicationListItem);
		try {
			setState(1569);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,116,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1563);
				match(STRING_LITERAL);
				setState(1564);
				match(COLON);
				setState(1565);
				match(STRING_LITERAL);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1566);
				match(STRING_LITERAL);
				setState(1567);
				match(COLON);
				setState(1568);
				match(DECIMAL_LITERAL);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DurableWritesContext extends ParserRuleContext {
		public KwDurableWritesContext kwDurableWrites() {
			return getRuleContext(KwDurableWritesContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public BooleanLiteralContext booleanLiteral() {
			return getRuleContext(BooleanLiteralContext.class,0);
		}
		public DurableWritesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_durableWrites; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDurableWrites(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDurableWrites(this);
		}
	}

	public final DurableWritesContext durableWrites() throws RecognitionException {
		DurableWritesContext _localctx = new DurableWritesContext(_ctx, getState());
		enterRule(_localctx, 196, RULE_durableWrites);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1571);
			kwDurableWrites();
			setState(1572);
			match(OPERATOR_EQ);
			setState(1573);
			booleanLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Use_Context extends ParserRuleContext {
		public KwUseContext kwUse() {
			return getRuleContext(KwUseContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public Use_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_use_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUse_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUse_(this);
		}
	}

	public final Use_Context use_() throws RecognitionException {
		Use_Context _localctx = new Use_Context(_ctx, getState());
		enterRule(_localctx, 198, RULE_use_);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1575);
			kwUse();
			setState(1576);
			keyspace();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TruncateContext extends ParserRuleContext {
		public KwTruncateContext kwTruncate() {
			return getRuleContext(KwTruncateContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public KwTableContext kwTable() {
			return getRuleContext(KwTableContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public TruncateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_truncate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTruncate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTruncate(this);
		}
	}

	public final TruncateContext truncate() throws RecognitionException {
		TruncateContext _localctx = new TruncateContext(_ctx, getState());
		enterRule(_localctx, 200, RULE_truncate);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1578);
			kwTruncate();
			setState(1580);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_TABLE) {
				{
				setState(1579);
				kwTable();
				}
			}

			setState(1585);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,118,_ctx) ) {
			case 1:
				{
				setState(1582);
				keyspace();
				setState(1583);
				match(DOT);
				}
				break;
			}
			setState(1587);
			table();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CreateIndexContext extends ParserRuleContext {
		public KwCreateContext kwCreate() {
			return getRuleContext(KwCreateContext.class,0);
		}
		public KwIndexContext kwIndex() {
			return getRuleContext(KwIndexContext.class,0);
		}
		public KwOnContext kwOn() {
			return getRuleContext(KwOnContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public IndexColumnSpecContext indexColumnSpec() {
			return getRuleContext(IndexColumnSpecContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public IndexNameContext indexName() {
			return getRuleContext(IndexNameContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public CreateIndexContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_createIndex; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterCreateIndex(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitCreateIndex(this);
		}
	}

	public final CreateIndexContext createIndex() throws RecognitionException {
		CreateIndexContext _localctx = new CreateIndexContext(_ctx, getState());
		enterRule(_localctx, 202, RULE_createIndex);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1589);
			kwCreate();
			setState(1590);
			kwIndex();
			setState(1592);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1591);
				ifNotExist();
				}
			}

			setState(1595);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==STRING_LITERAL || _la==OBJECT_NAME) {
				{
				setState(1594);
				indexName();
				}
			}

			setState(1597);
			kwOn();
			setState(1601);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,121,_ctx) ) {
			case 1:
				{
				setState(1598);
				keyspace();
				setState(1599);
				match(DOT);
				}
				break;
			}
			setState(1603);
			table();
			setState(1604);
			syntaxBracketLr();
			setState(1605);
			indexColumnSpec();
			setState(1606);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IndexNameContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public IndexNameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexName; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIndexName(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIndexName(this);
		}
	}

	public final IndexNameContext indexName() throws RecognitionException {
		IndexNameContext _localctx = new IndexNameContext(_ctx, getState());
		enterRule(_localctx, 204, RULE_indexName);
		try {
			setState(1610);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 1);
				{
				setState(1608);
				match(OBJECT_NAME);
				}
				break;
			case STRING_LITERAL:
				enterOuterAlt(_localctx, 2);
				{
				setState(1609);
				stringLiteral();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IndexColumnSpecContext extends ParserRuleContext {
		public ColumnContext column() {
			return getRuleContext(ColumnContext.class,0);
		}
		public IndexKeysSpecContext indexKeysSpec() {
			return getRuleContext(IndexKeysSpecContext.class,0);
		}
		public IndexEntriesSSpecContext indexEntriesSSpec() {
			return getRuleContext(IndexEntriesSSpecContext.class,0);
		}
		public IndexFullSpecContext indexFullSpec() {
			return getRuleContext(IndexFullSpecContext.class,0);
		}
		public IndexColumnSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexColumnSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIndexColumnSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIndexColumnSpec(this);
		}
	}

	public final IndexColumnSpecContext indexColumnSpec() throws RecognitionException {
		IndexColumnSpecContext _localctx = new IndexColumnSpecContext(_ctx, getState());
		enterRule(_localctx, 206, RULE_indexColumnSpec);
		try {
			setState(1616);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case DQUOTE:
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 1);
				{
				setState(1612);
				column();
				}
				break;
			case K_KEYS:
				enterOuterAlt(_localctx, 2);
				{
				setState(1613);
				indexKeysSpec();
				}
				break;
			case K_ENTRIES:
				enterOuterAlt(_localctx, 3);
				{
				setState(1614);
				indexEntriesSSpec();
				}
				break;
			case K_FULL:
				enterOuterAlt(_localctx, 4);
				{
				setState(1615);
				indexFullSpec();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IndexKeysSpecContext extends ParserRuleContext {
		public KwKeysContext kwKeys() {
			return getRuleContext(KwKeysContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IndexKeysSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexKeysSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIndexKeysSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIndexKeysSpec(this);
		}
	}

	public final IndexKeysSpecContext indexKeysSpec() throws RecognitionException {
		IndexKeysSpecContext _localctx = new IndexKeysSpecContext(_ctx, getState());
		enterRule(_localctx, 208, RULE_indexKeysSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1618);
			kwKeys();
			setState(1619);
			syntaxBracketLr();
			setState(1620);
			match(OBJECT_NAME);
			setState(1621);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IndexEntriesSSpecContext extends ParserRuleContext {
		public KwEntriesContext kwEntries() {
			return getRuleContext(KwEntriesContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IndexEntriesSSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexEntriesSSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIndexEntriesSSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIndexEntriesSSpec(this);
		}
	}

	public final IndexEntriesSSpecContext indexEntriesSSpec() throws RecognitionException {
		IndexEntriesSSpecContext _localctx = new IndexEntriesSSpecContext(_ctx, getState());
		enterRule(_localctx, 210, RULE_indexEntriesSSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1623);
			kwEntries();
			setState(1624);
			syntaxBracketLr();
			setState(1625);
			match(OBJECT_NAME);
			setState(1626);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IndexFullSpecContext extends ParserRuleContext {
		public KwFullContext kwFull() {
			return getRuleContext(KwFullContext.class,0);
		}
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public IndexFullSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_indexFullSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIndexFullSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIndexFullSpec(this);
		}
	}

	public final IndexFullSpecContext indexFullSpec() throws RecognitionException {
		IndexFullSpecContext _localctx = new IndexFullSpecContext(_ctx, getState());
		enterRule(_localctx, 212, RULE_indexFullSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1628);
			kwFull();
			setState(1629);
			syntaxBracketLr();
			setState(1630);
			match(OBJECT_NAME);
			setState(1631);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Delete_Context extends ParserRuleContext {
		public KwDeleteContext kwDelete() {
			return getRuleContext(KwDeleteContext.class,0);
		}
		public FromSpecContext fromSpec() {
			return getRuleContext(FromSpecContext.class,0);
		}
		public WhereSpecContext whereSpec() {
			return getRuleContext(WhereSpecContext.class,0);
		}
		public BeginBatchContext beginBatch() {
			return getRuleContext(BeginBatchContext.class,0);
		}
		public DeleteColumnListContext deleteColumnList() {
			return getRuleContext(DeleteColumnListContext.class,0);
		}
		public UsingTimestampSpecContext usingTimestampSpec() {
			return getRuleContext(UsingTimestampSpecContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public IfSpecContext ifSpec() {
			return getRuleContext(IfSpecContext.class,0);
		}
		public Delete_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_delete_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDelete_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDelete_(this);
		}
	}

	public final Delete_Context delete_() throws RecognitionException {
		Delete_Context _localctx = new Delete_Context(_ctx, getState());
		enterRule(_localctx, 214, RULE_delete_);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1634);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_BEGIN) {
				{
				setState(1633);
				beginBatch();
				}
			}

			setState(1636);
			kwDelete();
			setState(1638);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OBJECT_NAME) {
				{
				setState(1637);
				deleteColumnList();
				}
			}

			setState(1640);
			fromSpec();
			setState(1642);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_USING) {
				{
				setState(1641);
				usingTimestampSpec();
				}
			}

			setState(1644);
			whereSpec();
			setState(1647);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,127,_ctx) ) {
			case 1:
				{
				setState(1645);
				ifExist();
				}
				break;
			case 2:
				{
				setState(1646);
				ifSpec();
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DeleteColumnListContext extends ParserRuleContext {
		public List<DeleteColumnItemContext> deleteColumnItem() {
			return getRuleContexts(DeleteColumnItemContext.class);
		}
		public DeleteColumnItemContext deleteColumnItem(int i) {
			return getRuleContext(DeleteColumnItemContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public DeleteColumnListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_deleteColumnList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDeleteColumnList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDeleteColumnList(this);
		}
	}

	public final DeleteColumnListContext deleteColumnList() throws RecognitionException {
		DeleteColumnListContext _localctx = new DeleteColumnListContext(_ctx, getState());
		enterRule(_localctx, 216, RULE_deleteColumnList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1649);
			deleteColumnItem();
			}
			setState(1655);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1650);
				syntaxComma();
				setState(1651);
				deleteColumnItem();
				}
				}
				setState(1657);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DeleteColumnItemContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TerminalNode LS_BRACKET() { return getToken(CqlParser.LS_BRACKET, 0); }
		public TerminalNode RS_BRACKET() { return getToken(CqlParser.RS_BRACKET, 0); }
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public DeleteColumnItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_deleteColumnItem; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDeleteColumnItem(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDeleteColumnItem(this);
		}
	}

	public final DeleteColumnItemContext deleteColumnItem() throws RecognitionException {
		DeleteColumnItemContext _localctx = new DeleteColumnItemContext(_ctx, getState());
		enterRule(_localctx, 218, RULE_deleteColumnItem);
		try {
			setState(1667);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,130,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1658);
				match(OBJECT_NAME);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1659);
				match(OBJECT_NAME);
				setState(1660);
				match(LS_BRACKET);
				setState(1663);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case STRING_LITERAL:
					{
					setState(1661);
					stringLiteral();
					}
					break;
				case DECIMAL_LITERAL:
					{
					setState(1662);
					decimalLiteral();
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				setState(1665);
				match(RS_BRACKET);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UpdateContext extends ParserRuleContext {
		public KwUpdateContext kwUpdate() {
			return getRuleContext(KwUpdateContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public KwSetContext kwSet() {
			return getRuleContext(KwSetContext.class,0);
		}
		public AssignmentsContext assignments() {
			return getRuleContext(AssignmentsContext.class,0);
		}
		public WhereSpecContext whereSpec() {
			return getRuleContext(WhereSpecContext.class,0);
		}
		public BeginBatchContext beginBatch() {
			return getRuleContext(BeginBatchContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public UsingTtlTimestampContext usingTtlTimestamp() {
			return getRuleContext(UsingTtlTimestampContext.class,0);
		}
		public IfExistContext ifExist() {
			return getRuleContext(IfExistContext.class,0);
		}
		public IfSpecContext ifSpec() {
			return getRuleContext(IfSpecContext.class,0);
		}
		public UpdateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_update; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUpdate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUpdate(this);
		}
	}

	public final UpdateContext update() throws RecognitionException {
		UpdateContext _localctx = new UpdateContext(_ctx, getState());
		enterRule(_localctx, 220, RULE_update);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1670);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_BEGIN) {
				{
				setState(1669);
				beginBatch();
				}
			}

			setState(1672);
			kwUpdate();
			setState(1676);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,132,_ctx) ) {
			case 1:
				{
				setState(1673);
				keyspace();
				setState(1674);
				match(DOT);
				}
				break;
			}
			setState(1678);
			table();
			setState(1680);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_USING) {
				{
				setState(1679);
				usingTtlTimestamp();
				}
			}

			setState(1682);
			kwSet();
			setState(1683);
			assignments();
			setState(1684);
			whereSpec();
			setState(1687);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,134,_ctx) ) {
			case 1:
				{
				setState(1685);
				ifExist();
				}
				break;
			case 2:
				{
				setState(1686);
				ifSpec();
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IfSpecContext extends ParserRuleContext {
		public KwIfContext kwIf() {
			return getRuleContext(KwIfContext.class,0);
		}
		public IfConditionListContext ifConditionList() {
			return getRuleContext(IfConditionListContext.class,0);
		}
		public IfSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ifSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIfSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIfSpec(this);
		}
	}

	public final IfSpecContext ifSpec() throws RecognitionException {
		IfSpecContext _localctx = new IfSpecContext(_ctx, getState());
		enterRule(_localctx, 222, RULE_ifSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1689);
			kwIf();
			setState(1690);
			ifConditionList();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IfConditionListContext extends ParserRuleContext {
		public List<IfConditionContext> ifCondition() {
			return getRuleContexts(IfConditionContext.class);
		}
		public IfConditionContext ifCondition(int i) {
			return getRuleContext(IfConditionContext.class,i);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public IfConditionListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ifConditionList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIfConditionList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIfConditionList(this);
		}
	}

	public final IfConditionListContext ifConditionList() throws RecognitionException {
		IfConditionListContext _localctx = new IfConditionListContext(_ctx, getState());
		enterRule(_localctx, 224, RULE_ifConditionList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1692);
			ifCondition();
			}
			setState(1698);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==K_AND) {
				{
				{
				setState(1693);
				kwAnd();
				setState(1694);
				ifCondition();
				}
				}
				setState(1700);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IfConditionContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public IfConditionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ifCondition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIfCondition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIfCondition(this);
		}
	}

	public final IfConditionContext ifCondition() throws RecognitionException {
		IfConditionContext _localctx = new IfConditionContext(_ctx, getState());
		enterRule(_localctx, 226, RULE_ifCondition);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1701);
			match(OBJECT_NAME);
			setState(1702);
			match(OPERATOR_EQ);
			setState(1703);
			constant();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentsContext extends ParserRuleContext {
		public List<AssignmentElementContext> assignmentElement() {
			return getRuleContexts(AssignmentElementContext.class);
		}
		public AssignmentElementContext assignmentElement(int i) {
			return getRuleContext(AssignmentElementContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AssignmentsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignments; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignments(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignments(this);
		}
	}

	public final AssignmentsContext assignments() throws RecognitionException {
		AssignmentsContext _localctx = new AssignmentsContext(_ctx, getState());
		enterRule(_localctx, 228, RULE_assignments);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1705);
			assignmentElement();
			}
			setState(1711);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1706);
				syntaxComma();
				setState(1707);
				assignmentElement();
				}
				}
				setState(1713);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentElementContext extends ParserRuleContext {
		public List<TerminalNode> OBJECT_NAME() { return getTokens(CqlParser.OBJECT_NAME); }
		public TerminalNode OBJECT_NAME(int i) {
			return getToken(CqlParser.OBJECT_NAME, i);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public AssignmentMapContext assignmentMap() {
			return getRuleContext(AssignmentMapContext.class,0);
		}
		public AssignmentSetContext assignmentSet() {
			return getRuleContext(AssignmentSetContext.class,0);
		}
		public AssignmentListContext assignmentList() {
			return getRuleContext(AssignmentListContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public TerminalNode PLUS() { return getToken(CqlParser.PLUS, 0); }
		public TerminalNode MINUS() { return getToken(CqlParser.MINUS, 0); }
		public SyntaxBracketLsContext syntaxBracketLs() {
			return getRuleContext(SyntaxBracketLsContext.class,0);
		}
		public SyntaxBracketRsContext syntaxBracketRs() {
			return getRuleContext(SyntaxBracketRsContext.class,0);
		}
		public AssignmentElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignmentElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignmentElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignmentElement(this);
		}
	}

	public final AssignmentElementContext assignmentElement() throws RecognitionException {
		AssignmentElementContext _localctx = new AssignmentElementContext(_ctx, getState());
		enterRule(_localctx, 230, RULE_assignmentElement);
		int _la;
		try {
			setState(1767);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,138,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1714);
				match(OBJECT_NAME);
				setState(1715);
				match(OPERATOR_EQ);
				setState(1720);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,137,_ctx) ) {
				case 1:
					{
					setState(1716);
					constant();
					}
					break;
				case 2:
					{
					setState(1717);
					assignmentMap();
					}
					break;
				case 3:
					{
					setState(1718);
					assignmentSet();
					}
					break;
				case 4:
					{
					setState(1719);
					assignmentList();
					}
					break;
				}
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1722);
				match(OBJECT_NAME);
				setState(1723);
				match(OPERATOR_EQ);
				setState(1724);
				match(OBJECT_NAME);
				setState(1725);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1726);
				decimalLiteral();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1727);
				match(OBJECT_NAME);
				setState(1728);
				match(OPERATOR_EQ);
				setState(1729);
				match(OBJECT_NAME);
				setState(1730);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1731);
				assignmentSet();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(1732);
				match(OBJECT_NAME);
				setState(1733);
				match(OPERATOR_EQ);
				setState(1734);
				assignmentSet();
				setState(1735);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1736);
				match(OBJECT_NAME);
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(1738);
				match(OBJECT_NAME);
				setState(1739);
				match(OPERATOR_EQ);
				setState(1740);
				match(OBJECT_NAME);
				setState(1741);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1742);
				assignmentMap();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(1743);
				match(OBJECT_NAME);
				setState(1744);
				match(OPERATOR_EQ);
				setState(1745);
				assignmentMap();
				setState(1746);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1747);
				match(OBJECT_NAME);
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(1749);
				match(OBJECT_NAME);
				setState(1750);
				match(OPERATOR_EQ);
				setState(1751);
				match(OBJECT_NAME);
				setState(1752);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1753);
				assignmentList();
				}
				break;
			case 8:
				enterOuterAlt(_localctx, 8);
				{
				setState(1754);
				match(OBJECT_NAME);
				setState(1755);
				match(OPERATOR_EQ);
				setState(1756);
				assignmentList();
				setState(1757);
				_la = _input.LA(1);
				if ( !(_la==PLUS || _la==MINUS) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(1758);
				match(OBJECT_NAME);
				}
				break;
			case 9:
				enterOuterAlt(_localctx, 9);
				{
				setState(1760);
				match(OBJECT_NAME);
				setState(1761);
				syntaxBracketLs();
				setState(1762);
				decimalLiteral();
				setState(1763);
				syntaxBracketRs();
				setState(1764);
				match(OPERATOR_EQ);
				setState(1765);
				constant();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentSetContext extends ParserRuleContext {
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AssignmentSetContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignmentSet; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignmentSet(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignmentSet(this);
		}
	}

	public final AssignmentSetContext assignmentSet() throws RecognitionException {
		AssignmentSetContext _localctx = new AssignmentSetContext(_ctx, getState());
		enterRule(_localctx, 232, RULE_assignmentSet);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1769);
			syntaxBracketLc();
			setState(1779);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_FALSE || _la==K_NULL || ((((_la - 133)) & ~0x3f) == 0 && ((1L << (_la - 133)) & 87411174408193L) != 0)) {
				{
				setState(1770);
				constant();
				setState(1776);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(1771);
					syntaxComma();
					setState(1772);
					constant();
					}
					}
					setState(1778);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
			}

			setState(1781);
			syntaxBracketRc();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentMapContext extends ParserRuleContext {
		public SyntaxBracketLcContext syntaxBracketLc() {
			return getRuleContext(SyntaxBracketLcContext.class,0);
		}
		public SyntaxBracketRcContext syntaxBracketRc() {
			return getRuleContext(SyntaxBracketRcContext.class,0);
		}
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public List<SyntaxColonContext> syntaxColon() {
			return getRuleContexts(SyntaxColonContext.class);
		}
		public SyntaxColonContext syntaxColon(int i) {
			return getRuleContext(SyntaxColonContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AssignmentMapContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignmentMap; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignmentMap(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignmentMap(this);
		}
	}

	public final AssignmentMapContext assignmentMap() throws RecognitionException {
		AssignmentMapContext _localctx = new AssignmentMapContext(_ctx, getState());
		enterRule(_localctx, 234, RULE_assignmentMap);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1783);
			syntaxBracketLc();
			{
			setState(1784);
			constant();
			setState(1785);
			syntaxColon();
			setState(1786);
			constant();
			}
			setState(1795);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1788);
				syntaxComma();
				setState(1789);
				constant();
				setState(1790);
				syntaxColon();
				setState(1791);
				constant();
				}
				}
				setState(1797);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1798);
			syntaxBracketRc();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentListContext extends ParserRuleContext {
		public SyntaxBracketLsContext syntaxBracketLs() {
			return getRuleContext(SyntaxBracketLsContext.class,0);
		}
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public SyntaxBracketRsContext syntaxBracketRs() {
			return getRuleContext(SyntaxBracketRsContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AssignmentListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignmentList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignmentList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignmentList(this);
		}
	}

	public final AssignmentListContext assignmentList() throws RecognitionException {
		AssignmentListContext _localctx = new AssignmentListContext(_ctx, getState());
		enterRule(_localctx, 236, RULE_assignmentList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1800);
			syntaxBracketLs();
			setState(1801);
			constant();
			setState(1807);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1802);
				syntaxComma();
				setState(1803);
				constant();
				}
				}
				setState(1809);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(1810);
			syntaxBracketRs();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AssignmentTupleContext extends ParserRuleContext {
		public SyntaxBracketLrContext syntaxBracketLr() {
			return getRuleContext(SyntaxBracketLrContext.class,0);
		}
		public SyntaxBracketRrContext syntaxBracketRr() {
			return getRuleContext(SyntaxBracketRrContext.class,0);
		}
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public AssignmentTupleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_assignmentTuple; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAssignmentTuple(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAssignmentTuple(this);
		}
	}

	public final AssignmentTupleContext assignmentTuple() throws RecognitionException {
		AssignmentTupleContext _localctx = new AssignmentTupleContext(_ctx, getState());
		enterRule(_localctx, 238, RULE_assignmentTuple);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1812);
			syntaxBracketLr();
			{
			setState(1813);
			expression();
			setState(1819);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1814);
				syntaxComma();
				setState(1815);
				expression();
				}
				}
				setState(1821);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
			setState(1822);
			syntaxBracketRr();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InsertContext extends ParserRuleContext {
		public KwInsertContext kwInsert() {
			return getRuleContext(KwInsertContext.class,0);
		}
		public KwIntoContext kwInto() {
			return getRuleContext(KwIntoContext.class,0);
		}
		public TableContext table() {
			return getRuleContext(TableContext.class,0);
		}
		public InsertValuesSpecContext insertValuesSpec() {
			return getRuleContext(InsertValuesSpecContext.class,0);
		}
		public BeginBatchContext beginBatch() {
			return getRuleContext(BeginBatchContext.class,0);
		}
		public KeyspaceContext keyspace() {
			return getRuleContext(KeyspaceContext.class,0);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public InsertColumnSpecContext insertColumnSpec() {
			return getRuleContext(InsertColumnSpecContext.class,0);
		}
		public IfNotExistContext ifNotExist() {
			return getRuleContext(IfNotExistContext.class,0);
		}
		public UsingTtlTimestampContext usingTtlTimestamp() {
			return getRuleContext(UsingTtlTimestampContext.class,0);
		}
		public InsertContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_insert; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInsert(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInsert(this);
		}
	}

	public final InsertContext insert() throws RecognitionException {
		InsertContext _localctx = new InsertContext(_ctx, getState());
		enterRule(_localctx, 240, RULE_insert);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1825);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_BEGIN) {
				{
				setState(1824);
				beginBatch();
				}
			}

			setState(1827);
			kwInsert();
			setState(1828);
			kwInto();
			setState(1832);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,145,_ctx) ) {
			case 1:
				{
				setState(1829);
				keyspace();
				setState(1830);
				match(DOT);
				}
				break;
			}
			setState(1834);
			table();
			setState(1836);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LR_BRACKET) {
				{
				setState(1835);
				insertColumnSpec();
				}
			}

			setState(1838);
			insertValuesSpec();
			setState(1840);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_IF) {
				{
				setState(1839);
				ifNotExist();
				}
			}

			setState(1843);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_USING) {
				{
				setState(1842);
				usingTtlTimestamp();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UsingTtlTimestampContext extends ParserRuleContext {
		public KwUsingContext kwUsing() {
			return getRuleContext(KwUsingContext.class,0);
		}
		public TtlContext ttl() {
			return getRuleContext(TtlContext.class,0);
		}
		public KwAndContext kwAnd() {
			return getRuleContext(KwAndContext.class,0);
		}
		public TimestampContext timestamp() {
			return getRuleContext(TimestampContext.class,0);
		}
		public UsingTtlTimestampContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_usingTtlTimestamp; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUsingTtlTimestamp(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUsingTtlTimestamp(this);
		}
	}

	public final UsingTtlTimestampContext usingTtlTimestamp() throws RecognitionException {
		UsingTtlTimestampContext _localctx = new UsingTtlTimestampContext(_ctx, getState());
		enterRule(_localctx, 242, RULE_usingTtlTimestamp);
		try {
			setState(1861);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,149,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1845);
				kwUsing();
				setState(1846);
				ttl();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1848);
				kwUsing();
				setState(1849);
				ttl();
				setState(1850);
				kwAnd();
				setState(1851);
				timestamp();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1853);
				kwUsing();
				setState(1854);
				timestamp();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(1856);
				kwUsing();
				setState(1857);
				timestamp();
				setState(1858);
				kwAnd();
				setState(1859);
				ttl();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TimestampContext extends ParserRuleContext {
		public KwTimestampContext kwTimestamp() {
			return getRuleContext(KwTimestampContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public TimestampContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_timestamp; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTimestamp(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTimestamp(this);
		}
	}

	public final TimestampContext timestamp() throws RecognitionException {
		TimestampContext _localctx = new TimestampContext(_ctx, getState());
		enterRule(_localctx, 244, RULE_timestamp);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1863);
			kwTimestamp();
			setState(1864);
			decimalLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TtlContext extends ParserRuleContext {
		public KwTtlContext kwTtl() {
			return getRuleContext(KwTtlContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public TtlContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ttl; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTtl(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTtl(this);
		}
	}

	public final TtlContext ttl() throws RecognitionException {
		TtlContext _localctx = new TtlContext(_ctx, getState());
		enterRule(_localctx, 246, RULE_ttl);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1866);
			kwTtl();
			setState(1867);
			decimalLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UsingTimestampSpecContext extends ParserRuleContext {
		public KwUsingContext kwUsing() {
			return getRuleContext(KwUsingContext.class,0);
		}
		public TimestampContext timestamp() {
			return getRuleContext(TimestampContext.class,0);
		}
		public UsingTimestampSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_usingTimestampSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUsingTimestampSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUsingTimestampSpec(this);
		}
	}

	public final UsingTimestampSpecContext usingTimestampSpec() throws RecognitionException {
		UsingTimestampSpecContext _localctx = new UsingTimestampSpecContext(_ctx, getState());
		enterRule(_localctx, 248, RULE_usingTimestampSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1869);
			kwUsing();
			setState(1870);
			timestamp();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IfNotExistContext extends ParserRuleContext {
		public KwIfContext kwIf() {
			return getRuleContext(KwIfContext.class,0);
		}
		public KwNotContext kwNot() {
			return getRuleContext(KwNotContext.class,0);
		}
		public KwExistsContext kwExists() {
			return getRuleContext(KwExistsContext.class,0);
		}
		public IfNotExistContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ifNotExist; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIfNotExist(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIfNotExist(this);
		}
	}

	public final IfNotExistContext ifNotExist() throws RecognitionException {
		IfNotExistContext _localctx = new IfNotExistContext(_ctx, getState());
		enterRule(_localctx, 250, RULE_ifNotExist);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1872);
			kwIf();
			setState(1873);
			kwNot();
			setState(1874);
			kwExists();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class IfExistContext extends ParserRuleContext {
		public KwIfContext kwIf() {
			return getRuleContext(KwIfContext.class,0);
		}
		public KwExistsContext kwExists() {
			return getRuleContext(KwExistsContext.class,0);
		}
		public IfExistContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_ifExist; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterIfExist(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitIfExist(this);
		}
	}

	public final IfExistContext ifExist() throws RecognitionException {
		IfExistContext _localctx = new IfExistContext(_ctx, getState());
		enterRule(_localctx, 252, RULE_ifExist);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1876);
			kwIf();
			setState(1877);
			kwExists();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InsertValuesSpecContext extends ParserRuleContext {
		public KwValuesContext kwValues() {
			return getRuleContext(KwValuesContext.class,0);
		}
		public TerminalNode LR_BRACKET() { return getToken(CqlParser.LR_BRACKET, 0); }
		public ExpressionListContext expressionList() {
			return getRuleContext(ExpressionListContext.class,0);
		}
		public TerminalNode RR_BRACKET() { return getToken(CqlParser.RR_BRACKET, 0); }
		public KwJsonContext kwJson() {
			return getRuleContext(KwJsonContext.class,0);
		}
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public InsertValuesSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_insertValuesSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInsertValuesSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInsertValuesSpec(this);
		}
	}

	public final InsertValuesSpecContext insertValuesSpec() throws RecognitionException {
		InsertValuesSpecContext _localctx = new InsertValuesSpecContext(_ctx, getState());
		enterRule(_localctx, 254, RULE_insertValuesSpec);
		try {
			setState(1887);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_VALUES:
				enterOuterAlt(_localctx, 1);
				{
				setState(1879);
				kwValues();
				setState(1880);
				match(LR_BRACKET);
				setState(1881);
				expressionList();
				setState(1882);
				match(RR_BRACKET);
				}
				break;
			case K_JSON:
				enterOuterAlt(_localctx, 2);
				{
				setState(1884);
				kwJson();
				setState(1885);
				constant();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class InsertColumnSpecContext extends ParserRuleContext {
		public TerminalNode LR_BRACKET() { return getToken(CqlParser.LR_BRACKET, 0); }
		public ColumnListContext columnList() {
			return getRuleContext(ColumnListContext.class,0);
		}
		public TerminalNode RR_BRACKET() { return getToken(CqlParser.RR_BRACKET, 0); }
		public InsertColumnSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_insertColumnSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterInsertColumnSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitInsertColumnSpec(this);
		}
	}

	public final InsertColumnSpecContext insertColumnSpec() throws RecognitionException {
		InsertColumnSpecContext _localctx = new InsertColumnSpecContext(_ctx, getState());
		enterRule(_localctx, 256, RULE_insertColumnSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1889);
			match(LR_BRACKET);
			setState(1890);
			columnList();
			setState(1891);
			match(RR_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnListContext extends ParserRuleContext {
		public List<ColumnContext> column() {
			return getRuleContexts(ColumnContext.class);
		}
		public ColumnContext column(int i) {
			return getRuleContext(ColumnContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public ColumnListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_columnList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumnList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumnList(this);
		}
	}

	public final ColumnListContext columnList() throws RecognitionException {
		ColumnListContext _localctx = new ColumnListContext(_ctx, getState());
		enterRule(_localctx, 258, RULE_columnList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1893);
			column();
			setState(1899);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1894);
				syntaxComma();
				setState(1895);
				column();
				}
				}
				setState(1901);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ExpressionListContext extends ParserRuleContext {
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public ExpressionListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expressionList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterExpressionList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitExpressionList(this);
		}
	}

	public final ExpressionListContext expressionList() throws RecognitionException {
		ExpressionListContext _localctx = new ExpressionListContext(_ctx, getState());
		enterRule(_localctx, 260, RULE_expressionList);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1902);
			expression();
			setState(1908);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1903);
				syntaxComma();
				setState(1904);
				expression();
				}
				}
				setState(1910);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ExpressionContext extends ParserRuleContext {
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public FunctionCallContext functionCall() {
			return getRuleContext(FunctionCallContext.class,0);
		}
		public AssignmentMapContext assignmentMap() {
			return getRuleContext(AssignmentMapContext.class,0);
		}
		public AssignmentSetContext assignmentSet() {
			return getRuleContext(AssignmentSetContext.class,0);
		}
		public AssignmentListContext assignmentList() {
			return getRuleContext(AssignmentListContext.class,0);
		}
		public AssignmentTupleContext assignmentTuple() {
			return getRuleContext(AssignmentTupleContext.class,0);
		}
		public ExpressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterExpression(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitExpression(this);
		}
	}

	public final ExpressionContext expression() throws RecognitionException {
		ExpressionContext _localctx = new ExpressionContext(_ctx, getState());
		enterRule(_localctx, 262, RULE_expression);
		try {
			setState(1917);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,153,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1911);
				constant();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1912);
				functionCall();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1913);
				assignmentMap();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(1914);
				assignmentSet();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(1915);
				assignmentList();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(1916);
				assignmentTuple();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Select_Context extends ParserRuleContext {
		public KwSelectContext kwSelect() {
			return getRuleContext(KwSelectContext.class,0);
		}
		public SelectElementsContext selectElements() {
			return getRuleContext(SelectElementsContext.class,0);
		}
		public FromSpecContext fromSpec() {
			return getRuleContext(FromSpecContext.class,0);
		}
		public DistinctSpecContext distinctSpec() {
			return getRuleContext(DistinctSpecContext.class,0);
		}
		public KwJsonContext kwJson() {
			return getRuleContext(KwJsonContext.class,0);
		}
		public WhereSpecContext whereSpec() {
			return getRuleContext(WhereSpecContext.class,0);
		}
		public OrderSpecContext orderSpec() {
			return getRuleContext(OrderSpecContext.class,0);
		}
		public LimitSpecContext limitSpec() {
			return getRuleContext(LimitSpecContext.class,0);
		}
		public AllowFilteringSpecContext allowFilteringSpec() {
			return getRuleContext(AllowFilteringSpecContext.class,0);
		}
		public Select_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_select_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSelect_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSelect_(this);
		}
	}

	public final Select_Context select_() throws RecognitionException {
		Select_Context _localctx = new Select_Context(_ctx, getState());
		enterRule(_localctx, 264, RULE_select_);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1919);
			kwSelect();
			setState(1921);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_DISTINCT) {
				{
				setState(1920);
				distinctSpec();
				}
			}

			setState(1924);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_JSON) {
				{
				setState(1923);
				kwJson();
				}
			}

			setState(1926);
			selectElements();
			setState(1927);
			fromSpec();
			setState(1929);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_WHERE) {
				{
				setState(1928);
				whereSpec();
				}
			}

			setState(1932);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_ORDER) {
				{
				setState(1931);
				orderSpec();
				}
			}

			setState(1935);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_LIMIT) {
				{
				setState(1934);
				limitSpec();
				}
			}

			setState(1938);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==K_ALLOW) {
				{
				setState(1937);
				allowFilteringSpec();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AllowFilteringSpecContext extends ParserRuleContext {
		public KwAllowContext kwAllow() {
			return getRuleContext(KwAllowContext.class,0);
		}
		public KwFilteringContext kwFiltering() {
			return getRuleContext(KwFilteringContext.class,0);
		}
		public AllowFilteringSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_allowFilteringSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAllowFilteringSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAllowFilteringSpec(this);
		}
	}

	public final AllowFilteringSpecContext allowFilteringSpec() throws RecognitionException {
		AllowFilteringSpecContext _localctx = new AllowFilteringSpecContext(_ctx, getState());
		enterRule(_localctx, 266, RULE_allowFilteringSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1940);
			kwAllow();
			setState(1941);
			kwFiltering();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LimitSpecContext extends ParserRuleContext {
		public KwLimitContext kwLimit() {
			return getRuleContext(KwLimitContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public LimitSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_limitSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterLimitSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitLimitSpec(this);
		}
	}

	public final LimitSpecContext limitSpec() throws RecognitionException {
		LimitSpecContext _localctx = new LimitSpecContext(_ctx, getState());
		enterRule(_localctx, 268, RULE_limitSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1943);
			kwLimit();
			setState(1944);
			decimalLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class FromSpecContext extends ParserRuleContext {
		public KwFromContext kwFrom() {
			return getRuleContext(KwFromContext.class,0);
		}
		public FromSpecElementContext fromSpecElement() {
			return getRuleContext(FromSpecElementContext.class,0);
		}
		public FromSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fromSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFromSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFromSpec(this);
		}
	}

	public final FromSpecContext fromSpec() throws RecognitionException {
		FromSpecContext _localctx = new FromSpecContext(_ctx, getState());
		enterRule(_localctx, 270, RULE_fromSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1946);
			kwFrom();
			setState(1947);
			fromSpecElement();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class FromSpecElementContext extends ParserRuleContext {
		public List<TerminalNode> OBJECT_NAME() { return getTokens(CqlParser.OBJECT_NAME); }
		public TerminalNode OBJECT_NAME(int i) {
			return getToken(CqlParser.OBJECT_NAME, i);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public FromSpecElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_fromSpecElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFromSpecElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFromSpecElement(this);
		}
	}

	public final FromSpecElementContext fromSpecElement() throws RecognitionException {
		FromSpecElementContext _localctx = new FromSpecElementContext(_ctx, getState());
		enterRule(_localctx, 272, RULE_fromSpecElement);
		try {
			setState(1953);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,160,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1949);
				match(OBJECT_NAME);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1950);
				match(OBJECT_NAME);
				setState(1951);
				match(DOT);
				setState(1952);
				match(OBJECT_NAME);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrderSpecContext extends ParserRuleContext {
		public KwOrderContext kwOrder() {
			return getRuleContext(KwOrderContext.class,0);
		}
		public KwByContext kwBy() {
			return getRuleContext(KwByContext.class,0);
		}
		public OrderSpecElementContext orderSpecElement() {
			return getRuleContext(OrderSpecElementContext.class,0);
		}
		public OrderSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orderSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOrderSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOrderSpec(this);
		}
	}

	public final OrderSpecContext orderSpec() throws RecognitionException {
		OrderSpecContext _localctx = new OrderSpecContext(_ctx, getState());
		enterRule(_localctx, 274, RULE_orderSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1955);
			kwOrder();
			setState(1956);
			kwBy();
			setState(1957);
			orderSpecElement();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrderSpecElementContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public KwAscContext kwAsc() {
			return getRuleContext(KwAscContext.class,0);
		}
		public KwDescContext kwDesc() {
			return getRuleContext(KwDescContext.class,0);
		}
		public OrderSpecElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orderSpecElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOrderSpecElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOrderSpecElement(this);
		}
	}

	public final OrderSpecElementContext orderSpecElement() throws RecognitionException {
		OrderSpecElementContext _localctx = new OrderSpecElementContext(_ctx, getState());
		enterRule(_localctx, 276, RULE_orderSpecElement);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1959);
			match(OBJECT_NAME);
			setState(1962);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_ASC:
				{
				setState(1960);
				kwAsc();
				}
				break;
			case K_DESC:
				{
				setState(1961);
				kwDesc();
				}
				break;
			case EOF:
			case SEMI:
			case MINUSMINUS:
			case K_ALLOW:
			case K_LIMIT:
				break;
			default:
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class WhereSpecContext extends ParserRuleContext {
		public KwWhereContext kwWhere() {
			return getRuleContext(KwWhereContext.class,0);
		}
		public RelationElementsContext relationElements() {
			return getRuleContext(RelationElementsContext.class,0);
		}
		public WhereSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_whereSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterWhereSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitWhereSpec(this);
		}
	}

	public final WhereSpecContext whereSpec() throws RecognitionException {
		WhereSpecContext _localctx = new WhereSpecContext(_ctx, getState());
		enterRule(_localctx, 278, RULE_whereSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1964);
			kwWhere();
			setState(1965);
			relationElements();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DistinctSpecContext extends ParserRuleContext {
		public KwDistinctContext kwDistinct() {
			return getRuleContext(KwDistinctContext.class,0);
		}
		public DistinctSpecContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_distinctSpec; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDistinctSpec(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDistinctSpec(this);
		}
	}

	public final DistinctSpecContext distinctSpec() throws RecognitionException {
		DistinctSpecContext _localctx = new DistinctSpecContext(_ctx, getState());
		enterRule(_localctx, 280, RULE_distinctSpec);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1967);
			kwDistinct();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SelectElementsContext extends ParserRuleContext {
		public Token star;
		public List<SelectElementContext> selectElement() {
			return getRuleContexts(SelectElementContext.class);
		}
		public SelectElementContext selectElement(int i) {
			return getRuleContext(SelectElementContext.class,i);
		}
		public TerminalNode STAR() { return getToken(CqlParser.STAR, 0); }
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public SelectElementsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_selectElements; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSelectElements(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSelectElements(this);
		}
	}

	public final SelectElementsContext selectElements() throws RecognitionException {
		SelectElementsContext _localctx = new SelectElementsContext(_ctx, getState());
		enterRule(_localctx, 282, RULE_selectElements);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(1971);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STAR:
				{
				setState(1969);
				((SelectElementsContext)_localctx).star = match(STAR);
				}
				break;
			case K_UUID:
			case OBJECT_NAME:
				{
				setState(1970);
				selectElement();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(1978);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(1973);
				syntaxComma();
				setState(1974);
				selectElement();
				}
				}
				setState(1980);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SelectElementContext extends ParserRuleContext {
		public List<TerminalNode> OBJECT_NAME() { return getTokens(CqlParser.OBJECT_NAME); }
		public TerminalNode OBJECT_NAME(int i) {
			return getToken(CqlParser.OBJECT_NAME, i);
		}
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public TerminalNode STAR() { return getToken(CqlParser.STAR, 0); }
		public KwAsContext kwAs() {
			return getRuleContext(KwAsContext.class,0);
		}
		public FunctionCallContext functionCall() {
			return getRuleContext(FunctionCallContext.class,0);
		}
		public SelectElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_selectElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSelectElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSelectElement(this);
		}
	}

	public final SelectElementContext selectElement() throws RecognitionException {
		SelectElementContext _localctx = new SelectElementContext(_ctx, getState());
		enterRule(_localctx, 284, RULE_selectElement);
		int _la;
		try {
			setState(1996);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,166,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(1981);
				match(OBJECT_NAME);
				setState(1982);
				match(DOT);
				setState(1983);
				match(STAR);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(1984);
				match(OBJECT_NAME);
				setState(1988);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_AS) {
					{
					setState(1985);
					kwAs();
					setState(1986);
					match(OBJECT_NAME);
					}
				}

				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(1990);
				functionCall();
				setState(1994);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_AS) {
					{
					setState(1991);
					kwAs();
					setState(1992);
					match(OBJECT_NAME);
					}
				}

				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelationElementsContext extends ParserRuleContext {
		public List<RelationElementContext> relationElement() {
			return getRuleContexts(RelationElementContext.class);
		}
		public RelationElementContext relationElement(int i) {
			return getRuleContext(RelationElementContext.class,i);
		}
		public List<KwAndContext> kwAnd() {
			return getRuleContexts(KwAndContext.class);
		}
		public KwAndContext kwAnd(int i) {
			return getRuleContext(KwAndContext.class,i);
		}
		public RelationElementsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relationElements; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRelationElements(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRelationElements(this);
		}
	}

	public final RelationElementsContext relationElements() throws RecognitionException {
		RelationElementsContext _localctx = new RelationElementsContext(_ctx, getState());
		enterRule(_localctx, 286, RULE_relationElements);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(1998);
			relationElement();
			}
			setState(2004);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==K_AND) {
				{
				{
				setState(1999);
				kwAnd();
				setState(2000);
				relationElement();
				}
				}
				setState(2006);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelationElementContext extends ParserRuleContext {
		public List<TerminalNode> OBJECT_NAME() { return getTokens(CqlParser.OBJECT_NAME); }
		public TerminalNode OBJECT_NAME(int i) {
			return getToken(CqlParser.OBJECT_NAME, i);
		}
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public TerminalNode OPERATOR_EQ() { return getToken(CqlParser.OPERATOR_EQ, 0); }
		public TerminalNode OPERATOR_LT() { return getToken(CqlParser.OPERATOR_LT, 0); }
		public TerminalNode OPERATOR_GT() { return getToken(CqlParser.OPERATOR_GT, 0); }
		public TerminalNode OPERATOR_LTE() { return getToken(CqlParser.OPERATOR_LTE, 0); }
		public TerminalNode OPERATOR_GTE() { return getToken(CqlParser.OPERATOR_GTE, 0); }
		public TerminalNode DOT() { return getToken(CqlParser.DOT, 0); }
		public List<FunctionCallContext> functionCall() {
			return getRuleContexts(FunctionCallContext.class);
		}
		public FunctionCallContext functionCall(int i) {
			return getRuleContext(FunctionCallContext.class,i);
		}
		public KwInContext kwIn() {
			return getRuleContext(KwInContext.class,0);
		}
		public List<TerminalNode> LR_BRACKET() { return getTokens(CqlParser.LR_BRACKET); }
		public TerminalNode LR_BRACKET(int i) {
			return getToken(CqlParser.LR_BRACKET, i);
		}
		public List<TerminalNode> RR_BRACKET() { return getTokens(CqlParser.RR_BRACKET); }
		public TerminalNode RR_BRACKET(int i) {
			return getToken(CqlParser.RR_BRACKET, i);
		}
		public FunctionArgsContext functionArgs() {
			return getRuleContext(FunctionArgsContext.class,0);
		}
		public List<AssignmentTupleContext> assignmentTuple() {
			return getRuleContexts(AssignmentTupleContext.class);
		}
		public AssignmentTupleContext assignmentTuple(int i) {
			return getRuleContext(AssignmentTupleContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public RelalationContainsKeyContext relalationContainsKey() {
			return getRuleContext(RelalationContainsKeyContext.class,0);
		}
		public RelalationContainsContext relalationContains() {
			return getRuleContext(RelalationContainsContext.class,0);
		}
		public RelationElementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relationElement; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRelationElement(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRelationElement(this);
		}
	}

	public final RelationElementContext relationElement() throws RecognitionException {
		RelationElementContext _localctx = new RelationElementContext(_ctx, getState());
		enterRule(_localctx, 288, RULE_relationElement);
		int _la;
		try {
			setState(2078);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,173,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(2007);
				match(OBJECT_NAME);
				setState(2008);
				_la = _input.LA(1);
				if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16252928L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(2009);
				constant();
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(2010);
				match(OBJECT_NAME);
				setState(2011);
				match(DOT);
				setState(2012);
				match(OBJECT_NAME);
				setState(2013);
				_la = _input.LA(1);
				if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16252928L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(2014);
				constant();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(2015);
				functionCall();
				setState(2016);
				_la = _input.LA(1);
				if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16252928L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(2017);
				constant();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(2019);
				functionCall();
				setState(2020);
				_la = _input.LA(1);
				if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16252928L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				setState(2021);
				functionCall();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(2023);
				match(OBJECT_NAME);
				setState(2024);
				kwIn();
				setState(2025);
				match(LR_BRACKET);
				setState(2027);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_FALSE || _la==K_NULL || ((((_la - 133)) & ~0x3f) == 0 && ((1L << (_la - 133)) & 122595546499073L) != 0)) {
					{
					setState(2026);
					functionArgs();
					}
				}

				setState(2029);
				match(RR_BRACKET);
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(2031);
				match(LR_BRACKET);
				setState(2032);
				match(OBJECT_NAME);
				setState(2038);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(2033);
					syntaxComma();
					setState(2034);
					match(OBJECT_NAME);
					}
					}
					setState(2040);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(2041);
				match(RR_BRACKET);
				setState(2042);
				kwIn();
				setState(2043);
				match(LR_BRACKET);
				setState(2044);
				assignmentTuple();
				setState(2050);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(2045);
					syntaxComma();
					setState(2046);
					assignmentTuple();
					}
					}
					setState(2052);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(2053);
				match(RR_BRACKET);
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(2055);
				match(LR_BRACKET);
				setState(2056);
				match(OBJECT_NAME);
				setState(2062);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(2057);
					syntaxComma();
					setState(2058);
					match(OBJECT_NAME);
					}
					}
					setState(2064);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(2065);
				match(RR_BRACKET);
				setState(2066);
				_la = _input.LA(1);
				if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 16252928L) != 0)) ) {
				_errHandler.recoverInline(this);
				}
				else {
					if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
					_errHandler.reportMatch(this);
					consume();
				}
				{
				setState(2067);
				assignmentTuple();
				setState(2073);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(2068);
					syntaxComma();
					setState(2069);
					assignmentTuple();
					}
					}
					setState(2075);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
				}
				break;
			case 8:
				enterOuterAlt(_localctx, 8);
				{
				setState(2076);
				relalationContainsKey();
				}
				break;
			case 9:
				enterOuterAlt(_localctx, 9);
				{
				setState(2077);
				relalationContains();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelalationContainsContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public KwContainsContext kwContains() {
			return getRuleContext(KwContainsContext.class,0);
		}
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public RelalationContainsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relalationContains; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRelalationContains(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRelalationContains(this);
		}
	}

	public final RelalationContainsContext relalationContains() throws RecognitionException {
		RelalationContainsContext _localctx = new RelalationContainsContext(_ctx, getState());
		enterRule(_localctx, 290, RULE_relalationContains);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2080);
			match(OBJECT_NAME);
			setState(2081);
			kwContains();
			setState(2082);
			constant();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelalationContainsKeyContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public ConstantContext constant() {
			return getRuleContext(ConstantContext.class,0);
		}
		public KwContainsContext kwContains() {
			return getRuleContext(KwContainsContext.class,0);
		}
		public KwKeyContext kwKey() {
			return getRuleContext(KwKeyContext.class,0);
		}
		public RelalationContainsKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relalationContainsKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRelalationContainsKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRelalationContainsKey(this);
		}
	}

	public final RelalationContainsKeyContext relalationContainsKey() throws RecognitionException {
		RelalationContainsKeyContext _localctx = new RelalationContainsKeyContext(_ctx, getState());
		enterRule(_localctx, 292, RULE_relalationContainsKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2084);
			match(OBJECT_NAME);
			{
			setState(2085);
			kwContains();
			setState(2086);
			kwKey();
			}
			setState(2088);
			constant();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class FunctionCallContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TerminalNode LR_BRACKET() { return getToken(CqlParser.LR_BRACKET, 0); }
		public TerminalNode STAR() { return getToken(CqlParser.STAR, 0); }
		public TerminalNode RR_BRACKET() { return getToken(CqlParser.RR_BRACKET, 0); }
		public FunctionArgsContext functionArgs() {
			return getRuleContext(FunctionArgsContext.class,0);
		}
		public TerminalNode K_UUID() { return getToken(CqlParser.K_UUID, 0); }
		public FunctionCallContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_functionCall; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFunctionCall(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFunctionCall(this);
		}
	}

	public final FunctionCallContext functionCall() throws RecognitionException {
		FunctionCallContext _localctx = new FunctionCallContext(_ctx, getState());
		enterRule(_localctx, 294, RULE_functionCall);
		int _la;
		try {
			setState(2103);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,175,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(2090);
				match(OBJECT_NAME);
				setState(2091);
				match(LR_BRACKET);
				setState(2092);
				match(STAR);
				setState(2093);
				match(RR_BRACKET);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(2094);
				match(OBJECT_NAME);
				setState(2095);
				match(LR_BRACKET);
				setState(2097);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==K_FALSE || _la==K_NULL || ((((_la - 133)) & ~0x3f) == 0 && ((1L << (_la - 133)) & 122595546499073L) != 0)) {
					{
					setState(2096);
					functionArgs();
					}
				}

				setState(2099);
				match(RR_BRACKET);
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(2100);
				match(K_UUID);
				setState(2101);
				match(LR_BRACKET);
				setState(2102);
				match(RR_BRACKET);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class FunctionArgsContext extends ParserRuleContext {
		public List<ConstantContext> constant() {
			return getRuleContexts(ConstantContext.class);
		}
		public ConstantContext constant(int i) {
			return getRuleContext(ConstantContext.class,i);
		}
		public List<TerminalNode> OBJECT_NAME() { return getTokens(CqlParser.OBJECT_NAME); }
		public TerminalNode OBJECT_NAME(int i) {
			return getToken(CqlParser.OBJECT_NAME, i);
		}
		public List<FunctionCallContext> functionCall() {
			return getRuleContexts(FunctionCallContext.class);
		}
		public FunctionCallContext functionCall(int i) {
			return getRuleContext(FunctionCallContext.class,i);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public FunctionArgsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_functionArgs; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFunctionArgs(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFunctionArgs(this);
		}
	}

	public final FunctionArgsContext functionArgs() throws RecognitionException {
		FunctionArgsContext _localctx = new FunctionArgsContext(_ctx, getState());
		enterRule(_localctx, 296, RULE_functionArgs);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2108);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,176,_ctx) ) {
			case 1:
				{
				setState(2105);
				constant();
				}
				break;
			case 2:
				{
				setState(2106);
				match(OBJECT_NAME);
				}
				break;
			case 3:
				{
				setState(2107);
				functionCall();
				}
				break;
			}
			setState(2118);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(2110);
				syntaxComma();
				setState(2114);
				_errHandler.sync(this);
				switch ( getInterpreter().adaptivePredict(_input,177,_ctx) ) {
				case 1:
					{
					setState(2111);
					constant();
					}
					break;
				case 2:
					{
					setState(2112);
					match(OBJECT_NAME);
					}
					break;
				case 3:
					{
					setState(2113);
					functionCall();
					}
					break;
				}
				}
				}
				setState(2120);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ConstantContext extends ParserRuleContext {
		public TerminalNode UUID() { return getToken(CqlParser.UUID, 0); }
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public DecimalLiteralContext decimalLiteral() {
			return getRuleContext(DecimalLiteralContext.class,0);
		}
		public FloatLiteralContext floatLiteral() {
			return getRuleContext(FloatLiteralContext.class,0);
		}
		public HexadecimalLiteralContext hexadecimalLiteral() {
			return getRuleContext(HexadecimalLiteralContext.class,0);
		}
		public BooleanLiteralContext booleanLiteral() {
			return getRuleContext(BooleanLiteralContext.class,0);
		}
		public CodeBlockContext codeBlock() {
			return getRuleContext(CodeBlockContext.class,0);
		}
		public KwNullContext kwNull() {
			return getRuleContext(KwNullContext.class,0);
		}
		public ConstantContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_constant; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterConstant(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitConstant(this);
		}
	}

	public final ConstantContext constant() throws RecognitionException {
		ConstantContext _localctx = new ConstantContext(_ctx, getState());
		enterRule(_localctx, 298, RULE_constant);
		try {
			setState(2129);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,179,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(2121);
				match(UUID);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(2122);
				stringLiteral();
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(2123);
				decimalLiteral();
				}
				break;
			case 4:
				enterOuterAlt(_localctx, 4);
				{
				setState(2124);
				floatLiteral();
				}
				break;
			case 5:
				enterOuterAlt(_localctx, 5);
				{
				setState(2125);
				hexadecimalLiteral();
				}
				break;
			case 6:
				enterOuterAlt(_localctx, 6);
				{
				setState(2126);
				booleanLiteral();
				}
				break;
			case 7:
				enterOuterAlt(_localctx, 7);
				{
				setState(2127);
				codeBlock();
				}
				break;
			case 8:
				enterOuterAlt(_localctx, 8);
				{
				setState(2128);
				kwNull();
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DecimalLiteralContext extends ParserRuleContext {
		public TerminalNode DECIMAL_LITERAL() { return getToken(CqlParser.DECIMAL_LITERAL, 0); }
		public DecimalLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_decimalLiteral; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDecimalLiteral(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDecimalLiteral(this);
		}
	}

	public final DecimalLiteralContext decimalLiteral() throws RecognitionException {
		DecimalLiteralContext _localctx = new DecimalLiteralContext(_ctx, getState());
		enterRule(_localctx, 300, RULE_decimalLiteral);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2131);
			match(DECIMAL_LITERAL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class FloatLiteralContext extends ParserRuleContext {
		public TerminalNode DECIMAL_LITERAL() { return getToken(CqlParser.DECIMAL_LITERAL, 0); }
		public TerminalNode FLOAT_LITERAL() { return getToken(CqlParser.FLOAT_LITERAL, 0); }
		public FloatLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_floatLiteral; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFloatLiteral(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFloatLiteral(this);
		}
	}

	public final FloatLiteralContext floatLiteral() throws RecognitionException {
		FloatLiteralContext _localctx = new FloatLiteralContext(_ctx, getState());
		enterRule(_localctx, 302, RULE_floatLiteral);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2133);
			_la = _input.LA(1);
			if ( !(_la==DECIMAL_LITERAL || _la==FLOAT_LITERAL) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StringLiteralContext extends ParserRuleContext {
		public TerminalNode STRING_LITERAL() { return getToken(CqlParser.STRING_LITERAL, 0); }
		public StringLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_stringLiteral; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterStringLiteral(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitStringLiteral(this);
		}
	}

	public final StringLiteralContext stringLiteral() throws RecognitionException {
		StringLiteralContext _localctx = new StringLiteralContext(_ctx, getState());
		enterRule(_localctx, 304, RULE_stringLiteral);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2135);
			match(STRING_LITERAL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class BooleanLiteralContext extends ParserRuleContext {
		public TerminalNode K_TRUE() { return getToken(CqlParser.K_TRUE, 0); }
		public TerminalNode K_FALSE() { return getToken(CqlParser.K_FALSE, 0); }
		public BooleanLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_booleanLiteral; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterBooleanLiteral(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitBooleanLiteral(this);
		}
	}

	public final BooleanLiteralContext booleanLiteral() throws RecognitionException {
		BooleanLiteralContext _localctx = new BooleanLiteralContext(_ctx, getState());
		enterRule(_localctx, 306, RULE_booleanLiteral);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2137);
			_la = _input.LA(1);
			if ( !(_la==K_FALSE || _la==K_TRUE) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class HexadecimalLiteralContext extends ParserRuleContext {
		public TerminalNode HEXADECIMAL_LITERAL() { return getToken(CqlParser.HEXADECIMAL_LITERAL, 0); }
		public HexadecimalLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_hexadecimalLiteral; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterHexadecimalLiteral(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitHexadecimalLiteral(this);
		}
	}

	public final HexadecimalLiteralContext hexadecimalLiteral() throws RecognitionException {
		HexadecimalLiteralContext _localctx = new HexadecimalLiteralContext(_ctx, getState());
		enterRule(_localctx, 308, RULE_hexadecimalLiteral);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2139);
			match(HEXADECIMAL_LITERAL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KeyspaceContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public List<TerminalNode> DQUOTE() { return getTokens(CqlParser.DQUOTE); }
		public TerminalNode DQUOTE(int i) {
			return getToken(CqlParser.DQUOTE, i);
		}
		public KeyspaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_keyspace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKeyspace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKeyspace(this);
		}
	}

	public final KeyspaceContext keyspace() throws RecognitionException {
		KeyspaceContext _localctx = new KeyspaceContext(_ctx, getState());
		enterRule(_localctx, 310, RULE_keyspace);
		try {
			setState(2145);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 1);
				{
				setState(2141);
				match(OBJECT_NAME);
				}
				break;
			case DQUOTE:
				enterOuterAlt(_localctx, 2);
				{
				setState(2142);
				match(DQUOTE);
				setState(2143);
				match(OBJECT_NAME);
				setState(2144);
				match(DQUOTE);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TableContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public List<TerminalNode> DQUOTE() { return getTokens(CqlParser.DQUOTE); }
		public TerminalNode DQUOTE(int i) {
			return getToken(CqlParser.DQUOTE, i);
		}
		public TableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_table; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTable(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTable(this);
		}
	}

	public final TableContext table() throws RecognitionException {
		TableContext _localctx = new TableContext(_ctx, getState());
		enterRule(_localctx, 312, RULE_table);
		try {
			setState(2151);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 1);
				{
				setState(2147);
				match(OBJECT_NAME);
				}
				break;
			case DQUOTE:
				enterOuterAlt(_localctx, 2);
				{
				setState(2148);
				match(DQUOTE);
				setState(2149);
				match(OBJECT_NAME);
				setState(2150);
				match(DQUOTE);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ColumnContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public List<TerminalNode> DQUOTE() { return getTokens(CqlParser.DQUOTE); }
		public TerminalNode DQUOTE(int i) {
			return getToken(CqlParser.DQUOTE, i);
		}
		public ColumnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_column; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterColumn(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitColumn(this);
		}
	}

	public final ColumnContext column() throws RecognitionException {
		ColumnContext _localctx = new ColumnContext(_ctx, getState());
		enterRule(_localctx, 314, RULE_column);
		try {
			setState(2157);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case OBJECT_NAME:
				enterOuterAlt(_localctx, 1);
				{
				setState(2153);
				match(OBJECT_NAME);
				}
				break;
			case DQUOTE:
				enterOuterAlt(_localctx, 2);
				{
				setState(2154);
				match(DQUOTE);
				setState(2155);
				match(OBJECT_NAME);
				setState(2156);
				match(DQUOTE);
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DataTypeContext extends ParserRuleContext {
		public DataTypeNameContext dataTypeName() {
			return getRuleContext(DataTypeNameContext.class,0);
		}
		public DataTypeDefinitionContext dataTypeDefinition() {
			return getRuleContext(DataTypeDefinitionContext.class,0);
		}
		public DataTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dataType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDataType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDataType(this);
		}
	}

	public final DataTypeContext dataType() throws RecognitionException {
		DataTypeContext _localctx = new DataTypeContext(_ctx, getState());
		enterRule(_localctx, 316, RULE_dataType);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2159);
			dataTypeName();
			setState(2161);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==OPERATOR_LT) {
				{
				setState(2160);
				dataTypeDefinition();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DataTypeNameContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TerminalNode K_TIMESTAMP() { return getToken(CqlParser.K_TIMESTAMP, 0); }
		public TerminalNode K_SET() { return getToken(CqlParser.K_SET, 0); }
		public TerminalNode K_ASCII() { return getToken(CqlParser.K_ASCII, 0); }
		public TerminalNode K_BIGINT() { return getToken(CqlParser.K_BIGINT, 0); }
		public TerminalNode K_BLOB() { return getToken(CqlParser.K_BLOB, 0); }
		public TerminalNode K_BOOLEAN() { return getToken(CqlParser.K_BOOLEAN, 0); }
		public TerminalNode K_COUNTER() { return getToken(CqlParser.K_COUNTER, 0); }
		public TerminalNode K_DATE() { return getToken(CqlParser.K_DATE, 0); }
		public TerminalNode K_DECIMAL() { return getToken(CqlParser.K_DECIMAL, 0); }
		public TerminalNode K_DOUBLE() { return getToken(CqlParser.K_DOUBLE, 0); }
		public TerminalNode K_FLOAT() { return getToken(CqlParser.K_FLOAT, 0); }
		public TerminalNode K_FROZEN() { return getToken(CqlParser.K_FROZEN, 0); }
		public TerminalNode K_INET() { return getToken(CqlParser.K_INET, 0); }
		public TerminalNode K_INT() { return getToken(CqlParser.K_INT, 0); }
		public TerminalNode K_LIST() { return getToken(CqlParser.K_LIST, 0); }
		public TerminalNode K_MAP() { return getToken(CqlParser.K_MAP, 0); }
		public TerminalNode K_SMALLINT() { return getToken(CqlParser.K_SMALLINT, 0); }
		public TerminalNode K_TEXT() { return getToken(CqlParser.K_TEXT, 0); }
		public TerminalNode K_TIME() { return getToken(CqlParser.K_TIME, 0); }
		public TerminalNode K_TIMEUUID() { return getToken(CqlParser.K_TIMEUUID, 0); }
		public TerminalNode K_TINYINT() { return getToken(CqlParser.K_TINYINT, 0); }
		public TerminalNode K_TUPLE() { return getToken(CqlParser.K_TUPLE, 0); }
		public TerminalNode K_VARCHAR() { return getToken(CqlParser.K_VARCHAR, 0); }
		public TerminalNode K_VARINT() { return getToken(CqlParser.K_VARINT, 0); }
		public TerminalNode K_UUID() { return getToken(CqlParser.K_UUID, 0); }
		public DataTypeNameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dataTypeName; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDataTypeName(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDataTypeName(this);
		}
	}

	public final DataTypeNameContext dataTypeName() throws RecognitionException {
		DataTypeNameContext _localctx = new DataTypeNameContext(_ctx, getState());
		enterRule(_localctx, 318, RULE_dataTypeName);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2163);
			_la = _input.LA(1);
			if ( !(((((_la - 120)) & ~0x3f) == 0 && ((1L << (_la - 120)) & 292733974722118145L) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class DataTypeDefinitionContext extends ParserRuleContext {
		public SyntaxBracketLaContext syntaxBracketLa() {
			return getRuleContext(SyntaxBracketLaContext.class,0);
		}
		public List<DataTypeNameContext> dataTypeName() {
			return getRuleContexts(DataTypeNameContext.class);
		}
		public DataTypeNameContext dataTypeName(int i) {
			return getRuleContext(DataTypeNameContext.class,i);
		}
		public SyntaxBracketRaContext syntaxBracketRa() {
			return getRuleContext(SyntaxBracketRaContext.class,0);
		}
		public List<SyntaxCommaContext> syntaxComma() {
			return getRuleContexts(SyntaxCommaContext.class);
		}
		public SyntaxCommaContext syntaxComma(int i) {
			return getRuleContext(SyntaxCommaContext.class,i);
		}
		public DataTypeDefinitionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_dataTypeDefinition; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterDataTypeDefinition(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitDataTypeDefinition(this);
		}
	}

	public final DataTypeDefinitionContext dataTypeDefinition() throws RecognitionException {
		DataTypeDefinitionContext _localctx = new DataTypeDefinitionContext(_ctx, getState());
		enterRule(_localctx, 320, RULE_dataTypeDefinition);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2165);
			syntaxBracketLa();
			setState(2166);
			dataTypeName();
			setState(2172);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(2167);
				syntaxComma();
				setState(2168);
				dataTypeName();
				}
				}
				setState(2174);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(2175);
			syntaxBracketRa();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrderDirectionContext extends ParserRuleContext {
		public KwAscContext kwAsc() {
			return getRuleContext(KwAscContext.class,0);
		}
		public KwDescContext kwDesc() {
			return getRuleContext(KwDescContext.class,0);
		}
		public OrderDirectionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orderDirection; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterOrderDirection(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitOrderDirection(this);
		}
	}

	public final OrderDirectionContext orderDirection() throws RecognitionException {
		OrderDirectionContext _localctx = new OrderDirectionContext(_ctx, getState());
		enterRule(_localctx, 322, RULE_orderDirection);
		try {
			setState(2179);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case K_ASC:
				enterOuterAlt(_localctx, 1);
				{
				setState(2177);
				kwAsc();
				}
				break;
			case K_DESC:
				enterOuterAlt(_localctx, 2);
				{
				setState(2178);
				kwDesc();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RoleContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public RoleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_role; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterRole(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitRole(this);
		}
	}

	public final RoleContext role() throws RecognitionException {
		RoleContext _localctx = new RoleContext(_ctx, getState());
		enterRule(_localctx, 324, RULE_role);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2181);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TriggerContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TriggerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_trigger; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTrigger(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTrigger(this);
		}
	}

	public final TriggerContext trigger() throws RecognitionException {
		TriggerContext _localctx = new TriggerContext(_ctx, getState());
		enterRule(_localctx, 326, RULE_trigger);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2183);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TriggerClassContext extends ParserRuleContext {
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public TriggerClassContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_triggerClass; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterTriggerClass(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitTriggerClass(this);
		}
	}

	public final TriggerClassContext triggerClass() throws RecognitionException {
		TriggerClassContext _localctx = new TriggerClassContext(_ctx, getState());
		enterRule(_localctx, 328, RULE_triggerClass);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2185);
			stringLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class MaterializedViewContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public MaterializedViewContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_materializedView; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterMaterializedView(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitMaterializedView(this);
		}
	}

	public final MaterializedViewContext materializedView() throws RecognitionException {
		MaterializedViewContext _localctx = new MaterializedViewContext(_ctx, getState());
		enterRule(_localctx, 330, RULE_materializedView);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2187);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Type_Context extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public Type_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_type_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterType_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitType_(this);
		}
	}

	public final Type_Context type_() throws RecognitionException {
		Type_Context _localctx = new Type_Context(_ctx, getState());
		enterRule(_localctx, 332, RULE_type_);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2189);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AggregateContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public AggregateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_aggregate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterAggregate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitAggregate(this);
		}
	}

	public final AggregateContext aggregate() throws RecognitionException {
		AggregateContext _localctx = new AggregateContext(_ctx, getState());
		enterRule(_localctx, 334, RULE_aggregate);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2191);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Function_Context extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public Function_Context(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_function_; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterFunction_(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitFunction_(this);
		}
	}

	public final Function_Context function_() throws RecognitionException {
		Function_Context _localctx = new Function_Context(_ctx, getState());
		enterRule(_localctx, 336, RULE_function_);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2193);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LanguageContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public LanguageContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_language; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterLanguage(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitLanguage(this);
		}
	}

	public final LanguageContext language() throws RecognitionException {
		LanguageContext _localctx = new LanguageContext(_ctx, getState());
		enterRule(_localctx, 338, RULE_language);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2195);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class UserContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public UserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_user; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitUser(this);
		}
	}

	public final UserContext user() throws RecognitionException {
		UserContext _localctx = new UserContext(_ctx, getState());
		enterRule(_localctx, 340, RULE_user);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2197);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PasswordContext extends ParserRuleContext {
		public StringLiteralContext stringLiteral() {
			return getRuleContext(StringLiteralContext.class,0);
		}
		public PasswordContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_password; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterPassword(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitPassword(this);
		}
	}

	public final PasswordContext password() throws RecognitionException {
		PasswordContext _localctx = new PasswordContext(_ctx, getState());
		enterRule(_localctx, 342, RULE_password);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2199);
			stringLiteral();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class HashKeyContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public HashKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_hashKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterHashKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitHashKey(this);
		}
	}

	public final HashKeyContext hashKey() throws RecognitionException {
		HashKeyContext _localctx = new HashKeyContext(_ctx, getState());
		enterRule(_localctx, 344, RULE_hashKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2201);
			match(OBJECT_NAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ParamContext extends ParserRuleContext {
		public ParamNameContext paramName() {
			return getRuleContext(ParamNameContext.class,0);
		}
		public DataTypeContext dataType() {
			return getRuleContext(DataTypeContext.class,0);
		}
		public ParamContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_param; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterParam(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitParam(this);
		}
	}

	public final ParamContext param() throws RecognitionException {
		ParamContext _localctx = new ParamContext(_ctx, getState());
		enterRule(_localctx, 346, RULE_param);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2203);
			paramName();
			setState(2204);
			dataType();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ParamNameContext extends ParserRuleContext {
		public TerminalNode OBJECT_NAME() { return getToken(CqlParser.OBJECT_NAME, 0); }
		public TerminalNode K_INPUT() { return getToken(CqlParser.K_INPUT, 0); }
		public ParamNameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_paramName; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterParamName(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitParamName(this);
		}
	}

	public final ParamNameContext paramName() throws RecognitionException {
		ParamNameContext _localctx = new ParamNameContext(_ctx, getState());
		enterRule(_localctx, 348, RULE_paramName);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2206);
			_la = _input.LA(1);
			if ( !(_la==K_INPUT || _la==OBJECT_NAME) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAddContext extends ParserRuleContext {
		public TerminalNode K_ADD() { return getToken(CqlParser.K_ADD, 0); }
		public KwAddContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAdd; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAdd(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAdd(this);
		}
	}

	public final KwAddContext kwAdd() throws RecognitionException {
		KwAddContext _localctx = new KwAddContext(_ctx, getState());
		enterRule(_localctx, 350, RULE_kwAdd);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2208);
			match(K_ADD);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAggregateContext extends ParserRuleContext {
		public TerminalNode K_AGGREGATE() { return getToken(CqlParser.K_AGGREGATE, 0); }
		public KwAggregateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAggregate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAggregate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAggregate(this);
		}
	}

	public final KwAggregateContext kwAggregate() throws RecognitionException {
		KwAggregateContext _localctx = new KwAggregateContext(_ctx, getState());
		enterRule(_localctx, 352, RULE_kwAggregate);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2210);
			match(K_AGGREGATE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAggregatesContext extends ParserRuleContext {
		public TerminalNode K_AGGREGATES() { return getToken(CqlParser.K_AGGREGATES, 0); }
		public KwAggregatesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAggregates; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAggregates(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAggregates(this);
		}
	}

	public final KwAggregatesContext kwAggregates() throws RecognitionException {
		KwAggregatesContext _localctx = new KwAggregatesContext(_ctx, getState());
		enterRule(_localctx, 354, RULE_kwAggregates);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2212);
			match(K_AGGREGATES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAllContext extends ParserRuleContext {
		public TerminalNode K_ALL() { return getToken(CqlParser.K_ALL, 0); }
		public KwAllContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAll; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAll(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAll(this);
		}
	}

	public final KwAllContext kwAll() throws RecognitionException {
		KwAllContext _localctx = new KwAllContext(_ctx, getState());
		enterRule(_localctx, 356, RULE_kwAll);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2214);
			match(K_ALL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAllPermissionsContext extends ParserRuleContext {
		public TerminalNode K_ALL() { return getToken(CqlParser.K_ALL, 0); }
		public TerminalNode K_PERMISSIONS() { return getToken(CqlParser.K_PERMISSIONS, 0); }
		public KwAllPermissionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAllPermissions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAllPermissions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAllPermissions(this);
		}
	}

	public final KwAllPermissionsContext kwAllPermissions() throws RecognitionException {
		KwAllPermissionsContext _localctx = new KwAllPermissionsContext(_ctx, getState());
		enterRule(_localctx, 358, RULE_kwAllPermissions);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2216);
			match(K_ALL);
			setState(2217);
			match(K_PERMISSIONS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAllowContext extends ParserRuleContext {
		public TerminalNode K_ALLOW() { return getToken(CqlParser.K_ALLOW, 0); }
		public KwAllowContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAllow; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAllow(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAllow(this);
		}
	}

	public final KwAllowContext kwAllow() throws RecognitionException {
		KwAllowContext _localctx = new KwAllowContext(_ctx, getState());
		enterRule(_localctx, 360, RULE_kwAllow);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2219);
			match(K_ALLOW);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAlterContext extends ParserRuleContext {
		public TerminalNode K_ALTER() { return getToken(CqlParser.K_ALTER, 0); }
		public KwAlterContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAlter; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAlter(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAlter(this);
		}
	}

	public final KwAlterContext kwAlter() throws RecognitionException {
		KwAlterContext _localctx = new KwAlterContext(_ctx, getState());
		enterRule(_localctx, 362, RULE_kwAlter);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2221);
			match(K_ALTER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAndContext extends ParserRuleContext {
		public TerminalNode K_AND() { return getToken(CqlParser.K_AND, 0); }
		public KwAndContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAnd; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAnd(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAnd(this);
		}
	}

	public final KwAndContext kwAnd() throws RecognitionException {
		KwAndContext _localctx = new KwAndContext(_ctx, getState());
		enterRule(_localctx, 364, RULE_kwAnd);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2223);
			match(K_AND);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwApplyContext extends ParserRuleContext {
		public TerminalNode K_APPLY() { return getToken(CqlParser.K_APPLY, 0); }
		public KwApplyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwApply; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwApply(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwApply(this);
		}
	}

	public final KwApplyContext kwApply() throws RecognitionException {
		KwApplyContext _localctx = new KwApplyContext(_ctx, getState());
		enterRule(_localctx, 366, RULE_kwApply);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2225);
			match(K_APPLY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAsContext extends ParserRuleContext {
		public TerminalNode K_AS() { return getToken(CqlParser.K_AS, 0); }
		public KwAsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAs; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAs(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAs(this);
		}
	}

	public final KwAsContext kwAs() throws RecognitionException {
		KwAsContext _localctx = new KwAsContext(_ctx, getState());
		enterRule(_localctx, 368, RULE_kwAs);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2227);
			match(K_AS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAscContext extends ParserRuleContext {
		public TerminalNode K_ASC() { return getToken(CqlParser.K_ASC, 0); }
		public KwAscContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAsc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAsc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAsc(this);
		}
	}

	public final KwAscContext kwAsc() throws RecognitionException {
		KwAscContext _localctx = new KwAscContext(_ctx, getState());
		enterRule(_localctx, 370, RULE_kwAsc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2229);
			match(K_ASC);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwAuthorizeContext extends ParserRuleContext {
		public TerminalNode K_AUTHORIZE() { return getToken(CqlParser.K_AUTHORIZE, 0); }
		public KwAuthorizeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwAuthorize; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwAuthorize(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwAuthorize(this);
		}
	}

	public final KwAuthorizeContext kwAuthorize() throws RecognitionException {
		KwAuthorizeContext _localctx = new KwAuthorizeContext(_ctx, getState());
		enterRule(_localctx, 372, RULE_kwAuthorize);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2231);
			match(K_AUTHORIZE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwBatchContext extends ParserRuleContext {
		public TerminalNode K_BATCH() { return getToken(CqlParser.K_BATCH, 0); }
		public KwBatchContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwBatch; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwBatch(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwBatch(this);
		}
	}

	public final KwBatchContext kwBatch() throws RecognitionException {
		KwBatchContext _localctx = new KwBatchContext(_ctx, getState());
		enterRule(_localctx, 374, RULE_kwBatch);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2233);
			match(K_BATCH);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwBeginContext extends ParserRuleContext {
		public TerminalNode K_BEGIN() { return getToken(CqlParser.K_BEGIN, 0); }
		public KwBeginContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwBegin; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwBegin(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwBegin(this);
		}
	}

	public final KwBeginContext kwBegin() throws RecognitionException {
		KwBeginContext _localctx = new KwBeginContext(_ctx, getState());
		enterRule(_localctx, 376, RULE_kwBegin);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2235);
			match(K_BEGIN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwByContext extends ParserRuleContext {
		public TerminalNode K_BY() { return getToken(CqlParser.K_BY, 0); }
		public KwByContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwBy; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwBy(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwBy(this);
		}
	}

	public final KwByContext kwBy() throws RecognitionException {
		KwByContext _localctx = new KwByContext(_ctx, getState());
		enterRule(_localctx, 378, RULE_kwBy);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2237);
			match(K_BY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwCalledContext extends ParserRuleContext {
		public TerminalNode K_CALLED() { return getToken(CqlParser.K_CALLED, 0); }
		public KwCalledContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwCalled; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwCalled(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwCalled(this);
		}
	}

	public final KwCalledContext kwCalled() throws RecognitionException {
		KwCalledContext _localctx = new KwCalledContext(_ctx, getState());
		enterRule(_localctx, 380, RULE_kwCalled);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2239);
			match(K_CALLED);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwClusterContext extends ParserRuleContext {
		public TerminalNode K_CLUSTER() { return getToken(CqlParser.K_CLUSTER, 0); }
		public KwClusterContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwCluster; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwCluster(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwCluster(this);
		}
	}

	public final KwClusterContext kwCluster() throws RecognitionException {
		KwClusterContext _localctx = new KwClusterContext(_ctx, getState());
		enterRule(_localctx, 382, RULE_kwCluster);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2241);
			match(K_CLUSTER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwClusteringContext extends ParserRuleContext {
		public TerminalNode K_CLUSTERING() { return getToken(CqlParser.K_CLUSTERING, 0); }
		public KwClusteringContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwClustering; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwClustering(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwClustering(this);
		}
	}

	public final KwClusteringContext kwClustering() throws RecognitionException {
		KwClusteringContext _localctx = new KwClusteringContext(_ctx, getState());
		enterRule(_localctx, 384, RULE_kwClustering);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2243);
			match(K_CLUSTERING);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwCompactContext extends ParserRuleContext {
		public TerminalNode K_COMPACT() { return getToken(CqlParser.K_COMPACT, 0); }
		public KwCompactContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwCompact; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwCompact(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwCompact(this);
		}
	}

	public final KwCompactContext kwCompact() throws RecognitionException {
		KwCompactContext _localctx = new KwCompactContext(_ctx, getState());
		enterRule(_localctx, 386, RULE_kwCompact);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2245);
			match(K_COMPACT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwConnectionContext extends ParserRuleContext {
		public TerminalNode K_CONNECTION() { return getToken(CqlParser.K_CONNECTION, 0); }
		public KwConnectionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwConnection; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwConnection(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwConnection(this);
		}
	}

	public final KwConnectionContext kwConnection() throws RecognitionException {
		KwConnectionContext _localctx = new KwConnectionContext(_ctx, getState());
		enterRule(_localctx, 388, RULE_kwConnection);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2247);
			match(K_CONNECTION);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwConsistencyContext extends ParserRuleContext {
		public TerminalNode K_CONSISTENCY() { return getToken(CqlParser.K_CONSISTENCY, 0); }
		public KwConsistencyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwConsistency; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwConsistency(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwConsistency(this);
		}
	}

	public final KwConsistencyContext kwConsistency() throws RecognitionException {
		KwConsistencyContext _localctx = new KwConsistencyContext(_ctx, getState());
		enterRule(_localctx, 390, RULE_kwConsistency);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2249);
			match(K_CONSISTENCY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwConsistencyLevelContext extends ParserRuleContext {
		public TerminalNode K_ONE() { return getToken(CqlParser.K_ONE, 0); }
		public TerminalNode K_TWO() { return getToken(CqlParser.K_TWO, 0); }
		public TerminalNode K_THREE() { return getToken(CqlParser.K_THREE, 0); }
		public TerminalNode K_QUORUM() { return getToken(CqlParser.K_QUORUM, 0); }
		public TerminalNode K_ALL() { return getToken(CqlParser.K_ALL, 0); }
		public TerminalNode K_LOCAL_QUORUM() { return getToken(CqlParser.K_LOCAL_QUORUM, 0); }
		public TerminalNode K_EACH_QUORUM() { return getToken(CqlParser.K_EACH_QUORUM, 0); }
		public TerminalNode K_LOCAL_ONE() { return getToken(CqlParser.K_LOCAL_ONE, 0); }
		public TerminalNode K_SERIAL() { return getToken(CqlParser.K_SERIAL, 0); }
		public TerminalNode K_LOCAL_SERIAL() { return getToken(CqlParser.K_LOCAL_SERIAL, 0); }
		public TerminalNode K_ANY() { return getToken(CqlParser.K_ANY, 0); }
		public KwConsistencyLevelContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwConsistencyLevel; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwConsistencyLevel(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwConsistencyLevel(this);
		}
	}

	public final KwConsistencyLevelContext kwConsistencyLevel() throws RecognitionException {
		KwConsistencyLevelContext _localctx = new KwConsistencyLevelContext(_ctx, getState());
		enterRule(_localctx, 392, RULE_kwConsistencyLevel);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2251);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 36028799300665344L) != 0) || ((((_la - 84)) & ~0x3f) == 0 && ((1L << (_la - 84)) & 4521226206724103L) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOutputContext extends ParserRuleContext {
		public TerminalNode K_OUTPUT() { return getToken(CqlParser.K_OUTPUT, 0); }
		public KwOutputContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOutput; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOutput(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOutput(this);
		}
	}

	public final KwOutputContext kwOutput() throws RecognitionException {
		KwOutputContext _localctx = new KwOutputContext(_ctx, getState());
		enterRule(_localctx, 394, RULE_kwOutput);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2253);
			match(K_OUTPUT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOutputFormatTypeContext extends ParserRuleContext {
		public TerminalNode K_TABLE() { return getToken(CqlParser.K_TABLE, 0); }
		public TerminalNode K_ASCII() { return getToken(CqlParser.K_ASCII, 0); }
		public KwOutputFormatTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOutputFormatType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOutputFormatType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOutputFormatType(this);
		}
	}

	public final KwOutputFormatTypeContext kwOutputFormatType() throws RecognitionException {
		KwOutputFormatTypeContext _localctx = new KwOutputFormatTypeContext(_ctx, getState());
		enterRule(_localctx, 396, RULE_kwOutputFormatType);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2255);
			_la = _input.LA(1);
			if ( !(_la==K_TABLE || _la==K_ASCII) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwContainsContext extends ParserRuleContext {
		public TerminalNode K_CONTAINS() { return getToken(CqlParser.K_CONTAINS, 0); }
		public KwContainsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwContains; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwContains(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwContains(this);
		}
	}

	public final KwContainsContext kwContains() throws RecognitionException {
		KwContainsContext _localctx = new KwContainsContext(_ctx, getState());
		enterRule(_localctx, 398, RULE_kwContains);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2257);
			match(K_CONTAINS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwCreateContext extends ParserRuleContext {
		public TerminalNode K_CREATE() { return getToken(CqlParser.K_CREATE, 0); }
		public KwCreateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwCreate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwCreate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwCreate(this);
		}
	}

	public final KwCreateContext kwCreate() throws RecognitionException {
		KwCreateContext _localctx = new KwCreateContext(_ctx, getState());
		enterRule(_localctx, 400, RULE_kwCreate);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2259);
			match(K_CREATE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDeleteContext extends ParserRuleContext {
		public TerminalNode K_DELETE() { return getToken(CqlParser.K_DELETE, 0); }
		public KwDeleteContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDelete; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDelete(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDelete(this);
		}
	}

	public final KwDeleteContext kwDelete() throws RecognitionException {
		KwDeleteContext _localctx = new KwDeleteContext(_ctx, getState());
		enterRule(_localctx, 402, RULE_kwDelete);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2261);
			match(K_DELETE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDescContext extends ParserRuleContext {
		public TerminalNode K_DESC() { return getToken(CqlParser.K_DESC, 0); }
		public KwDescContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDesc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDesc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDesc(this);
		}
	}

	public final KwDescContext kwDesc() throws RecognitionException {
		KwDescContext _localctx = new KwDescContext(_ctx, getState());
		enterRule(_localctx, 404, RULE_kwDesc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2263);
			match(K_DESC);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDescribeContext extends ParserRuleContext {
		public TerminalNode K_DESCRIBE() { return getToken(CqlParser.K_DESCRIBE, 0); }
		public TerminalNode K_DESC() { return getToken(CqlParser.K_DESC, 0); }
		public KwDescribeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDescribe; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDescribe(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDescribe(this);
		}
	}

	public final KwDescribeContext kwDescribe() throws RecognitionException {
		KwDescribeContext _localctx = new KwDescribeContext(_ctx, getState());
		enterRule(_localctx, 406, RULE_kwDescribe);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2265);
			_la = _input.LA(1);
			if ( !(_la==K_DESC || _la==K_DESCRIBE) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDistinctContext extends ParserRuleContext {
		public TerminalNode K_DISTINCT() { return getToken(CqlParser.K_DISTINCT, 0); }
		public KwDistinctContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDistinct; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDistinct(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDistinct(this);
		}
	}

	public final KwDistinctContext kwDistinct() throws RecognitionException {
		KwDistinctContext _localctx = new KwDistinctContext(_ctx, getState());
		enterRule(_localctx, 408, RULE_kwDistinct);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2267);
			match(K_DISTINCT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDropContext extends ParserRuleContext {
		public TerminalNode K_DROP() { return getToken(CqlParser.K_DROP, 0); }
		public KwDropContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDrop; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDrop(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDrop(this);
		}
	}

	public final KwDropContext kwDrop() throws RecognitionException {
		KwDropContext _localctx = new KwDropContext(_ctx, getState());
		enterRule(_localctx, 410, RULE_kwDrop);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2269);
			match(K_DROP);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwDurableWritesContext extends ParserRuleContext {
		public TerminalNode K_DURABLE_WRITES() { return getToken(CqlParser.K_DURABLE_WRITES, 0); }
		public KwDurableWritesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwDurableWrites; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwDurableWrites(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwDurableWrites(this);
		}
	}

	public final KwDurableWritesContext kwDurableWrites() throws RecognitionException {
		KwDurableWritesContext _localctx = new KwDurableWritesContext(_ctx, getState());
		enterRule(_localctx, 412, RULE_kwDurableWrites);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2271);
			match(K_DURABLE_WRITES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwEntriesContext extends ParserRuleContext {
		public TerminalNode K_ENTRIES() { return getToken(CqlParser.K_ENTRIES, 0); }
		public KwEntriesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwEntries; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwEntries(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwEntries(this);
		}
	}

	public final KwEntriesContext kwEntries() throws RecognitionException {
		KwEntriesContext _localctx = new KwEntriesContext(_ctx, getState());
		enterRule(_localctx, 414, RULE_kwEntries);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2273);
			match(K_ENTRIES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwExecuteContext extends ParserRuleContext {
		public TerminalNode K_EXECUTE() { return getToken(CqlParser.K_EXECUTE, 0); }
		public KwExecuteContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwExecute; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwExecute(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwExecute(this);
		}
	}

	public final KwExecuteContext kwExecute() throws RecognitionException {
		KwExecuteContext _localctx = new KwExecuteContext(_ctx, getState());
		enterRule(_localctx, 416, RULE_kwExecute);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2275);
			match(K_EXECUTE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwExistsContext extends ParserRuleContext {
		public TerminalNode K_EXISTS() { return getToken(CqlParser.K_EXISTS, 0); }
		public KwExistsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwExists; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwExists(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwExists(this);
		}
	}

	public final KwExistsContext kwExists() throws RecognitionException {
		KwExistsContext _localctx = new KwExistsContext(_ctx, getState());
		enterRule(_localctx, 418, RULE_kwExists);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2277);
			match(K_EXISTS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFilteringContext extends ParserRuleContext {
		public TerminalNode K_FILTERING() { return getToken(CqlParser.K_FILTERING, 0); }
		public KwFilteringContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFiltering; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFiltering(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFiltering(this);
		}
	}

	public final KwFilteringContext kwFiltering() throws RecognitionException {
		KwFilteringContext _localctx = new KwFilteringContext(_ctx, getState());
		enterRule(_localctx, 420, RULE_kwFiltering);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2279);
			match(K_FILTERING);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFinalfuncContext extends ParserRuleContext {
		public TerminalNode K_FINALFUNC() { return getToken(CqlParser.K_FINALFUNC, 0); }
		public KwFinalfuncContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFinalfunc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFinalfunc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFinalfunc(this);
		}
	}

	public final KwFinalfuncContext kwFinalfunc() throws RecognitionException {
		KwFinalfuncContext _localctx = new KwFinalfuncContext(_ctx, getState());
		enterRule(_localctx, 422, RULE_kwFinalfunc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2281);
			match(K_FINALFUNC);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFromContext extends ParserRuleContext {
		public TerminalNode K_FROM() { return getToken(CqlParser.K_FROM, 0); }
		public KwFromContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFrom; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFrom(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFrom(this);
		}
	}

	public final KwFromContext kwFrom() throws RecognitionException {
		KwFromContext _localctx = new KwFromContext(_ctx, getState());
		enterRule(_localctx, 424, RULE_kwFrom);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2283);
			match(K_FROM);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFullContext extends ParserRuleContext {
		public TerminalNode K_FULL() { return getToken(CqlParser.K_FULL, 0); }
		public KwFullContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFull; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFull(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFull(this);
		}
	}

	public final KwFullContext kwFull() throws RecognitionException {
		KwFullContext _localctx = new KwFullContext(_ctx, getState());
		enterRule(_localctx, 426, RULE_kwFull);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2285);
			match(K_FULL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFunctionContext extends ParserRuleContext {
		public TerminalNode K_FUNCTION() { return getToken(CqlParser.K_FUNCTION, 0); }
		public KwFunctionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFunction; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFunction(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFunction(this);
		}
	}

	public final KwFunctionContext kwFunction() throws RecognitionException {
		KwFunctionContext _localctx = new KwFunctionContext(_ctx, getState());
		enterRule(_localctx, 428, RULE_kwFunction);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2287);
			match(K_FUNCTION);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwFunctionsContext extends ParserRuleContext {
		public TerminalNode K_FUNCTIONS() { return getToken(CqlParser.K_FUNCTIONS, 0); }
		public KwFunctionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwFunctions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwFunctions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwFunctions(this);
		}
	}

	public final KwFunctionsContext kwFunctions() throws RecognitionException {
		KwFunctionsContext _localctx = new KwFunctionsContext(_ctx, getState());
		enterRule(_localctx, 430, RULE_kwFunctions);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2289);
			match(K_FUNCTIONS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwGrantContext extends ParserRuleContext {
		public TerminalNode K_GRANT() { return getToken(CqlParser.K_GRANT, 0); }
		public KwGrantContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwGrant; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwGrant(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwGrant(this);
		}
	}

	public final KwGrantContext kwGrant() throws RecognitionException {
		KwGrantContext _localctx = new KwGrantContext(_ctx, getState());
		enterRule(_localctx, 432, RULE_kwGrant);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2291);
			match(K_GRANT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwIfContext extends ParserRuleContext {
		public TerminalNode K_IF() { return getToken(CqlParser.K_IF, 0); }
		public KwIfContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwIf; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwIf(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwIf(this);
		}
	}

	public final KwIfContext kwIf() throws RecognitionException {
		KwIfContext _localctx = new KwIfContext(_ctx, getState());
		enterRule(_localctx, 434, RULE_kwIf);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2293);
			match(K_IF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwInContext extends ParserRuleContext {
		public TerminalNode K_IN() { return getToken(CqlParser.K_IN, 0); }
		public KwInContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwIn; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwIn(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwIn(this);
		}
	}

	public final KwInContext kwIn() throws RecognitionException {
		KwInContext _localctx = new KwInContext(_ctx, getState());
		enterRule(_localctx, 436, RULE_kwIn);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2295);
			match(K_IN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwIndexContext extends ParserRuleContext {
		public TerminalNode K_INDEX() { return getToken(CqlParser.K_INDEX, 0); }
		public KwIndexContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwIndex; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwIndex(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwIndex(this);
		}
	}

	public final KwIndexContext kwIndex() throws RecognitionException {
		KwIndexContext _localctx = new KwIndexContext(_ctx, getState());
		enterRule(_localctx, 438, RULE_kwIndex);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2297);
			match(K_INDEX);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwInitcondContext extends ParserRuleContext {
		public TerminalNode K_INITCOND() { return getToken(CqlParser.K_INITCOND, 0); }
		public KwInitcondContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwInitcond; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwInitcond(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwInitcond(this);
		}
	}

	public final KwInitcondContext kwInitcond() throws RecognitionException {
		KwInitcondContext _localctx = new KwInitcondContext(_ctx, getState());
		enterRule(_localctx, 440, RULE_kwInitcond);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2299);
			match(K_INITCOND);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwInputContext extends ParserRuleContext {
		public TerminalNode K_INPUT() { return getToken(CqlParser.K_INPUT, 0); }
		public KwInputContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwInput; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwInput(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwInput(this);
		}
	}

	public final KwInputContext kwInput() throws RecognitionException {
		KwInputContext _localctx = new KwInputContext(_ctx, getState());
		enterRule(_localctx, 442, RULE_kwInput);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2301);
			match(K_INPUT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwInsertContext extends ParserRuleContext {
		public TerminalNode K_INSERT() { return getToken(CqlParser.K_INSERT, 0); }
		public KwInsertContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwInsert; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwInsert(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwInsert(this);
		}
	}

	public final KwInsertContext kwInsert() throws RecognitionException {
		KwInsertContext _localctx = new KwInsertContext(_ctx, getState());
		enterRule(_localctx, 444, RULE_kwInsert);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2303);
			match(K_INSERT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwIntoContext extends ParserRuleContext {
		public TerminalNode K_INTO() { return getToken(CqlParser.K_INTO, 0); }
		public KwIntoContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwInto; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwInto(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwInto(this);
		}
	}

	public final KwIntoContext kwInto() throws RecognitionException {
		KwIntoContext _localctx = new KwIntoContext(_ctx, getState());
		enterRule(_localctx, 446, RULE_kwInto);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2305);
			match(K_INTO);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwIsContext extends ParserRuleContext {
		public TerminalNode K_IS() { return getToken(CqlParser.K_IS, 0); }
		public KwIsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwIs; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwIs(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwIs(this);
		}
	}

	public final KwIsContext kwIs() throws RecognitionException {
		KwIsContext _localctx = new KwIsContext(_ctx, getState());
		enterRule(_localctx, 448, RULE_kwIs);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2307);
			match(K_IS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwJsonContext extends ParserRuleContext {
		public TerminalNode K_JSON() { return getToken(CqlParser.K_JSON, 0); }
		public KwJsonContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwJson; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwJson(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwJson(this);
		}
	}

	public final KwJsonContext kwJson() throws RecognitionException {
		KwJsonContext _localctx = new KwJsonContext(_ctx, getState());
		enterRule(_localctx, 450, RULE_kwJson);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2309);
			match(K_JSON);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwKeyContext extends ParserRuleContext {
		public TerminalNode K_KEY() { return getToken(CqlParser.K_KEY, 0); }
		public KwKeyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwKey; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwKey(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwKey(this);
		}
	}

	public final KwKeyContext kwKey() throws RecognitionException {
		KwKeyContext _localctx = new KwKeyContext(_ctx, getState());
		enterRule(_localctx, 452, RULE_kwKey);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2311);
			match(K_KEY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwKeysContext extends ParserRuleContext {
		public TerminalNode K_KEYS() { return getToken(CqlParser.K_KEYS, 0); }
		public KwKeysContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwKeys; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwKeys(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwKeys(this);
		}
	}

	public final KwKeysContext kwKeys() throws RecognitionException {
		KwKeysContext _localctx = new KwKeysContext(_ctx, getState());
		enterRule(_localctx, 454, RULE_kwKeys);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2313);
			match(K_KEYS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwKeyspaceContext extends ParserRuleContext {
		public TerminalNode K_KEYSPACE() { return getToken(CqlParser.K_KEYSPACE, 0); }
		public KwKeyspaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwKeyspace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwKeyspace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwKeyspace(this);
		}
	}

	public final KwKeyspaceContext kwKeyspace() throws RecognitionException {
		KwKeyspaceContext _localctx = new KwKeyspaceContext(_ctx, getState());
		enterRule(_localctx, 456, RULE_kwKeyspace);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2315);
			match(K_KEYSPACE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwKeyspacesContext extends ParserRuleContext {
		public TerminalNode K_KEYSPACES() { return getToken(CqlParser.K_KEYSPACES, 0); }
		public KwKeyspacesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwKeyspaces; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwKeyspaces(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwKeyspaces(this);
		}
	}

	public final KwKeyspacesContext kwKeyspaces() throws RecognitionException {
		KwKeyspacesContext _localctx = new KwKeyspacesContext(_ctx, getState());
		enterRule(_localctx, 458, RULE_kwKeyspaces);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2317);
			match(K_KEYSPACES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwLanguageContext extends ParserRuleContext {
		public TerminalNode K_LANGUAGE() { return getToken(CqlParser.K_LANGUAGE, 0); }
		public KwLanguageContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwLanguage; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwLanguage(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwLanguage(this);
		}
	}

	public final KwLanguageContext kwLanguage() throws RecognitionException {
		KwLanguageContext _localctx = new KwLanguageContext(_ctx, getState());
		enterRule(_localctx, 460, RULE_kwLanguage);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2319);
			match(K_LANGUAGE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwLimitContext extends ParserRuleContext {
		public TerminalNode K_LIMIT() { return getToken(CqlParser.K_LIMIT, 0); }
		public KwLimitContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwLimit; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwLimit(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwLimit(this);
		}
	}

	public final KwLimitContext kwLimit() throws RecognitionException {
		KwLimitContext _localctx = new KwLimitContext(_ctx, getState());
		enterRule(_localctx, 462, RULE_kwLimit);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2321);
			match(K_LIMIT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwListContext extends ParserRuleContext {
		public TerminalNode K_LIST() { return getToken(CqlParser.K_LIST, 0); }
		public KwListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwList; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwList(this);
		}
	}

	public final KwListContext kwList() throws RecognitionException {
		KwListContext _localctx = new KwListContext(_ctx, getState());
		enterRule(_localctx, 464, RULE_kwList);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2323);
			match(K_LIST);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwLoggedContext extends ParserRuleContext {
		public TerminalNode K_LOGGED() { return getToken(CqlParser.K_LOGGED, 0); }
		public KwLoggedContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwLogged; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwLogged(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwLogged(this);
		}
	}

	public final KwLoggedContext kwLogged() throws RecognitionException {
		KwLoggedContext _localctx = new KwLoggedContext(_ctx, getState());
		enterRule(_localctx, 466, RULE_kwLogged);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2325);
			match(K_LOGGED);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwLocalSerialContext extends ParserRuleContext {
		public TerminalNode K_LOCAL_SERIAL() { return getToken(CqlParser.K_LOCAL_SERIAL, 0); }
		public KwLocalSerialContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwLocalSerial; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwLocalSerial(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwLocalSerial(this);
		}
	}

	public final KwLocalSerialContext kwLocalSerial() throws RecognitionException {
		KwLocalSerialContext _localctx = new KwLocalSerialContext(_ctx, getState());
		enterRule(_localctx, 468, RULE_kwLocalSerial);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2327);
			match(K_LOCAL_SERIAL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwLoginContext extends ParserRuleContext {
		public TerminalNode K_LOGIN() { return getToken(CqlParser.K_LOGIN, 0); }
		public KwLoginContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwLogin; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwLogin(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwLogin(this);
		}
	}

	public final KwLoginContext kwLogin() throws RecognitionException {
		KwLoginContext _localctx = new KwLoginContext(_ctx, getState());
		enterRule(_localctx, 470, RULE_kwLogin);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2329);
			match(K_LOGIN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwMaterializedContext extends ParserRuleContext {
		public TerminalNode K_MATERIALIZED() { return getToken(CqlParser.K_MATERIALIZED, 0); }
		public KwMaterializedContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwMaterialized; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwMaterialized(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwMaterialized(this);
		}
	}

	public final KwMaterializedContext kwMaterialized() throws RecognitionException {
		KwMaterializedContext _localctx = new KwMaterializedContext(_ctx, getState());
		enterRule(_localctx, 472, RULE_kwMaterialized);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2331);
			match(K_MATERIALIZED);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwModifyContext extends ParserRuleContext {
		public TerminalNode K_MODIFY() { return getToken(CqlParser.K_MODIFY, 0); }
		public KwModifyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwModify; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwModify(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwModify(this);
		}
	}

	public final KwModifyContext kwModify() throws RecognitionException {
		KwModifyContext _localctx = new KwModifyContext(_ctx, getState());
		enterRule(_localctx, 474, RULE_kwModify);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2333);
			match(K_MODIFY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwNosuperuserContext extends ParserRuleContext {
		public TerminalNode K_NOSUPERUSER() { return getToken(CqlParser.K_NOSUPERUSER, 0); }
		public KwNosuperuserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwNosuperuser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwNosuperuser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwNosuperuser(this);
		}
	}

	public final KwNosuperuserContext kwNosuperuser() throws RecognitionException {
		KwNosuperuserContext _localctx = new KwNosuperuserContext(_ctx, getState());
		enterRule(_localctx, 476, RULE_kwNosuperuser);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2335);
			match(K_NOSUPERUSER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwNorecursiveContext extends ParserRuleContext {
		public TerminalNode K_NORECURSIVE() { return getToken(CqlParser.K_NORECURSIVE, 0); }
		public KwNorecursiveContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwNorecursive; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwNorecursive(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwNorecursive(this);
		}
	}

	public final KwNorecursiveContext kwNorecursive() throws RecognitionException {
		KwNorecursiveContext _localctx = new KwNorecursiveContext(_ctx, getState());
		enterRule(_localctx, 478, RULE_kwNorecursive);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2337);
			match(K_NORECURSIVE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwNotContext extends ParserRuleContext {
		public TerminalNode K_NOT() { return getToken(CqlParser.K_NOT, 0); }
		public KwNotContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwNot; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwNot(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwNot(this);
		}
	}

	public final KwNotContext kwNot() throws RecognitionException {
		KwNotContext _localctx = new KwNotContext(_ctx, getState());
		enterRule(_localctx, 480, RULE_kwNot);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2339);
			match(K_NOT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwNullContext extends ParserRuleContext {
		public TerminalNode K_NULL() { return getToken(CqlParser.K_NULL, 0); }
		public KwNullContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwNull; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwNull(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwNull(this);
		}
	}

	public final KwNullContext kwNull() throws RecognitionException {
		KwNullContext _localctx = new KwNullContext(_ctx, getState());
		enterRule(_localctx, 482, RULE_kwNull);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2341);
			match(K_NULL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOfContext extends ParserRuleContext {
		public TerminalNode K_OF() { return getToken(CqlParser.K_OF, 0); }
		public KwOfContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOf; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOf(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOf(this);
		}
	}

	public final KwOfContext kwOf() throws RecognitionException {
		KwOfContext _localctx = new KwOfContext(_ctx, getState());
		enterRule(_localctx, 484, RULE_kwOf);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2343);
			match(K_OF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOnContext extends ParserRuleContext {
		public TerminalNode K_ON() { return getToken(CqlParser.K_ON, 0); }
		public KwOnContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOn; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOn(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOn(this);
		}
	}

	public final KwOnContext kwOn() throws RecognitionException {
		KwOnContext _localctx = new KwOnContext(_ctx, getState());
		enterRule(_localctx, 486, RULE_kwOn);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2345);
			match(K_ON);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOptionsContext extends ParserRuleContext {
		public TerminalNode K_OPTIONS() { return getToken(CqlParser.K_OPTIONS, 0); }
		public KwOptionsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOptions; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOptions(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOptions(this);
		}
	}

	public final KwOptionsContext kwOptions() throws RecognitionException {
		KwOptionsContext _localctx = new KwOptionsContext(_ctx, getState());
		enterRule(_localctx, 488, RULE_kwOptions);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2347);
			match(K_OPTIONS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOrContext extends ParserRuleContext {
		public TerminalNode K_OR() { return getToken(CqlParser.K_OR, 0); }
		public KwOrContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOr; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOr(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOr(this);
		}
	}

	public final KwOrContext kwOr() throws RecognitionException {
		KwOrContext _localctx = new KwOrContext(_ctx, getState());
		enterRule(_localctx, 490, RULE_kwOr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2349);
			match(K_OR);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwOrderContext extends ParserRuleContext {
		public TerminalNode K_ORDER() { return getToken(CqlParser.K_ORDER, 0); }
		public KwOrderContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwOrder; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwOrder(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwOrder(this);
		}
	}

	public final KwOrderContext kwOrder() throws RecognitionException {
		KwOrderContext _localctx = new KwOrderContext(_ctx, getState());
		enterRule(_localctx, 492, RULE_kwOrder);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2351);
			match(K_ORDER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwPasswordContext extends ParserRuleContext {
		public TerminalNode K_PASSWORD() { return getToken(CqlParser.K_PASSWORD, 0); }
		public KwPasswordContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwPassword; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwPassword(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwPassword(this);
		}
	}

	public final KwPasswordContext kwPassword() throws RecognitionException {
		KwPasswordContext _localctx = new KwPasswordContext(_ctx, getState());
		enterRule(_localctx, 494, RULE_kwPassword);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2353);
			match(K_PASSWORD);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwPrimaryContext extends ParserRuleContext {
		public TerminalNode K_PRIMARY() { return getToken(CqlParser.K_PRIMARY, 0); }
		public KwPrimaryContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwPrimary; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwPrimary(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwPrimary(this);
		}
	}

	public final KwPrimaryContext kwPrimary() throws RecognitionException {
		KwPrimaryContext _localctx = new KwPrimaryContext(_ctx, getState());
		enterRule(_localctx, 496, RULE_kwPrimary);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2355);
			match(K_PRIMARY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwRenameContext extends ParserRuleContext {
		public TerminalNode K_RENAME() { return getToken(CqlParser.K_RENAME, 0); }
		public KwRenameContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwRename; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwRename(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwRename(this);
		}
	}

	public final KwRenameContext kwRename() throws RecognitionException {
		KwRenameContext _localctx = new KwRenameContext(_ctx, getState());
		enterRule(_localctx, 498, RULE_kwRename);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2357);
			match(K_RENAME);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwReplaceContext extends ParserRuleContext {
		public TerminalNode K_REPLACE() { return getToken(CqlParser.K_REPLACE, 0); }
		public KwReplaceContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwReplace; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwReplace(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwReplace(this);
		}
	}

	public final KwReplaceContext kwReplace() throws RecognitionException {
		KwReplaceContext _localctx = new KwReplaceContext(_ctx, getState());
		enterRule(_localctx, 500, RULE_kwReplace);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2359);
			match(K_REPLACE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwReplicationContext extends ParserRuleContext {
		public TerminalNode K_REPLICATION() { return getToken(CqlParser.K_REPLICATION, 0); }
		public KwReplicationContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwReplication; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwReplication(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwReplication(this);
		}
	}

	public final KwReplicationContext kwReplication() throws RecognitionException {
		KwReplicationContext _localctx = new KwReplicationContext(_ctx, getState());
		enterRule(_localctx, 502, RULE_kwReplication);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2361);
			match(K_REPLICATION);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwReturnsContext extends ParserRuleContext {
		public TerminalNode K_RETURNS() { return getToken(CqlParser.K_RETURNS, 0); }
		public KwReturnsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwReturns; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwReturns(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwReturns(this);
		}
	}

	public final KwReturnsContext kwReturns() throws RecognitionException {
		KwReturnsContext _localctx = new KwReturnsContext(_ctx, getState());
		enterRule(_localctx, 504, RULE_kwReturns);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2363);
			match(K_RETURNS);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwRoleContext extends ParserRuleContext {
		public TerminalNode K_ROLE() { return getToken(CqlParser.K_ROLE, 0); }
		public KwRoleContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwRole; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwRole(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwRole(this);
		}
	}

	public final KwRoleContext kwRole() throws RecognitionException {
		KwRoleContext _localctx = new KwRoleContext(_ctx, getState());
		enterRule(_localctx, 506, RULE_kwRole);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2365);
			match(K_ROLE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwRolesContext extends ParserRuleContext {
		public TerminalNode K_ROLES() { return getToken(CqlParser.K_ROLES, 0); }
		public KwRolesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwRoles; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwRoles(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwRoles(this);
		}
	}

	public final KwRolesContext kwRoles() throws RecognitionException {
		KwRolesContext _localctx = new KwRolesContext(_ctx, getState());
		enterRule(_localctx, 508, RULE_kwRoles);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2367);
			match(K_ROLES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwSelectContext extends ParserRuleContext {
		public TerminalNode K_SELECT() { return getToken(CqlParser.K_SELECT, 0); }
		public KwSelectContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwSelect; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwSelect(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwSelect(this);
		}
	}

	public final KwSelectContext kwSelect() throws RecognitionException {
		KwSelectContext _localctx = new KwSelectContext(_ctx, getState());
		enterRule(_localctx, 510, RULE_kwSelect);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2369);
			match(K_SELECT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwSerialContext extends ParserRuleContext {
		public TerminalNode K_SERIAL() { return getToken(CqlParser.K_SERIAL, 0); }
		public KwSerialContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwSerial; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwSerial(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwSerial(this);
		}
	}

	public final KwSerialContext kwSerial() throws RecognitionException {
		KwSerialContext _localctx = new KwSerialContext(_ctx, getState());
		enterRule(_localctx, 512, RULE_kwSerial);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2371);
			match(K_SERIAL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwSetContext extends ParserRuleContext {
		public TerminalNode K_SET() { return getToken(CqlParser.K_SET, 0); }
		public KwSetContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwSet; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwSet(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwSet(this);
		}
	}

	public final KwSetContext kwSet() throws RecognitionException {
		KwSetContext _localctx = new KwSetContext(_ctx, getState());
		enterRule(_localctx, 514, RULE_kwSet);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2373);
			match(K_SET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwSfuncContext extends ParserRuleContext {
		public TerminalNode K_SFUNC() { return getToken(CqlParser.K_SFUNC, 0); }
		public KwSfuncContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwSfunc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwSfunc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwSfunc(this);
		}
	}

	public final KwSfuncContext kwSfunc() throws RecognitionException {
		KwSfuncContext _localctx = new KwSfuncContext(_ctx, getState());
		enterRule(_localctx, 516, RULE_kwSfunc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2375);
			match(K_SFUNC);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwStorageContext extends ParserRuleContext {
		public TerminalNode K_STORAGE() { return getToken(CqlParser.K_STORAGE, 0); }
		public KwStorageContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwStorage; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwStorage(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwStorage(this);
		}
	}

	public final KwStorageContext kwStorage() throws RecognitionException {
		KwStorageContext _localctx = new KwStorageContext(_ctx, getState());
		enterRule(_localctx, 518, RULE_kwStorage);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2377);
			match(K_STORAGE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwStypeContext extends ParserRuleContext {
		public TerminalNode K_STYPE() { return getToken(CqlParser.K_STYPE, 0); }
		public KwStypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwStype; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwStype(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwStype(this);
		}
	}

	public final KwStypeContext kwStype() throws RecognitionException {
		KwStypeContext _localctx = new KwStypeContext(_ctx, getState());
		enterRule(_localctx, 520, RULE_kwStype);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2379);
			match(K_STYPE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwSuperuserContext extends ParserRuleContext {
		public TerminalNode K_SUPERUSER() { return getToken(CqlParser.K_SUPERUSER, 0); }
		public KwSuperuserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwSuperuser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwSuperuser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwSuperuser(this);
		}
	}

	public final KwSuperuserContext kwSuperuser() throws RecognitionException {
		KwSuperuserContext _localctx = new KwSuperuserContext(_ctx, getState());
		enterRule(_localctx, 522, RULE_kwSuperuser);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2381);
			match(K_SUPERUSER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTableContext extends ParserRuleContext {
		public TerminalNode K_TABLE() { return getToken(CqlParser.K_TABLE, 0); }
		public KwTableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTable; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTable(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTable(this);
		}
	}

	public final KwTableContext kwTable() throws RecognitionException {
		KwTableContext _localctx = new KwTableContext(_ctx, getState());
		enterRule(_localctx, 524, RULE_kwTable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2383);
			match(K_TABLE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTablesContext extends ParserRuleContext {
		public TerminalNode K_TABLES() { return getToken(CqlParser.K_TABLES, 0); }
		public KwTablesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTables; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTables(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTables(this);
		}
	}

	public final KwTablesContext kwTables() throws RecognitionException {
		KwTablesContext _localctx = new KwTablesContext(_ctx, getState());
		enterRule(_localctx, 526, RULE_kwTables);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2385);
			match(K_TABLES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTimestampContext extends ParserRuleContext {
		public TerminalNode K_TIMESTAMP() { return getToken(CqlParser.K_TIMESTAMP, 0); }
		public KwTimestampContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTimestamp; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTimestamp(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTimestamp(this);
		}
	}

	public final KwTimestampContext kwTimestamp() throws RecognitionException {
		KwTimestampContext _localctx = new KwTimestampContext(_ctx, getState());
		enterRule(_localctx, 528, RULE_kwTimestamp);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2387);
			match(K_TIMESTAMP);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwToContext extends ParserRuleContext {
		public TerminalNode K_TO() { return getToken(CqlParser.K_TO, 0); }
		public KwToContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTo; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTo(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTo(this);
		}
	}

	public final KwToContext kwTo() throws RecognitionException {
		KwToContext _localctx = new KwToContext(_ctx, getState());
		enterRule(_localctx, 530, RULE_kwTo);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2389);
			match(K_TO);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTriggerContext extends ParserRuleContext {
		public TerminalNode K_TRIGGER() { return getToken(CqlParser.K_TRIGGER, 0); }
		public KwTriggerContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTrigger; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTrigger(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTrigger(this);
		}
	}

	public final KwTriggerContext kwTrigger() throws RecognitionException {
		KwTriggerContext _localctx = new KwTriggerContext(_ctx, getState());
		enterRule(_localctx, 532, RULE_kwTrigger);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2391);
			match(K_TRIGGER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTruncateContext extends ParserRuleContext {
		public TerminalNode K_TRUNCATE() { return getToken(CqlParser.K_TRUNCATE, 0); }
		public KwTruncateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTruncate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTruncate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTruncate(this);
		}
	}

	public final KwTruncateContext kwTruncate() throws RecognitionException {
		KwTruncateContext _localctx = new KwTruncateContext(_ctx, getState());
		enterRule(_localctx, 534, RULE_kwTruncate);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2393);
			match(K_TRUNCATE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTtlContext extends ParserRuleContext {
		public TerminalNode K_TTL() { return getToken(CqlParser.K_TTL, 0); }
		public KwTtlContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTtl; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTtl(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTtl(this);
		}
	}

	public final KwTtlContext kwTtl() throws RecognitionException {
		KwTtlContext _localctx = new KwTtlContext(_ctx, getState());
		enterRule(_localctx, 536, RULE_kwTtl);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2395);
			match(K_TTL);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTypeContext extends ParserRuleContext {
		public TerminalNode K_TYPE() { return getToken(CqlParser.K_TYPE, 0); }
		public KwTypeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwType; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwType(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwType(this);
		}
	}

	public final KwTypeContext kwType() throws RecognitionException {
		KwTypeContext _localctx = new KwTypeContext(_ctx, getState());
		enterRule(_localctx, 538, RULE_kwType);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2397);
			match(K_TYPE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwTypesContext extends ParserRuleContext {
		public TerminalNode K_TYPES() { return getToken(CqlParser.K_TYPES, 0); }
		public KwTypesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwTypes; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwTypes(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwTypes(this);
		}
	}

	public final KwTypesContext kwTypes() throws RecognitionException {
		KwTypesContext _localctx = new KwTypesContext(_ctx, getState());
		enterRule(_localctx, 540, RULE_kwTypes);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2399);
			match(K_TYPES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwUnloggedContext extends ParserRuleContext {
		public TerminalNode K_UNLOGGED() { return getToken(CqlParser.K_UNLOGGED, 0); }
		public KwUnloggedContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwUnlogged; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwUnlogged(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwUnlogged(this);
		}
	}

	public final KwUnloggedContext kwUnlogged() throws RecognitionException {
		KwUnloggedContext _localctx = new KwUnloggedContext(_ctx, getState());
		enterRule(_localctx, 542, RULE_kwUnlogged);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2401);
			match(K_UNLOGGED);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwUpdateContext extends ParserRuleContext {
		public TerminalNode K_UPDATE() { return getToken(CqlParser.K_UPDATE, 0); }
		public KwUpdateContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwUpdate; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwUpdate(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwUpdate(this);
		}
	}

	public final KwUpdateContext kwUpdate() throws RecognitionException {
		KwUpdateContext _localctx = new KwUpdateContext(_ctx, getState());
		enterRule(_localctx, 544, RULE_kwUpdate);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2403);
			match(K_UPDATE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwUseContext extends ParserRuleContext {
		public TerminalNode K_USE() { return getToken(CqlParser.K_USE, 0); }
		public KwUseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwUse; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwUse(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwUse(this);
		}
	}

	public final KwUseContext kwUse() throws RecognitionException {
		KwUseContext _localctx = new KwUseContext(_ctx, getState());
		enterRule(_localctx, 546, RULE_kwUse);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2405);
			match(K_USE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwUserContext extends ParserRuleContext {
		public TerminalNode K_USER() { return getToken(CqlParser.K_USER, 0); }
		public KwUserContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwUser; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwUser(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwUser(this);
		}
	}

	public final KwUserContext kwUser() throws RecognitionException {
		KwUserContext _localctx = new KwUserContext(_ctx, getState());
		enterRule(_localctx, 548, RULE_kwUser);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2407);
			match(K_USER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwUsingContext extends ParserRuleContext {
		public TerminalNode K_USING() { return getToken(CqlParser.K_USING, 0); }
		public KwUsingContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwUsing; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwUsing(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwUsing(this);
		}
	}

	public final KwUsingContext kwUsing() throws RecognitionException {
		KwUsingContext _localctx = new KwUsingContext(_ctx, getState());
		enterRule(_localctx, 550, RULE_kwUsing);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2409);
			match(K_USING);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwValuesContext extends ParserRuleContext {
		public TerminalNode K_VALUES() { return getToken(CqlParser.K_VALUES, 0); }
		public KwValuesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwValues; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwValues(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwValues(this);
		}
	}

	public final KwValuesContext kwValues() throws RecognitionException {
		KwValuesContext _localctx = new KwValuesContext(_ctx, getState());
		enterRule(_localctx, 552, RULE_kwValues);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2411);
			match(K_VALUES);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwViewContext extends ParserRuleContext {
		public TerminalNode K_VIEW() { return getToken(CqlParser.K_VIEW, 0); }
		public KwViewContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwView; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwView(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwView(this);
		}
	}

	public final KwViewContext kwView() throws RecognitionException {
		KwViewContext _localctx = new KwViewContext(_ctx, getState());
		enterRule(_localctx, 554, RULE_kwView);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2413);
			match(K_VIEW);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwWhereContext extends ParserRuleContext {
		public TerminalNode K_WHERE() { return getToken(CqlParser.K_WHERE, 0); }
		public KwWhereContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwWhere; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwWhere(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwWhere(this);
		}
	}

	public final KwWhereContext kwWhere() throws RecognitionException {
		KwWhereContext _localctx = new KwWhereContext(_ctx, getState());
		enterRule(_localctx, 556, RULE_kwWhere);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2415);
			match(K_WHERE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwWithContext extends ParserRuleContext {
		public TerminalNode K_WITH() { return getToken(CqlParser.K_WITH, 0); }
		public KwWithContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwWith; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwWith(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwWith(this);
		}
	}

	public final KwWithContext kwWith() throws RecognitionException {
		KwWithContext _localctx = new KwWithContext(_ctx, getState());
		enterRule(_localctx, 558, RULE_kwWith);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2417);
			match(K_WITH);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class KwRevokeContext extends ParserRuleContext {
		public TerminalNode K_REVOKE() { return getToken(CqlParser.K_REVOKE, 0); }
		public KwRevokeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_kwRevoke; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterKwRevoke(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitKwRevoke(this);
		}
	}

	public final KwRevokeContext kwRevoke() throws RecognitionException {
		KwRevokeContext _localctx = new KwRevokeContext(_ctx, getState());
		enterRule(_localctx, 560, RULE_kwRevoke);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2419);
			match(K_REVOKE);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketLrContext extends ParserRuleContext {
		public TerminalNode LR_BRACKET() { return getToken(CqlParser.LR_BRACKET, 0); }
		public SyntaxBracketLrContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketLr; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketLr(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketLr(this);
		}
	}

	public final SyntaxBracketLrContext syntaxBracketLr() throws RecognitionException {
		SyntaxBracketLrContext _localctx = new SyntaxBracketLrContext(_ctx, getState());
		enterRule(_localctx, 562, RULE_syntaxBracketLr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2421);
			match(LR_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketRrContext extends ParserRuleContext {
		public TerminalNode RR_BRACKET() { return getToken(CqlParser.RR_BRACKET, 0); }
		public SyntaxBracketRrContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketRr; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketRr(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketRr(this);
		}
	}

	public final SyntaxBracketRrContext syntaxBracketRr() throws RecognitionException {
		SyntaxBracketRrContext _localctx = new SyntaxBracketRrContext(_ctx, getState());
		enterRule(_localctx, 564, RULE_syntaxBracketRr);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2423);
			match(RR_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketLcContext extends ParserRuleContext {
		public TerminalNode LC_BRACKET() { return getToken(CqlParser.LC_BRACKET, 0); }
		public SyntaxBracketLcContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketLc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketLc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketLc(this);
		}
	}

	public final SyntaxBracketLcContext syntaxBracketLc() throws RecognitionException {
		SyntaxBracketLcContext _localctx = new SyntaxBracketLcContext(_ctx, getState());
		enterRule(_localctx, 566, RULE_syntaxBracketLc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2425);
			match(LC_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketRcContext extends ParserRuleContext {
		public TerminalNode RC_BRACKET() { return getToken(CqlParser.RC_BRACKET, 0); }
		public SyntaxBracketRcContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketRc; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketRc(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketRc(this);
		}
	}

	public final SyntaxBracketRcContext syntaxBracketRc() throws RecognitionException {
		SyntaxBracketRcContext _localctx = new SyntaxBracketRcContext(_ctx, getState());
		enterRule(_localctx, 568, RULE_syntaxBracketRc);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2427);
			match(RC_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketLaContext extends ParserRuleContext {
		public TerminalNode OPERATOR_LT() { return getToken(CqlParser.OPERATOR_LT, 0); }
		public SyntaxBracketLaContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketLa; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketLa(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketLa(this);
		}
	}

	public final SyntaxBracketLaContext syntaxBracketLa() throws RecognitionException {
		SyntaxBracketLaContext _localctx = new SyntaxBracketLaContext(_ctx, getState());
		enterRule(_localctx, 570, RULE_syntaxBracketLa);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2429);
			match(OPERATOR_LT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketRaContext extends ParserRuleContext {
		public TerminalNode OPERATOR_GT() { return getToken(CqlParser.OPERATOR_GT, 0); }
		public SyntaxBracketRaContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketRa; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketRa(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketRa(this);
		}
	}

	public final SyntaxBracketRaContext syntaxBracketRa() throws RecognitionException {
		SyntaxBracketRaContext _localctx = new SyntaxBracketRaContext(_ctx, getState());
		enterRule(_localctx, 572, RULE_syntaxBracketRa);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2431);
			match(OPERATOR_GT);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketLsContext extends ParserRuleContext {
		public TerminalNode LS_BRACKET() { return getToken(CqlParser.LS_BRACKET, 0); }
		public SyntaxBracketLsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketLs; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketLs(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketLs(this);
		}
	}

	public final SyntaxBracketLsContext syntaxBracketLs() throws RecognitionException {
		SyntaxBracketLsContext _localctx = new SyntaxBracketLsContext(_ctx, getState());
		enterRule(_localctx, 574, RULE_syntaxBracketLs);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2433);
			match(LS_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxBracketRsContext extends ParserRuleContext {
		public TerminalNode RS_BRACKET() { return getToken(CqlParser.RS_BRACKET, 0); }
		public SyntaxBracketRsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxBracketRs; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxBracketRs(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxBracketRs(this);
		}
	}

	public final SyntaxBracketRsContext syntaxBracketRs() throws RecognitionException {
		SyntaxBracketRsContext _localctx = new SyntaxBracketRsContext(_ctx, getState());
		enterRule(_localctx, 576, RULE_syntaxBracketRs);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2435);
			match(RS_BRACKET);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxCommaContext extends ParserRuleContext {
		public TerminalNode COMMA() { return getToken(CqlParser.COMMA, 0); }
		public SyntaxCommaContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxComma; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxComma(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxComma(this);
		}
	}

	public final SyntaxCommaContext syntaxComma() throws RecognitionException {
		SyntaxCommaContext _localctx = new SyntaxCommaContext(_ctx, getState());
		enterRule(_localctx, 578, RULE_syntaxComma);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2437);
			match(COMMA);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SyntaxColonContext extends ParserRuleContext {
		public TerminalNode COLON() { return getToken(CqlParser.COLON, 0); }
		public SyntaxColonContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_syntaxColon; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).enterSyntaxColon(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof CqlParserListener ) ((CqlParserListener)listener).exitSyntaxColon(this);
		}
	}

	public final SyntaxColonContext syntaxColon() throws RecognitionException {
		SyntaxColonContext _localctx = new SyntaxColonContext(_ctx, getState());
		enterRule(_localctx, 580, RULE_syntaxColon);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(2439);
			match(COLON);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static final String _serializedATN =
		"\u0004\u0001\u00b7\u098a\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001"+
		"\u0002\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004"+
		"\u0002\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007"+
		"\u0002\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b"+
		"\u0002\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007"+
		"\u000f\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007"+
		"\u0012\u0002\u0013\u0007\u0013\u0002\u0014\u0007\u0014\u0002\u0015\u0007"+
		"\u0015\u0002\u0016\u0007\u0016\u0002\u0017\u0007\u0017\u0002\u0018\u0007"+
		"\u0018\u0002\u0019\u0007\u0019\u0002\u001a\u0007\u001a\u0002\u001b\u0007"+
		"\u001b\u0002\u001c\u0007\u001c\u0002\u001d\u0007\u001d\u0002\u001e\u0007"+
		"\u001e\u0002\u001f\u0007\u001f\u0002 \u0007 \u0002!\u0007!\u0002\"\u0007"+
		"\"\u0002#\u0007#\u0002$\u0007$\u0002%\u0007%\u0002&\u0007&\u0002\'\u0007"+
		"\'\u0002(\u0007(\u0002)\u0007)\u0002*\u0007*\u0002+\u0007+\u0002,\u0007"+
		",\u0002-\u0007-\u0002.\u0007.\u0002/\u0007/\u00020\u00070\u00021\u0007"+
		"1\u00022\u00072\u00023\u00073\u00024\u00074\u00025\u00075\u00026\u0007"+
		"6\u00027\u00077\u00028\u00078\u00029\u00079\u0002:\u0007:\u0002;\u0007"+
		";\u0002<\u0007<\u0002=\u0007=\u0002>\u0007>\u0002?\u0007?\u0002@\u0007"+
		"@\u0002A\u0007A\u0002B\u0007B\u0002C\u0007C\u0002D\u0007D\u0002E\u0007"+
		"E\u0002F\u0007F\u0002G\u0007G\u0002H\u0007H\u0002I\u0007I\u0002J\u0007"+
		"J\u0002K\u0007K\u0002L\u0007L\u0002M\u0007M\u0002N\u0007N\u0002O\u0007"+
		"O\u0002P\u0007P\u0002Q\u0007Q\u0002R\u0007R\u0002S\u0007S\u0002T\u0007"+
		"T\u0002U\u0007U\u0002V\u0007V\u0002W\u0007W\u0002X\u0007X\u0002Y\u0007"+
		"Y\u0002Z\u0007Z\u0002[\u0007[\u0002\\\u0007\\\u0002]\u0007]\u0002^\u0007"+
		"^\u0002_\u0007_\u0002`\u0007`\u0002a\u0007a\u0002b\u0007b\u0002c\u0007"+
		"c\u0002d\u0007d\u0002e\u0007e\u0002f\u0007f\u0002g\u0007g\u0002h\u0007"+
		"h\u0002i\u0007i\u0002j\u0007j\u0002k\u0007k\u0002l\u0007l\u0002m\u0007"+
		"m\u0002n\u0007n\u0002o\u0007o\u0002p\u0007p\u0002q\u0007q\u0002r\u0007"+
		"r\u0002s\u0007s\u0002t\u0007t\u0002u\u0007u\u0002v\u0007v\u0002w\u0007"+
		"w\u0002x\u0007x\u0002y\u0007y\u0002z\u0007z\u0002{\u0007{\u0002|\u0007"+
		"|\u0002}\u0007}\u0002~\u0007~\u0002\u007f\u0007\u007f\u0002\u0080\u0007"+
		"\u0080\u0002\u0081\u0007\u0081\u0002\u0082\u0007\u0082\u0002\u0083\u0007"+
		"\u0083\u0002\u0084\u0007\u0084\u0002\u0085\u0007\u0085\u0002\u0086\u0007"+
		"\u0086\u0002\u0087\u0007\u0087\u0002\u0088\u0007\u0088\u0002\u0089\u0007"+
		"\u0089\u0002\u008a\u0007\u008a\u0002\u008b\u0007\u008b\u0002\u008c\u0007"+
		"\u008c\u0002\u008d\u0007\u008d\u0002\u008e\u0007\u008e\u0002\u008f\u0007"+
		"\u008f\u0002\u0090\u0007\u0090\u0002\u0091\u0007\u0091\u0002\u0092\u0007"+
		"\u0092\u0002\u0093\u0007\u0093\u0002\u0094\u0007\u0094\u0002\u0095\u0007"+
		"\u0095\u0002\u0096\u0007\u0096\u0002\u0097\u0007\u0097\u0002\u0098\u0007"+
		"\u0098\u0002\u0099\u0007\u0099\u0002\u009a\u0007\u009a\u0002\u009b\u0007"+
		"\u009b\u0002\u009c\u0007\u009c\u0002\u009d\u0007\u009d\u0002\u009e\u0007"+
		"\u009e\u0002\u009f\u0007\u009f\u0002\u00a0\u0007\u00a0\u0002\u00a1\u0007"+
		"\u00a1\u0002\u00a2\u0007\u00a2\u0002\u00a3\u0007\u00a3\u0002\u00a4\u0007"+
		"\u00a4\u0002\u00a5\u0007\u00a5\u0002\u00a6\u0007\u00a6\u0002\u00a7\u0007"+
		"\u00a7\u0002\u00a8\u0007\u00a8\u0002\u00a9\u0007\u00a9\u0002\u00aa\u0007"+
		"\u00aa\u0002\u00ab\u0007\u00ab\u0002\u00ac\u0007\u00ac\u0002\u00ad\u0007"+
		"\u00ad\u0002\u00ae\u0007\u00ae\u0002\u00af\u0007\u00af\u0002\u00b0\u0007"+
		"\u00b0\u0002\u00b1\u0007\u00b1\u0002\u00b2\u0007\u00b2\u0002\u00b3\u0007"+
		"\u00b3\u0002\u00b4\u0007\u00b4\u0002\u00b5\u0007\u00b5\u0002\u00b6\u0007"+
		"\u00b6\u0002\u00b7\u0007\u00b7\u0002\u00b8\u0007\u00b8\u0002\u00b9\u0007"+
		"\u00b9\u0002\u00ba\u0007\u00ba\u0002\u00bb\u0007\u00bb\u0002\u00bc\u0007"+
		"\u00bc\u0002\u00bd\u0007\u00bd\u0002\u00be\u0007\u00be\u0002\u00bf\u0007"+
		"\u00bf\u0002\u00c0\u0007\u00c0\u0002\u00c1\u0007\u00c1\u0002\u00c2\u0007"+
		"\u00c2\u0002\u00c3\u0007\u00c3\u0002\u00c4\u0007\u00c4\u0002\u00c5\u0007"+
		"\u00c5\u0002\u00c6\u0007\u00c6\u0002\u00c7\u0007\u00c7\u0002\u00c8\u0007"+
		"\u00c8\u0002\u00c9\u0007\u00c9\u0002\u00ca\u0007\u00ca\u0002\u00cb\u0007"+
		"\u00cb\u0002\u00cc\u0007\u00cc\u0002\u00cd\u0007\u00cd\u0002\u00ce\u0007"+
		"\u00ce\u0002\u00cf\u0007\u00cf\u0002\u00d0\u0007\u00d0\u0002\u00d1\u0007"+
		"\u00d1\u0002\u00d2\u0007\u00d2\u0002\u00d3\u0007\u00d3\u0002\u00d4\u0007"+
		"\u00d4\u0002\u00d5\u0007\u00d5\u0002\u00d6\u0007\u00d6\u0002\u00d7\u0007"+
		"\u00d7\u0002\u00d8\u0007\u00d8\u0002\u00d9\u0007\u00d9\u0002\u00da\u0007"+
		"\u00da\u0002\u00db\u0007\u00db\u0002\u00dc\u0007\u00dc\u0002\u00dd\u0007"+
		"\u00dd\u0002\u00de\u0007\u00de\u0002\u00df\u0007\u00df\u0002\u00e0\u0007"+
		"\u00e0\u0002\u00e1\u0007\u00e1\u0002\u00e2\u0007\u00e2\u0002\u00e3\u0007"+
		"\u00e3\u0002\u00e4\u0007\u00e4\u0002\u00e5\u0007\u00e5\u0002\u00e6\u0007"+
		"\u00e6\u0002\u00e7\u0007\u00e7\u0002\u00e8\u0007\u00e8\u0002\u00e9\u0007"+
		"\u00e9\u0002\u00ea\u0007\u00ea\u0002\u00eb\u0007\u00eb\u0002\u00ec\u0007"+
		"\u00ec\u0002\u00ed\u0007\u00ed\u0002\u00ee\u0007\u00ee\u0002\u00ef\u0007"+
		"\u00ef\u0002\u00f0\u0007\u00f0\u0002\u00f1\u0007\u00f1\u0002\u00f2\u0007"+
		"\u00f2\u0002\u00f3\u0007\u00f3\u0002\u00f4\u0007\u00f4\u0002\u00f5\u0007"+
		"\u00f5\u0002\u00f6\u0007\u00f6\u0002\u00f7\u0007\u00f7\u0002\u00f8\u0007"+
		"\u00f8\u0002\u00f9\u0007\u00f9\u0002\u00fa\u0007\u00fa\u0002\u00fb\u0007"+
		"\u00fb\u0002\u00fc\u0007\u00fc\u0002\u00fd\u0007\u00fd\u0002\u00fe\u0007"+
		"\u00fe\u0002\u00ff\u0007\u00ff\u0002\u0100\u0007\u0100\u0002\u0101\u0007"+
		"\u0101\u0002\u0102\u0007\u0102\u0002\u0103\u0007\u0103\u0002\u0104\u0007"+
		"\u0104\u0002\u0105\u0007\u0105\u0002\u0106\u0007\u0106\u0002\u0107\u0007"+
		"\u0107\u0002\u0108\u0007\u0108\u0002\u0109\u0007\u0109\u0002\u010a\u0007"+
		"\u010a\u0002\u010b\u0007\u010b\u0002\u010c\u0007\u010c\u0002\u010d\u0007"+
		"\u010d\u0002\u010e\u0007\u010e\u0002\u010f\u0007\u010f\u0002\u0110\u0007"+
		"\u0110\u0002\u0111\u0007\u0111\u0002\u0112\u0007\u0112\u0002\u0113\u0007"+
		"\u0113\u0002\u0114\u0007\u0114\u0002\u0115\u0007\u0115\u0002\u0116\u0007"+
		"\u0116\u0002\u0117\u0007\u0117\u0002\u0118\u0007\u0118\u0002\u0119\u0007"+
		"\u0119\u0002\u011a\u0007\u011a\u0002\u011b\u0007\u011b\u0002\u011c\u0007"+
		"\u011c\u0002\u011d\u0007\u011d\u0002\u011e\u0007\u011e\u0002\u011f\u0007"+
		"\u011f\u0002\u0120\u0007\u0120\u0002\u0121\u0007\u0121\u0002\u0122\u0007"+
		"\u0122\u0001\u0000\u0003\u0000\u0248\b\u0000\u0001\u0000\u0003\u0000\u024b"+
		"\b\u0000\u0001\u0000\u0001\u0000\u0001\u0001\u0001\u0001\u0003\u0001\u0251"+
		"\b\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0005\u0001\u0256\b\u0001"+
		"\n\u0001\f\u0001\u0259\t\u0001\u0001\u0001\u0001\u0001\u0003\u0001\u025d"+
		"\b\u0001\u0001\u0001\u0003\u0001\u0260\b\u0001\u0001\u0001\u0003\u0001"+
		"\u0263\b\u0001\u0001\u0002\u0001\u0002\u0001\u0003\u0001\u0003\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004\u0001\u0004"+
		"\u0001\u0004\u0001\u0004\u0001\u0004\u0003\u0004\u0291\b\u0004\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0003\u0005\u029c\b\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0003\u0005"+
		"\u02a5\b\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0003\u0005\u02ae\b\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005"+
		"\u0003\u0005\u02b8\b\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0003\u0005\u02c2\b\u0005"+
		"\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005\u0001\u0005"+
		"\u0001\u0005\u0003\u0005\u02cb\b\u0005\u0001\u0005\u0001\u0005\u0003\u0005"+
		"\u02cf\b\u0005\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0007\u0001\u0007"+
		"\u0003\u0007\u02d6\b\u0007\u0001\b\u0001\b\u0001\b\u0001\b\u0001\b\u0001"+
		"\b\u0001\b\u0001\t\u0001\t\u0001\t\u0001\t\u0001\t\u0003\t\u02e4\b\t\u0001"+
		"\t\u0003\t\u02e7\b\t\u0001\n\u0001\n\u0001\n\u0001\n\u0001\n\u0003\n\u02ee"+
		"\b\n\u0001\n\u0001\n\u0001\n\u0003\n\u02f3\b\n\u0001\u000b\u0001\u000b"+
		"\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\f\u0001"+
		"\f\u0003\f\u02fe\b\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001\f\u0001"+
		"\f\u0001\f\u0003\f\u0308\b\f\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u0317"+
		"\b\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001"+
		"\r\u0003\r\u0322\b\r\u0001\r\u0001\r\u0001\r\u0003\r\u0327\b\r\u0001\r"+
		"\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0001\r\u0003\r\u0330\b\r\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0003\u000e\u0335\b\u000e\u0001\u000e\u0001"+
		"\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e\u0003\u000e\u033d"+
		"\b\u000e\u0001\u000f\u0001\u000f\u0001\u000f\u0003\u000f\u0342\b\u000f"+
		"\u0001\u000f\u0001\u000f\u0003\u000f\u0346\b\u000f\u0001\u0010\u0001\u0010"+
		"\u0001\u0010\u0003\u0010\u034b\b\u0010\u0001\u0010\u0001\u0010\u0001\u0010"+
		"\u0003\u0010\u0350\b\u0010\u0001\u0010\u0001\u0010\u0001\u0010\u0001\u0010"+
		"\u0001\u0010\u0001\u0011\u0001\u0011\u0001\u0011\u0001\u0011\u0001\u0011"+
		"\u0001\u0011\u0005\u0011\u035d\b\u0011\n\u0011\f\u0011\u0360\t\u0011\u0001"+
		"\u0012\u0001\u0012\u0001\u0012\u0003\u0012\u0365\b\u0012\u0001\u0012\u0001"+
		"\u0012\u0001\u0012\u0003\u0012\u036a\b\u0012\u0001\u0012\u0001\u0012\u0001"+
		"\u0012\u0001\u0012\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0003"+
		"\u0013\u0374\b\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0003\u0013\u0379"+
		"\b\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0003\u0013\u0383\b\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0003\u0013\u038f\b\u0013\u0001\u0014\u0001"+
		"\u0014\u0001\u0014\u0001\u0014\u0001\u0014\u0003\u0014\u0396\b\u0014\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0005\u0015\u039c\b\u0015\n"+
		"\u0015\f\u0015\u039f\t\u0015\u0001\u0016\u0001\u0016\u0001\u0016\u0001"+
		"\u0016\u0001\u0016\u0001\u0017\u0001\u0017\u0001\u0017\u0001\u0017\u0001"+
		"\u0017\u0001\u0017\u0003\u0017\u03ac\b\u0017\u0001\u0018\u0001\u0018\u0001"+
		"\u0018\u0003\u0018\u03b1\b\u0018\u0001\u0018\u0001\u0018\u0001\u0018\u0001"+
		"\u0018\u0001\u0018\u0001\u0018\u0001\u0018\u0001\u0018\u0001\u0018\u0001"+
		"\u0018\u0003\u0018\u03bd\b\u0018\u0001\u0019\u0001\u0019\u0003\u0019\u03c1"+
		"\b\u0019\u0001\u0019\u0001\u0019\u0003\u0019\u03c5\b\u0019\u0001\u0019"+
		"\u0001\u0019\u0001\u0019\u0003\u0019\u03ca\b\u0019\u0001\u0019\u0001\u0019"+
		"\u0001\u0019\u0003\u0019\u03cf\b\u0019\u0001\u0019\u0001\u0019\u0001\u0019"+
		"\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019\u0001\u0019"+
		"\u0001\u001a\u0001\u001a\u0001\u001b\u0001\u001b\u0001\u001b\u0001\u001b"+
		"\u0005\u001b\u03e0\b\u001b\n\u001b\f\u001b\u03e3\t\u001b\u0001\u001c\u0001"+
		"\u001c\u0001\u001c\u0001\u001c\u0003\u001c\u03e9\b\u001c\u0001\u001c\u0001"+
		"\u001c\u0001\u001c\u0001\u001c\u0001\u001d\u0001\u001d\u0003\u001d\u03f1"+
		"\b\u001d\u0001\u001d\u0001\u001d\u0003\u001d\u03f5\b\u001d\u0001\u001d"+
		"\u0001\u001d\u0001\u001d\u0003\u001d\u03fa\b\u001d\u0001\u001d\u0001\u001d"+
		"\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001d"+
		"\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001d\u0001\u001e"+
		"\u0001\u001e\u0001\u001e\u0001\u001e\u0003\u001e\u040d\b\u001e\u0001\u001f"+
		"\u0001\u001f\u0001\u001f\u0001\u001f\u0001\u001f\u0005\u001f\u0414\b\u001f"+
		"\n\u001f\f\u001f\u0417\t\u001f\u0001\u001f\u0001\u001f\u0001 \u0001 \u0001"+
		" \u0001 \u0001!\u0001!\u0001!\u0001!\u0001!\u0001!\u0005!\u0425\b!\n!"+
		"\f!\u0428\t!\u0001!\u0001!\u0001\"\u0001\"\u0001\"\u0001\"\u0001\"\u0005"+
		"\"\u0431\b\"\n\"\f\"\u0434\t\"\u0001\"\u0001\"\u0001#\u0001#\u0001#\u0001"+
		"$\u0001$\u0001$\u0001$\u0001$\u0001$\u0003$\u0441\b$\u0001%\u0001%\u0001"+
		"%\u0001&\u0001&\u0003&\u0448\b&\u0001\'\u0001\'\u0001\'\u0001\'\u0001"+
		"\'\u0003\'\u044f\b\'\u0001\'\u0001\'\u0001\'\u0001(\u0001(\u0001(\u0003"+
		"(\u0457\b(\u0001)\u0001)\u0001)\u0001*\u0001*\u0001*\u0001*\u0005*\u0460"+
		"\b*\n*\f*\u0463\t*\u0001+\u0001+\u0001+\u0001+\u0001,\u0001,\u0001,\u0001"+
		",\u0001,\u0001,\u0001,\u0005,\u0470\b,\n,\f,\u0473\t,\u0001-\u0001-\u0001"+
		"-\u0001-\u0001-\u0001.\u0001.\u0001.\u0001.\u0001.\u0003.\u047f\b.\u0001"+
		".\u0001.\u0001.\u0001/\u0001/\u0001/\u0001/\u0001/\u0003/\u0489\b/\u0001"+
		"0\u00010\u00010\u00011\u00011\u00011\u00011\u00011\u00012\u00012\u0001"+
		"2\u00012\u00013\u00013\u00013\u00014\u00014\u00014\u00014\u00054\u049e"+
		"\b4\n4\f4\u04a1\t4\u00015\u00015\u00015\u00016\u00016\u00016\u00016\u0001"+
		"6\u00016\u00056\u04ac\b6\n6\f6\u04af\t6\u00017\u00017\u00017\u00017\u0003"+
		"7\u04b5\b7\u00018\u00018\u00018\u00018\u00018\u00058\u04bc\b8\n8\f8\u04bf"+
		"\t8\u00019\u00019\u00019\u00019\u00019\u00019\u00019\u00019\u00019\u0001"+
		"9\u00019\u00019\u00019\u00019\u00019\u00019\u00039\u04d1\b9\u0001:\u0001"+
		":\u0001:\u0001:\u0001:\u0001:\u0003:\u04d9\b:\u0001:\u0001:\u0001:\u0001"+
		":\u0003:\u04df\b:\u0001;\u0001;\u0001;\u0003;\u04e4\b;\u0001;\u0001;\u0001"+
		"<\u0001<\u0001<\u0003<\u04eb\b<\u0001<\u0001<\u0001<\u0003<\u04f0\b<\u0001"+
		"<\u0001<\u0001=\u0001=\u0001=\u0001=\u0003=\u04f8\b=\u0001=\u0001=\u0001"+
		"=\u0003=\u04fd\b=\u0001=\u0001=\u0001>\u0001>\u0001>\u0003>\u0504\b>\u0001"+
		">\u0001>\u0001>\u0003>\u0509\b>\u0001>\u0001>\u0001?\u0001?\u0001?\u0003"+
		"?\u0510\b?\u0001?\u0001?\u0001?\u0003?\u0515\b?\u0001?\u0001?\u0001@\u0001"+
		"@\u0001@\u0003@\u051c\b@\u0001@\u0001@\u0001@\u0001@\u0001@\u0003@\u0523"+
		"\b@\u0001@\u0001@\u0001A\u0001A\u0001A\u0003A\u052a\bA\u0001A\u0001A\u0001"+
		"B\u0001B\u0001B\u0003B\u0531\bB\u0001B\u0001B\u0001B\u0003B\u0536\bB\u0001"+
		"B\u0001B\u0001C\u0001C\u0001C\u0003C\u053d\bC\u0001C\u0001C\u0001D\u0001"+
		"D\u0001D\u0003D\u0544\bD\u0001D\u0001D\u0001D\u0003D\u0549\bD\u0001D\u0001"+
		"D\u0001E\u0001E\u0001E\u0003E\u0550\bE\u0001E\u0001E\u0001E\u0003E\u0555"+
		"\bE\u0001E\u0001E\u0001E\u0001E\u0001E\u0003E\u055c\bE\u0001F\u0001F\u0001"+
		"F\u0001G\u0001G\u0001G\u0001G\u0001G\u0003G\u0566\bG\u0001G\u0001G\u0001"+
		"G\u0001G\u0003G\u056c\bG\u0001G\u0001G\u0001G\u0001G\u0005G\u0572\bG\n"+
		"G\fG\u0575\tG\u0003G\u0577\bG\u0001H\u0001H\u0001H\u0001H\u0001H\u0001"+
		"H\u0003H\u057f\bH\u0001H\u0001H\u0001H\u0003H\u0584\bH\u0005H\u0586\b"+
		"H\nH\fH\u0589\tH\u0001H\u0001H\u0001I\u0001I\u0001I\u0001I\u0001I\u0001"+
		"I\u0001I\u0001I\u0003I\u0595\bI\u0001J\u0001J\u0001K\u0001K\u0003K\u059b"+
		"\bK\u0001L\u0001L\u0001L\u0001L\u0001L\u0005L\u05a2\bL\nL\fL\u05a5\tL"+
		"\u0001L\u0001L\u0001M\u0001M\u0001M\u0001M\u0001N\u0001N\u0001O\u0001"+
		"O\u0003O\u05b1\bO\u0001P\u0001P\u0001P\u0001P\u0005P\u05b7\bP\nP\fP\u05ba"+
		"\tP\u0001P\u0001P\u0001P\u0003P\u05bf\bP\u0001Q\u0001Q\u0001Q\u0003Q\u05c4"+
		"\bQ\u0001R\u0001R\u0001R\u0001S\u0001S\u0001S\u0001S\u0001S\u0001S\u0001"+
		"T\u0001T\u0001T\u0003T\u05d2\bT\u0001U\u0001U\u0001V\u0001V\u0001V\u0001"+
		"V\u0001W\u0001W\u0001W\u0001W\u0001W\u0001W\u0001X\u0001X\u0001X\u0001"+
		"X\u0005X\u05e4\bX\nX\fX\u05e7\tX\u0001Y\u0001Y\u0001Y\u0001Y\u0005Y\u05ed"+
		"\bY\nY\fY\u05f0\tY\u0001Z\u0001Z\u0001[\u0001[\u0001\\\u0001\\\u0001\\"+
		"\u0001]\u0001]\u0003]\u05fb\b]\u0001]\u0001]\u0003]\u05ff\b]\u0001^\u0001"+
		"^\u0003^\u0603\b^\u0001_\u0001_\u0001_\u0001_\u0001_\u0001_\u0001_\u0001"+
		"_\u0001_\u0001_\u0001_\u0001_\u0003_\u0611\b_\u0001`\u0001`\u0001`\u0001"+
		"`\u0005`\u0617\b`\n`\f`\u061a\t`\u0001a\u0001a\u0001a\u0001a\u0001a\u0001"+
		"a\u0003a\u0622\ba\u0001b\u0001b\u0001b\u0001b\u0001c\u0001c\u0001c\u0001"+
		"d\u0001d\u0003d\u062d\bd\u0001d\u0001d\u0001d\u0003d\u0632\bd\u0001d\u0001"+
		"d\u0001e\u0001e\u0001e\u0003e\u0639\be\u0001e\u0003e\u063c\be\u0001e\u0001"+
		"e\u0001e\u0001e\u0003e\u0642\be\u0001e\u0001e\u0001e\u0001e\u0001e\u0001"+
		"f\u0001f\u0003f\u064b\bf\u0001g\u0001g\u0001g\u0001g\u0003g\u0651\bg\u0001"+
		"h\u0001h\u0001h\u0001h\u0001h\u0001i\u0001i\u0001i\u0001i\u0001i\u0001"+
		"j\u0001j\u0001j\u0001j\u0001j\u0001k\u0003k\u0663\bk\u0001k\u0001k\u0003"+
		"k\u0667\bk\u0001k\u0001k\u0003k\u066b\bk\u0001k\u0001k\u0001k\u0003k\u0670"+
		"\bk\u0001l\u0001l\u0001l\u0001l\u0005l\u0676\bl\nl\fl\u0679\tl\u0001m"+
		"\u0001m\u0001m\u0001m\u0001m\u0003m\u0680\bm\u0001m\u0001m\u0003m\u0684"+
		"\bm\u0001n\u0003n\u0687\bn\u0001n\u0001n\u0001n\u0001n\u0003n\u068d\b"+
		"n\u0001n\u0001n\u0003n\u0691\bn\u0001n\u0001n\u0001n\u0001n\u0001n\u0003"+
		"n\u0698\bn\u0001o\u0001o\u0001o\u0001p\u0001p\u0001p\u0001p\u0005p\u06a1"+
		"\bp\np\fp\u06a4\tp\u0001q\u0001q\u0001q\u0001q\u0001r\u0001r\u0001r\u0001"+
		"r\u0005r\u06ae\br\nr\fr\u06b1\tr\u0001s\u0001s\u0001s\u0001s\u0001s\u0001"+
		"s\u0003s\u06b9\bs\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001"+
		"s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001"+
		"s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001"+
		"s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001"+
		"s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0001s\u0003s\u06e8\bs\u0001"+
		"t\u0001t\u0001t\u0001t\u0001t\u0005t\u06ef\bt\nt\ft\u06f2\tt\u0003t\u06f4"+
		"\bt\u0001t\u0001t\u0001u\u0001u\u0001u\u0001u\u0001u\u0001u\u0001u\u0001"+
		"u\u0001u\u0001u\u0005u\u0702\bu\nu\fu\u0705\tu\u0001u\u0001u\u0001v\u0001"+
		"v\u0001v\u0001v\u0001v\u0005v\u070e\bv\nv\fv\u0711\tv\u0001v\u0001v\u0001"+
		"w\u0001w\u0001w\u0001w\u0001w\u0005w\u071a\bw\nw\fw\u071d\tw\u0001w\u0001"+
		"w\u0001x\u0003x\u0722\bx\u0001x\u0001x\u0001x\u0001x\u0001x\u0003x\u0729"+
		"\bx\u0001x\u0001x\u0003x\u072d\bx\u0001x\u0001x\u0003x\u0731\bx\u0001"+
		"x\u0003x\u0734\bx\u0001y\u0001y\u0001y\u0001y\u0001y\u0001y\u0001y\u0001"+
		"y\u0001y\u0001y\u0001y\u0001y\u0001y\u0001y\u0001y\u0001y\u0003y\u0746"+
		"\by\u0001z\u0001z\u0001z\u0001{\u0001{\u0001{\u0001|\u0001|\u0001|\u0001"+
		"}\u0001}\u0001}\u0001}\u0001~\u0001~\u0001~\u0001\u007f\u0001\u007f\u0001"+
		"\u007f\u0001\u007f\u0001\u007f\u0001\u007f\u0001\u007f\u0001\u007f\u0003"+
		"\u007f\u0760\b\u007f\u0001\u0080\u0001\u0080\u0001\u0080\u0001\u0080\u0001"+
		"\u0081\u0001\u0081\u0001\u0081\u0001\u0081\u0005\u0081\u076a\b\u0081\n"+
		"\u0081\f\u0081\u076d\t\u0081\u0001\u0082\u0001\u0082\u0001\u0082\u0001"+
		"\u0082\u0005\u0082\u0773\b\u0082\n\u0082\f\u0082\u0776\t\u0082\u0001\u0083"+
		"\u0001\u0083\u0001\u0083\u0001\u0083\u0001\u0083\u0001\u0083\u0003\u0083"+
		"\u077e\b\u0083\u0001\u0084\u0001\u0084\u0003\u0084\u0782\b\u0084\u0001"+
		"\u0084\u0003\u0084\u0785\b\u0084\u0001\u0084\u0001\u0084\u0001\u0084\u0003"+
		"\u0084\u078a\b\u0084\u0001\u0084\u0003\u0084\u078d\b\u0084\u0001\u0084"+
		"\u0003\u0084\u0790\b\u0084\u0001\u0084\u0003\u0084\u0793\b\u0084\u0001"+
		"\u0085\u0001\u0085\u0001\u0085\u0001\u0086\u0001\u0086\u0001\u0086\u0001"+
		"\u0087\u0001\u0087\u0001\u0087\u0001\u0088\u0001\u0088\u0001\u0088\u0001"+
		"\u0088\u0003\u0088\u07a2\b\u0088\u0001\u0089\u0001\u0089\u0001\u0089\u0001"+
		"\u0089\u0001\u008a\u0001\u008a\u0001\u008a\u0003\u008a\u07ab\b\u008a\u0001"+
		"\u008b\u0001\u008b\u0001\u008b\u0001\u008c\u0001\u008c\u0001\u008d\u0001"+
		"\u008d\u0003\u008d\u07b4\b\u008d\u0001\u008d\u0001\u008d\u0001\u008d\u0005"+
		"\u008d\u07b9\b\u008d\n\u008d\f\u008d\u07bc\t\u008d\u0001\u008e\u0001\u008e"+
		"\u0001\u008e\u0001\u008e\u0001\u008e\u0001\u008e\u0001\u008e\u0003\u008e"+
		"\u07c5\b\u008e\u0001\u008e\u0001\u008e\u0001\u008e\u0001\u008e\u0003\u008e"+
		"\u07cb\b\u008e\u0003\u008e\u07cd\b\u008e\u0001\u008f\u0001\u008f\u0001"+
		"\u008f\u0001\u008f\u0005\u008f\u07d3\b\u008f\n\u008f\f\u008f\u07d6\t\u008f"+
		"\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090"+
		"\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090"+
		"\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090"+
		"\u0001\u0090\u0001\u0090\u0003\u0090\u07ec\b\u0090\u0001\u0090\u0001\u0090"+
		"\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0005\u0090"+
		"\u07f5\b\u0090\n\u0090\f\u0090\u07f8\t\u0090\u0001\u0090\u0001\u0090\u0001"+
		"\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0005\u0090\u0801"+
		"\b\u0090\n\u0090\f\u0090\u0804\t\u0090\u0001\u0090\u0001\u0090\u0001\u0090"+
		"\u0001\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0005\u0090\u080d\b\u0090"+
		"\n\u0090\f\u0090\u0810\t\u0090\u0001\u0090\u0001\u0090\u0001\u0090\u0001"+
		"\u0090\u0001\u0090\u0001\u0090\u0005\u0090\u0818\b\u0090\n\u0090\f\u0090"+
		"\u081b\t\u0090\u0001\u0090\u0001\u0090\u0003\u0090\u081f\b\u0090\u0001"+
		"\u0091\u0001\u0091\u0001\u0091\u0001\u0091\u0001\u0092\u0001\u0092\u0001"+
		"\u0092\u0001\u0092\u0001\u0092\u0001\u0092\u0001\u0093\u0001\u0093\u0001"+
		"\u0093\u0001\u0093\u0001\u0093\u0001\u0093\u0001\u0093\u0003\u0093\u0832"+
		"\b\u0093\u0001\u0093\u0001\u0093\u0001\u0093\u0001\u0093\u0003\u0093\u0838"+
		"\b\u0093\u0001\u0094\u0001\u0094\u0001\u0094\u0003\u0094\u083d\b\u0094"+
		"\u0001\u0094\u0001\u0094\u0001\u0094\u0001\u0094\u0003\u0094\u0843\b\u0094"+
		"\u0005\u0094\u0845\b\u0094\n\u0094\f\u0094\u0848\t\u0094\u0001\u0095\u0001"+
		"\u0095\u0001\u0095\u0001\u0095\u0001\u0095\u0001\u0095\u0001\u0095\u0001"+
		"\u0095\u0003\u0095\u0852\b\u0095\u0001\u0096\u0001\u0096\u0001\u0097\u0001"+
		"\u0097\u0001\u0098\u0001\u0098\u0001\u0099\u0001\u0099\u0001\u009a\u0001"+
		"\u009a\u0001\u009b\u0001\u009b\u0001\u009b\u0001\u009b\u0003\u009b\u0862"+
		"\b\u009b\u0001\u009c\u0001\u009c\u0001\u009c\u0001\u009c\u0003\u009c\u0868"+
		"\b\u009c\u0001\u009d\u0001\u009d\u0001\u009d\u0001\u009d\u0003\u009d\u086e"+
		"\b\u009d\u0001\u009e\u0001\u009e\u0003\u009e\u0872\b\u009e\u0001\u009f"+
		"\u0001\u009f\u0001\u00a0\u0001\u00a0\u0001\u00a0\u0001\u00a0\u0001\u00a0"+
		"\u0005\u00a0\u087b\b\u00a0\n\u00a0\f\u00a0\u087e\t\u00a0\u0001\u00a0\u0001"+
		"\u00a0\u0001\u00a1\u0001\u00a1\u0003\u00a1\u0884\b\u00a1\u0001\u00a2\u0001"+
		"\u00a2\u0001\u00a3\u0001\u00a3\u0001\u00a4\u0001\u00a4\u0001\u00a5\u0001"+
		"\u00a5\u0001\u00a6\u0001\u00a6\u0001\u00a7\u0001\u00a7\u0001\u00a8\u0001"+
		"\u00a8\u0001\u00a9\u0001\u00a9\u0001\u00aa\u0001\u00aa\u0001\u00ab\u0001"+
		"\u00ab\u0001\u00ac\u0001\u00ac\u0001\u00ad\u0001\u00ad\u0001\u00ad\u0001"+
		"\u00ae\u0001\u00ae\u0001\u00af\u0001\u00af\u0001\u00b0\u0001\u00b0\u0001"+
		"\u00b1\u0001\u00b1\u0001\u00b2\u0001\u00b2\u0001\u00b3\u0001\u00b3\u0001"+
		"\u00b3\u0001\u00b4\u0001\u00b4\u0001\u00b5\u0001\u00b5\u0001\u00b6\u0001"+
		"\u00b6\u0001\u00b7\u0001\u00b7\u0001\u00b8\u0001\u00b8\u0001\u00b9\u0001"+
		"\u00b9\u0001\u00ba\u0001\u00ba\u0001\u00bb\u0001\u00bb\u0001\u00bc\u0001"+
		"\u00bc\u0001\u00bd\u0001\u00bd\u0001\u00be\u0001\u00be\u0001\u00bf\u0001"+
		"\u00bf\u0001\u00c0\u0001\u00c0\u0001\u00c1\u0001\u00c1\u0001\u00c2\u0001"+
		"\u00c2\u0001\u00c3\u0001\u00c3\u0001\u00c4\u0001\u00c4\u0001\u00c5\u0001"+
		"\u00c5\u0001\u00c6\u0001\u00c6\u0001\u00c7\u0001\u00c7\u0001\u00c8\u0001"+
		"\u00c8\u0001\u00c9\u0001\u00c9\u0001\u00ca\u0001\u00ca\u0001\u00cb\u0001"+
		"\u00cb\u0001\u00cc\u0001\u00cc\u0001\u00cd\u0001\u00cd\u0001\u00ce\u0001"+
		"\u00ce\u0001\u00cf\u0001\u00cf\u0001\u00d0\u0001\u00d0\u0001\u00d1\u0001"+
		"\u00d1\u0001\u00d2\u0001\u00d2\u0001\u00d3\u0001\u00d3\u0001\u00d4\u0001"+
		"\u00d4\u0001\u00d5\u0001\u00d5\u0001\u00d6\u0001\u00d6\u0001\u00d7\u0001"+
		"\u00d7\u0001\u00d8\u0001\u00d8\u0001\u00d9\u0001\u00d9\u0001\u00da\u0001"+
		"\u00da\u0001\u00db\u0001\u00db\u0001\u00dc\u0001\u00dc\u0001\u00dd\u0001"+
		"\u00dd\u0001\u00de\u0001\u00de\u0001\u00df\u0001\u00df\u0001\u00e0\u0001"+
		"\u00e0\u0001\u00e1\u0001\u00e1\u0001\u00e2\u0001\u00e2\u0001\u00e3\u0001"+
		"\u00e3\u0001\u00e4\u0001\u00e4\u0001\u00e5\u0001\u00e5\u0001\u00e6\u0001"+
		"\u00e6\u0001\u00e7\u0001\u00e7\u0001\u00e8\u0001\u00e8\u0001\u00e9\u0001"+
		"\u00e9\u0001\u00ea\u0001\u00ea\u0001\u00eb\u0001\u00eb\u0001\u00ec\u0001"+
		"\u00ec\u0001\u00ed\u0001\u00ed\u0001\u00ee\u0001\u00ee\u0001\u00ef\u0001"+
		"\u00ef\u0001\u00f0\u0001\u00f0\u0001\u00f1\u0001\u00f1\u0001\u00f2\u0001"+
		"\u00f2\u0001\u00f3\u0001\u00f3\u0001\u00f4\u0001\u00f4\u0001\u00f5\u0001"+
		"\u00f5\u0001\u00f6\u0001\u00f6\u0001\u00f7\u0001\u00f7\u0001\u00f8\u0001"+
		"\u00f8\u0001\u00f9\u0001\u00f9\u0001\u00fa\u0001\u00fa\u0001\u00fb\u0001"+
		"\u00fb\u0001\u00fc\u0001\u00fc\u0001\u00fd\u0001\u00fd\u0001\u00fe\u0001"+
		"\u00fe\u0001\u00ff\u0001\u00ff\u0001\u0100\u0001\u0100\u0001\u0101\u0001"+
		"\u0101\u0001\u0102\u0001\u0102\u0001\u0103\u0001\u0103\u0001\u0104\u0001"+
		"\u0104\u0001\u0105\u0001\u0105\u0001\u0106\u0001\u0106\u0001\u0107\u0001"+
		"\u0107\u0001\u0108\u0001\u0108\u0001\u0109\u0001\u0109\u0001\u010a\u0001"+
		"\u010a\u0001\u010b\u0001\u010b\u0001\u010c\u0001\u010c\u0001\u010d\u0001"+
		"\u010d\u0001\u010e\u0001\u010e\u0001\u010f\u0001\u010f\u0001\u0110\u0001"+
		"\u0110\u0001\u0111\u0001\u0111\u0001\u0112\u0001\u0112\u0001\u0113\u0001"+
		"\u0113\u0001\u0114\u0001\u0114\u0001\u0115\u0001\u0115\u0001\u0116\u0001"+
		"\u0116\u0001\u0117\u0001\u0117\u0001\u0118\u0001\u0118\u0001\u0119\u0001"+
		"\u0119\u0001\u011a\u0001\u011a\u0001\u011b\u0001\u011b\u0001\u011c\u0001"+
		"\u011c\u0001\u011d\u0001\u011d\u0001\u011e\u0001\u011e\u0001\u011f\u0001"+
		"\u011f\u0001\u0120\u0001\u0120\u0001\u0121\u0001\u0121\u0001\u0122\u0001"+
		"\u0122\u0001\u0122\u0000\u0000\u0123\u0000\u0002\u0004\u0006\b\n\f\u000e"+
		"\u0010\u0012\u0014\u0016\u0018\u001a\u001c\u001e \"$&(*,.02468:<>@BDF"+
		"HJLNPRTVXZ\\^`bdfhjlnprtvxz|~\u0080\u0082\u0084\u0086\u0088\u008a\u008c"+
		"\u008e\u0090\u0092\u0094\u0096\u0098\u009a\u009c\u009e\u00a0\u00a2\u00a4"+
		"\u00a6\u00a8\u00aa\u00ac\u00ae\u00b0\u00b2\u00b4\u00b6\u00b8\u00ba\u00bc"+
		"\u00be\u00c0\u00c2\u00c4\u00c6\u00c8\u00ca\u00cc\u00ce\u00d0\u00d2\u00d4"+
		"\u00d6\u00d8\u00da\u00dc\u00de\u00e0\u00e2\u00e4\u00e6\u00e8\u00ea\u00ec"+
		"\u00ee\u00f0\u00f2\u00f4\u00f6\u00f8\u00fa\u00fc\u00fe\u0100\u0102\u0104"+
		"\u0106\u0108\u010a\u010c\u010e\u0110\u0112\u0114\u0116\u0118\u011a\u011c"+
		"\u011e\u0120\u0122\u0124\u0126\u0128\u012a\u012c\u012e\u0130\u0132\u0134"+
		"\u0136\u0138\u013a\u013c\u013e\u0140\u0142\u0144\u0146\u0148\u014a\u014c"+
		"\u014e\u0150\u0152\u0154\u0156\u0158\u015a\u015c\u015e\u0160\u0162\u0164"+
		"\u0166\u0168\u016a\u016c\u016e\u0170\u0172\u0174\u0176\u0178\u017a\u017c"+
		"\u017e\u0180\u0182\u0184\u0186\u0188\u018a\u018c\u018e\u0190\u0192\u0194"+
		"\u0196\u0198\u019a\u019c\u019e\u01a0\u01a2\u01a4\u01a6\u01a8\u01aa\u01ac"+
		"\u01ae\u01b0\u01b2\u01b4\u01b6\u01b8\u01ba\u01bc\u01be\u01c0\u01c2\u01c4"+
		"\u01c6\u01c8\u01ca\u01cc\u01ce\u01d0\u01d2\u01d4\u01d6\u01d8\u01da\u01dc"+
		"\u01de\u01e0\u01e2\u01e4\u01e6\u01e8\u01ea\u01ec\u01ee\u01f0\u01f2\u01f4"+
		"\u01f6\u01f8\u01fa\u01fc\u01fe\u0200\u0202\u0204\u0206\u0208\u020a\u020c"+
		"\u020e\u0210\u0212\u0214\u0216\u0218\u021a\u021c\u021e\u0220\u0222\u0224"+
		"\u0226\u0228\u022a\u022c\u022e\u0230\u0232\u0234\u0236\u0238\u023a\u023c"+
		"\u023e\u0240\u0242\u0244\u0000\n\u0001\u0000\u00ac\u00ad\u0002\u0000\u000e"+
		"\u000e\u0010\u0010\u0001\u0000\u0013\u0017\u0001\u0000\u00ae\u00af\u0002"+
		"\u0000;;\u0085\u0085\u0005\u0000xx\u0081\u0081\u0090\u0090\u0096\u00ab"+
		"\u00b2\u00b2\u0002\u0000HH\u00b2\u00b2\t\u0000\u001b\u001b\u001f\u001f"+
		"77TVbbmmww\u0080\u0080\u0088\u0088\u0002\u0000~~\u0096\u0096\u0001\u0000"+
		"23\u0990\u0000\u0247\u0001\u0000\u0000\u0000\u0002\u0257\u0001\u0000\u0000"+
		"\u0000\u0004\u0264\u0001\u0000\u0000\u0000\u0006\u0266\u0001\u0000\u0000"+
		"\u0000\b\u0290\u0001\u0000\u0000\u0000\n\u0292\u0001\u0000\u0000\u0000"+
		"\f\u02d0\u0001\u0000\u0000\u0000\u000e\u02d3\u0001\u0000\u0000\u0000\u0010"+
		"\u02d7\u0001\u0000\u0000\u0000\u0012\u02de\u0001\u0000\u0000\u0000\u0014"+
		"\u02e8\u0001\u0000\u0000\u0000\u0016\u02f4\u0001\u0000\u0000\u0000\u0018"+
		"\u0307\u0001\u0000\u0000\u0000\u001a\u032f\u0001\u0000\u0000\u0000\u001c"+
		"\u0331\u0001\u0000\u0000\u0000\u001e\u033e\u0001\u0000\u0000\u0000 \u0347"+
		"\u0001\u0000\u0000\u0000\"\u0356\u0001\u0000\u0000\u0000$\u0361\u0001"+
		"\u0000\u0000\u0000&\u036f\u0001\u0000\u0000\u0000(\u0390\u0001\u0000\u0000"+
		"\u0000*\u0397\u0001\u0000\u0000\u0000,\u03a0\u0001\u0000\u0000\u0000."+
		"\u03ab\u0001\u0000\u0000\u00000\u03ad\u0001\u0000\u0000\u00002\u03be\u0001"+
		"\u0000\u0000\u00004\u03d9\u0001\u0000\u0000\u00006\u03db\u0001\u0000\u0000"+
		"\u00008\u03e8\u0001\u0000\u0000\u0000:\u03ee\u0001\u0000\u0000\u0000<"+
		"\u040c\u0001\u0000\u0000\u0000>\u040e\u0001\u0000\u0000\u0000@\u041a\u0001"+
		"\u0000\u0000\u0000B\u041e\u0001\u0000\u0000\u0000D\u042b\u0001\u0000\u0000"+
		"\u0000F\u0437\u0001\u0000\u0000\u0000H\u043a\u0001\u0000\u0000\u0000J"+
		"\u0442\u0001\u0000\u0000\u0000L\u0447\u0001\u0000\u0000\u0000N\u0449\u0001"+
		"\u0000\u0000\u0000P\u0456\u0001\u0000\u0000\u0000R\u0458\u0001\u0000\u0000"+
		"\u0000T\u045b\u0001\u0000\u0000\u0000V\u0464\u0001\u0000\u0000\u0000X"+
		"\u0468\u0001\u0000\u0000\u0000Z\u0474\u0001\u0000\u0000\u0000\\\u0479"+
		"\u0001\u0000\u0000\u0000^\u0488\u0001\u0000\u0000\u0000`\u048a\u0001\u0000"+
		"\u0000\u0000b\u048d\u0001\u0000\u0000\u0000d\u0492\u0001\u0000\u0000\u0000"+
		"f\u0496\u0001\u0000\u0000\u0000h\u0499\u0001\u0000\u0000\u0000j\u04a2"+
		"\u0001\u0000\u0000\u0000l\u04a5\u0001\u0000\u0000\u0000n\u04b0\u0001\u0000"+
		"\u0000\u0000p\u04b6\u0001\u0000\u0000\u0000r\u04d0\u0001\u0000\u0000\u0000"+
		"t\u04d2\u0001\u0000\u0000\u0000v\u04e0\u0001\u0000\u0000\u0000x\u04e7"+
		"\u0001\u0000\u0000\u0000z\u04f3\u0001\u0000\u0000\u0000|\u0500\u0001\u0000"+
		"\u0000\u0000~\u050c\u0001\u0000\u0000\u0000\u0080\u0518\u0001\u0000\u0000"+
		"\u0000\u0082\u0526\u0001\u0000\u0000\u0000\u0084\u052d\u0001\u0000\u0000"+
		"\u0000\u0086\u0539\u0001\u0000\u0000\u0000\u0088\u0540\u0001\u0000\u0000"+
		"\u0000\u008a\u054c\u0001\u0000\u0000\u0000\u008c\u055d\u0001\u0000\u0000"+
		"\u0000\u008e\u0576\u0001\u0000\u0000\u0000\u0090\u0578\u0001\u0000\u0000"+
		"\u0000\u0092\u0594\u0001\u0000\u0000\u0000\u0094\u0596\u0001\u0000\u0000"+
		"\u0000\u0096\u059a\u0001\u0000\u0000\u0000\u0098\u059c\u0001\u0000\u0000"+
		"\u0000\u009a\u05a8\u0001\u0000\u0000\u0000\u009c\u05ac\u0001\u0000\u0000"+
		"\u0000\u009e\u05b0\u0001\u0000\u0000\u0000\u00a0\u05b2\u0001\u0000\u0000"+
		"\u0000\u00a2\u05c0\u0001\u0000\u0000\u0000\u00a4\u05c5\u0001\u0000\u0000"+
		"\u0000\u00a6\u05c8\u0001\u0000\u0000\u0000\u00a8\u05d1\u0001\u0000\u0000"+
		"\u0000\u00aa\u05d3\u0001\u0000\u0000\u0000\u00ac\u05d5\u0001\u0000\u0000"+
		"\u0000\u00ae\u05d9\u0001\u0000\u0000\u0000\u00b0\u05df\u0001\u0000\u0000"+
		"\u0000\u00b2\u05e8\u0001\u0000\u0000\u0000\u00b4\u05f1\u0001\u0000\u0000"+
		"\u0000\u00b6\u05f3\u0001\u0000\u0000\u0000\u00b8\u05f5\u0001\u0000\u0000"+
		"\u0000\u00ba\u05f8\u0001\u0000\u0000\u0000\u00bc\u0602\u0001\u0000\u0000"+
		"\u0000\u00be\u0604\u0001\u0000\u0000\u0000\u00c0\u0612\u0001\u0000\u0000"+
		"\u0000\u00c2\u0621\u0001\u0000\u0000\u0000\u00c4\u0623\u0001\u0000\u0000"+
		"\u0000\u00c6\u0627\u0001\u0000\u0000\u0000\u00c8\u062a\u0001\u0000\u0000"+
		"\u0000\u00ca\u0635\u0001\u0000\u0000\u0000\u00cc\u064a\u0001\u0000\u0000"+
		"\u0000\u00ce\u0650\u0001\u0000\u0000\u0000\u00d0\u0652\u0001\u0000\u0000"+
		"\u0000\u00d2\u0657\u0001\u0000\u0000\u0000\u00d4\u065c\u0001\u0000\u0000"+
		"\u0000\u00d6\u0662\u0001\u0000\u0000\u0000\u00d8\u0671\u0001\u0000\u0000"+
		"\u0000\u00da\u0683\u0001\u0000\u0000\u0000\u00dc\u0686\u0001\u0000\u0000"+
		"\u0000\u00de\u0699\u0001\u0000\u0000\u0000\u00e0\u069c\u0001\u0000\u0000"+
		"\u0000\u00e2\u06a5\u0001\u0000\u0000\u0000\u00e4\u06a9\u0001\u0000\u0000"+
		"\u0000\u00e6\u06e7\u0001\u0000\u0000\u0000\u00e8\u06e9\u0001\u0000\u0000"+
		"\u0000\u00ea\u06f7\u0001\u0000\u0000\u0000\u00ec\u0708\u0001\u0000\u0000"+
		"\u0000\u00ee\u0714\u0001\u0000\u0000\u0000\u00f0\u0721\u0001\u0000\u0000"+
		"\u0000\u00f2\u0745\u0001\u0000\u0000\u0000\u00f4\u0747\u0001\u0000\u0000"+
		"\u0000\u00f6\u074a\u0001\u0000\u0000\u0000\u00f8\u074d\u0001\u0000\u0000"+
		"\u0000\u00fa\u0750\u0001\u0000\u0000\u0000\u00fc\u0754\u0001\u0000\u0000"+
		"\u0000\u00fe\u075f\u0001\u0000\u0000\u0000\u0100\u0761\u0001\u0000\u0000"+
		"\u0000\u0102\u0765\u0001\u0000\u0000\u0000\u0104\u076e\u0001\u0000\u0000"+
		"\u0000\u0106\u077d\u0001\u0000\u0000\u0000\u0108\u077f\u0001\u0000\u0000"+
		"\u0000\u010a\u0794\u0001\u0000\u0000\u0000\u010c\u0797\u0001\u0000\u0000"+
		"\u0000\u010e\u079a\u0001\u0000\u0000\u0000\u0110\u07a1\u0001\u0000\u0000"+
		"\u0000\u0112\u07a3\u0001\u0000\u0000\u0000\u0114\u07a7\u0001\u0000\u0000"+
		"\u0000\u0116\u07ac\u0001\u0000\u0000\u0000\u0118\u07af\u0001\u0000\u0000"+
		"\u0000\u011a\u07b3\u0001\u0000\u0000\u0000\u011c\u07cc\u0001\u0000\u0000"+
		"\u0000\u011e\u07ce\u0001\u0000\u0000\u0000\u0120\u081e\u0001\u0000\u0000"+
		"\u0000\u0122\u0820\u0001\u0000\u0000\u0000\u0124\u0824\u0001\u0000\u0000"+
		"\u0000\u0126\u0837\u0001\u0000\u0000\u0000\u0128\u083c\u0001\u0000\u0000"+
		"\u0000\u012a\u0851\u0001\u0000\u0000\u0000\u012c\u0853\u0001\u0000\u0000"+
		"\u0000\u012e\u0855\u0001\u0000\u0000\u0000\u0130\u0857\u0001\u0000\u0000"+
		"\u0000\u0132\u0859\u0001\u0000\u0000\u0000\u0134\u085b\u0001\u0000\u0000"+
		"\u0000\u0136\u0861\u0001\u0000\u0000\u0000\u0138\u0867\u0001\u0000\u0000"+
		"\u0000\u013a\u086d\u0001\u0000\u0000\u0000\u013c\u086f\u0001\u0000\u0000"+
		"\u0000\u013e\u0873\u0001\u0000\u0000\u0000\u0140\u0875\u0001\u0000\u0000"+
		"\u0000\u0142\u0883\u0001\u0000\u0000\u0000\u0144\u0885\u0001\u0000\u0000"+
		"\u0000\u0146\u0887\u0001\u0000\u0000\u0000\u0148\u0889\u0001\u0000\u0000"+
		"\u0000\u014a\u088b\u0001\u0000\u0000\u0000\u014c\u088d\u0001\u0000\u0000"+
		"\u0000\u014e\u088f\u0001\u0000\u0000\u0000\u0150\u0891\u0001\u0000\u0000"+
		"\u0000\u0152\u0893\u0001\u0000\u0000\u0000\u0154\u0895\u0001\u0000\u0000"+
		"\u0000\u0156\u0897\u0001\u0000\u0000\u0000\u0158\u0899\u0001\u0000\u0000"+
		"\u0000\u015a\u089b\u0001\u0000\u0000\u0000\u015c\u089e\u0001\u0000\u0000"+
		"\u0000\u015e\u08a0\u0001\u0000\u0000\u0000\u0160\u08a2\u0001\u0000\u0000"+
		"\u0000\u0162\u08a4\u0001\u0000\u0000\u0000\u0164\u08a6\u0001\u0000\u0000"+
		"\u0000\u0166\u08a8\u0001\u0000\u0000\u0000\u0168\u08ab\u0001\u0000\u0000"+
		"\u0000\u016a\u08ad\u0001\u0000\u0000\u0000\u016c\u08af\u0001\u0000\u0000"+
		"\u0000\u016e\u08b1\u0001\u0000\u0000\u0000\u0170\u08b3\u0001\u0000\u0000"+
		"\u0000\u0172\u08b5\u0001\u0000\u0000\u0000\u0174\u08b7\u0001\u0000\u0000"+
		"\u0000\u0176\u08b9\u0001\u0000\u0000\u0000\u0178\u08bb\u0001\u0000\u0000"+
		"\u0000\u017a\u08bd\u0001\u0000\u0000\u0000\u017c\u08bf\u0001\u0000\u0000"+
		"\u0000\u017e\u08c1\u0001\u0000\u0000\u0000\u0180\u08c3\u0001\u0000\u0000"+
		"\u0000\u0182\u08c5\u0001\u0000\u0000\u0000\u0184\u08c7\u0001\u0000\u0000"+
		"\u0000\u0186\u08c9\u0001\u0000\u0000\u0000\u0188\u08cb\u0001\u0000\u0000"+
		"\u0000\u018a\u08cd\u0001\u0000\u0000\u0000\u018c\u08cf\u0001\u0000\u0000"+
		"\u0000\u018e\u08d1\u0001\u0000\u0000\u0000\u0190\u08d3\u0001\u0000\u0000"+
		"\u0000\u0192\u08d5\u0001\u0000\u0000\u0000\u0194\u08d7\u0001\u0000\u0000"+
		"\u0000\u0196\u08d9\u0001\u0000\u0000\u0000\u0198\u08db\u0001\u0000\u0000"+
		"\u0000\u019a\u08dd\u0001\u0000\u0000\u0000\u019c\u08df\u0001\u0000\u0000"+
		"\u0000\u019e\u08e1\u0001\u0000\u0000\u0000\u01a0\u08e3\u0001\u0000\u0000"+
		"\u0000\u01a2\u08e5\u0001\u0000\u0000\u0000\u01a4\u08e7\u0001\u0000\u0000"+
		"\u0000\u01a6\u08e9\u0001\u0000\u0000\u0000\u01a8\u08eb\u0001\u0000\u0000"+
		"\u0000\u01aa\u08ed\u0001\u0000\u0000\u0000\u01ac\u08ef\u0001\u0000\u0000"+
		"\u0000\u01ae\u08f1\u0001\u0000\u0000\u0000\u01b0\u08f3\u0001\u0000\u0000"+
		"\u0000\u01b2\u08f5\u0001\u0000\u0000\u0000\u01b4\u08f7\u0001\u0000\u0000"+
		"\u0000\u01b6\u08f9\u0001\u0000\u0000\u0000\u01b8\u08fb\u0001\u0000\u0000"+
		"\u0000\u01ba\u08fd\u0001\u0000\u0000\u0000\u01bc\u08ff\u0001\u0000\u0000"+
		"\u0000\u01be\u0901\u0001\u0000\u0000\u0000\u01c0\u0903\u0001\u0000\u0000"+
		"\u0000\u01c2\u0905\u0001\u0000\u0000\u0000\u01c4\u0907\u0001\u0000\u0000"+
		"\u0000\u01c6\u0909\u0001\u0000\u0000\u0000\u01c8\u090b\u0001\u0000\u0000"+
		"\u0000\u01ca\u090d\u0001\u0000\u0000\u0000\u01cc\u090f\u0001\u0000\u0000"+
		"\u0000\u01ce\u0911\u0001\u0000\u0000\u0000\u01d0\u0913\u0001\u0000\u0000"+
		"\u0000\u01d2\u0915\u0001\u0000\u0000\u0000\u01d4\u0917\u0001\u0000\u0000"+
		"\u0000\u01d6\u0919\u0001\u0000\u0000\u0000\u01d8\u091b\u0001\u0000\u0000"+
		"\u0000\u01da\u091d\u0001\u0000\u0000\u0000\u01dc\u091f\u0001\u0000\u0000"+
		"\u0000\u01de\u0921\u0001\u0000\u0000\u0000\u01e0\u0923\u0001\u0000\u0000"+
		"\u0000\u01e2\u0925\u0001\u0000\u0000\u0000\u01e4\u0927\u0001\u0000\u0000"+
		"\u0000\u01e6\u0929\u0001\u0000\u0000\u0000\u01e8\u092b\u0001\u0000\u0000"+
		"\u0000\u01ea\u092d\u0001\u0000\u0000\u0000\u01ec\u092f\u0001\u0000\u0000"+
		"\u0000\u01ee\u0931\u0001\u0000\u0000\u0000\u01f0\u0933\u0001\u0000\u0000"+
		"\u0000\u01f2\u0935\u0001\u0000\u0000\u0000\u01f4\u0937\u0001\u0000\u0000"+
		"\u0000\u01f6\u0939\u0001\u0000\u0000\u0000\u01f8\u093b\u0001\u0000\u0000"+
		"\u0000\u01fa\u093d\u0001\u0000\u0000\u0000\u01fc\u093f\u0001\u0000\u0000"+
		"\u0000\u01fe\u0941\u0001\u0000\u0000\u0000\u0200\u0943\u0001\u0000\u0000"+
		"\u0000\u0202\u0945\u0001\u0000\u0000\u0000\u0204\u0947\u0001\u0000\u0000"+
		"\u0000\u0206\u0949\u0001\u0000\u0000\u0000\u0208\u094b\u0001\u0000\u0000"+
		"\u0000\u020a\u094d\u0001\u0000\u0000\u0000\u020c\u094f\u0001\u0000\u0000"+
		"\u0000\u020e\u0951\u0001\u0000\u0000\u0000\u0210\u0953\u0001\u0000\u0000"+
		"\u0000\u0212\u0955\u0001\u0000\u0000\u0000\u0214\u0957\u0001\u0000\u0000"+
		"\u0000\u0216\u0959\u0001\u0000\u0000\u0000\u0218\u095b\u0001\u0000\u0000"+
		"\u0000\u021a\u095d\u0001\u0000\u0000\u0000\u021c\u095f\u0001\u0000\u0000"+
		"\u0000\u021e\u0961\u0001\u0000\u0000\u0000\u0220\u0963\u0001\u0000\u0000"+
		"\u0000\u0222\u0965\u0001\u0000\u0000\u0000\u0224\u0967\u0001\u0000\u0000"+
		"\u0000\u0226\u0969\u0001\u0000\u0000\u0000\u0228\u096b\u0001\u0000\u0000"+
		"\u0000\u022a\u096d\u0001\u0000\u0000\u0000\u022c\u096f\u0001\u0000\u0000"+
		"\u0000\u022e\u0971\u0001\u0000\u0000\u0000\u0230\u0973\u0001\u0000\u0000"+
		"\u0000\u0232\u0975\u0001\u0000\u0000\u0000\u0234\u0977\u0001\u0000\u0000"+
		"\u0000\u0236\u0979\u0001\u0000\u0000\u0000\u0238\u097b\u0001\u0000\u0000"+
		"\u0000\u023a\u097d\u0001\u0000\u0000\u0000\u023c\u097f\u0001\u0000\u0000"+
		"\u0000\u023e\u0981\u0001\u0000\u0000\u0000\u0240\u0983\u0001\u0000\u0000"+
		"\u0000\u0242\u0985\u0001\u0000\u0000\u0000\u0244\u0987\u0001\u0000\u0000"+
		"\u0000\u0246\u0248\u0003\u0002\u0001\u0000\u0247\u0246\u0001\u0000\u0000"+
		"\u0000\u0247\u0248\u0001\u0000\u0000\u0000\u0248\u024a\u0001\u0000\u0000"+
		"\u0000\u0249\u024b\u0005\u000f\u0000\u0000\u024a\u0249\u0001\u0000\u0000"+
		"\u0000\u024a\u024b\u0001\u0000\u0000\u0000\u024b\u024c\u0001\u0000\u0000"+
		"\u0000\u024c\u024d\u0005\u0000\u0000\u0001\u024d\u0001\u0001\u0000\u0000"+
		"\u0000\u024e\u0250\u0003\b\u0004\u0000\u024f\u0251\u0005\u000f\u0000\u0000"+
		"\u0250\u024f\u0001\u0000\u0000\u0000\u0250\u0251\u0001\u0000\u0000\u0000"+
		"\u0251\u0252\u0001\u0000\u0000\u0000\u0252\u0253\u0003\u0004\u0002\u0000"+
		"\u0253\u0256\u0001\u0000\u0000\u0000\u0254\u0256\u0003\u0006\u0003\u0000"+
		"\u0255\u024e\u0001\u0000\u0000\u0000\u0255\u0254\u0001\u0000\u0000\u0000"+
		"\u0256\u0259\u0001\u0000\u0000\u0000\u0257\u0255\u0001\u0000\u0000\u0000"+
		"\u0257\u0258\u0001\u0000\u0000\u0000\u0258\u0262\u0001\u0000\u0000\u0000"+
		"\u0259\u0257\u0001\u0000\u0000\u0000\u025a\u025f\u0003\b\u0004\u0000\u025b"+
		"\u025d\u0005\u000f\u0000\u0000\u025c\u025b\u0001\u0000\u0000\u0000\u025c"+
		"\u025d\u0001\u0000\u0000\u0000\u025d\u025e\u0001\u0000\u0000\u0000\u025e"+
		"\u0260\u0003\u0004\u0002\u0000\u025f\u025c\u0001\u0000\u0000\u0000\u025f"+
		"\u0260\u0001\u0000\u0000\u0000\u0260\u0263\u0001\u0000\u0000\u0000\u0261"+
		"\u0263\u0003\u0006\u0003\u0000\u0262\u025a\u0001\u0000\u0000\u0000\u0262"+
		"\u0261\u0001\u0000\u0000\u0000\u0263\u0003\u0001\u0000\u0000\u0000\u0264"+
		"\u0265\u0005\b\u0000\u0000\u0265\u0005\u0001\u0000\u0000\u0000\u0266\u0267"+
		"\u0003\u0004\u0002\u0000\u0267\u0007\u0001\u0000\u0000\u0000\u0268\u0291"+
		"\u0003\u00be_\u0000\u0269\u0291\u0003t:\u0000\u026a\u0291\u0003n7\u0000"+
		"\u026b\u0291\u0003\\.\u0000\u026c\u0291\u0003N\'\u0000\u026d\u0291\u0003"+
		"H$\u0000\u026e\u0291\u0003\u00b8\\\u0000\u026f\u0291\u0003\f\u0006\u0000"+
		"\u0270\u0291\u0003\u000e\u0007\u0000\u0271\u0291\u0003:\u001d\u0000\u0272"+
		"\u0291\u00032\u0019\u0000\u0273\u0291\u0003\u00cae\u0000\u0274\u0291\u0003"+
		"0\u0018\u0000\u0275\u0291\u0003&\u0013\u0000\u0276\u0291\u0003\u001e\u000f"+
		"\u0000\u0277\u0291\u0003\u008aE\u0000\u0278\u0291\u0003$\u0012\u0000\u0279"+
		"\u0291\u0003 \u0010\u0000\u027a\u0291\u0003\u001c\u000e\u0000\u027b\u0291"+
		"\u0003\u00d6k\u0000\u027c\u0291\u0003\n\u0005\u0000\u027d\u0291\u0003"+
		"|>\u0000\u027e\u0291\u0003~?\u0000\u027f\u0291\u0003\u0088D\u0000\u0280"+
		"\u0291\u0003\u0086C\u0000\u0281\u0291\u0003z=\u0000\u0282\u0291\u0003"+
		"\u0082A\u0000\u0283\u0291\u0003\u0084B\u0000\u0284\u0291\u0003\u0080@"+
		"\u0000\u0285\u0291\u0003x<\u0000\u0286\u0291\u0003v;\u0000\u0287\u0291"+
		"\u0003\u0016\u000b\u0000\u0288\u0291\u0003\u00f0x\u0000\u0289\u0291\u0003"+
		"\u0014\n\u0000\u028a\u0291\u0003\u0012\t\u0000\u028b\u0291\u0003\u0010"+
		"\b\u0000\u028c\u0291\u0003\u0108\u0084\u0000\u028d\u0291\u0003\u00c8d"+
		"\u0000\u028e\u0291\u0003\u00dcn\u0000\u028f\u0291\u0003\u00c6c\u0000\u0290"+
		"\u0268\u0001\u0000\u0000\u0000\u0290\u0269\u0001\u0000\u0000\u0000\u0290"+
		"\u026a\u0001\u0000\u0000\u0000\u0290\u026b\u0001\u0000\u0000\u0000\u0290"+
		"\u026c\u0001\u0000\u0000\u0000\u0290\u026d\u0001\u0000\u0000\u0000\u0290"+
		"\u026e\u0001\u0000\u0000\u0000\u0290\u026f\u0001\u0000\u0000\u0000\u0290"+
		"\u0270\u0001\u0000\u0000\u0000\u0290\u0271\u0001\u0000\u0000\u0000\u0290"+
		"\u0272\u0001\u0000\u0000\u0000\u0290\u0273\u0001\u0000\u0000\u0000\u0290"+
		"\u0274\u0001\u0000\u0000\u0000\u0290\u0275\u0001\u0000\u0000\u0000\u0290"+
		"\u0276\u0001\u0000\u0000\u0000\u0290\u0277\u0001\u0000\u0000\u0000\u0290"+
		"\u0278\u0001\u0000\u0000\u0000\u0290\u0279\u0001\u0000\u0000\u0000\u0290"+
		"\u027a\u0001\u0000\u0000\u0000\u0290\u027b\u0001\u0000\u0000\u0000\u0290"+
		"\u027c\u0001\u0000\u0000\u0000\u0290\u027d\u0001\u0000\u0000\u0000\u0290"+
		"\u027e\u0001\u0000\u0000\u0000\u0290\u027f\u0001\u0000\u0000\u0000\u0290"+
		"\u0280\u0001\u0000\u0000\u0000\u0290\u0281\u0001\u0000\u0000\u0000\u0290"+
		"\u0282\u0001\u0000\u0000\u0000\u0290\u0283\u0001\u0000\u0000\u0000\u0290"+
		"\u0284\u0001\u0000\u0000\u0000\u0290\u0285\u0001\u0000\u0000\u0000\u0290"+
		"\u0286\u0001\u0000\u0000\u0000\u0290\u0287\u0001\u0000\u0000\u0000\u0290"+
		"\u0288\u0001\u0000\u0000\u0000\u0290\u0289\u0001\u0000\u0000\u0000\u0290"+
		"\u028a\u0001\u0000\u0000\u0000\u0290\u028b\u0001\u0000\u0000\u0000\u0290"+
		"\u028c\u0001\u0000\u0000\u0000\u0290\u028d\u0001\u0000\u0000\u0000\u0290"+
		"\u028e\u0001\u0000\u0000\u0000\u0290\u028f\u0001\u0000\u0000\u0000\u0291"+
		"\t\u0001\u0000\u0000\u0000\u0292\u02ce\u0003\u0196\u00cb\u0000\u0293\u02cf"+
		"\u0003\u01ca\u00e5\u0000\u0294\u0295\u0003\u01c8\u00e4\u0000\u0295\u0296"+
		"\u0003\u0136\u009b\u0000\u0296\u02cf\u0001\u0000\u0000\u0000\u0297\u029b"+
		"\u0003\u020c\u0106\u0000\u0298\u0299\u0003\u0136\u009b\u0000\u0299\u029a"+
		"\u0005\n\u0000\u0000\u029a\u029c\u0001\u0000\u0000\u0000\u029b\u0298\u0001"+
		"\u0000\u0000\u0000\u029b\u029c\u0001\u0000\u0000\u0000\u029c\u029d\u0001"+
		"\u0000\u0000\u0000\u029d\u029e\u0003\u0138\u009c\u0000\u029e\u02cf\u0001"+
		"\u0000\u0000\u0000\u029f\u02cf\u0003\u020e\u0107\u0000\u02a0\u02a4\u0003"+
		"\u021a\u010d\u0000\u02a1\u02a2\u0003\u0136\u009b\u0000\u02a2\u02a3\u0005"+
		"\n\u0000\u0000\u02a3\u02a5\u0001\u0000\u0000\u0000\u02a4\u02a1\u0001\u0000"+
		"\u0000\u0000\u02a4\u02a5\u0001\u0000\u0000\u0000\u02a5\u02a6\u0001\u0000"+
		"\u0000\u0000\u02a6\u02a7\u0003\u014c\u00a6\u0000\u02a7\u02cf\u0001\u0000"+
		"\u0000\u0000\u02a8\u02cf\u0003\u021c\u010e\u0000\u02a9\u02ad\u0003\u01ac"+
		"\u00d6\u0000\u02aa\u02ab\u0003\u0136\u009b\u0000\u02ab\u02ac\u0005\n\u0000"+
		"\u0000\u02ac\u02ae\u0001\u0000\u0000\u0000\u02ad\u02aa\u0001\u0000\u0000"+
		"\u0000\u02ad\u02ae\u0001\u0000\u0000\u0000\u02ae\u02af\u0001\u0000\u0000"+
		"\u0000\u02af\u02b0\u0003\u0150\u00a8\u0000\u02b0\u02cf\u0001\u0000\u0000"+
		"\u0000\u02b1\u02cf\u0003\u01ae\u00d7\u0000\u02b2\u02cf\u0003\u0162\u00b1"+
		"\u0000\u02b3\u02b7\u0003\u0160\u00b0\u0000\u02b4\u02b5\u0003\u0136\u009b"+
		"\u0000\u02b5\u02b6\u0005\n\u0000\u0000\u02b6\u02b8\u0001\u0000\u0000\u0000"+
		"\u02b7\u02b4\u0001\u0000\u0000\u0000\u02b7\u02b8\u0001\u0000\u0000\u0000"+
		"\u02b8\u02b9\u0001\u0000\u0000\u0000\u02b9\u02ba\u0003\u014e\u00a7\u0000"+
		"\u02ba\u02cf\u0001\u0000\u0000\u0000\u02bb\u02cf\u0003\u017e\u00bf\u0000"+
		"\u02bc\u02cf\u0003\u0184\u00c2\u0000\u02bd\u02c1\u0003\u01b6\u00db\u0000"+
		"\u02be\u02bf\u0003\u0136\u009b\u0000\u02bf\u02c0\u0005\n\u0000\u0000\u02c0"+
		"\u02c2\u0001\u0000\u0000\u0000\u02c1\u02be\u0001\u0000\u0000\u0000\u02c1"+
		"\u02c2\u0001\u0000\u0000\u0000\u02c2\u02c3\u0001\u0000\u0000\u0000\u02c3"+
		"\u02c4\u0003\u00ccf\u0000\u02c4\u02cf\u0001\u0000\u0000\u0000\u02c5\u02c6"+
		"\u0003\u01d8\u00ec\u0000\u02c6\u02ca\u0003\u022a\u0115\u0000\u02c7\u02c8"+
		"\u0003\u0136\u009b\u0000\u02c8\u02c9\u0005\n\u0000\u0000\u02c9\u02cb\u0001"+
		"\u0000\u0000\u0000\u02ca\u02c7\u0001\u0000\u0000\u0000\u02ca\u02cb\u0001"+
		"\u0000\u0000\u0000\u02cb\u02cc\u0001\u0000\u0000\u0000\u02cc\u02cd\u0003"+
		"\u014a\u00a5\u0000\u02cd\u02cf\u0001\u0000\u0000\u0000\u02ce\u0293\u0001"+
		"\u0000\u0000\u0000\u02ce\u0294\u0001\u0000\u0000\u0000\u02ce\u0297\u0001"+
		"\u0000\u0000\u0000\u02ce\u029f\u0001\u0000\u0000\u0000\u02ce\u02a0\u0001"+
		"\u0000\u0000\u0000\u02ce\u02a8\u0001\u0000\u0000\u0000\u02ce\u02a9\u0001"+
		"\u0000\u0000\u0000\u02ce\u02b1\u0001\u0000\u0000\u0000\u02ce\u02b2\u0001"+
		"\u0000\u0000\u0000\u02ce\u02b3\u0001\u0000\u0000\u0000\u02ce\u02bb\u0001"+
		"\u0000\u0000\u0000\u02ce\u02bc\u0001\u0000\u0000\u0000\u02ce\u02bd\u0001"+
		"\u0000\u0000\u0000\u02ce\u02c5\u0001\u0000\u0000\u0000\u02cf\u000b\u0001"+
		"\u0000\u0000\u0000\u02d0\u02d1\u0003\u0186\u00c3\u0000\u02d1\u02d2\u0003"+
		"\u0188\u00c4\u0000\u02d2\r\u0001\u0000\u0000\u0000\u02d3\u02d5\u0003\u018a"+
		"\u00c5\u0000\u02d4\u02d6\u0003\u018c\u00c6\u0000\u02d5\u02d4\u0001\u0000"+
		"\u0000\u0000\u02d5\u02d6\u0001\u0000\u0000\u0000\u02d6\u000f\u0001\u0000"+
		"\u0000\u0000\u02d7\u02d8\u0003\u0230\u0118\u0000\u02d8\u02d9\u0003\u0018"+
		"\f\u0000\u02d9\u02da\u0003\u01e6\u00f3\u0000\u02da\u02db\u0003\u001a\r"+
		"\u0000\u02db\u02dc\u0003\u01a8\u00d4\u0000\u02dc\u02dd\u0003\u0144\u00a2"+
		"\u0000\u02dd\u0011\u0001\u0000\u0000\u0000\u02de\u02df\u0003\u01d0\u00e8"+
		"\u0000\u02df\u02e3\u0003\u01fc\u00fe\u0000\u02e0\u02e1\u0003\u01e4\u00f2"+
		"\u0000\u02e1\u02e2\u0003\u0144\u00a2\u0000\u02e2\u02e4\u0001\u0000\u0000"+
		"\u0000\u02e3\u02e0\u0001\u0000\u0000\u0000\u02e3\u02e4\u0001\u0000\u0000"+
		"\u0000\u02e4\u02e6\u0001\u0000\u0000\u0000\u02e5\u02e7\u0003\u01de\u00ef"+
		"\u0000\u02e6\u02e5\u0001\u0000\u0000\u0000\u02e6\u02e7\u0001\u0000\u0000"+
		"\u0000\u02e7\u0013\u0001\u0000\u0000\u0000\u02e8\u02e9\u0003\u01d0\u00e8"+
		"\u0000\u02e9\u02ed\u0003\u0018\f\u0000\u02ea\u02eb\u0003\u01e6\u00f3\u0000"+
		"\u02eb\u02ec\u0003\u001a\r\u0000\u02ec\u02ee\u0001\u0000\u0000\u0000\u02ed"+
		"\u02ea\u0001\u0000\u0000\u0000\u02ed\u02ee\u0001\u0000\u0000\u0000\u02ee"+
		"\u02f2\u0001\u0000\u0000\u0000\u02ef\u02f0\u0003\u01e4\u00f2\u0000\u02f0"+
		"\u02f1\u0003\u0144\u00a2\u0000\u02f1\u02f3\u0001\u0000\u0000\u0000\u02f2"+
		"\u02ef\u0001\u0000\u0000\u0000\u02f2\u02f3\u0001\u0000\u0000\u0000\u02f3"+
		"\u0015\u0001\u0000\u0000\u0000\u02f4\u02f5\u0003\u01b0\u00d8\u0000\u02f5"+
		"\u02f6\u0003\u0018\f\u0000\u02f6\u02f7\u0003\u01e6\u00f3\u0000\u02f7\u02f8"+
		"\u0003\u001a\r\u0000\u02f8\u02f9\u0003\u0212\u0109\u0000\u02f9\u02fa\u0003"+
		"\u0144\u00a2\u0000\u02fa\u0017\u0001\u0000\u0000\u0000\u02fb\u02fe\u0003"+
		"\u0164\u00b2\u0000\u02fc\u02fe\u0003\u0166\u00b3\u0000\u02fd\u02fb\u0001"+
		"\u0000\u0000\u0000\u02fd\u02fc\u0001\u0000\u0000\u0000\u02fe\u0308\u0001"+
		"\u0000\u0000\u0000\u02ff\u0308\u0003\u016a\u00b5\u0000\u0300\u0308\u0003"+
		"\u0174\u00ba\u0000\u0301\u0308\u0003\u0196\u00cb\u0000\u0302\u0308\u0003"+
		"\u01a0\u00d0\u0000\u0303\u0308\u0003\u0190\u00c8\u0000\u0304\u0308\u0003"+
		"\u019a\u00cd\u0000\u0305\u0308\u0003\u01da\u00ed\u0000\u0306\u0308\u0003"+
		"\u01fe\u00ff\u0000\u0307\u02fd\u0001\u0000\u0000\u0000\u0307\u02ff\u0001"+
		"\u0000\u0000\u0000\u0307\u0300\u0001\u0000\u0000\u0000\u0307\u0301\u0001"+
		"\u0000\u0000\u0000\u0307\u0302\u0001\u0000\u0000\u0000\u0307\u0303\u0001"+
		"\u0000\u0000\u0000\u0307\u0304\u0001\u0000\u0000\u0000\u0307\u0305\u0001"+
		"\u0000\u0000\u0000\u0307\u0306\u0001\u0000\u0000\u0000\u0308\u0019\u0001"+
		"\u0000\u0000\u0000\u0309\u030a\u0003\u0164\u00b2\u0000\u030a\u030b\u0003"+
		"\u01ae\u00d7\u0000\u030b\u0330\u0001\u0000\u0000\u0000\u030c\u030d\u0003"+
		"\u0164\u00b2\u0000\u030d\u030e\u0003\u01ae\u00d7\u0000\u030e\u030f\u0003"+
		"\u01b4\u00da\u0000\u030f\u0310\u0003\u01c8\u00e4\u0000\u0310\u0311\u0003"+
		"\u0136\u009b\u0000\u0311\u0330\u0001\u0000\u0000\u0000\u0312\u0316\u0003"+
		"\u01ac\u00d6\u0000\u0313\u0314\u0003\u0136\u009b\u0000\u0314\u0315\u0005"+
		"\n\u0000\u0000\u0315\u0317\u0001\u0000\u0000\u0000\u0316\u0313\u0001\u0000"+
		"\u0000\u0000\u0316\u0317\u0001\u0000\u0000\u0000\u0317\u0318\u0001\u0000"+
		"\u0000\u0000\u0318\u0319\u0003\u0150\u00a8\u0000\u0319\u0330\u0001\u0000"+
		"\u0000\u0000\u031a\u031b\u0003\u0164\u00b2\u0000\u031b\u031c\u0003\u01ca"+
		"\u00e5\u0000\u031c\u0330\u0001\u0000\u0000\u0000\u031d\u031e\u0003\u01c8"+
		"\u00e4\u0000\u031e\u031f\u0003\u0136\u009b\u0000\u031f\u0330\u0001\u0000"+
		"\u0000\u0000\u0320\u0322\u0003\u020c\u0106\u0000\u0321\u0320\u0001\u0000"+
		"\u0000\u0000\u0321\u0322\u0001\u0000\u0000\u0000\u0322\u0326\u0001\u0000"+
		"\u0000\u0000\u0323\u0324\u0003\u0136\u009b\u0000\u0324\u0325\u0005\n\u0000"+
		"\u0000\u0325\u0327\u0001\u0000\u0000\u0000\u0326\u0323\u0001\u0000\u0000"+
		"\u0000\u0326\u0327\u0001\u0000\u0000\u0000\u0327\u0328\u0001\u0000\u0000"+
		"\u0000\u0328\u0330\u0003\u0138\u009c\u0000\u0329\u032a\u0003\u0164\u00b2"+
		"\u0000\u032a\u032b\u0003\u01fc\u00fe\u0000\u032b\u0330\u0001\u0000\u0000"+
		"\u0000\u032c\u032d\u0003\u01fa\u00fd\u0000\u032d\u032e\u0003\u0144\u00a2"+
		"\u0000\u032e\u0330\u0001\u0000\u0000\u0000\u032f\u0309\u0001\u0000\u0000"+
		"\u0000\u032f\u030c\u0001\u0000\u0000\u0000\u032f\u0312\u0001\u0000\u0000"+
		"\u0000\u032f\u031a\u0001\u0000\u0000\u0000\u032f\u031d\u0001\u0000\u0000"+
		"\u0000\u032f\u0321\u0001\u0000\u0000\u0000\u032f\u0329\u0001\u0000\u0000"+
		"\u0000\u032f\u032c\u0001\u0000\u0000\u0000\u0330\u001b\u0001\u0000\u0000"+
		"\u0000\u0331\u0332\u0003\u0190\u00c8\u0000\u0332\u0334\u0003\u0224\u0112"+
		"\u0000\u0333\u0335\u0003\u00fa}\u0000\u0334\u0333\u0001\u0000\u0000\u0000"+
		"\u0334\u0335\u0001\u0000\u0000\u0000\u0335\u0336\u0001\u0000\u0000\u0000"+
		"\u0336\u0337\u0003\u0154\u00aa\u0000\u0337\u0338\u0003\u022e\u0117\u0000"+
		"\u0338\u0339\u0003\u01ee\u00f7\u0000\u0339\u033c\u0003\u0130\u0098\u0000"+
		"\u033a\u033d\u0003\u020a\u0105\u0000\u033b\u033d\u0003\u01dc\u00ee\u0000"+
		"\u033c\u033a\u0001\u0000\u0000\u0000\u033c\u033b\u0001\u0000\u0000\u0000"+
		"\u033c\u033d\u0001\u0000\u0000\u0000\u033d\u001d\u0001\u0000\u0000\u0000"+
		"\u033e\u033f\u0003\u0190\u00c8\u0000\u033f\u0341\u0003\u01fa\u00fd\u0000"+
		"\u0340\u0342\u0003\u00fa}\u0000\u0341\u0340\u0001\u0000\u0000\u0000\u0341"+
		"\u0342\u0001\u0000\u0000\u0000\u0342\u0343\u0001\u0000\u0000\u0000\u0343"+
		"\u0345\u0003\u0144\u00a2\u0000\u0344\u0346\u0003p8\u0000\u0345\u0344\u0001"+
		"\u0000\u0000\u0000\u0345\u0346\u0001\u0000\u0000\u0000\u0346\u001f\u0001"+
		"\u0000\u0000\u0000\u0347\u0348\u0003\u0190\u00c8\u0000\u0348\u034a\u0003"+
		"\u021a\u010d\u0000\u0349\u034b\u0003\u00fa}\u0000\u034a\u0349\u0001\u0000"+
		"\u0000\u0000\u034a\u034b\u0001\u0000\u0000\u0000\u034b\u034f\u0001\u0000"+
		"\u0000\u0000\u034c\u034d\u0003\u0136\u009b\u0000\u034d\u034e\u0005\n\u0000"+
		"\u0000\u034e\u0350\u0001\u0000\u0000\u0000\u034f\u034c\u0001\u0000\u0000"+
		"\u0000\u034f\u0350\u0001\u0000\u0000\u0000\u0350\u0351\u0001\u0000\u0000"+
		"\u0000\u0351\u0352\u0003\u014c\u00a6\u0000\u0352\u0353\u0003\u0232\u0119"+
		"\u0000\u0353\u0354\u0003\"\u0011\u0000\u0354\u0355\u0003\u0234\u011a\u0000"+
		"\u0355!\u0001\u0000\u0000\u0000\u0356\u0357\u0003\u013a\u009d\u0000\u0357"+
		"\u035e\u0003\u013c\u009e\u0000\u0358\u0359\u0003\u0242\u0121\u0000\u0359"+
		"\u035a\u0003\u013a\u009d\u0000\u035a\u035b\u0003\u013c\u009e\u0000\u035b"+
		"\u035d\u0001\u0000\u0000\u0000\u035c\u0358\u0001\u0000\u0000\u0000\u035d"+
		"\u0360\u0001\u0000\u0000\u0000\u035e\u035c\u0001\u0000\u0000\u0000\u035e"+
		"\u035f\u0001\u0000\u0000\u0000\u035f#\u0001\u0000\u0000\u0000\u0360\u035e"+
		"\u0001\u0000\u0000\u0000\u0361\u0362\u0003\u0190\u00c8\u0000\u0362\u0364"+
		"\u0003\u0214\u010a\u0000\u0363\u0365\u0003\u00fa}\u0000\u0364\u0363\u0001"+
		"\u0000\u0000\u0000\u0364\u0365\u0001\u0000\u0000\u0000\u0365\u0369\u0001"+
		"\u0000\u0000\u0000\u0366\u0367\u0003\u0136\u009b\u0000\u0367\u0368\u0005"+
		"\n\u0000\u0000\u0368\u036a\u0001\u0000\u0000\u0000\u0369\u0366\u0001\u0000"+
		"\u0000\u0000\u0369\u036a\u0001\u0000\u0000\u0000\u036a\u036b\u0001\u0000"+
		"\u0000\u0000\u036b\u036c\u0003\u0146\u00a3\u0000\u036c\u036d\u0003\u0226"+
		"\u0113\u0000\u036d\u036e\u0003\u0148\u00a4\u0000\u036e%\u0001\u0000\u0000"+
		"\u0000\u036f\u0370\u0003\u0190\u00c8\u0000\u0370\u0371\u0003\u01d8\u00ec"+
		"\u0000\u0371\u0373\u0003\u022a\u0115\u0000\u0372\u0374\u0003\u00fa}\u0000"+
		"\u0373\u0372\u0001\u0000\u0000\u0000\u0373\u0374\u0001\u0000\u0000\u0000"+
		"\u0374\u0378\u0001\u0000\u0000\u0000\u0375\u0376\u0003\u0136\u009b\u0000"+
		"\u0376\u0377\u0005\n\u0000\u0000\u0377\u0379\u0001\u0000\u0000\u0000\u0378"+
		"\u0375\u0001\u0000\u0000\u0000\u0378\u0379\u0001\u0000\u0000\u0000\u0379"+
		"\u037a\u0001\u0000\u0000\u0000\u037a\u037b\u0003\u014a\u00a5\u0000\u037b"+
		"\u037c\u0003\u0170\u00b8\u0000\u037c\u037d\u0003\u01fe\u00ff\u0000\u037d"+
		"\u037e\u0003\u0102\u0081\u0000\u037e\u0382\u0003\u01a8\u00d4\u0000\u037f"+
		"\u0380\u0003\u0136\u009b\u0000\u0380\u0381\u0005\n\u0000\u0000\u0381\u0383"+
		"\u0001\u0000\u0000\u0000\u0382\u037f\u0001\u0000\u0000\u0000\u0382\u0383"+
		"\u0001\u0000\u0000\u0000\u0383\u0384\u0001\u0000\u0000\u0000\u0384\u0385"+
		"\u0003\u0138\u009c\u0000\u0385\u0386\u0003(\u0014\u0000\u0386\u0387\u0003"+
		"\u01f0\u00f8\u0000\u0387\u0388\u0003\u01c4\u00e2\u0000\u0388\u0389\u0003"+
		"\u0232\u0119\u0000\u0389\u038a\u0003\u0102\u0081\u0000\u038a\u038e\u0003"+
		"\u0234\u011a\u0000\u038b\u038c\u0003\u022e\u0117\u0000\u038c\u038d\u0003"+
		".\u0017\u0000\u038d\u038f\u0001\u0000\u0000\u0000\u038e\u038b\u0001\u0000"+
		"\u0000\u0000\u038e\u038f\u0001\u0000\u0000\u0000\u038f\'\u0001\u0000\u0000"+
		"\u0000\u0390\u0391\u0003\u022c\u0116\u0000\u0391\u0395\u0003*\u0015\u0000"+
		"\u0392\u0393\u0003\u016c\u00b6\u0000\u0393\u0394\u0003\u011e\u008f\u0000"+
		"\u0394\u0396\u0001\u0000\u0000\u0000\u0395\u0392\u0001\u0000\u0000\u0000"+
		"\u0395\u0396\u0001\u0000\u0000\u0000\u0396)\u0001\u0000\u0000\u0000\u0397"+
		"\u039d\u0003,\u0016\u0000\u0398\u0399\u0003\u016c\u00b6\u0000\u0399\u039a"+
		"\u0003,\u0016\u0000\u039a\u039c\u0001\u0000\u0000\u0000\u039b\u0398\u0001"+
		"\u0000\u0000\u0000\u039c\u039f\u0001\u0000\u0000\u0000\u039d\u039b\u0001"+
		"\u0000\u0000\u0000\u039d\u039e\u0001\u0000\u0000\u0000\u039e+\u0001\u0000"+
		"\u0000\u0000\u039f\u039d\u0001\u0000\u0000\u0000\u03a0\u03a1\u0003\u013a"+
		"\u009d\u0000\u03a1\u03a2\u0003\u01c0\u00e0\u0000\u03a2\u03a3\u0003\u01e0"+
		"\u00f0\u0000\u03a3\u03a4\u0003\u01e2\u00f1\u0000\u03a4-\u0001\u0000\u0000"+
		"\u0000\u03a5\u03ac\u0003\u008eG\u0000\u03a6\u03a7\u0003\u008eG\u0000\u03a7"+
		"\u03a8\u0003\u016c\u00b6\u0000\u03a8\u03a9\u0003\u0090H\u0000\u03a9\u03ac"+
		"\u0001\u0000\u0000\u0000\u03aa\u03ac\u0003\u0090H\u0000\u03ab\u03a5\u0001"+
		"\u0000\u0000\u0000\u03ab\u03a6\u0001\u0000\u0000\u0000\u03ab\u03aa\u0001"+
		"\u0000\u0000\u0000\u03ac/\u0001\u0000\u0000\u0000\u03ad\u03ae\u0003\u0190"+
		"\u00c8\u0000\u03ae\u03b0\u0003\u01c8\u00e4\u0000\u03af\u03b1\u0003\u00fa"+
		"}\u0000\u03b0\u03af\u0001\u0000\u0000\u0000\u03b0\u03b1\u0001\u0000\u0000"+
		"\u0000\u03b1\u03b2\u0001\u0000\u0000\u0000\u03b2\u03b3\u0003\u0136\u009b"+
		"\u0000\u03b3\u03b4\u0003\u022e\u0117\u0000\u03b4\u03b5\u0003\u01f6\u00fb"+
		"\u0000\u03b5\u03b6\u0005\u0013\u0000\u0000\u03b6\u03b7\u0003\u0236\u011b"+
		"\u0000\u03b7\u03b8\u0003\u00c0`\u0000\u03b8\u03bc\u0003\u0238\u011c\u0000"+
		"\u03b9\u03ba\u0003\u016c\u00b6\u0000\u03ba\u03bb\u0003\u00c4b\u0000\u03bb"+
		"\u03bd\u0001\u0000\u0000\u0000\u03bc\u03b9\u0001\u0000\u0000\u0000\u03bc"+
		"\u03bd\u0001\u0000\u0000\u0000\u03bd1\u0001\u0000\u0000\u0000\u03be\u03c0"+
		"\u0003\u0190\u00c8\u0000\u03bf\u03c1\u0003F#\u0000\u03c0\u03bf\u0001\u0000"+
		"\u0000\u0000\u03c0\u03c1\u0001\u0000\u0000\u0000\u03c1\u03c2\u0001\u0000"+
		"\u0000\u0000\u03c2\u03c4\u0003\u01ac\u00d6\u0000\u03c3\u03c5\u0003\u00fa"+
		"}\u0000\u03c4\u03c3\u0001\u0000\u0000\u0000\u03c4\u03c5\u0001\u0000\u0000"+
		"\u0000\u03c5\u03c9\u0001\u0000\u0000\u0000\u03c6\u03c7\u0003\u0136\u009b"+
		"\u0000\u03c7\u03c8\u0005\n\u0000\u0000\u03c8\u03ca\u0001\u0000\u0000\u0000"+
		"\u03c9\u03c6\u0001\u0000\u0000\u0000\u03c9\u03ca\u0001\u0000\u0000\u0000"+
		"\u03ca\u03cb\u0001\u0000\u0000\u0000\u03cb\u03cc\u0003\u0150\u00a8\u0000"+
		"\u03cc\u03ce\u0003\u0232\u0119\u0000\u03cd\u03cf\u00036\u001b\u0000\u03ce"+
		"\u03cd\u0001\u0000\u0000\u0000\u03ce\u03cf\u0001\u0000\u0000\u0000\u03cf"+
		"\u03d0\u0001\u0000\u0000\u0000\u03d0\u03d1\u0003\u0234\u011a\u0000\u03d1"+
		"\u03d2\u00038\u001c\u0000\u03d2\u03d3\u0003\u01f8\u00fc\u0000\u03d3\u03d4"+
		"\u0003\u013c\u009e\u0000\u03d4\u03d5\u0003\u01cc\u00e6\u0000\u03d5\u03d6"+
		"\u0003\u0152\u00a9\u0000\u03d6\u03d7\u0003\u0170\u00b8\u0000\u03d7\u03d8"+
		"\u00034\u001a\u0000\u03d83\u0001\u0000\u0000\u0000\u03d9\u03da\u0007\u0000"+
		"\u0000\u0000\u03da5\u0001\u0000\u0000\u0000\u03db\u03e1\u0003\u015a\u00ad"+
		"\u0000\u03dc\u03dd\u0003\u0242\u0121\u0000\u03dd\u03de\u0003\u015a\u00ad"+
		"\u0000\u03de\u03e0\u0001\u0000\u0000\u0000\u03df\u03dc\u0001\u0000\u0000"+
		"\u0000\u03e0\u03e3\u0001\u0000\u0000\u0000\u03e1\u03df\u0001\u0000\u0000"+
		"\u0000\u03e1\u03e2\u0001\u0000\u0000\u0000\u03e27\u0001\u0000\u0000\u0000"+
		"\u03e3\u03e1\u0001\u0000\u0000\u0000\u03e4\u03e9\u0003\u017c\u00be\u0000"+
		"\u03e5\u03e6\u0003\u01f8\u00fc\u0000\u03e6\u03e7\u0003\u01e2\u00f1\u0000"+
		"\u03e7\u03e9\u0001\u0000\u0000\u0000\u03e8\u03e4\u0001\u0000\u0000\u0000"+
		"\u03e8\u03e5\u0001\u0000\u0000\u0000\u03e9\u03ea\u0001\u0000\u0000\u0000"+
		"\u03ea\u03eb\u0003\u01e6\u00f3\u0000\u03eb\u03ec\u0003\u01e2\u00f1\u0000"+
		"\u03ec\u03ed\u0003\u01ba\u00dd\u0000\u03ed9\u0001\u0000\u0000\u0000\u03ee"+
		"\u03f0\u0003\u0190\u00c8\u0000\u03ef\u03f1\u0003F#\u0000\u03f0\u03ef\u0001"+
		"\u0000\u0000\u0000\u03f0\u03f1\u0001\u0000\u0000\u0000\u03f1\u03f2\u0001"+
		"\u0000\u0000\u0000\u03f2\u03f4\u0003\u0160\u00b0\u0000\u03f3\u03f5\u0003"+
		"\u00fa}\u0000\u03f4\u03f3\u0001\u0000\u0000\u0000\u03f4\u03f5\u0001\u0000"+
		"\u0000\u0000\u03f5\u03f9\u0001\u0000\u0000\u0000\u03f6\u03f7\u0003\u0136"+
		"\u009b\u0000\u03f7\u03f8\u0005\n\u0000\u0000\u03f8\u03fa\u0001\u0000\u0000"+
		"\u0000\u03f9\u03f6\u0001\u0000\u0000\u0000\u03f9\u03fa\u0001\u0000\u0000"+
		"\u0000\u03fa\u03fb\u0001\u0000\u0000\u0000\u03fb\u03fc\u0003\u014e\u00a7"+
		"\u0000\u03fc\u03fd\u0003\u0232\u0119\u0000\u03fd\u03fe\u0003\u013c\u009e"+
		"\u0000\u03fe\u03ff\u0003\u0234\u011a\u0000\u03ff\u0400\u0003\u0204\u0102"+
		"\u0000\u0400\u0401\u0003\u0150\u00a8\u0000\u0401\u0402\u0003\u0208\u0104"+
		"\u0000\u0402\u0403\u0003\u013c\u009e\u0000\u0403\u0404\u0003\u01a6\u00d3"+
		"\u0000\u0404\u0405\u0003\u0150\u00a8\u0000\u0405\u0406\u0003\u01b8\u00dc"+
		"\u0000\u0406\u0407\u0003<\u001e\u0000\u0407;\u0001\u0000\u0000\u0000\u0408"+
		"\u040d\u0003\u012a\u0095\u0000\u0409\u040d\u0003D\"\u0000\u040a\u040d"+
		"\u0003B!\u0000\u040b\u040d\u0003>\u001f\u0000\u040c\u0408\u0001\u0000"+
		"\u0000\u0000\u040c\u0409\u0001\u0000\u0000\u0000\u040c\u040a\u0001\u0000"+
		"\u0000\u0000\u040c\u040b\u0001\u0000\u0000\u0000\u040d=\u0001\u0000\u0000"+
		"\u0000\u040e\u040f\u0003\u0236\u011b\u0000\u040f\u0415\u0003@ \u0000\u0410"+
		"\u0411\u0003\u0242\u0121\u0000\u0411\u0412\u0003@ \u0000\u0412\u0414\u0001"+
		"\u0000\u0000\u0000\u0413\u0410\u0001\u0000\u0000\u0000\u0414\u0417\u0001"+
		"\u0000\u0000\u0000\u0415\u0413\u0001\u0000\u0000\u0000\u0415\u0416\u0001"+
		"\u0000\u0000\u0000\u0416\u0418\u0001\u0000\u0000\u0000\u0417\u0415\u0001"+
		"\u0000\u0000\u0000\u0418\u0419\u0003\u0238\u011c\u0000\u0419?\u0001\u0000"+
		"\u0000\u0000\u041a\u041b\u0003\u0158\u00ac\u0000\u041b\u041c\u0005\t\u0000"+
		"\u0000\u041c\u041d\u0003<\u001e\u0000\u041dA\u0001\u0000\u0000\u0000\u041e"+
		"\u041f\u0003\u0232\u0119\u0000\u041f\u0426\u0003D\"\u0000\u0420\u0421"+
		"\u0003\u0242\u0121\u0000\u0421\u0422\u0003\u012a\u0095\u0000\u0422\u0425"+
		"\u0001\u0000\u0000\u0000\u0423\u0425\u0003D\"\u0000\u0424\u0420\u0001"+
		"\u0000\u0000\u0000\u0424\u0423\u0001\u0000\u0000\u0000\u0425\u0428\u0001"+
		"\u0000\u0000\u0000\u0426\u0424\u0001\u0000\u0000\u0000\u0426\u0427\u0001"+
		"\u0000\u0000\u0000\u0427\u0429\u0001\u0000\u0000\u0000\u0428\u0426\u0001"+
		"\u0000\u0000\u0000\u0429\u042a\u0003\u0234\u011a\u0000\u042aC\u0001\u0000"+
		"\u0000\u0000\u042b\u042c\u0003\u0232\u0119\u0000\u042c\u0432\u0003\u012a"+
		"\u0095\u0000\u042d\u042e\u0003\u0242\u0121\u0000\u042e\u042f\u0003\u012a"+
		"\u0095\u0000\u042f\u0431\u0001\u0000\u0000\u0000\u0430\u042d\u0001\u0000"+
		"\u0000\u0000\u0431\u0434\u0001\u0000\u0000\u0000\u0432\u0430\u0001\u0000"+
		"\u0000\u0000\u0432\u0433\u0001\u0000\u0000\u0000\u0433\u0435\u0001\u0000"+
		"\u0000\u0000\u0434\u0432\u0001\u0000\u0000\u0000\u0435\u0436\u0003\u0234"+
		"\u011a\u0000\u0436E\u0001\u0000\u0000\u0000\u0437\u0438\u0003\u01ea\u00f5"+
		"\u0000\u0438\u0439\u0003\u01f4\u00fa\u0000\u0439G\u0001\u0000\u0000\u0000"+
		"\u043a\u043b\u0003\u016a\u00b5\u0000\u043b\u043c\u0003\u0224\u0112\u0000"+
		"\u043c\u043d\u0003\u0154\u00aa\u0000\u043d\u043e\u0003\u022e\u0117\u0000"+
		"\u043e\u0440\u0003J%\u0000\u043f\u0441\u0003L&\u0000\u0440\u043f\u0001"+
		"\u0000\u0000\u0000\u0440\u0441\u0001\u0000\u0000\u0000\u0441I\u0001\u0000"+
		"\u0000\u0000\u0442\u0443\u0003\u01ee\u00f7\u0000\u0443\u0444\u0003\u0130"+
		"\u0098\u0000\u0444K\u0001\u0000\u0000\u0000\u0445\u0448\u0003\u020a\u0105"+
		"\u0000\u0446\u0448\u0003\u01dc\u00ee\u0000\u0447\u0445\u0001\u0000\u0000"+
		"\u0000\u0447\u0446\u0001\u0000\u0000\u0000\u0448M\u0001\u0000\u0000\u0000"+
		"\u0449\u044a\u0003\u016a\u00b5\u0000\u044a\u044e\u0003\u021a\u010d\u0000"+
		"\u044b\u044c\u0003\u0136\u009b\u0000\u044c\u044d\u0005\n\u0000\u0000\u044d"+
		"\u044f\u0001\u0000\u0000\u0000\u044e\u044b\u0001\u0000\u0000\u0000\u044e"+
		"\u044f\u0001\u0000\u0000\u0000\u044f\u0450\u0001\u0000\u0000\u0000\u0450"+
		"\u0451\u0003\u014c\u00a6\u0000\u0451\u0452\u0003P(\u0000\u0452O\u0001"+
		"\u0000\u0000\u0000\u0453\u0457\u0003Z-\u0000\u0454\u0457\u0003X,\u0000"+
		"\u0455\u0457\u0003R)\u0000\u0456\u0453\u0001\u0000\u0000\u0000\u0456\u0454"+
		"\u0001\u0000\u0000\u0000\u0456\u0455\u0001\u0000\u0000\u0000\u0457Q\u0001"+
		"\u0000\u0000\u0000\u0458\u0459\u0003\u01f2\u00f9\u0000\u0459\u045a\u0003"+
		"T*\u0000\u045aS\u0001\u0000\u0000\u0000\u045b\u0461\u0003V+\u0000\u045c"+
		"\u045d\u0003\u016c\u00b6\u0000\u045d\u045e\u0003V+\u0000\u045e\u0460\u0001"+
		"\u0000\u0000\u0000\u045f\u045c\u0001\u0000\u0000\u0000\u0460\u0463\u0001"+
		"\u0000\u0000\u0000\u0461\u045f\u0001\u0000\u0000\u0000\u0461\u0462\u0001"+
		"\u0000\u0000\u0000\u0462U\u0001\u0000\u0000\u0000\u0463\u0461\u0001\u0000"+
		"\u0000\u0000\u0464\u0465\u0003\u013a\u009d\u0000\u0465\u0466\u0003\u0212"+
		"\u0109\u0000\u0466\u0467\u0003\u013a\u009d\u0000\u0467W\u0001\u0000\u0000"+
		"\u0000\u0468\u0469\u0003\u015e\u00af\u0000\u0469\u046a\u0003\u013a\u009d"+
		"\u0000\u046a\u0471\u0003\u013c\u009e\u0000\u046b\u046c\u0003\u0242\u0121"+
		"\u0000\u046c\u046d\u0003\u013a\u009d\u0000\u046d\u046e\u0003\u013c\u009e"+
		"\u0000\u046e\u0470\u0001\u0000\u0000\u0000\u046f\u046b\u0001\u0000\u0000"+
		"\u0000\u0470\u0473\u0001\u0000\u0000\u0000\u0471\u046f\u0001\u0000\u0000"+
		"\u0000\u0471\u0472\u0001\u0000\u0000\u0000\u0472Y\u0001\u0000\u0000\u0000"+
		"\u0473\u0471\u0001\u0000\u0000\u0000\u0474\u0475\u0003\u016a\u00b5\u0000"+
		"\u0475\u0476\u0003\u013a\u009d\u0000\u0476\u0477\u0003\u021a\u010d\u0000"+
		"\u0477\u0478\u0003\u013c\u009e\u0000\u0478[\u0001\u0000\u0000\u0000\u0479"+
		"\u047a\u0003\u016a\u00b5\u0000\u047a\u047e\u0003\u020c\u0106\u0000\u047b"+
		"\u047c\u0003\u0136\u009b\u0000\u047c\u047d\u0005\n\u0000\u0000\u047d\u047f"+
		"\u0001\u0000\u0000\u0000\u047e\u047b\u0001\u0000\u0000\u0000\u047e\u047f"+
		"\u0001\u0000\u0000\u0000\u047f\u0480\u0001\u0000\u0000\u0000\u0480\u0481"+
		"\u0003\u0138\u009c\u0000\u0481\u0482\u0003^/\u0000\u0482]\u0001\u0000"+
		"\u0000\u0000\u0483\u0489\u0003j5\u0000\u0484\u0489\u0003f3\u0000\u0485"+
		"\u0489\u0003d2\u0000\u0486\u0489\u0003b1\u0000\u0487\u0489\u0003`0\u0000"+
		"\u0488\u0483\u0001\u0000\u0000\u0000\u0488\u0484\u0001\u0000\u0000\u0000"+
		"\u0488\u0485\u0001\u0000\u0000\u0000\u0488\u0486\u0001\u0000\u0000\u0000"+
		"\u0488\u0487\u0001\u0000\u0000\u0000\u0489_\u0001\u0000\u0000\u0000\u048a"+
		"\u048b\u0003\u022e\u0117\u0000\u048b\u048c\u0003\u008eG\u0000\u048ca\u0001"+
		"\u0000\u0000\u0000\u048d\u048e\u0003\u01f2\u00f9\u0000\u048e\u048f\u0003"+
		"\u013a\u009d\u0000\u048f\u0490\u0003\u0212\u0109\u0000\u0490\u0491\u0003"+
		"\u013a\u009d\u0000\u0491c\u0001\u0000\u0000\u0000\u0492\u0493\u0003\u019a"+
		"\u00cd\u0000\u0493\u0494\u0003\u0182\u00c1\u0000\u0494\u0495\u0003\u0206"+
		"\u0103\u0000\u0495e\u0001\u0000\u0000\u0000\u0496\u0497\u0003\u019a\u00cd"+
		"\u0000\u0497\u0498\u0003h4\u0000\u0498g\u0001\u0000\u0000\u0000\u0499"+
		"\u049f\u0003\u013a\u009d\u0000\u049a\u049b\u0003\u0242\u0121\u0000\u049b"+
		"\u049c\u0003\u013a\u009d\u0000\u049c\u049e\u0001\u0000\u0000\u0000\u049d"+
		"\u049a\u0001\u0000\u0000\u0000\u049e\u04a1\u0001\u0000\u0000\u0000\u049f"+
		"\u049d\u0001\u0000\u0000\u0000\u049f\u04a0\u0001\u0000\u0000\u0000\u04a0"+
		"i\u0001\u0000\u0000\u0000\u04a1\u049f\u0001\u0000\u0000\u0000\u04a2\u04a3"+
		"\u0003\u015e\u00af\u0000\u04a3\u04a4\u0003l6\u0000\u04a4k\u0001\u0000"+
		"\u0000\u0000\u04a5\u04a6\u0003\u013a\u009d\u0000\u04a6\u04ad\u0003\u013c"+
		"\u009e\u0000\u04a7\u04a8\u0003\u0242\u0121\u0000\u04a8\u04a9\u0003\u013a"+
		"\u009d\u0000\u04a9\u04aa\u0003\u013c\u009e\u0000\u04aa\u04ac\u0001\u0000"+
		"\u0000\u0000\u04ab\u04a7\u0001\u0000\u0000\u0000\u04ac\u04af\u0001\u0000"+
		"\u0000\u0000\u04ad\u04ab\u0001\u0000\u0000\u0000\u04ad\u04ae\u0001\u0000"+
		"\u0000\u0000\u04aem\u0001\u0000\u0000\u0000\u04af\u04ad\u0001\u0000\u0000"+
		"\u0000\u04b0\u04b1\u0003\u016a\u00b5\u0000\u04b1\u04b2\u0003\u01fa\u00fd"+
		"\u0000\u04b2\u04b4\u0003\u0144\u00a2\u0000\u04b3\u04b5\u0003p8\u0000\u04b4"+
		"\u04b3\u0001\u0000\u0000\u0000\u04b4\u04b5\u0001\u0000\u0000\u0000\u04b5"+
		"o\u0001\u0000\u0000\u0000\u04b6\u04b7\u0003\u022e\u0117\u0000\u04b7\u04bd"+
		"\u0003r9\u0000\u04b8\u04b9\u0003\u016c\u00b6\u0000\u04b9\u04ba\u0003r"+
		"9\u0000\u04ba\u04bc\u0001\u0000\u0000\u0000\u04bb\u04b8\u0001\u0000\u0000"+
		"\u0000\u04bc\u04bf\u0001\u0000\u0000\u0000\u04bd\u04bb\u0001\u0000\u0000"+
		"\u0000\u04bd\u04be\u0001\u0000\u0000\u0000\u04beq\u0001\u0000\u0000\u0000"+
		"\u04bf\u04bd\u0001\u0000\u0000\u0000\u04c0\u04c1\u0003\u01ee\u00f7\u0000"+
		"\u04c1\u04c2\u0005\u0013\u0000\u0000\u04c2\u04c3\u0003\u0130\u0098\u0000"+
		"\u04c3\u04d1\u0001\u0000\u0000\u0000\u04c4\u04c5\u0003\u01d6\u00eb\u0000"+
		"\u04c5\u04c6\u0005\u0013\u0000\u0000\u04c6\u04c7\u0003\u0132\u0099\u0000"+
		"\u04c7\u04d1\u0001\u0000\u0000\u0000\u04c8\u04c9\u0003\u020a\u0105\u0000"+
		"\u04c9\u04ca\u0005\u0013\u0000\u0000\u04ca\u04cb\u0003\u0132\u0099\u0000"+
		"\u04cb\u04d1\u0001\u0000\u0000\u0000\u04cc\u04cd\u0003\u01e8\u00f4\u0000"+
		"\u04cd\u04ce\u0005\u0013\u0000\u0000\u04ce\u04cf\u0003\u0098L\u0000\u04cf"+
		"\u04d1\u0001\u0000\u0000\u0000\u04d0\u04c0\u0001\u0000\u0000\u0000\u04d0"+
		"\u04c4\u0001\u0000\u0000\u0000\u04d0\u04c8\u0001\u0000\u0000\u0000\u04d0"+
		"\u04cc\u0001\u0000\u0000\u0000\u04d1s\u0001\u0000\u0000\u0000\u04d2\u04d3"+
		"\u0003\u016a\u00b5\u0000\u04d3\u04d4\u0003\u01d8\u00ec\u0000\u04d4\u04d8"+
		"\u0003\u022a\u0115\u0000\u04d5\u04d6\u0003\u0136\u009b\u0000\u04d6\u04d7"+
		"\u0005\n\u0000\u0000\u04d7\u04d9\u0001\u0000\u0000\u0000\u04d8\u04d5\u0001"+
		"\u0000\u0000\u0000\u04d8\u04d9\u0001\u0000\u0000\u0000\u04d9\u04da\u0001"+
		"\u0000\u0000\u0000\u04da\u04de\u0003\u014a\u00a5\u0000\u04db\u04dc\u0003"+
		"\u022e\u0117\u0000\u04dc\u04dd\u0003\u008eG\u0000\u04dd\u04df\u0001\u0000"+
		"\u0000\u0000\u04de\u04db\u0001\u0000\u0000\u0000\u04de\u04df\u0001\u0000"+
		"\u0000\u0000\u04dfu\u0001\u0000\u0000\u0000\u04e0\u04e1\u0003\u019a\u00cd"+
		"\u0000\u04e1\u04e3\u0003\u0224\u0112\u0000\u04e2\u04e4\u0003\u00fc~\u0000"+
		"\u04e3\u04e2\u0001\u0000\u0000\u0000\u04e3\u04e4\u0001\u0000\u0000\u0000"+
		"\u04e4\u04e5\u0001\u0000\u0000\u0000\u04e5\u04e6\u0003\u0154\u00aa\u0000"+
		"\u04e6w\u0001\u0000\u0000\u0000\u04e7\u04e8\u0003\u019a\u00cd\u0000\u04e8"+
		"\u04ea\u0003\u021a\u010d\u0000\u04e9\u04eb\u0003\u00fc~\u0000\u04ea\u04e9"+
		"\u0001\u0000\u0000\u0000\u04ea\u04eb\u0001\u0000\u0000\u0000\u04eb\u04ef"+
		"\u0001\u0000\u0000\u0000\u04ec\u04ed\u0003\u0136\u009b\u0000\u04ed\u04ee"+
		"\u0005\n\u0000\u0000\u04ee\u04f0\u0001\u0000\u0000\u0000\u04ef\u04ec\u0001"+
		"\u0000\u0000\u0000\u04ef\u04f0\u0001\u0000\u0000\u0000\u04f0\u04f1\u0001"+
		"\u0000\u0000\u0000\u04f1\u04f2\u0003\u014c\u00a6\u0000\u04f2y\u0001\u0000"+
		"\u0000\u0000\u04f3\u04f4\u0003\u019a\u00cd\u0000\u04f4\u04f5\u0003\u01d8"+
		"\u00ec\u0000\u04f5\u04f7\u0003\u022a\u0115\u0000\u04f6\u04f8\u0003\u00fc"+
		"~\u0000\u04f7\u04f6\u0001\u0000\u0000\u0000\u04f7\u04f8\u0001\u0000\u0000"+
		"\u0000\u04f8\u04fc\u0001\u0000\u0000\u0000\u04f9\u04fa\u0003\u0136\u009b"+
		"\u0000\u04fa\u04fb\u0005\n\u0000\u0000\u04fb\u04fd\u0001\u0000\u0000\u0000"+
		"\u04fc\u04f9\u0001\u0000\u0000\u0000\u04fc\u04fd\u0001\u0000\u0000\u0000"+
		"\u04fd\u04fe\u0001\u0000\u0000\u0000\u04fe\u04ff\u0003\u014a\u00a5\u0000"+
		"\u04ff{\u0001\u0000\u0000\u0000\u0500\u0501\u0003\u019a\u00cd\u0000\u0501"+
		"\u0503\u0003\u0160\u00b0\u0000\u0502\u0504\u0003\u00fc~\u0000\u0503\u0502"+
		"\u0001\u0000\u0000\u0000\u0503\u0504\u0001\u0000\u0000\u0000\u0504\u0508"+
		"\u0001\u0000\u0000\u0000\u0505\u0506\u0003\u0136\u009b\u0000\u0506\u0507"+
		"\u0005\n\u0000\u0000\u0507\u0509\u0001\u0000\u0000\u0000\u0508\u0505\u0001"+
		"\u0000\u0000\u0000\u0508\u0509\u0001\u0000\u0000\u0000\u0509\u050a\u0001"+
		"\u0000\u0000\u0000\u050a\u050b\u0003\u014e\u00a7\u0000\u050b}\u0001\u0000"+
		"\u0000\u0000\u050c\u050d\u0003\u019a\u00cd\u0000\u050d\u050f\u0003\u01ac"+
		"\u00d6\u0000\u050e\u0510\u0003\u00fc~\u0000\u050f\u050e\u0001\u0000\u0000"+
		"\u0000\u050f\u0510\u0001\u0000\u0000\u0000\u0510\u0514\u0001\u0000\u0000"+
		"\u0000\u0511\u0512\u0003\u0136\u009b\u0000\u0512\u0513\u0005\n\u0000\u0000"+
		"\u0513\u0515\u0001\u0000\u0000\u0000\u0514\u0511\u0001\u0000\u0000\u0000"+
		"\u0514\u0515\u0001\u0000\u0000\u0000\u0515\u0516\u0001\u0000\u0000\u0000"+
		"\u0516\u0517\u0003\u0150\u00a8\u0000\u0517\u007f\u0001\u0000\u0000\u0000"+
		"\u0518\u0519\u0003\u019a\u00cd\u0000\u0519\u051b\u0003\u0214\u010a\u0000"+
		"\u051a\u051c\u0003\u00fc~\u0000\u051b\u051a\u0001\u0000\u0000\u0000\u051b"+
		"\u051c\u0001\u0000\u0000\u0000\u051c\u051d\u0001\u0000\u0000\u0000\u051d"+
		"\u051e\u0003\u0146\u00a3\u0000\u051e\u0522\u0003\u01e6\u00f3\u0000\u051f"+
		"\u0520\u0003\u0136\u009b\u0000\u0520\u0521\u0005\n\u0000\u0000\u0521\u0523"+
		"\u0001\u0000\u0000\u0000\u0522\u051f\u0001\u0000\u0000\u0000\u0522\u0523"+
		"\u0001\u0000\u0000\u0000\u0523\u0524\u0001\u0000\u0000\u0000\u0524\u0525"+
		"\u0003\u0138\u009c\u0000\u0525\u0081\u0001\u0000\u0000\u0000\u0526\u0527"+
		"\u0003\u019a\u00cd\u0000\u0527\u0529\u0003\u01fa\u00fd\u0000\u0528\u052a"+
		"\u0003\u00fc~\u0000\u0529\u0528\u0001\u0000\u0000\u0000\u0529\u052a\u0001"+
		"\u0000\u0000\u0000\u052a\u052b\u0001\u0000\u0000\u0000\u052b\u052c\u0003"+
		"\u0144\u00a2\u0000\u052c\u0083\u0001\u0000\u0000\u0000\u052d\u052e\u0003"+
		"\u019a\u00cd\u0000\u052e\u0530\u0003\u020c\u0106\u0000\u052f\u0531\u0003"+
		"\u00fc~\u0000\u0530\u052f\u0001\u0000\u0000\u0000\u0530\u0531\u0001\u0000"+
		"\u0000\u0000\u0531\u0535\u0001\u0000\u0000\u0000\u0532\u0533\u0003\u0136"+
		"\u009b\u0000\u0533\u0534\u0005\n\u0000\u0000\u0534\u0536\u0001\u0000\u0000"+
		"\u0000\u0535\u0532\u0001\u0000\u0000\u0000\u0535\u0536\u0001\u0000\u0000"+
		"\u0000\u0536\u0537\u0001\u0000\u0000\u0000\u0537\u0538\u0003\u0138\u009c"+
		"\u0000\u0538\u0085\u0001\u0000\u0000\u0000\u0539\u053a\u0003\u019a\u00cd"+
		"\u0000\u053a\u053c\u0003\u01c8\u00e4\u0000\u053b\u053d\u0003\u00fc~\u0000"+
		"\u053c\u053b\u0001\u0000\u0000\u0000\u053c\u053d\u0001\u0000\u0000\u0000"+
		"\u053d\u053e\u0001\u0000\u0000\u0000\u053e\u053f\u0003\u0136\u009b\u0000"+
		"\u053f\u0087\u0001\u0000\u0000\u0000\u0540\u0541\u0003\u019a\u00cd\u0000"+
		"\u0541\u0543\u0003\u01b6\u00db\u0000\u0542\u0544\u0003\u00fc~\u0000\u0543"+
		"\u0542\u0001\u0000\u0000\u0000\u0543\u0544\u0001\u0000\u0000\u0000\u0544"+
		"\u0548\u0001\u0000\u0000\u0000\u0545\u0546\u0003\u0136\u009b\u0000\u0546"+
		"\u0547\u0005\n\u0000\u0000\u0547\u0549\u0001\u0000\u0000\u0000\u0548\u0545"+
		"\u0001\u0000\u0000\u0000\u0548\u0549\u0001\u0000\u0000\u0000\u0549\u054a"+
		"\u0001\u0000\u0000\u0000\u054a\u054b\u0003\u00ccf\u0000\u054b\u0089\u0001"+
		"\u0000\u0000\u0000\u054c\u054d\u0003\u0190\u00c8\u0000\u054d\u054f\u0003"+
		"\u020c\u0106\u0000\u054e\u0550\u0003\u00fa}\u0000\u054f\u054e\u0001\u0000"+
		"\u0000\u0000\u054f\u0550\u0001\u0000\u0000\u0000\u0550\u0554\u0001\u0000"+
		"\u0000\u0000\u0551\u0552\u0003\u0136\u009b\u0000\u0552\u0553\u0005\n\u0000"+
		"\u0000\u0553\u0555\u0001\u0000\u0000\u0000\u0554\u0551\u0001\u0000\u0000"+
		"\u0000\u0554\u0555\u0001\u0000\u0000\u0000\u0555\u0556\u0001\u0000\u0000"+
		"\u0000\u0556\u0557\u0003\u0138\u009c\u0000\u0557\u0558\u0003\u0232\u0119"+
		"\u0000\u0558\u0559\u0003\u00a0P\u0000\u0559\u055b\u0003\u0234\u011a\u0000"+
		"\u055a\u055c\u0003\u008cF\u0000\u055b\u055a\u0001\u0000\u0000\u0000\u055b"+
		"\u055c\u0001\u0000\u0000\u0000\u055c\u008b\u0001\u0000\u0000\u0000\u055d"+
		"\u055e\u0003\u022e\u0117\u0000\u055e\u055f\u0003\u008eG\u0000\u055f\u008d"+
		"\u0001\u0000\u0000\u0000\u0560\u0561\u0003\u0182\u00c1\u0000\u0561\u0565"+
		"\u0003\u0206\u0103\u0000\u0562\u0563\u0003\u016c\u00b6\u0000\u0563\u0564"+
		"\u0003\u008eG\u0000\u0564\u0566\u0001\u0000\u0000\u0000\u0565\u0562\u0001"+
		"\u0000\u0000\u0000\u0565\u0566\u0001\u0000\u0000\u0000\u0566\u0577\u0001"+
		"\u0000\u0000\u0000\u0567\u056b\u0003\u0090H\u0000\u0568\u0569\u0003\u016c"+
		"\u00b6\u0000\u0569\u056a\u0003\u008eG\u0000\u056a\u056c\u0001\u0000\u0000"+
		"\u0000\u056b\u0568\u0001\u0000\u0000\u0000\u056b\u056c\u0001\u0000\u0000"+
		"\u0000\u056c\u0577\u0001\u0000\u0000\u0000\u056d\u0573\u0003\u0092I\u0000"+
		"\u056e\u056f\u0003\u016c\u00b6\u0000\u056f\u0570\u0003\u0092I\u0000\u0570"+
		"\u0572\u0001\u0000\u0000\u0000\u0571\u056e\u0001\u0000\u0000\u0000\u0572"+
		"\u0575\u0001\u0000\u0000\u0000\u0573\u0571\u0001\u0000\u0000\u0000\u0573"+
		"\u0574\u0001\u0000\u0000\u0000\u0574\u0577\u0001\u0000\u0000\u0000\u0575"+
		"\u0573\u0001\u0000\u0000\u0000\u0576\u0560\u0001\u0000\u0000\u0000\u0576"+
		"\u0567\u0001\u0000\u0000\u0000\u0576\u056d\u0001\u0000\u0000\u0000\u0577"+
		"\u008f\u0001\u0000\u0000\u0000\u0578\u0579\u0003\u0180\u00c0\u0000\u0579"+
		"\u057a\u0003\u01ec\u00f6\u0000\u057a\u057b\u0003\u017a\u00bd\u0000\u057b"+
		"\u057c\u0003\u0232\u0119\u0000\u057c\u057e\u0003\u013a\u009d\u0000\u057d"+
		"\u057f\u0003\u0142\u00a1\u0000\u057e\u057d\u0001\u0000\u0000\u0000\u057e"+
		"\u057f\u0001\u0000\u0000\u0000\u057f\u0587\u0001\u0000\u0000\u0000\u0580"+
		"\u0581\u0003\u0242\u0121\u0000\u0581\u0583\u0003\u013a\u009d\u0000\u0582"+
		"\u0584\u0003\u0142\u00a1\u0000\u0583\u0582\u0001\u0000\u0000\u0000\u0583"+
		"\u0584\u0001\u0000\u0000\u0000\u0584\u0586\u0001\u0000\u0000\u0000\u0585"+
		"\u0580\u0001\u0000\u0000\u0000\u0586\u0589\u0001\u0000\u0000\u0000\u0587"+
		"\u0585\u0001\u0000\u0000\u0000\u0587\u0588\u0001\u0000\u0000\u0000\u0588"+
		"\u058a\u0001\u0000\u0000\u0000\u0589\u0587\u0001\u0000\u0000\u0000\u058a"+
		"\u058b\u0003\u0234\u011a\u0000\u058b\u0091\u0001\u0000\u0000\u0000\u058c"+
		"\u058d\u0003\u0094J\u0000\u058d\u058e\u0005\u0013\u0000\u0000\u058e\u058f"+
		"\u0003\u0096K\u0000\u058f\u0595\u0001\u0000\u0000\u0000\u0590\u0591\u0003"+
		"\u0094J\u0000\u0591\u0592\u0005\u0013\u0000\u0000\u0592\u0593\u0003\u0098"+
		"L\u0000\u0593\u0595\u0001\u0000\u0000\u0000\u0594\u058c\u0001\u0000\u0000"+
		"\u0000\u0594\u0590\u0001\u0000\u0000\u0000\u0595\u0093\u0001\u0000\u0000"+
		"\u0000\u0596\u0597\u0005\u00b2\u0000\u0000\u0597\u0095\u0001\u0000\u0000"+
		"\u0000\u0598\u059b\u0003\u0130\u0098\u0000\u0599\u059b\u0003\u012e\u0097"+
		"\u0000\u059a\u0598\u0001\u0000\u0000\u0000\u059a\u0599\u0001\u0000\u0000"+
		"\u0000\u059b\u0097\u0001\u0000\u0000\u0000\u059c\u059d\u0003\u0236\u011b"+
		"\u0000\u059d\u05a3\u0003\u009aM\u0000\u059e\u059f\u0003\u0242\u0121\u0000"+
		"\u059f\u05a0\u0003\u009aM\u0000\u05a0\u05a2\u0001\u0000\u0000\u0000\u05a1"+
		"\u059e\u0001\u0000\u0000\u0000\u05a2\u05a5\u0001\u0000\u0000\u0000\u05a3"+
		"\u05a1\u0001\u0000\u0000\u0000\u05a3\u05a4\u0001\u0000\u0000\u0000\u05a4"+
		"\u05a6\u0001\u0000\u0000\u0000\u05a5\u05a3\u0001\u0000\u0000\u0000\u05a6"+
		"\u05a7\u0003\u0238\u011c\u0000\u05a7\u0099\u0001\u0000\u0000\u0000\u05a8"+
		"\u05a9\u0003\u009cN\u0000\u05a9\u05aa\u0005\t\u0000\u0000\u05aa\u05ab"+
		"\u0003\u009eO\u0000\u05ab\u009b\u0001\u0000\u0000\u0000\u05ac\u05ad\u0003"+
		"\u0130\u0098\u0000\u05ad\u009d\u0001\u0000\u0000\u0000\u05ae\u05b1\u0003"+
		"\u0130\u0098\u0000\u05af\u05b1\u0003\u012e\u0097\u0000\u05b0\u05ae\u0001"+
		"\u0000\u0000\u0000\u05b0\u05af\u0001\u0000\u0000\u0000\u05b1\u009f\u0001"+
		"\u0000\u0000\u0000\u05b2\u05b8\u0003\u00a2Q\u0000\u05b3\u05b4\u0003\u0242"+
		"\u0121\u0000\u05b4\u05b5\u0003\u00a2Q\u0000\u05b5\u05b7\u0001\u0000\u0000"+
		"\u0000\u05b6\u05b3\u0001\u0000\u0000\u0000\u05b7\u05ba\u0001\u0000\u0000"+
		"\u0000\u05b8\u05b6\u0001\u0000\u0000\u0000\u05b8\u05b9\u0001\u0000\u0000"+
		"\u0000\u05b9\u05be\u0001\u0000\u0000\u0000\u05ba\u05b8\u0001\u0000\u0000"+
		"\u0000\u05bb\u05bc\u0003\u0242\u0121\u0000\u05bc\u05bd\u0003\u00a6S\u0000"+
		"\u05bd\u05bf\u0001\u0000\u0000\u0000\u05be\u05bb\u0001\u0000\u0000\u0000"+
		"\u05be\u05bf\u0001\u0000\u0000\u0000\u05bf\u00a1\u0001\u0000\u0000\u0000"+
		"\u05c0\u05c1\u0003\u013a\u009d\u0000\u05c1\u05c3\u0003\u013c\u009e\u0000"+
		"\u05c2\u05c4\u0003\u00a4R\u0000\u05c3\u05c2\u0001\u0000\u0000\u0000\u05c3"+
		"\u05c4\u0001\u0000\u0000\u0000\u05c4\u00a3\u0001\u0000\u0000\u0000\u05c5"+
		"\u05c6\u0003\u01f0\u00f8\u0000\u05c6\u05c7\u0003\u01c4\u00e2\u0000\u05c7"+
		"\u00a5\u0001\u0000\u0000\u0000\u05c8\u05c9\u0003\u01f0\u00f8\u0000\u05c9"+
		"\u05ca\u0003\u01c4\u00e2\u0000\u05ca\u05cb\u0003\u0232\u0119\u0000\u05cb"+
		"\u05cc\u0003\u00a8T\u0000\u05cc\u05cd\u0003\u0234\u011a\u0000\u05cd\u00a7"+
		"\u0001\u0000\u0000\u0000\u05ce\u05d2\u0003\u00aaU\u0000\u05cf\u05d2\u0003"+
		"\u00acV\u0000\u05d0\u05d2\u0003\u00aeW\u0000\u05d1\u05ce\u0001\u0000\u0000"+
		"\u0000\u05d1\u05cf\u0001\u0000\u0000\u0000\u05d1\u05d0\u0001\u0000\u0000"+
		"\u0000\u05d2\u00a9\u0001\u0000\u0000\u0000\u05d3\u05d4\u0003\u013a\u009d"+
		"\u0000\u05d4\u00ab\u0001\u0000\u0000\u0000\u05d5\u05d6\u0003\u00b4Z\u0000"+
		"\u05d6\u05d7\u0003\u0242\u0121\u0000\u05d7\u05d8\u0003\u00b2Y\u0000\u05d8"+
		"\u00ad\u0001\u0000\u0000\u0000\u05d9\u05da\u0003\u0232\u0119\u0000\u05da"+
		"\u05db\u0003\u00b0X\u0000\u05db\u05dc\u0003\u0234\u011a\u0000\u05dc\u05dd"+
		"\u0003\u0242\u0121\u0000\u05dd\u05de\u0003\u00b2Y\u0000\u05de\u00af\u0001"+
		"\u0000\u0000\u0000\u05df\u05e5\u0003\u00b4Z\u0000\u05e0\u05e1\u0003\u0242"+
		"\u0121\u0000\u05e1\u05e2\u0003\u00b4Z\u0000\u05e2\u05e4\u0001\u0000\u0000"+
		"\u0000\u05e3\u05e0\u0001\u0000\u0000\u0000\u05e4\u05e7\u0001\u0000\u0000"+
		"\u0000\u05e5\u05e3\u0001\u0000\u0000\u0000\u05e5\u05e6\u0001\u0000\u0000"+
		"\u0000\u05e6\u00b1\u0001\u0000\u0000\u0000\u05e7\u05e5\u0001\u0000\u0000"+
		"\u0000\u05e8\u05ee\u0003\u00b6[\u0000\u05e9\u05ea\u0003\u0242\u0121\u0000"+
		"\u05ea\u05eb\u0003\u00b6[\u0000\u05eb\u05ed\u0001\u0000\u0000\u0000\u05ec"+
		"\u05e9\u0001\u0000\u0000\u0000\u05ed\u05f0\u0001\u0000\u0000\u0000\u05ee"+
		"\u05ec\u0001\u0000\u0000\u0000\u05ee\u05ef\u0001\u0000\u0000\u0000\u05ef"+
		"\u00b3\u0001\u0000\u0000\u0000\u05f0\u05ee\u0001\u0000\u0000\u0000\u05f1"+
		"\u05f2\u0003\u013a\u009d\u0000\u05f2\u00b5\u0001\u0000\u0000\u0000\u05f3"+
		"\u05f4\u0003\u013a\u009d\u0000\u05f4\u00b7\u0001\u0000\u0000\u0000\u05f5"+
		"\u05f6\u0003\u016e\u00b7\u0000\u05f6\u05f7\u0003\u0176\u00bb\u0000\u05f7"+
		"\u00b9\u0001\u0000\u0000\u0000\u05f8\u05fa\u0003\u0178\u00bc\u0000\u05f9"+
		"\u05fb\u0003\u00bc^\u0000\u05fa\u05f9\u0001\u0000\u0000\u0000\u05fa\u05fb"+
		"\u0001\u0000\u0000\u0000\u05fb\u05fc\u0001\u0000\u0000\u0000\u05fc\u05fe"+
		"\u0003\u0176\u00bb\u0000\u05fd\u05ff\u0003\u00f8|\u0000\u05fe\u05fd\u0001"+
		"\u0000\u0000\u0000\u05fe\u05ff\u0001\u0000\u0000\u0000\u05ff\u00bb\u0001"+
		"\u0000\u0000\u0000\u0600\u0603\u0003\u01d2\u00e9\u0000\u0601\u0603\u0003"+
		"\u021e\u010f\u0000\u0602\u0600\u0001\u0000\u0000\u0000\u0602\u0601\u0001"+
		"\u0000\u0000\u0000\u0603\u00bd\u0001\u0000\u0000\u0000\u0604\u0605\u0003"+
		"\u016a\u00b5\u0000\u0605\u0606\u0003\u01c8\u00e4\u0000\u0606\u0607\u0003"+
		"\u0136\u009b\u0000\u0607\u0608\u0003\u022e\u0117\u0000\u0608\u0609\u0003"+
		"\u01f6\u00fb\u0000\u0609\u060a\u0005\u0013\u0000\u0000\u060a\u060b\u0003"+
		"\u0236\u011b\u0000\u060b\u060c\u0003\u00c0`\u0000\u060c\u0610\u0003\u0238"+
		"\u011c\u0000\u060d\u060e\u0003\u016c\u00b6\u0000\u060e\u060f\u0003\u00c4"+
		"b\u0000\u060f\u0611\u0001\u0000\u0000\u0000\u0610\u060d\u0001\u0000\u0000"+
		"\u0000\u0610\u0611\u0001\u0000\u0000\u0000\u0611\u00bf\u0001\u0000\u0000"+
		"\u0000\u0612\u0618\u0003\u00c2a\u0000\u0613\u0614\u0003\u0242\u0121\u0000"+
		"\u0614\u0615\u0003\u00c2a\u0000\u0615\u0617\u0001\u0000\u0000\u0000\u0616"+
		"\u0613\u0001\u0000\u0000\u0000\u0617\u061a\u0001\u0000\u0000\u0000\u0618"+
		"\u0616\u0001\u0000\u0000\u0000\u0618\u0619\u0001\u0000\u0000\u0000\u0619"+
		"\u00c1\u0001\u0000\u0000\u0000\u061a\u0618\u0001\u0000\u0000\u0000\u061b"+
		"\u061c\u0005\u00ad\u0000\u0000\u061c\u061d\u0005\t\u0000\u0000\u061d\u0622"+
		"\u0005\u00ad\u0000\u0000\u061e\u061f\u0005\u00ad\u0000\u0000\u061f\u0620"+
		"\u0005\t\u0000\u0000\u0620\u0622\u0005\u00ae\u0000\u0000\u0621\u061b\u0001"+
		"\u0000\u0000\u0000\u0621\u061e\u0001\u0000\u0000\u0000\u0622\u00c3\u0001"+
		"\u0000\u0000\u0000\u0623\u0624\u0003\u019c\u00ce\u0000\u0624\u0625\u0005"+
		"\u0013\u0000\u0000\u0625\u0626\u0003\u0132\u0099\u0000\u0626\u00c5\u0001"+
		"\u0000\u0000\u0000\u0627\u0628\u0003\u0222\u0111\u0000\u0628\u0629\u0003"+
		"\u0136\u009b\u0000\u0629\u00c7\u0001\u0000\u0000\u0000\u062a\u062c\u0003"+
		"\u0216\u010b\u0000\u062b\u062d\u0003\u020c\u0106\u0000\u062c\u062b\u0001"+
		"\u0000\u0000\u0000\u062c\u062d\u0001\u0000\u0000\u0000\u062d\u0631\u0001"+
		"\u0000\u0000\u0000\u062e\u062f\u0003\u0136\u009b\u0000\u062f\u0630\u0005"+
		"\n\u0000\u0000\u0630\u0632\u0001\u0000\u0000\u0000\u0631\u062e\u0001\u0000"+
		"\u0000\u0000\u0631\u0632\u0001\u0000\u0000\u0000\u0632\u0633\u0001\u0000"+
		"\u0000\u0000\u0633\u0634\u0003\u0138\u009c\u0000\u0634\u00c9\u0001\u0000"+
		"\u0000\u0000\u0635\u0636\u0003\u0190\u00c8\u0000\u0636\u0638\u0003\u01b6"+
		"\u00db\u0000\u0637\u0639\u0003\u00fa}\u0000\u0638\u0637\u0001\u0000\u0000"+
		"\u0000\u0638\u0639\u0001\u0000\u0000\u0000\u0639\u063b\u0001\u0000\u0000"+
		"\u0000\u063a\u063c\u0003\u00ccf\u0000\u063b\u063a\u0001\u0000\u0000\u0000"+
		"\u063b\u063c\u0001\u0000\u0000\u0000\u063c\u063d\u0001\u0000\u0000\u0000"+
		"\u063d\u0641\u0003\u01e6\u00f3\u0000\u063e\u063f\u0003\u0136\u009b\u0000"+
		"\u063f\u0640\u0005\n\u0000\u0000\u0640\u0642\u0001\u0000\u0000\u0000\u0641"+
		"\u063e\u0001\u0000\u0000\u0000\u0641\u0642\u0001\u0000\u0000\u0000\u0642"+
		"\u0643\u0001\u0000\u0000\u0000\u0643\u0644\u0003\u0138\u009c\u0000\u0644"+
		"\u0645\u0003\u0232\u0119\u0000\u0645\u0646\u0003\u00ceg\u0000\u0646\u0647"+
		"\u0003\u0234\u011a\u0000\u0647\u00cb\u0001\u0000\u0000\u0000\u0648\u064b"+
		"\u0005\u00b2\u0000\u0000\u0649\u064b\u0003\u0130\u0098\u0000\u064a\u0648"+
		"\u0001\u0000\u0000\u0000\u064a\u0649\u0001\u0000\u0000\u0000\u064b\u00cd"+
		"\u0001\u0000\u0000\u0000\u064c\u0651\u0003\u013a\u009d\u0000\u064d\u0651"+
		"\u0003\u00d0h\u0000\u064e\u0651\u0003\u00d2i\u0000\u064f\u0651\u0003\u00d4"+
		"j\u0000\u0650\u064c\u0001\u0000\u0000\u0000\u0650\u064d\u0001\u0000\u0000"+
		"\u0000\u0650\u064e\u0001\u0000\u0000\u0000\u0650\u064f\u0001\u0000\u0000"+
		"\u0000\u0651\u00cf\u0001\u0000\u0000\u0000\u0652\u0653\u0003\u01c6\u00e3"+
		"\u0000\u0653\u0654\u0003\u0232\u0119\u0000\u0654\u0655\u0005\u00b2\u0000"+
		"\u0000\u0655\u0656\u0003\u0234\u011a\u0000\u0656\u00d1\u0001\u0000\u0000"+
		"\u0000\u0657\u0658\u0003\u019e\u00cf\u0000\u0658\u0659\u0003\u0232\u0119"+
		"\u0000\u0659\u065a\u0005\u00b2\u0000\u0000\u065a\u065b\u0003\u0234\u011a"+
		"\u0000\u065b\u00d3\u0001\u0000\u0000\u0000\u065c\u065d\u0003\u01aa\u00d5"+
		"\u0000\u065d\u065e\u0003\u0232\u0119\u0000\u065e\u065f\u0005\u00b2\u0000"+
		"\u0000\u065f\u0660\u0003\u0234\u011a\u0000\u0660\u00d5\u0001\u0000\u0000"+
		"\u0000\u0661\u0663\u0003\u00ba]\u0000\u0662\u0661\u0001\u0000\u0000\u0000"+
		"\u0662\u0663\u0001\u0000\u0000\u0000\u0663\u0664\u0001\u0000\u0000\u0000"+
		"\u0664\u0666\u0003\u0192\u00c9\u0000\u0665\u0667\u0003\u00d8l\u0000\u0666"+
		"\u0665\u0001\u0000\u0000\u0000\u0666\u0667\u0001\u0000\u0000\u0000\u0667"+
		"\u0668\u0001\u0000\u0000\u0000\u0668\u066a\u0003\u010e\u0087\u0000\u0669"+
		"\u066b\u0003\u00f8|\u0000\u066a\u0669\u0001\u0000\u0000\u0000\u066a\u066b"+
		"\u0001\u0000\u0000\u0000\u066b\u066c\u0001\u0000\u0000\u0000\u066c\u066f"+
		"\u0003\u0116\u008b\u0000\u066d\u0670\u0003\u00fc~\u0000\u066e\u0670\u0003"+
		"\u00deo\u0000\u066f\u066d\u0001\u0000\u0000\u0000\u066f\u066e\u0001\u0000"+
		"\u0000\u0000\u066f\u0670\u0001\u0000\u0000\u0000\u0670\u00d7\u0001\u0000"+
		"\u0000\u0000\u0671\u0677\u0003\u00dam\u0000\u0672\u0673\u0003\u0242\u0121"+
		"\u0000\u0673\u0674\u0003\u00dam\u0000\u0674\u0676\u0001\u0000\u0000\u0000"+
		"\u0675\u0672\u0001\u0000\u0000\u0000\u0676\u0679\u0001\u0000\u0000\u0000"+
		"\u0677\u0675\u0001\u0000\u0000\u0000\u0677\u0678\u0001\u0000\u0000\u0000"+
		"\u0678\u00d9\u0001\u0000\u0000\u0000\u0679\u0677\u0001\u0000\u0000\u0000"+
		"\u067a\u0684\u0005\u00b2\u0000\u0000\u067b\u067c\u0005\u00b2\u0000\u0000"+
		"\u067c\u067f\u0005\u0005\u0000\u0000\u067d\u0680\u0003\u0130\u0098\u0000"+
		"\u067e\u0680\u0003\u012c\u0096\u0000\u067f\u067d\u0001\u0000\u0000\u0000"+
		"\u067f\u067e\u0001\u0000\u0000\u0000\u0680\u0681\u0001\u0000\u0000\u0000"+
		"\u0681\u0682\u0005\u0006\u0000\u0000\u0682\u0684\u0001\u0000\u0000\u0000"+
		"\u0683\u067a\u0001\u0000\u0000\u0000\u0683\u067b\u0001\u0000\u0000\u0000"+
		"\u0684\u00db\u0001\u0000\u0000\u0000\u0685\u0687\u0003\u00ba]\u0000\u0686"+
		"\u0685\u0001\u0000\u0000\u0000\u0686\u0687\u0001\u0000\u0000\u0000\u0687"+
		"\u0688\u0001\u0000\u0000\u0000\u0688\u068c\u0003\u0220\u0110\u0000\u0689"+
		"\u068a\u0003\u0136\u009b\u0000\u068a\u068b\u0005\n\u0000\u0000\u068b\u068d"+
		"\u0001\u0000\u0000\u0000\u068c\u0689\u0001\u0000\u0000\u0000\u068c\u068d"+
		"\u0001\u0000\u0000\u0000\u068d\u068e\u0001\u0000\u0000\u0000\u068e\u0690"+
		"\u0003\u0138\u009c\u0000\u068f\u0691\u0003\u00f2y\u0000\u0690\u068f\u0001"+
		"\u0000\u0000\u0000\u0690\u0691\u0001\u0000\u0000\u0000\u0691\u0692\u0001"+
		"\u0000\u0000\u0000\u0692\u0693\u0003\u0202\u0101\u0000\u0693\u0694\u0003"+
		"\u00e4r\u0000\u0694\u0697\u0003\u0116\u008b\u0000\u0695\u0698\u0003\u00fc"+
		"~\u0000\u0696\u0698\u0003\u00deo\u0000\u0697\u0695\u0001\u0000\u0000\u0000"+
		"\u0697\u0696\u0001\u0000\u0000\u0000\u0697\u0698\u0001\u0000\u0000\u0000"+
		"\u0698\u00dd\u0001\u0000\u0000\u0000\u0699\u069a\u0003\u01b2\u00d9\u0000"+
		"\u069a\u069b\u0003\u00e0p\u0000\u069b\u00df\u0001\u0000\u0000\u0000\u069c"+
		"\u06a2\u0003\u00e2q\u0000\u069d\u069e\u0003\u016c\u00b6\u0000\u069e\u069f"+
		"\u0003\u00e2q\u0000\u069f\u06a1\u0001\u0000\u0000\u0000\u06a0\u069d\u0001"+
		"\u0000\u0000\u0000\u06a1\u06a4\u0001\u0000\u0000\u0000\u06a2\u06a0\u0001"+
		"\u0000\u0000\u0000\u06a2\u06a3\u0001\u0000\u0000\u0000\u06a3\u00e1\u0001"+
		"\u0000\u0000\u0000\u06a4\u06a2\u0001\u0000\u0000\u0000\u06a5\u06a6\u0005"+
		"\u00b2\u0000\u0000\u06a6\u06a7\u0005\u0013\u0000\u0000\u06a7\u06a8\u0003"+
		"\u012a\u0095\u0000\u06a8\u00e3\u0001\u0000\u0000\u0000\u06a9\u06af\u0003"+
		"\u00e6s\u0000\u06aa\u06ab\u0003\u0242\u0121\u0000\u06ab\u06ac\u0003\u00e6"+
		"s\u0000\u06ac\u06ae\u0001\u0000\u0000\u0000\u06ad\u06aa\u0001\u0000\u0000"+
		"\u0000\u06ae\u06b1\u0001\u0000\u0000\u0000\u06af\u06ad\u0001\u0000\u0000"+
		"\u0000\u06af\u06b0\u0001\u0000\u0000\u0000\u06b0\u00e5\u0001\u0000\u0000"+
		"\u0000\u06b1\u06af\u0001\u0000\u0000\u0000\u06b2\u06b3\u0005\u00b2\u0000"+
		"\u0000\u06b3\u06b8\u0005\u0013\u0000\u0000\u06b4\u06b9\u0003\u012a\u0095"+
		"\u0000\u06b5\u06b9\u0003\u00eau\u0000\u06b6\u06b9\u0003\u00e8t\u0000\u06b7"+
		"\u06b9\u0003\u00ecv\u0000\u06b8\u06b4\u0001\u0000\u0000\u0000\u06b8\u06b5"+
		"\u0001\u0000\u0000\u0000\u06b8\u06b6\u0001\u0000\u0000\u0000\u06b8\u06b7"+
		"\u0001\u0000\u0000\u0000\u06b9\u06e8\u0001\u0000\u0000\u0000\u06ba\u06bb"+
		"\u0005\u00b2\u0000\u0000\u06bb\u06bc\u0005\u0013\u0000\u0000\u06bc\u06bd"+
		"\u0005\u00b2\u0000\u0000\u06bd\u06be\u0007\u0001\u0000\u0000\u06be\u06e8"+
		"\u0003\u012c\u0096\u0000\u06bf\u06c0\u0005\u00b2\u0000\u0000\u06c0\u06c1"+
		"\u0005\u0013\u0000\u0000\u06c1\u06c2\u0005\u00b2\u0000\u0000\u06c2\u06c3"+
		"\u0007\u0001\u0000\u0000\u06c3\u06e8\u0003\u00e8t\u0000\u06c4\u06c5\u0005"+
		"\u00b2\u0000\u0000\u06c5\u06c6\u0005\u0013\u0000\u0000\u06c6\u06c7\u0003"+
		"\u00e8t\u0000\u06c7\u06c8\u0007\u0001\u0000\u0000\u06c8\u06c9\u0005\u00b2"+
		"\u0000\u0000\u06c9\u06e8\u0001\u0000\u0000\u0000\u06ca\u06cb\u0005\u00b2"+
		"\u0000\u0000\u06cb\u06cc\u0005\u0013\u0000\u0000\u06cc\u06cd\u0005\u00b2"+
		"\u0000\u0000\u06cd\u06ce\u0007\u0001\u0000\u0000\u06ce\u06e8\u0003\u00ea"+
		"u\u0000\u06cf\u06d0\u0005\u00b2\u0000\u0000\u06d0\u06d1\u0005\u0013\u0000"+
		"\u0000\u06d1\u06d2\u0003\u00eau\u0000\u06d2\u06d3\u0007\u0001\u0000\u0000"+
		"\u06d3\u06d4\u0005\u00b2\u0000\u0000\u06d4\u06e8\u0001\u0000\u0000\u0000"+
		"\u06d5\u06d6\u0005\u00b2\u0000\u0000\u06d6\u06d7\u0005\u0013\u0000\u0000"+
		"\u06d7\u06d8\u0005\u00b2\u0000\u0000\u06d8\u06d9\u0007\u0001\u0000\u0000"+
		"\u06d9\u06e8\u0003\u00ecv\u0000\u06da\u06db\u0005\u00b2\u0000\u0000\u06db"+
		"\u06dc\u0005\u0013\u0000\u0000\u06dc\u06dd\u0003\u00ecv\u0000\u06dd\u06de"+
		"\u0007\u0001\u0000\u0000\u06de\u06df\u0005\u00b2\u0000\u0000\u06df\u06e8"+
		"\u0001\u0000\u0000\u0000\u06e0\u06e1\u0005\u00b2\u0000\u0000\u06e1\u06e2"+
		"\u0003\u023e\u011f\u0000\u06e2\u06e3\u0003\u012c\u0096\u0000\u06e3\u06e4"+
		"\u0003\u0240\u0120\u0000\u06e4\u06e5\u0005\u0013\u0000\u0000\u06e5\u06e6"+
		"\u0003\u012a\u0095\u0000\u06e6\u06e8\u0001\u0000\u0000\u0000\u06e7\u06b2"+
		"\u0001\u0000\u0000\u0000\u06e7\u06ba\u0001\u0000\u0000\u0000\u06e7\u06bf"+
		"\u0001\u0000\u0000\u0000\u06e7\u06c4\u0001\u0000\u0000\u0000\u06e7\u06ca"+
		"\u0001\u0000\u0000\u0000\u06e7\u06cf\u0001\u0000\u0000\u0000\u06e7\u06d5"+
		"\u0001\u0000\u0000\u0000\u06e7\u06da\u0001\u0000\u0000\u0000\u06e7\u06e0"+
		"\u0001\u0000\u0000\u0000\u06e8\u00e7\u0001\u0000\u0000\u0000\u06e9\u06f3"+
		"\u0003\u0236\u011b\u0000\u06ea\u06f0\u0003\u012a\u0095\u0000\u06eb\u06ec"+
		"\u0003\u0242\u0121\u0000\u06ec\u06ed\u0003\u012a\u0095\u0000\u06ed\u06ef"+
		"\u0001\u0000\u0000\u0000\u06ee\u06eb\u0001\u0000\u0000\u0000\u06ef\u06f2"+
		"\u0001\u0000\u0000\u0000\u06f0\u06ee\u0001\u0000\u0000\u0000\u06f0\u06f1"+
		"\u0001\u0000\u0000\u0000\u06f1\u06f4\u0001\u0000\u0000\u0000\u06f2\u06f0"+
		"\u0001\u0000\u0000\u0000\u06f3\u06ea\u0001\u0000\u0000\u0000\u06f3\u06f4"+
		"\u0001\u0000\u0000\u0000\u06f4\u06f5\u0001\u0000\u0000\u0000\u06f5\u06f6"+
		"\u0003\u0238\u011c\u0000\u06f6\u00e9\u0001\u0000\u0000\u0000\u06f7\u06f8"+
		"\u0003\u0236\u011b\u0000\u06f8\u06f9\u0003\u012a\u0095\u0000\u06f9\u06fa"+
		"\u0003\u0244\u0122\u0000\u06fa\u06fb\u0003\u012a\u0095\u0000\u06fb\u0703"+
		"\u0001\u0000\u0000\u0000\u06fc\u06fd\u0003\u0242\u0121\u0000\u06fd\u06fe"+
		"\u0003\u012a\u0095\u0000\u06fe\u06ff\u0003\u0244\u0122\u0000\u06ff\u0700"+
		"\u0003\u012a\u0095\u0000\u0700\u0702\u0001\u0000\u0000\u0000\u0701\u06fc"+
		"\u0001\u0000\u0000\u0000\u0702\u0705\u0001\u0000\u0000\u0000\u0703\u0701"+
		"\u0001\u0000\u0000\u0000\u0703\u0704\u0001\u0000\u0000\u0000\u0704\u0706"+
		"\u0001\u0000\u0000\u0000\u0705\u0703\u0001\u0000\u0000\u0000\u0706\u0707"+
		"\u0003\u0238\u011c\u0000\u0707\u00eb\u0001\u0000\u0000\u0000\u0708\u0709"+
		"\u0003\u023e\u011f\u0000\u0709\u070f\u0003\u012a\u0095\u0000\u070a\u070b"+
		"\u0003\u0242\u0121\u0000\u070b\u070c\u0003\u012a\u0095\u0000\u070c\u070e"+
		"\u0001\u0000\u0000\u0000\u070d\u070a\u0001\u0000\u0000\u0000\u070e\u0711"+
		"\u0001\u0000\u0000\u0000\u070f\u070d\u0001\u0000\u0000\u0000\u070f\u0710"+
		"\u0001\u0000\u0000\u0000\u0710\u0712\u0001\u0000\u0000\u0000\u0711\u070f"+
		"\u0001\u0000\u0000\u0000\u0712\u0713\u0003\u0240\u0120\u0000\u0713\u00ed"+
		"\u0001\u0000\u0000\u0000\u0714\u0715\u0003\u0232\u0119\u0000\u0715\u071b"+
		"\u0003\u0106\u0083\u0000\u0716\u0717\u0003\u0242\u0121\u0000\u0717\u0718"+
		"\u0003\u0106\u0083\u0000\u0718\u071a\u0001\u0000\u0000\u0000\u0719\u0716"+
		"\u0001\u0000\u0000\u0000\u071a\u071d\u0001\u0000\u0000\u0000\u071b\u0719"+
		"\u0001\u0000\u0000\u0000\u071b\u071c\u0001\u0000\u0000\u0000\u071c\u071e"+
		"\u0001\u0000\u0000\u0000\u071d\u071b\u0001\u0000\u0000\u0000\u071e\u071f"+
		"\u0003\u0234\u011a\u0000\u071f\u00ef\u0001\u0000\u0000\u0000\u0720\u0722"+
		"\u0003\u00ba]\u0000\u0721\u0720\u0001\u0000\u0000\u0000\u0721\u0722\u0001"+
		"\u0000\u0000\u0000\u0722\u0723\u0001\u0000\u0000\u0000\u0723\u0724\u0003"+
		"\u01bc\u00de\u0000\u0724\u0728\u0003\u01be\u00df\u0000\u0725\u0726\u0003"+
		"\u0136\u009b\u0000\u0726\u0727\u0005\n\u0000\u0000\u0727\u0729\u0001\u0000"+
		"\u0000\u0000\u0728\u0725\u0001\u0000\u0000\u0000\u0728\u0729\u0001\u0000"+
		"\u0000\u0000\u0729\u072a\u0001\u0000\u0000\u0000\u072a\u072c\u0003\u0138"+
		"\u009c\u0000\u072b\u072d\u0003\u0100\u0080\u0000\u072c\u072b\u0001\u0000"+
		"\u0000\u0000\u072c\u072d\u0001\u0000\u0000\u0000\u072d\u072e\u0001\u0000"+
		"\u0000\u0000\u072e\u0730\u0003\u00fe\u007f\u0000\u072f\u0731\u0003\u00fa"+
		"}\u0000\u0730\u072f\u0001\u0000\u0000\u0000\u0730\u0731\u0001\u0000\u0000"+
		"\u0000\u0731\u0733\u0001\u0000\u0000\u0000\u0732\u0734\u0003\u00f2y\u0000"+
		"\u0733\u0732\u0001\u0000\u0000\u0000\u0733\u0734\u0001\u0000\u0000\u0000"+
		"\u0734\u00f1\u0001\u0000\u0000\u0000\u0735\u0736\u0003\u0226\u0113\u0000"+
		"\u0736\u0737\u0003\u00f6{\u0000\u0737\u0746\u0001\u0000\u0000\u0000\u0738"+
		"\u0739\u0003\u0226\u0113\u0000\u0739\u073a\u0003\u00f6{\u0000\u073a\u073b"+
		"\u0003\u016c\u00b6\u0000\u073b\u073c\u0003\u00f4z\u0000\u073c\u0746\u0001"+
		"\u0000\u0000\u0000\u073d\u073e\u0003\u0226\u0113\u0000\u073e\u073f\u0003"+
		"\u00f4z\u0000\u073f\u0746\u0001\u0000\u0000\u0000\u0740\u0741\u0003\u0226"+
		"\u0113\u0000\u0741\u0742\u0003\u00f4z\u0000\u0742\u0743\u0003\u016c\u00b6"+
		"\u0000\u0743\u0744\u0003\u00f6{\u0000\u0744\u0746\u0001\u0000\u0000\u0000"+
		"\u0745\u0735\u0001\u0000\u0000\u0000\u0745\u0738\u0001\u0000\u0000\u0000"+
		"\u0745\u073d\u0001\u0000\u0000\u0000\u0745\u0740\u0001\u0000\u0000\u0000"+
		"\u0746\u00f3\u0001\u0000\u0000\u0000\u0747\u0748\u0003\u0210\u0108\u0000"+
		"\u0748\u0749\u0003\u012c\u0096\u0000\u0749\u00f5\u0001\u0000\u0000\u0000"+
		"\u074a\u074b\u0003\u0218\u010c\u0000\u074b\u074c\u0003\u012c\u0096\u0000"+
		"\u074c\u00f7\u0001\u0000\u0000\u0000\u074d\u074e\u0003\u0226\u0113\u0000"+
		"\u074e\u074f\u0003\u00f4z\u0000\u074f\u00f9\u0001\u0000\u0000\u0000\u0750"+
		"\u0751\u0003\u01b2\u00d9\u0000\u0751\u0752\u0003\u01e0\u00f0\u0000\u0752"+
		"\u0753\u0003\u01a2\u00d1\u0000\u0753\u00fb\u0001\u0000\u0000\u0000\u0754"+
		"\u0755\u0003\u01b2\u00d9\u0000\u0755\u0756\u0003\u01a2\u00d1\u0000\u0756"+
		"\u00fd\u0001\u0000\u0000\u0000\u0757\u0758\u0003\u0228\u0114\u0000\u0758"+
		"\u0759\u0005\u0001\u0000\u0000\u0759\u075a\u0003\u0104\u0082\u0000\u075a"+
		"\u075b\u0005\u0002\u0000\u0000\u075b\u0760\u0001\u0000\u0000\u0000\u075c"+
		"\u075d\u0003\u01c2\u00e1\u0000\u075d\u075e\u0003\u012a\u0095\u0000\u075e"+
		"\u0760\u0001\u0000\u0000\u0000\u075f\u0757\u0001\u0000\u0000\u0000\u075f"+
		"\u075c\u0001\u0000\u0000\u0000\u0760\u00ff\u0001\u0000\u0000\u0000\u0761"+
		"\u0762\u0005\u0001\u0000\u0000\u0762\u0763\u0003\u0102\u0081\u0000\u0763"+
		"\u0764\u0005\u0002\u0000\u0000\u0764\u0101\u0001\u0000\u0000\u0000\u0765"+
		"\u076b\u0003\u013a\u009d\u0000\u0766\u0767\u0003\u0242\u0121\u0000\u0767"+
		"\u0768\u0003\u013a\u009d\u0000\u0768\u076a\u0001\u0000\u0000\u0000\u0769"+
		"\u0766\u0001\u0000\u0000\u0000\u076a\u076d\u0001\u0000\u0000\u0000\u076b"+
		"\u0769\u0001\u0000\u0000\u0000\u076b\u076c\u0001\u0000\u0000\u0000\u076c"+
		"\u0103\u0001\u0000\u0000\u0000\u076d\u076b\u0001\u0000\u0000\u0000\u076e"+
		"\u0774\u0003\u0106\u0083\u0000\u076f\u0770\u0003\u0242\u0121\u0000\u0770"+
		"\u0771\u0003\u0106\u0083\u0000\u0771\u0773\u0001\u0000\u0000\u0000\u0772"+
		"\u076f\u0001\u0000\u0000\u0000\u0773\u0776\u0001\u0000\u0000\u0000\u0774"+
		"\u0772\u0001\u0000\u0000\u0000\u0774\u0775\u0001\u0000\u0000\u0000\u0775"+
		"\u0105\u0001\u0000\u0000\u0000\u0776\u0774\u0001\u0000\u0000\u0000\u0777"+
		"\u077e\u0003\u012a\u0095\u0000\u0778\u077e\u0003\u0126\u0093\u0000\u0779"+
		"\u077e\u0003\u00eau\u0000\u077a\u077e\u0003\u00e8t\u0000\u077b\u077e\u0003"+
		"\u00ecv\u0000\u077c\u077e\u0003\u00eew\u0000\u077d\u0777\u0001\u0000\u0000"+
		"\u0000\u077d\u0778\u0001\u0000\u0000\u0000\u077d\u0779\u0001\u0000\u0000"+
		"\u0000\u077d\u077a\u0001\u0000\u0000\u0000\u077d\u077b\u0001\u0000\u0000"+
		"\u0000\u077d\u077c\u0001\u0000\u0000\u0000\u077e\u0107\u0001\u0000\u0000"+
		"\u0000\u077f\u0781\u0003\u01fe\u00ff\u0000\u0780\u0782\u0003\u0118\u008c"+
		"\u0000\u0781\u0780\u0001\u0000\u0000\u0000\u0781\u0782\u0001\u0000\u0000"+
		"\u0000\u0782\u0784\u0001\u0000\u0000\u0000\u0783\u0785\u0003\u01c2\u00e1"+
		"\u0000\u0784\u0783\u0001\u0000\u0000\u0000\u0784\u0785\u0001\u0000\u0000"+
		"\u0000\u0785\u0786\u0001\u0000\u0000\u0000\u0786\u0787\u0003\u011a\u008d"+
		"\u0000\u0787\u0789\u0003\u010e\u0087\u0000\u0788\u078a\u0003\u0116\u008b"+
		"\u0000\u0789\u0788\u0001\u0000\u0000\u0000\u0789\u078a\u0001\u0000\u0000"+
		"\u0000\u078a\u078c\u0001\u0000\u0000\u0000\u078b\u078d\u0003\u0112\u0089"+
		"\u0000\u078c\u078b\u0001\u0000\u0000\u0000\u078c\u078d\u0001\u0000\u0000"+
		"\u0000\u078d\u078f\u0001\u0000\u0000\u0000\u078e\u0790\u0003\u010c\u0086"+
		"\u0000\u078f\u078e\u0001\u0000\u0000\u0000\u078f\u0790\u0001\u0000\u0000"+
		"\u0000\u0790\u0792\u0001\u0000\u0000\u0000\u0791\u0793\u0003\u010a\u0085"+
		"\u0000\u0792\u0791\u0001\u0000\u0000\u0000\u0792\u0793\u0001\u0000\u0000"+
		"\u0000\u0793\u0109\u0001\u0000\u0000\u0000\u0794\u0795\u0003\u0168\u00b4"+
		"\u0000\u0795\u0796\u0003\u01a4\u00d2\u0000\u0796\u010b\u0001\u0000\u0000"+
		"\u0000\u0797\u0798\u0003\u01ce\u00e7\u0000\u0798\u0799\u0003\u012c\u0096"+
		"\u0000\u0799\u010d\u0001\u0000\u0000\u0000\u079a\u079b\u0003\u01a8\u00d4"+
		"\u0000\u079b\u079c\u0003\u0110\u0088\u0000\u079c\u010f\u0001\u0000\u0000"+
		"\u0000\u079d\u07a2\u0005\u00b2\u0000\u0000\u079e\u079f\u0005\u00b2\u0000"+
		"\u0000\u079f\u07a0\u0005\n\u0000\u0000\u07a0\u07a2\u0005\u00b2\u0000\u0000"+
		"\u07a1\u079d\u0001\u0000\u0000\u0000\u07a1\u079e\u0001\u0000\u0000\u0000"+
		"\u07a2\u0111\u0001\u0000\u0000\u0000\u07a3\u07a4\u0003\u01ec\u00f6\u0000"+
		"\u07a4\u07a5\u0003\u017a\u00bd\u0000\u07a5\u07a6\u0003\u0114\u008a\u0000"+
		"\u07a6\u0113\u0001\u0000\u0000\u0000\u07a7\u07aa\u0005\u00b2\u0000\u0000"+
		"\u07a8\u07ab\u0003\u0172\u00b9\u0000\u07a9\u07ab\u0003\u0194\u00ca\u0000"+
		"\u07aa\u07a8\u0001\u0000\u0000\u0000\u07aa\u07a9\u0001\u0000\u0000\u0000"+
		"\u07aa\u07ab\u0001\u0000\u0000\u0000\u07ab\u0115\u0001\u0000\u0000\u0000"+
		"\u07ac\u07ad\u0003\u022c\u0116\u0000\u07ad\u07ae\u0003\u011e\u008f\u0000"+
		"\u07ae\u0117\u0001\u0000\u0000\u0000\u07af\u07b0\u0003\u0198\u00cc\u0000"+
		"\u07b0\u0119\u0001\u0000\u0000\u0000\u07b1\u07b4\u0005\u000b\u0000\u0000"+
		"\u07b2\u07b4\u0003\u011c\u008e\u0000\u07b3\u07b1\u0001\u0000\u0000\u0000"+
		"\u07b3\u07b2\u0001\u0000\u0000\u0000\u07b4\u07ba\u0001\u0000\u0000\u0000"+
		"\u07b5\u07b6\u0003\u0242\u0121\u0000\u07b6\u07b7\u0003\u011c\u008e\u0000"+
		"\u07b7\u07b9\u0001\u0000\u0000\u0000\u07b8\u07b5\u0001\u0000\u0000\u0000"+
		"\u07b9\u07bc\u0001\u0000\u0000\u0000\u07ba\u07b8\u0001\u0000\u0000\u0000"+
		"\u07ba\u07bb\u0001\u0000\u0000\u0000\u07bb\u011b\u0001\u0000\u0000\u0000"+
		"\u07bc\u07ba\u0001\u0000\u0000\u0000\u07bd\u07be\u0005\u00b2\u0000\u0000"+
		"\u07be\u07bf\u0005\n\u0000\u0000\u07bf\u07cd\u0005\u000b\u0000\u0000\u07c0"+
		"\u07c4\u0005\u00b2\u0000\u0000\u07c1\u07c2\u0003\u0170\u00b8\u0000\u07c2"+
		"\u07c3\u0005\u00b2\u0000\u0000\u07c3\u07c5\u0001\u0000\u0000\u0000\u07c4"+
		"\u07c1\u0001\u0000\u0000\u0000\u07c4\u07c5\u0001\u0000\u0000\u0000\u07c5"+
		"\u07cd\u0001\u0000\u0000\u0000\u07c6\u07ca\u0003\u0126\u0093\u0000\u07c7"+
		"\u07c8\u0003\u0170\u00b8\u0000\u07c8\u07c9\u0005\u00b2\u0000\u0000\u07c9"+
		"\u07cb\u0001\u0000\u0000\u0000\u07ca\u07c7\u0001\u0000\u0000\u0000\u07ca"+
		"\u07cb\u0001\u0000\u0000\u0000\u07cb\u07cd\u0001\u0000\u0000\u0000\u07cc"+
		"\u07bd\u0001\u0000\u0000\u0000\u07cc\u07c0\u0001\u0000\u0000\u0000\u07cc"+
		"\u07c6\u0001\u0000\u0000\u0000\u07cd\u011d\u0001\u0000\u0000\u0000\u07ce"+
		"\u07d4\u0003\u0120\u0090\u0000\u07cf\u07d0\u0003\u016c\u00b6\u0000\u07d0"+
		"\u07d1\u0003\u0120\u0090\u0000\u07d1\u07d3\u0001\u0000\u0000\u0000\u07d2"+
		"\u07cf\u0001\u0000\u0000\u0000\u07d3\u07d6\u0001\u0000\u0000\u0000\u07d4"+
		"\u07d2\u0001\u0000\u0000\u0000\u07d4\u07d5\u0001\u0000\u0000\u0000\u07d5"+
		"\u011f\u0001\u0000\u0000\u0000\u07d6\u07d4\u0001\u0000\u0000\u0000\u07d7"+
		"\u07d8\u0005\u00b2\u0000\u0000\u07d8\u07d9\u0007\u0002\u0000\u0000\u07d9"+
		"\u081f\u0003\u012a\u0095\u0000\u07da\u07db\u0005\u00b2\u0000\u0000\u07db"+
		"\u07dc\u0005\n\u0000\u0000\u07dc\u07dd\u0005\u00b2\u0000\u0000\u07dd\u07de"+
		"\u0007\u0002\u0000\u0000\u07de\u081f\u0003\u012a\u0095\u0000\u07df\u07e0"+
		"\u0003\u0126\u0093\u0000\u07e0\u07e1\u0007\u0002\u0000\u0000\u07e1\u07e2"+
		"\u0003\u012a\u0095\u0000\u07e2\u081f\u0001\u0000\u0000\u0000\u07e3\u07e4"+
		"\u0003\u0126\u0093\u0000\u07e4\u07e5\u0007\u0002\u0000\u0000\u07e5\u07e6"+
		"\u0003\u0126\u0093\u0000\u07e6\u081f\u0001\u0000\u0000\u0000\u07e7\u07e8"+
		"\u0005\u00b2\u0000\u0000\u07e8\u07e9\u0003\u01b4\u00da\u0000\u07e9\u07eb"+
		"\u0005\u0001\u0000\u0000\u07ea\u07ec\u0003\u0128\u0094\u0000\u07eb\u07ea"+
		"\u0001\u0000\u0000\u0000\u07eb\u07ec\u0001\u0000\u0000\u0000\u07ec\u07ed"+
		"\u0001\u0000\u0000\u0000\u07ed\u07ee\u0005\u0002\u0000\u0000\u07ee\u081f"+
		"\u0001\u0000\u0000\u0000\u07ef\u07f0\u0005\u0001\u0000\u0000\u07f0\u07f6"+
		"\u0005\u00b2\u0000\u0000\u07f1\u07f2\u0003\u0242\u0121\u0000\u07f2\u07f3"+
		"\u0005\u00b2\u0000\u0000\u07f3\u07f5\u0001\u0000\u0000\u0000\u07f4\u07f1"+
		"\u0001\u0000\u0000\u0000\u07f5\u07f8\u0001\u0000\u0000\u0000\u07f6\u07f4"+
		"\u0001\u0000\u0000\u0000\u07f6\u07f7\u0001\u0000\u0000\u0000\u07f7\u07f9"+
		"\u0001\u0000\u0000\u0000\u07f8\u07f6\u0001\u0000\u0000\u0000\u07f9\u07fa"+
		"\u0005\u0002\u0000\u0000\u07fa\u07fb\u0003\u01b4\u00da\u0000\u07fb\u07fc"+
		"\u0005\u0001\u0000\u0000\u07fc\u0802\u0003\u00eew\u0000\u07fd\u07fe\u0003"+
		"\u0242\u0121\u0000\u07fe\u07ff\u0003\u00eew\u0000\u07ff\u0801\u0001\u0000"+
		"\u0000\u0000\u0800\u07fd\u0001\u0000\u0000\u0000\u0801\u0804\u0001\u0000"+
		"\u0000\u0000\u0802\u0800\u0001\u0000\u0000\u0000\u0802\u0803\u0001\u0000"+
		"\u0000\u0000\u0803\u0805\u0001\u0000\u0000\u0000\u0804\u0802\u0001\u0000"+
		"\u0000\u0000\u0805\u0806\u0005\u0002\u0000\u0000\u0806\u081f\u0001\u0000"+
		"\u0000\u0000\u0807\u0808\u0005\u0001\u0000\u0000\u0808\u080e\u0005\u00b2"+
		"\u0000\u0000\u0809\u080a\u0003\u0242\u0121\u0000\u080a\u080b\u0005\u00b2"+
		"\u0000\u0000\u080b\u080d\u0001\u0000\u0000\u0000\u080c\u0809\u0001\u0000"+
		"\u0000\u0000\u080d\u0810\u0001\u0000\u0000\u0000\u080e\u080c\u0001\u0000"+
		"\u0000\u0000\u080e\u080f\u0001\u0000\u0000\u0000\u080f\u0811\u0001\u0000"+
		"\u0000\u0000\u0810\u080e\u0001\u0000\u0000\u0000\u0811\u0812\u0005\u0002"+
		"\u0000\u0000\u0812\u0813\u0007\u0002\u0000\u0000\u0813\u0819\u0003\u00ee"+
		"w\u0000\u0814\u0815\u0003\u0242\u0121\u0000\u0815\u0816\u0003\u00eew\u0000"+
		"\u0816\u0818\u0001\u0000\u0000\u0000\u0817\u0814\u0001\u0000\u0000\u0000"+
		"\u0818\u081b\u0001\u0000\u0000\u0000\u0819\u0817\u0001\u0000\u0000\u0000"+
		"\u0819\u081a\u0001\u0000\u0000\u0000\u081a\u081f\u0001\u0000\u0000\u0000"+
		"\u081b\u0819\u0001\u0000\u0000\u0000\u081c\u081f\u0003\u0124\u0092\u0000"+
		"\u081d\u081f\u0003\u0122\u0091\u0000\u081e\u07d7\u0001\u0000\u0000\u0000"+
		"\u081e\u07da\u0001\u0000\u0000\u0000\u081e\u07df\u0001\u0000\u0000\u0000"+
		"\u081e\u07e3\u0001\u0000\u0000\u0000\u081e\u07e7\u0001\u0000\u0000\u0000"+
		"\u081e\u07ef\u0001\u0000\u0000\u0000\u081e\u0807\u0001\u0000\u0000\u0000"+
		"\u081e\u081c\u0001\u0000\u0000\u0000\u081e\u081d\u0001\u0000\u0000\u0000"+
		"\u081f\u0121\u0001\u0000\u0000\u0000\u0820\u0821\u0005\u00b2\u0000\u0000"+
		"\u0821\u0822\u0003\u018e\u00c7\u0000\u0822\u0823\u0003\u012a\u0095\u0000"+
		"\u0823\u0123\u0001\u0000\u0000\u0000\u0824\u0825\u0005\u00b2\u0000\u0000"+
		"\u0825\u0826\u0003\u018e\u00c7\u0000\u0826\u0827\u0003\u01c4\u00e2\u0000"+
		"\u0827\u0828\u0001\u0000\u0000\u0000\u0828\u0829\u0003\u012a\u0095\u0000"+
		"\u0829\u0125\u0001\u0000\u0000\u0000\u082a\u082b\u0005\u00b2\u0000\u0000"+
		"\u082b\u082c\u0005\u0001\u0000\u0000\u082c\u082d\u0005\u000b\u0000\u0000"+
		"\u082d\u0838\u0005\u0002\u0000\u0000\u082e\u082f\u0005\u00b2\u0000\u0000"+
		"\u082f\u0831\u0005\u0001\u0000\u0000\u0830\u0832\u0003\u0128\u0094\u0000"+
		"\u0831\u0830\u0001\u0000\u0000\u0000\u0831\u0832\u0001\u0000\u0000\u0000"+
		"\u0832\u0833\u0001\u0000\u0000\u0000\u0833\u0838\u0005\u0002\u0000\u0000"+
		"\u0834\u0835\u0005\u0090\u0000\u0000\u0835\u0836\u0005\u0001\u0000\u0000"+
		"\u0836\u0838\u0005\u0002\u0000\u0000\u0837\u082a\u0001\u0000\u0000\u0000"+
		"\u0837\u082e\u0001\u0000\u0000\u0000\u0837\u0834\u0001\u0000\u0000\u0000"+
		"\u0838\u0127\u0001\u0000\u0000\u0000\u0839\u083d\u0003\u012a\u0095\u0000"+
		"\u083a\u083d\u0005\u00b2\u0000\u0000\u083b\u083d\u0003\u0126\u0093\u0000"+
		"\u083c\u0839\u0001\u0000\u0000\u0000\u083c\u083a\u0001\u0000\u0000\u0000"+
		"\u083c\u083b\u0001\u0000\u0000\u0000\u083d\u0846\u0001\u0000\u0000\u0000"+
		"\u083e\u0842\u0003\u0242\u0121\u0000\u083f\u0843\u0003\u012a\u0095\u0000"+
		"\u0840\u0843\u0005\u00b2\u0000\u0000\u0841\u0843\u0003\u0126\u0093\u0000"+
		"\u0842\u083f\u0001\u0000\u0000\u0000\u0842\u0840\u0001\u0000\u0000\u0000"+
		"\u0842\u0841\u0001\u0000\u0000\u0000\u0843\u0845\u0001\u0000\u0000\u0000"+
		"\u0844\u083e\u0001\u0000\u0000\u0000\u0845\u0848\u0001\u0000\u0000\u0000"+
		"\u0846\u0844\u0001\u0000\u0000\u0000\u0846\u0847\u0001\u0000\u0000\u0000"+
		"\u0847\u0129\u0001\u0000\u0000\u0000\u0848\u0846\u0001\u0000\u0000\u0000"+
		"\u0849\u0852\u0005\u00b3\u0000\u0000\u084a\u0852\u0003\u0130\u0098\u0000"+
		"\u084b\u0852\u0003\u012c\u0096\u0000\u084c\u0852\u0003\u012e\u0097\u0000"+
		"\u084d\u0852\u0003\u0134\u009a\u0000\u084e\u0852\u0003\u0132\u0099\u0000"+
		"\u084f\u0852\u00034\u001a\u0000\u0850\u0852\u0003\u01e2\u00f1\u0000\u0851"+
		"\u0849\u0001\u0000\u0000\u0000\u0851\u084a\u0001\u0000\u0000\u0000\u0851"+
		"\u084b\u0001\u0000\u0000\u0000\u0851\u084c\u0001\u0000\u0000\u0000\u0851"+
		"\u084d\u0001\u0000\u0000\u0000\u0851\u084e\u0001\u0000\u0000\u0000\u0851"+
		"\u084f\u0001\u0000\u0000\u0000\u0851\u0850\u0001\u0000\u0000\u0000\u0852"+
		"\u012b\u0001\u0000\u0000\u0000\u0853\u0854\u0005\u00ae\u0000\u0000\u0854"+
		"\u012d\u0001\u0000\u0000\u0000\u0855\u0856\u0007\u0003\u0000\u0000\u0856"+
		"\u012f\u0001\u0000\u0000\u0000\u0857\u0858\u0005\u00ad\u0000\u0000\u0858"+
		"\u0131\u0001\u0000\u0000\u0000\u0859\u085a\u0007\u0004\u0000\u0000\u085a"+
		"\u0133\u0001\u0000\u0000\u0000\u085b\u085c\u0005\u00b0\u0000\u0000\u085c"+
		"\u0135\u0001\u0000\u0000\u0000\u085d\u0862\u0005\u00b2\u0000\u0000\u085e"+
		"\u085f\u0005\u0011\u0000\u0000\u085f\u0860\u0005\u00b2\u0000\u0000\u0860"+
		"\u0862\u0005\u0011\u0000\u0000\u0861\u085d\u0001\u0000\u0000\u0000\u0861"+
		"\u085e\u0001\u0000\u0000\u0000\u0862\u0137\u0001\u0000\u0000\u0000\u0863"+
		"\u0868\u0005\u00b2\u0000\u0000\u0864\u0865\u0005\u0011\u0000\u0000\u0865"+
		"\u0866\u0005\u00b2\u0000\u0000\u0866\u0868\u0005\u0011\u0000\u0000\u0867"+
		"\u0863\u0001\u0000\u0000\u0000\u0867\u0864\u0001\u0000\u0000\u0000\u0868"+
		"\u0139\u0001\u0000\u0000\u0000\u0869\u086e\u0005\u00b2\u0000\u0000\u086a"+
		"\u086b\u0005\u0011\u0000\u0000\u086b\u086c\u0005\u00b2\u0000\u0000\u086c"+
		"\u086e\u0005\u0011\u0000\u0000\u086d\u0869\u0001\u0000\u0000\u0000\u086d"+
		"\u086a\u0001\u0000\u0000\u0000\u086e\u013b\u0001\u0000\u0000\u0000\u086f"+
		"\u0871\u0003\u013e\u009f\u0000\u0870\u0872\u0003\u0140\u00a0\u0000\u0871"+
		"\u0870\u0001\u0000\u0000\u0000\u0871\u0872\u0001\u0000\u0000\u0000\u0872"+
		"\u013d\u0001\u0000\u0000\u0000\u0873\u0874\u0007\u0005\u0000\u0000\u0874"+
		"\u013f\u0001\u0000\u0000\u0000\u0875\u0876\u0003\u023a\u011d\u0000\u0876"+
		"\u087c\u0003\u013e\u009f\u0000\u0877\u0878\u0003\u0242\u0121\u0000\u0878"+
		"\u0879\u0003\u013e\u009f\u0000\u0879\u087b\u0001\u0000\u0000\u0000\u087a"+
		"\u0877\u0001\u0000\u0000\u0000\u087b\u087e\u0001\u0000\u0000\u0000\u087c"+
		"\u087a\u0001\u0000\u0000\u0000\u087c\u087d\u0001\u0000\u0000\u0000\u087d"+
		"\u087f\u0001\u0000\u0000\u0000\u087e\u087c\u0001\u0000\u0000\u0000\u087f"+
		"\u0880\u0003\u023c\u011e\u0000\u0880\u0141\u0001\u0000\u0000\u0000\u0881"+
		"\u0884\u0003\u0172\u00b9\u0000\u0882\u0884\u0003\u0194\u00ca\u0000\u0883"+
		"\u0881\u0001\u0000\u0000\u0000\u0883\u0882\u0001\u0000\u0000\u0000\u0884"+
		"\u0143\u0001\u0000\u0000\u0000\u0885\u0886\u0005\u00b2\u0000\u0000\u0886"+
		"\u0145\u0001\u0000\u0000\u0000\u0887\u0888\u0005\u00b2\u0000\u0000\u0888"+
		"\u0147\u0001\u0000\u0000\u0000\u0889\u088a\u0003\u0130\u0098\u0000\u088a"+
		"\u0149\u0001\u0000\u0000\u0000\u088b\u088c\u0005\u00b2\u0000\u0000\u088c"+
		"\u014b\u0001\u0000\u0000\u0000\u088d\u088e\u0005\u00b2\u0000\u0000\u088e"+
		"\u014d\u0001\u0000\u0000\u0000\u088f\u0890\u0005\u00b2\u0000\u0000\u0890"+
		"\u014f\u0001\u0000\u0000\u0000\u0891\u0892\u0005\u00b2\u0000\u0000\u0892"+
		"\u0151\u0001\u0000\u0000\u0000\u0893\u0894\u0005\u00b2\u0000\u0000\u0894"+
		"\u0153\u0001\u0000\u0000\u0000\u0895\u0896\u0005\u00b2\u0000\u0000\u0896"+
		"\u0155\u0001\u0000\u0000\u0000\u0897\u0898\u0003\u0130\u0098\u0000\u0898"+
		"\u0157\u0001\u0000\u0000\u0000\u0899\u089a\u0005\u00b2\u0000\u0000\u089a"+
		"\u0159\u0001\u0000\u0000\u0000\u089b\u089c\u0003\u015c\u00ae\u0000\u089c"+
		"\u089d\u0003\u013c\u009e\u0000\u089d\u015b\u0001\u0000\u0000\u0000\u089e"+
		"\u089f\u0007\u0006\u0000\u0000\u089f\u015d\u0001\u0000\u0000\u0000\u08a0"+
		"\u08a1\u0005\u0018\u0000\u0000\u08a1\u015f\u0001\u0000\u0000\u0000\u08a2"+
		"\u08a3\u0005\u0019\u0000\u0000\u08a3\u0161\u0001\u0000\u0000\u0000\u08a4"+
		"\u08a5\u0005\u001a\u0000\u0000\u08a5\u0163\u0001\u0000\u0000\u0000\u08a6"+
		"\u08a7\u0005\u001b\u0000\u0000\u08a7\u0165\u0001\u0000\u0000\u0000\u08a8"+
		"\u08a9\u0005\u001b\u0000\u0000\u08a9\u08aa\u0005k\u0000\u0000\u08aa\u0167"+
		"\u0001\u0000\u0000\u0000\u08ab\u08ac\u0005\u001c\u0000\u0000\u08ac\u0169"+
		"\u0001\u0000\u0000\u0000\u08ad\u08ae\u0005\u001d\u0000\u0000\u08ae\u016b"+
		"\u0001\u0000\u0000\u0000\u08af\u08b0\u0005\u001e\u0000\u0000\u08b0\u016d"+
		"\u0001\u0000\u0000\u0000\u08b1\u08b2\u0005 \u0000\u0000\u08b2\u016f\u0001"+
		"\u0000\u0000\u0000\u08b3\u08b4\u0005!\u0000\u0000\u08b4\u0171\u0001\u0000"+
		"\u0000\u0000\u08b5\u08b6\u0005\"\u0000\u0000\u08b6\u0173\u0001\u0000\u0000"+
		"\u0000\u08b7\u08b8\u0005#\u0000\u0000\u08b8\u0175\u0001\u0000\u0000\u0000"+
		"\u08b9\u08ba\u0005$\u0000\u0000\u08ba\u0177\u0001\u0000\u0000\u0000\u08bb"+
		"\u08bc\u0005%\u0000\u0000\u08bc\u0179\u0001\u0000\u0000\u0000\u08bd\u08be"+
		"\u0005&\u0000\u0000\u08be\u017b\u0001\u0000\u0000\u0000\u08bf\u08c0\u0005"+
		"\'\u0000\u0000\u08c0\u017d\u0001\u0000\u0000\u0000\u08c1\u08c2\u0005("+
		"\u0000\u0000\u08c2\u017f\u0001\u0000\u0000\u0000\u08c3\u08c4\u0005)\u0000"+
		"\u0000\u08c4\u0181\u0001\u0000\u0000\u0000\u08c5\u08c6\u0005+\u0000\u0000"+
		"\u08c6\u0183\u0001\u0000\u0000\u0000\u08c7\u08c8\u0005,\u0000\u0000\u08c8"+
		"\u0185\u0001\u0000\u0000\u0000\u08c9\u08ca\u0005-\u0000\u0000\u08ca\u0187"+
		"\u0001\u0000\u0000\u0000\u08cb\u08cc\u0007\u0007\u0000\u0000\u08cc\u0189"+
		"\u0001\u0000\u0000\u0000\u08cd\u08ce\u0005f\u0000\u0000\u08ce\u018b\u0001"+
		"\u0000\u0000\u0000\u08cf\u08d0\u0007\b\u0000\u0000\u08d0\u018d\u0001\u0000"+
		"\u0000\u0000\u08d1\u08d2\u0005.\u0000\u0000\u08d2\u018f\u0001\u0000\u0000"+
		"\u0000\u08d3\u08d4\u0005/\u0000\u0000\u08d4\u0191\u0001\u0000\u0000\u0000"+
		"\u08d5\u08d6\u00051\u0000\u0000\u08d6\u0193\u0001\u0000\u0000\u0000\u08d7"+
		"\u08d8\u00052\u0000\u0000\u08d8\u0195\u0001\u0000\u0000\u0000\u08d9\u08da"+
		"\u0007\t\u0000\u0000\u08da\u0197\u0001\u0000\u0000\u0000\u08db\u08dc\u0005"+
		"4\u0000\u0000\u08dc\u0199\u0001\u0000\u0000\u0000\u08dd\u08de\u00055\u0000"+
		"\u0000\u08de\u019b\u0001\u0000\u0000\u0000\u08df\u08e0\u00056\u0000\u0000"+
		"\u08e0\u019d\u0001\u0000\u0000\u0000\u08e1\u08e2\u00058\u0000\u0000\u08e2"+
		"\u019f\u0001\u0000\u0000\u0000\u08e3\u08e4\u00059\u0000\u0000\u08e4\u01a1"+
		"\u0001\u0000\u0000\u0000\u08e5\u08e6\u0005:\u0000\u0000\u08e6\u01a3\u0001"+
		"\u0000\u0000\u0000\u08e7\u08e8\u0005<\u0000\u0000\u08e8\u01a5\u0001\u0000"+
		"\u0000\u0000\u08e9\u08ea\u0005=\u0000\u0000\u08ea\u01a7\u0001\u0000\u0000"+
		"\u0000\u08eb\u08ec\u0005>\u0000\u0000\u08ec\u01a9\u0001\u0000\u0000\u0000"+
		"\u08ed\u08ee\u0005?\u0000\u0000\u08ee\u01ab\u0001\u0000\u0000\u0000\u08ef"+
		"\u08f0\u0005@\u0000\u0000\u08f0\u01ad\u0001\u0000\u0000\u0000\u08f1\u08f2"+
		"\u0005A\u0000\u0000\u08f2\u01af\u0001\u0000\u0000\u0000\u08f3\u08f4\u0005"+
		"B\u0000\u0000\u08f4\u01b1\u0001\u0000\u0000\u0000\u08f5\u08f6\u0005C\u0000"+
		"\u0000\u08f6\u01b3\u0001\u0000\u0000\u0000\u08f7\u08f8\u0005D\u0000\u0000"+
		"\u08f8\u01b5\u0001\u0000\u0000\u0000\u08f9\u08fa\u0005E\u0000\u0000\u08fa"+
		"\u01b7\u0001\u0000\u0000\u0000\u08fb\u08fc\u0005G\u0000\u0000\u08fc\u01b9"+
		"\u0001\u0000\u0000\u0000\u08fd\u08fe\u0005H\u0000\u0000\u08fe\u01bb\u0001"+
		"\u0000\u0000\u0000\u08ff\u0900\u0005I\u0000\u0000\u0900\u01bd\u0001\u0000"+
		"\u0000\u0000\u0901\u0902\u0005J\u0000\u0000\u0902\u01bf\u0001\u0000\u0000"+
		"\u0000\u0903\u0904\u0005K\u0000\u0000\u0904\u01c1\u0001\u0000\u0000\u0000"+
		"\u0905\u0906\u0005L\u0000\u0000\u0906\u01c3\u0001\u0000\u0000\u0000\u0907"+
		"\u0908\u0005M\u0000\u0000\u0908\u01c5\u0001\u0000\u0000\u0000\u0909\u090a"+
		"\u0005N\u0000\u0000\u090a\u01c7\u0001\u0000\u0000\u0000\u090b\u090c\u0005"+
		"O\u0000\u0000\u090c\u01c9\u0001\u0000\u0000\u0000\u090d\u090e\u0005P\u0000"+
		"\u0000\u090e\u01cb\u0001\u0000\u0000\u0000\u090f\u0910\u0005Q\u0000\u0000"+
		"\u0910\u01cd\u0001\u0000\u0000\u0000\u0911\u0912\u0005S\u0000\u0000\u0912"+
		"\u01cf\u0001\u0000\u0000\u0000\u0913\u0914\u0005\u00a2\u0000\u0000\u0914"+
		"\u01d1\u0001\u0000\u0000\u0000\u0915\u0916\u0005W\u0000\u0000\u0916\u01d3"+
		"\u0001\u0000\u0000\u0000\u0917\u0918\u0005V\u0000\u0000\u0918\u01d5\u0001"+
		"\u0000\u0000\u0000\u0919\u091a\u0005X\u0000\u0000\u091a\u01d7\u0001\u0000"+
		"\u0000\u0000\u091b\u091c\u0005Y\u0000\u0000\u091c\u01d9\u0001\u0000\u0000"+
		"\u0000\u091d\u091e\u0005Z\u0000\u0000\u091e\u01db\u0001\u0000\u0000\u0000"+
		"\u091f\u0920\u0005]\u0000\u0000\u0920\u01dd\u0001\u0000\u0000\u0000\u0921"+
		"\u0922\u0005\\\u0000\u0000\u0922\u01df\u0001\u0000\u0000\u0000\u0923\u0924"+
		"\u0005^\u0000\u0000\u0924\u01e1\u0001\u0000\u0000\u0000\u0925\u0926\u0005"+
		"_\u0000\u0000\u0926\u01e3\u0001\u0000\u0000\u0000\u0927\u0928\u0005`\u0000"+
		"\u0000\u0928\u01e5\u0001\u0000\u0000\u0000\u0929\u092a\u0005a\u0000\u0000"+
		"\u092a\u01e7\u0001\u0000\u0000\u0000\u092b\u092c\u0005c\u0000\u0000\u092c"+
		"\u01e9\u0001\u0000\u0000\u0000\u092d\u092e\u0005d\u0000\u0000\u092e\u01eb"+
		"\u0001\u0000\u0000\u0000\u092f\u0930\u0005e\u0000\u0000\u0930\u01ed\u0001"+
		"\u0000\u0000\u0000\u0931\u0932\u0005h\u0000\u0000\u0932\u01ef\u0001\u0000"+
		"\u0000\u0000\u0933\u0934\u0005l\u0000\u0000\u0934\u01f1\u0001\u0000\u0000"+
		"\u0000\u0935\u0936\u0005n\u0000\u0000\u0936\u01f3\u0001\u0000\u0000\u0000"+
		"\u0937\u0938\u0005o\u0000\u0000\u0938\u01f5\u0001\u0000\u0000\u0000\u0939"+
		"\u093a\u0005p\u0000\u0000\u093a\u01f7\u0001\u0000\u0000\u0000\u093b\u093c"+
		"\u0005q\u0000\u0000\u093c\u01f9\u0001\u0000\u0000\u0000\u093d\u093e\u0005"+
		"s\u0000\u0000\u093e\u01fb\u0001\u0000\u0000\u0000\u093f\u0940\u0005t\u0000"+
		"\u0000\u0940\u01fd\u0001\u0000\u0000\u0000\u0941\u0942\u0005v\u0000\u0000"+
		"\u0942\u01ff\u0001\u0000\u0000\u0000\u0943\u0944\u0005w\u0000\u0000\u0944"+
		"\u0201\u0001\u0000\u0000\u0000\u0945\u0946\u0005x\u0000\u0000\u0946\u0203"+
		"\u0001\u0000\u0000\u0000\u0947\u0948\u0005y\u0000\u0000\u0948\u0205\u0001"+
		"\u0000\u0000\u0000\u0949\u094a\u0005{\u0000\u0000\u094a\u0207\u0001\u0000"+
		"\u0000\u0000\u094b\u094c\u0005|\u0000\u0000\u094c\u0209\u0001\u0000\u0000"+
		"\u0000\u094d\u094e\u0005}\u0000\u0000\u094e\u020b\u0001\u0000\u0000\u0000"+
		"\u094f\u0950\u0005~\u0000\u0000\u0950\u020d\u0001\u0000\u0000\u0000\u0951"+
		"\u0952\u0005\u007f\u0000\u0000\u0952\u020f\u0001\u0000\u0000\u0000\u0953"+
		"\u0954\u0005\u0081\u0000\u0000\u0954\u0211\u0001\u0000\u0000\u0000\u0955"+
		"\u0956\u0005\u0082\u0000\u0000\u0956\u0213\u0001\u0000\u0000\u0000\u0957"+
		"\u0958\u0005\u0084\u0000\u0000\u0958\u0215\u0001\u0000\u0000\u0000\u0959"+
		"\u095a\u0005\u0086\u0000\u0000\u095a\u0217\u0001\u0000\u0000\u0000\u095b"+
		"\u095c\u0005\u0087\u0000\u0000\u095c\u0219\u0001\u0000\u0000\u0000\u095d"+
		"\u095e\u0005\u0089\u0000\u0000\u095e\u021b\u0001\u0000\u0000\u0000\u095f"+
		"\u0960\u0005\u008a\u0000\u0000\u0960\u021d\u0001\u0000\u0000\u0000\u0961"+
		"\u0962\u0005\u008b\u0000\u0000\u0962\u021f\u0001\u0000\u0000\u0000\u0963"+
		"\u0964\u0005\u008c\u0000\u0000\u0964\u0221\u0001\u0000\u0000\u0000\u0965"+
		"\u0966\u0005\u008d\u0000\u0000\u0966\u0223\u0001\u0000\u0000\u0000\u0967"+
		"\u0968\u0005\u008e\u0000\u0000\u0968\u0225\u0001\u0000\u0000\u0000\u0969"+
		"\u096a\u0005\u008f\u0000\u0000\u096a\u0227\u0001\u0000\u0000\u0000\u096b"+
		"\u096c\u0005\u0091\u0000\u0000\u096c\u0229\u0001\u0000\u0000\u0000\u096d"+
		"\u096e\u0005\u0092\u0000\u0000\u096e\u022b\u0001\u0000\u0000\u0000\u096f"+
		"\u0970\u0005\u0093\u0000\u0000\u0970\u022d\u0001\u0000\u0000\u0000\u0971"+
		"\u0972\u0005\u0094\u0000\u0000\u0972\u022f\u0001\u0000\u0000\u0000\u0973"+
		"\u0974\u0005r\u0000\u0000\u0974\u0231\u0001\u0000\u0000\u0000\u0975\u0976"+
		"\u0005\u0001\u0000\u0000\u0976\u0233\u0001\u0000\u0000\u0000\u0977\u0978"+
		"\u0005\u0002\u0000\u0000\u0978\u0235\u0001\u0000\u0000\u0000\u0979\u097a"+
		"\u0005\u0003\u0000\u0000\u097a\u0237\u0001\u0000\u0000\u0000\u097b\u097c"+
		"\u0005\u0004\u0000\u0000\u097c\u0239\u0001\u0000\u0000\u0000\u097d\u097e"+
		"\u0005\u0014\u0000\u0000\u097e\u023b\u0001\u0000\u0000\u0000\u097f\u0980"+
		"\u0005\u0015\u0000\u0000\u0980\u023d\u0001\u0000\u0000\u0000\u0981\u0982"+
		"\u0005\u0005\u0000\u0000\u0982\u023f\u0001\u0000\u0000\u0000\u0983\u0984"+
		"\u0005\u0006\u0000\u0000\u0984\u0241\u0001\u0000\u0000\u0000\u0985\u0986"+
		"\u0005\u0007\u0000\u0000\u0986\u0243\u0001\u0000\u0000\u0000\u0987\u0988"+
		"\u0005\t\u0000\u0000\u0988\u0245\u0001\u0000\u0000\u0000\u00ba\u0247\u024a"+
		"\u0250\u0255\u0257\u025c\u025f\u0262\u0290\u029b\u02a4\u02ad\u02b7\u02c1"+
		"\u02ca\u02ce\u02d5\u02e3\u02e6\u02ed\u02f2\u02fd\u0307\u0316\u0321\u0326"+
		"\u032f\u0334\u033c\u0341\u0345\u034a\u034f\u035e\u0364\u0369\u0373\u0378"+
		"\u0382\u038e\u0395\u039d\u03ab\u03b0\u03bc\u03c0\u03c4\u03c9\u03ce\u03e1"+
		"\u03e8\u03f0\u03f4\u03f9\u040c\u0415\u0424\u0426\u0432\u0440\u0447\u044e"+
		"\u0456\u0461\u0471\u047e\u0488\u049f\u04ad\u04b4\u04bd\u04d0\u04d8\u04de"+
		"\u04e3\u04ea\u04ef\u04f7\u04fc\u0503\u0508\u050f\u0514\u051b\u0522\u0529"+
		"\u0530\u0535\u053c\u0543\u0548\u054f\u0554\u055b\u0565\u056b\u0573\u0576"+
		"\u057e\u0583\u0587\u0594\u059a\u05a3\u05b0\u05b8\u05be\u05c3\u05d1\u05e5"+
		"\u05ee\u05fa\u05fe\u0602\u0610\u0618\u0621\u062c\u0631\u0638\u063b\u0641"+
		"\u064a\u0650\u0662\u0666\u066a\u066f\u0677\u067f\u0683\u0686\u068c\u0690"+
		"\u0697\u06a2\u06af\u06b8\u06e7\u06f0\u06f3\u0703\u070f\u071b\u0721\u0728"+
		"\u072c\u0730\u0733\u0745\u075f\u076b\u0774\u077d\u0781\u0784\u0789\u078c"+
		"\u078f\u0792\u07a1\u07aa\u07b3\u07ba\u07c4\u07ca\u07cc\u07d4\u07eb\u07f6"+
		"\u0802\u080e\u0819\u081e\u0831\u0837\u083c\u0842\u0846\u0851\u0861\u0867"+
		"\u086d\u0871\u087c\u0883";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}