package cluster

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// ============================================================================
// Translation Functions: gocql → our wrapper types
//
// CRITICAL: These functions translate on-demand. They DO NOT cache.
// Each call creates NEW wrapper instances from gocql metadata.
// ============================================================================

// translateKeyspace converts gocql.KeyspaceMetadata to our KeyspaceMetadata
func (m *GocqlMetadataManager) translateKeyspace(gocqlKs *gocql.KeyspaceMetadata) *KeyspaceMetadata {
	if gocqlKs == nil {
		return nil
	}

	ksMeta := &KeyspaceMetadata{
		Name:          gocqlKs.Name,
		DurableWrites: gocqlKs.DurableWrites,
		Replication:   m.translateReplication(gocqlKs),
		Tables:        make(map[string]*TableMetadata),
		UserTypes:     make(map[string]*UserType),
		Functions:     make(map[string]*FunctionMetadata),
		Aggregates:    make(map[string]*AggregateMetadata),
		MaterializedViews: make(map[string]*MaterializedViewMetadata),
	}

	// Translate tables
	for name, gocqlTable := range gocqlKs.Tables {
		ksMeta.Tables[name] = m.translateTable(gocqlTable)
	}

	// Translate user types
	for name, gocqlType := range gocqlKs.UserTypes {
		ksMeta.UserTypes[name] = m.translateUserType(gocqlType)
	}

	// Translate functions
	for name, gocqlFunc := range gocqlKs.Functions {
		ksMeta.Functions[name] = m.translateFunction(gocqlFunc)
	}

	// Translate aggregates
	for name, gocqlAgg := range gocqlKs.Aggregates {
		ksMeta.Aggregates[name] = m.translateAggregate(gocqlAgg)
	}

	// Translate materialized views
	for name, gocqlView := range gocqlKs.MaterializedViews {
		ksMeta.MaterializedViews[name] = m.translateMaterializedView(gocqlView)
	}

	return ksMeta
}

// translateReplication extracts replication strategy from gocql metadata
func (m *GocqlMetadataManager) translateReplication(gocqlKs *gocql.KeyspaceMetadata) *ReplicationStrategy {
	if gocqlKs == nil {
		return nil
	}

	strategy := &ReplicationStrategy{
		DataCenters: make(map[string]int),
	}

	// Determine strategy class
	switch gocqlKs.StrategyClass {
	case "org.apache.cassandra.locator.SimpleStrategy":
		strategy.Class = ReplicationStrategySimple
	case "org.apache.cassandra.locator.NetworkTopologyStrategy":
		strategy.Class = ReplicationStrategyNetworkTopology
	case "org.apache.cassandra.locator.LocalStrategy":
		strategy.Class = ReplicationStrategyLocal
	case "org.apache.cassandra.locator.EverywhereStrategy":
		strategy.Class = ReplicationStrategyEverywhere
	default:
		strategy.Class = ReplicationStrategyClass(gocqlKs.StrategyClass)
	}

	// Extract replication factor
	if rfVal, ok := gocqlKs.StrategyOptions["replication_factor"]; ok {
		if rfStr, ok := rfVal.(string); ok {
			if rf, err := strconv.Atoi(rfStr); err == nil {
				strategy.ReplicationFactor = rf
			}
		}
	}

	// Extract per-DC replication for NetworkTopologyStrategy
	for dc, rfVal := range gocqlKs.StrategyOptions {
		if dc != "class" && dc != "replication_factor" {
			if rfStr, ok := rfVal.(string); ok {
				if rf, err := strconv.Atoi(rfStr); err == nil {
					strategy.DataCenters[dc] = rf
				}
			}
		}
	}

	return strategy
}

// translateTable converts gocql.TableMetadata to our TableMetadata
func (m *GocqlMetadataManager) translateTable(gocqlTable *gocql.TableMetadata) *TableMetadata {
	if gocqlTable == nil {
		return nil
	}

	tableMeta := &TableMetadata{
		Keyspace:       gocqlTable.Keyspace,
		Name:           gocqlTable.Name,
		Columns:        make(map[string]*ColumnInfo),
		PartitionKeys:  make([]*ColumnInfo, 0, len(gocqlTable.PartitionKey)),
		ClusteringKeys: make([]*ColumnInfo, 0, len(gocqlTable.ClusteringColumns)),
		StaticColumns:  make([]*ColumnInfo, 0),
		Options:        m.translateTableOptions(nil), // gocql doesn't expose table options directly
	}

	// Translate all columns
	for name, gocqlCol := range gocqlTable.Columns {
		colInfo := m.translateColumn(gocqlCol)
		tableMeta.Columns[name] = colInfo

		// Collect static columns
		if colInfo.IsStatic() {
			tableMeta.StaticColumns = append(tableMeta.StaticColumns, colInfo)
		}
	}

	// Translate partition keys (preserve order)
	for _, gocqlPK := range gocqlTable.PartitionKey {
		if colInfo, ok := tableMeta.Columns[gocqlPK.Name]; ok {
			tableMeta.PartitionKeys = append(tableMeta.PartitionKeys, colInfo)
		}
	}

	// Translate clustering keys (preserve order)
	for _, gocqlCK := range gocqlTable.ClusteringColumns {
		if colInfo, ok := tableMeta.Columns[gocqlCK.Name]; ok {
			tableMeta.ClusteringKeys = append(tableMeta.ClusteringKeys, colInfo)
		}
	}

	return tableMeta
}

// translateColumn converts gocql.ColumnMetadata to our ColumnInfo
func (m *GocqlMetadataManager) translateColumn(gocqlCol *gocql.ColumnMetadata) *ColumnInfo {
	if gocqlCol == nil {
		return nil
	}

	colInfo := &ColumnInfo{
		Name:            gocqlCol.Name,
		Type:            m.translateType(gocqlCol.Type),
		Kind:            m.translateColumnKind(gocqlCol.Kind),
		ComponentIndex:  gocqlCol.ComponentIndex,
		ClusteringOrder: m.translateClusteringOrder(gocqlCol.ClusteringOrder),
		Index:           m.translateIndex(&gocqlCol.Index),
	}

	return colInfo
}

// translateColumnKind converts gocql column kind to our ColumnKind
func (m *GocqlMetadataManager) translateColumnKind(kind gocql.ColumnKind) ColumnKind {
	switch kind {
	case gocql.ColumnPartitionKey:
		return ColumnKindPartitionKey
	case gocql.ColumnClusteringKey:
		return ColumnKindClusteringKey
	case gocql.ColumnRegular:
		return ColumnKindRegular
	case gocql.ColumnStatic:
		return ColumnKindStatic
	case gocql.ColumnCompact:
		return ColumnKindCompact
	default:
		return ColumnKindRegular
	}
}

// translateClusteringOrder converts gocql clustering order to our ColumnOrder
func (m *GocqlMetadataManager) translateClusteringOrder(order string) ColumnOrder {
	if strings.ToUpper(order) == "DESC" {
		return ColumnOrderDESC
	}
	return ColumnOrderASC
}

// translateIndex converts gocql index metadata to our ColumnIndexInfo
func (m *GocqlMetadataManager) translateIndex(gocqlIdx *gocql.ColumnIndexMetadata) *ColumnIndexInfo {
	if gocqlIdx == nil || gocqlIdx.Name == "" {
		return nil
	}

	idxInfo := &ColumnIndexInfo{
		Name:           gocqlIdx.Name,
		Target:         "", // gocql doesn't expose Target field
		Options:        m.convertIndexOptions(gocqlIdx.Options),
		IndexClassName: gocqlIdx.Type, // gocql stores class name in Type field
	}

	// Detect index type
	kind, isNative, isSAI, isVector := m.detectIndexType(gocqlIdx.Type)
	idxInfo.Kind = kind
	idxInfo.IsNativeIndex = isNative
	idxInfo.IsSAIIndex = isSAI
	idxInfo.IsVectorIndex = isVector

	// Set category
	if idxInfo.IsNativeIndex || idxInfo.IsSAIIndex {
		idxInfo.IndexCategory = IndexCategoryNative
	} else {
		idxInfo.IndexCategory = IndexCategoryCustom
	}

	return idxInfo
}

// convertIndexOptions converts map[string]interface{} to map[string]string
func (m *GocqlMetadataManager) convertIndexOptions(opts map[string]interface{}) map[string]string {
	if opts == nil {
		return nil
	}
	result := make(map[string]string, len(opts))
	for k, v := range opts {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// detectIndexType determines index kind and properties from class name
func (m *GocqlMetadataManager) detectIndexType(className string) (IndexKind, bool, bool, bool) {
	classLower := strings.ToLower(className)

	// Check for SAI (Cassandra 5.0+)
	if strings.Contains(classLower, "storageattachedindex") || strings.Contains(classLower, "sai") {
		// Could be regular SAI or vector SAI - need to check column type
		return IndexKindSAI, true, true, false
	}

	// Check for native index types
	switch {
	case strings.Contains(classLower, "composites"):
		return IndexKindComposites, true, false, false
	case strings.Contains(classLower, "keys"):
		return IndexKindKeys, true, false, false
	case strings.Contains(classLower, "full"):
		return IndexKindFull, true, false, false
	case strings.Contains(classLower, "values"):
		return IndexKindValues, true, false, false
	case strings.Contains(classLower, "entries"):
		return IndexKindEntries, true, false, false
	}

	// Custom index
	return IndexKindCustom, false, false, false
}

// formatTypeInfo converts TypeInfo to string representation
func (m *GocqlMetadataManager) formatTypeInfo(typeInfo gocql.TypeInfo) string {
	if typeInfo == nil {
		return "unknown"
	}

	// Similar to formatTypeInfo in db/metadata.go
	baseType := typeInfo.Type()

	// Handle collection types
	if collType, ok := typeInfo.(gocql.CollectionType); ok {
		switch collType.Type() {
		case gocql.TypeList:
			return fmt.Sprintf("list<%s>", m.formatTypeInfo(collType.Elem))
		case gocql.TypeSet:
			return fmt.Sprintf("set<%s>", m.formatTypeInfo(collType.Elem))
		case gocql.TypeMap:
			return fmt.Sprintf("map<%s, %s>", m.formatTypeInfo(collType.Key), m.formatTypeInfo(collType.Elem))
		}
	}

	// Handle UDT types
	if udtType, ok := typeInfo.(gocql.UDTTypeInfo); ok {
		if udtType.Keyspace != "" {
			return fmt.Sprintf("%s.%s", udtType.Keyspace, udtType.Name)
		}
		return udtType.Name
	}

	// Handle tuple types
	if tupleType, ok := typeInfo.(gocql.TupleTypeInfo); ok {
		var elements []string
		for _, elem := range tupleType.Elems {
			elements = append(elements, m.formatTypeInfo(elem))
		}
		return fmt.Sprintf("tuple<%s>", strings.Join(elements, ", "))
	}

	// Native type
	return m.typeNameFromType(baseType)
}

// typeNameFromType converts gocql.Type to string name
func (m *GocqlMetadataManager) typeNameFromType(t gocql.Type) string {
	switch t {
	case gocql.TypeAscii:
		return "ascii"
	case gocql.TypeBigInt:
		return "bigint"
	case gocql.TypeBlob:
		return "blob"
	case gocql.TypeBoolean:
		return "boolean"
	case gocql.TypeCounter:
		return "counter"
	case gocql.TypeDecimal:
		return "decimal"
	case gocql.TypeDouble:
		return "double"
	case gocql.TypeFloat:
		return "float"
	case gocql.TypeInt:
		return "int"
	case gocql.TypeText, gocql.TypeVarchar:
		return "text"
	case gocql.TypeTimestamp:
		return "timestamp"
	case gocql.TypeUUID:
		return "uuid"
	case gocql.TypeVarint:
		return "varint"
	case gocql.TypeTimeUUID:
		return "timeuuid"
	case gocql.TypeInet:
		return "inet"
	case gocql.TypeDate:
		return "date"
	case gocql.TypeDuration:
		return "duration"
	case gocql.TypeTime:
		return "time"
	case gocql.TypeSmallInt:
		return "smallint"
	case gocql.TypeTinyInt:
		return "tinyint"
	default:
		return "unknown"
	}
}

// translateType converts gocql.TypeInfo to our ColumnType
func (m *GocqlMetadataManager) translateType(typeInfo gocql.TypeInfo) *ColumnType {
	if typeInfo == nil {
		return nil
	}

	typeStr := m.formatTypeInfo(typeInfo)

	colType := &ColumnType{
		Name:     typeStr,
		IsNative: true,
	}

	// Check for vector type (Cassandra 5.0+)
	if m.config.DetectVectorColumns {
		if isVec, dim, elemType := m.parseVectorType(typeStr); isVec {
			colType.Category = TypeCategoryVector
			colType.IsVector = true
			colType.VectorDimension = dim
			colType.VectorElementType = elemType
			colType.VectorSimilarityFunction = "euclidean" // Default
			return colType
		}
	}

	// Handle collection types
	if collType, ok := typeInfo.(gocql.CollectionType); ok {
		colType.Category = TypeCategoryCollection
		switch collType.Type() {
		case gocql.TypeList:
			colType.ElementType = m.translateType(collType.Elem)
		case gocql.TypeSet:
			colType.ElementType = m.translateType(collType.Elem)
		case gocql.TypeMap:
			colType.KeyType = m.translateType(collType.Key)
			colType.ValueType = m.translateType(collType.Elem)
		}
		return colType
	}

	// Handle UDT types
	if udtType, ok := typeInfo.(gocql.UDTTypeInfo); ok {
		colType.Category = TypeCategoryUDT
		colType.UDTKeyspace = udtType.Keyspace
		colType.UDTName = udtType.Name
		colType.UDTFields = make(map[string]*ColumnType)
		for _, field := range udtType.Elements {
			colType.UDTFields[field.Name] = m.translateType(field.Type)
		}
		return colType
	}

	// Handle tuple types
	if tupleType, ok := typeInfo.(gocql.TupleTypeInfo); ok {
		colType.Category = TypeCategoryTuple
		colType.TupleTypes = make([]*ColumnType, len(tupleType.Elems))
		for i, elem := range tupleType.Elems {
			colType.TupleTypes[i] = m.translateType(elem)
		}
		return colType
	}

	// Simple native type
	colType.Category = TypeCategorySimple
	return colType
}

// parseVectorType parses vector type string (e.g., "vector<float, 384>")
// Returns: (isVector bool, dimension int, elementType string)
func (m *GocqlMetadataManager) parseVectorType(typeStr string) (bool, int, string) {
	// Pattern: vector<element_type, dimension>
	re := regexp.MustCompile(`(?i)vector<\s*(\w+)\s*,\s*(\d+)\s*>`)
	matches := re.FindStringSubmatch(typeStr)

	if len(matches) != 3 {
		return false, 0, ""
	}

	elementType := matches[1]
	dimension, err := strconv.Atoi(matches[2])
	if err != nil {
		return false, 0, ""
	}

	return true, dimension, elementType
}

// translateTableOptions converts gocql table options to our TableOptions
func (m *GocqlMetadataManager) translateTableOptions(options map[string]interface{}) *TableOptions {
	if options == nil {
		return &TableOptions{
			Caching:     make(map[string]string),
			Compression: make(map[string]string),
			Compaction:  make(map[string]string),
		}
	}

	opts := &TableOptions{
		Caching:     make(map[string]string),
		Compression: make(map[string]string),
		Compaction:  make(map[string]string),
	}

	// Extract known options
	if comment, ok := options["comment"].(string); ok {
		opts.Comment = comment
	}

	if rrc, ok := options["read_repair_chance"].(float64); ok {
		opts.ReadRepairChance = rrc
	}

	if gc, ok := options["gc_grace_seconds"].(int); ok {
		opts.GcGraceSeconds = gc
	}

	// Extract maps
	if caching, ok := options["caching"].(map[string]string); ok {
		opts.Caching = caching
	}

	if compression, ok := options["compression"].(map[string]string); ok {
		opts.Compression = compression
	}

	if compaction, ok := options["compaction"].(map[string]string); ok {
		opts.Compaction = compaction
	}

	return opts
}

// translateUserType converts gocql.UserTypeMetadata to our UserType
func (m *GocqlMetadataManager) translateUserType(gocqlType *gocql.UserTypeMetadata) *UserType {
	if gocqlType == nil {
		return nil
	}

	userType := &UserType{
		Keyspace:   gocqlType.Keyspace,
		Name:       gocqlType.Name,
		Fields:     make(map[string]*ColumnType),
		FieldNames: gocqlType.FieldNames,
	}

	// Translate field types
	for i, fieldName := range gocqlType.FieldNames {
		if i < len(gocqlType.FieldTypes) {
			userType.Fields[fieldName] = m.translateType(gocqlType.FieldTypes[i])
		}
	}

	return userType
}

// translateFunction converts gocql.FunctionMetadata to our FunctionMetadata
func (m *GocqlMetadataManager) translateFunction(gocqlFunc *gocql.FunctionMetadata) *FunctionMetadata {
	if gocqlFunc == nil {
		return nil
	}

	funcMeta := &FunctionMetadata{
		Keyspace:          gocqlFunc.Keyspace,
		Name:              gocqlFunc.Name,
		ArgumentNames:     gocqlFunc.ArgumentNames,
		ReturnType:        m.translateType(gocqlFunc.ReturnType),
		Language:          gocqlFunc.Language,
		Body:              gocqlFunc.Body,
		CalledOnNullInput: gocqlFunc.CalledOnNullInput,
	}

	funcMeta.ArgumentTypes = make([]*ColumnType, len(gocqlFunc.ArgumentTypes))
	for i, argType := range gocqlFunc.ArgumentTypes {
		funcMeta.ArgumentTypes[i] = m.translateType(argType)
	}

	return funcMeta
}

// translateAggregate converts gocql.AggregateMetadata to our AggregateMetadata
func (m *GocqlMetadataManager) translateAggregate(gocqlAgg *gocql.AggregateMetadata) *AggregateMetadata {
	if gocqlAgg == nil {
		return nil
	}

	aggMeta := &AggregateMetadata{
		Keyspace:         gocqlAgg.Keyspace,
		Name:             gocqlAgg.Name,
		StateFunction:    gocqlAgg.StateFunc.Name,
		StateType:        m.translateType(gocqlAgg.StateType),
		FinalFunction:    gocqlAgg.FinalFunc.Name,
		InitialCondition: gocqlAgg.InitCond,
	}

	aggMeta.ArgumentTypes = make([]*ColumnType, len(gocqlAgg.ArgumentTypes))
	for i, argType := range gocqlAgg.ArgumentTypes {
		aggMeta.ArgumentTypes[i] = m.translateType(argType)
	}

	return aggMeta
}

// translateMaterializedView converts gocql.MaterializedViewMetadata to our MaterializedViewMetadata
func (m *GocqlMetadataManager) translateMaterializedView(gocqlView *gocql.MaterializedViewMetadata) *MaterializedViewMetadata {
	if gocqlView == nil {
		return nil
	}

	// Get base table name from BaseTable metadata
	baseTableName := ""
	if gocqlView.BaseTable != nil {
		baseTableName = gocqlView.BaseTable.Name
	}

	viewMeta := &MaterializedViewMetadata{
		Keyspace:          gocqlView.Keyspace,
		Name:              gocqlView.Name,
		BaseTable:         baseTableName,
		Columns:           make(map[string]*ColumnInfo),
		PartitionKeys:     make([]*ColumnInfo, 0),
		ClusteringKeys:    make([]*ColumnInfo, 0),
		WhereClause:       "", // gocql doesn't expose WHERE clause
		IncludeAllColumns: gocqlView.IncludeAllColumns,
	}

	// Get columns from base table if available
	// Note: gocql doesn't expose view columns directly
	if gocqlView.BaseTable != nil {
		for name, gocqlCol := range gocqlView.BaseTable.Columns {
			colInfo := m.translateColumn(gocqlCol)
			viewMeta.Columns[name] = colInfo

			// Collect partition/clustering keys
			if colInfo.IsPartitionKey() {
				viewMeta.PartitionKeys = append(viewMeta.PartitionKeys, colInfo)
			} else if colInfo.IsClusteringKey() {
				viewMeta.ClusteringKeys = append(viewMeta.ClusteringKeys, colInfo)
			}
		}
	}

	return viewMeta
}
