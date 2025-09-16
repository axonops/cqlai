package db

import (
	"fmt"
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// UDTField represents a field within a User-Defined Type
type UDTField struct {
	Name     string
	TypeStr  string
	TypeInfo *CQLTypeInfo // Parsed type information
}

// UDTDefinition represents a complete User-Defined Type definition
type UDTDefinition struct {
	Keyspace string
	Name     string
	Fields   []UDTField
}

// UDTRegistry manages UDT definitions using gocql's cached metadata
// This simplified version relies on gocql's internal caching instead of maintaining our own cache
type UDTRegistry struct {
	session *gocql.Session
}

// NewUDTRegistry creates a new UDT registry with the given session
func NewUDTRegistry(session *gocql.Session) *UDTRegistry {
	return &UDTRegistry{
		session: session,
	}
}

// GetUDTDefinition retrieves a UDT definition from gocql's cached metadata
func (r *UDTRegistry) GetUDTDefinition(keyspace, udtName string) (*UDTDefinition, error) {
	if r.session == nil {
		return nil, fmt.Errorf("no session available")
	}

	// Get keyspace metadata from gocql (this is cached internally by gocql)
	ksMetadata, err := r.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyspace metadata: %w", err)
	}

	// Look for the UDT in the metadata
	udtMeta, exists := ksMetadata.UserTypes[udtName]
	if !exists {
		return nil, fmt.Errorf("UDT %s.%s not found", keyspace, udtName)
	}

	// Convert to our UDTDefinition format
	udtDef := &UDTDefinition{
		Keyspace: keyspace,
		Name:     udtName,
		Fields:   make([]UDTField, len(udtMeta.FieldNames)),
	}

	// Convert each field
	for i := range udtMeta.FieldNames {
		// Convert TypeInfo to string representation
		typeStr := formatGocqlTypeInfo(udtMeta.FieldTypes[i])

		// Parse the type string to get our CQLTypeInfo
		typeInfo, err := ParseCQLType(typeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type for field %s in UDT %s.%s: %w",
				udtMeta.FieldNames[i], keyspace, udtName, err)
		}

		udtDef.Fields[i] = UDTField{
			Name:     udtMeta.FieldNames[i],
			TypeStr:  typeStr,
			TypeInfo: typeInfo,
		}
	}

	return udtDef, nil
}

// GetUDTDefinitionOrLoad is now just an alias for GetUDTDefinition
// since gocql handles the loading and caching internally
func (r *UDTRegistry) GetUDTDefinitionOrLoad(keyspace, udtName string) (*UDTDefinition, error) {
	return r.GetUDTDefinition(keyspace, udtName)
}

// LoadKeyspaceUDTs is now a no-op since gocql handles loading automatically
// Kept for backward compatibility
func (r *UDTRegistry) LoadKeyspaceUDTs(keyspace string) error {
	if r.session == nil {
		return fmt.Errorf("no session available")
	}
	// gocql will load and cache metadata when KeyspaceMetadata is called
	_, err := r.session.KeyspaceMetadata(keyspace)
	return err
}

// LoadKeyspaceUDTsUsingMetadata is now a no-op since we always use metadata
// Kept for backward compatibility
func (r *UDTRegistry) LoadKeyspaceUDTsUsingMetadata(keyspace string) error {
	return r.LoadKeyspaceUDTs(keyspace)
}

// Clear is a no-op since we don't maintain our own cache
// gocql's cache is managed internally
func (r *UDTRegistry) Clear() {
	// No-op - gocql manages its own cache
}

// ClearKeyspace is a no-op since we don't maintain our own cache
func (r *UDTRegistry) ClearKeyspace(keyspace string) {
	// No-op - gocql manages its own cache
}

// GetAllUDTs returns all UDT definitions for a keyspace from gocql's cached metadata
func (r *UDTRegistry) GetAllUDTs(keyspace string) map[string]*UDTDefinition {
	if r.session == nil {
		return nil
	}
	ksMetadata, err := r.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil
	}

	result := make(map[string]*UDTDefinition)
	for udtName := range ksMetadata.UserTypes {
		if udtDef, err := r.GetUDTDefinition(keyspace, udtName); err == nil {
			result[udtName] = udtDef
		}
	}

	return result
}

// HasUDT checks if a UDT exists in the keyspace
func (r *UDTRegistry) HasUDT(keyspace, udtName string) bool {
	if r.session == nil {
		return false
	}
	ksMetadata, err := r.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return false
	}

	_, exists := ksMetadata.UserTypes[udtName]
	return exists
}

// String returns a string representation of a UDT definition
func (d *UDTDefinition) String() string {
	result := fmt.Sprintf("%s.%s {\n", d.Keyspace, d.Name)
	for _, field := range d.Fields {
		result += fmt.Sprintf("  %s: %s\n", field.Name, field.TypeStr)
	}
	result += "}"
	return result
}

// GetFieldByName returns a field by name from the UDT definition
func (d *UDTDefinition) GetFieldByName(name string) (*UDTField, int, error) {
	for i, field := range d.Fields {
		if field.Name == name {
			return &field, i, nil
		}
	}
	return nil, -1, fmt.Errorf("field %s not found in UDT %s.%s", name, d.Keyspace, d.Name)
}

// formatGocqlTypeInfo converts gocql.TypeInfo to a string representation for UDT fields
func formatGocqlTypeInfo(typeInfo gocql.TypeInfo) string {
	if typeInfo == nil {
		return "unknown"
	}

	// Check the base type first
	baseType := typeInfo.Type()

	// Handle simple types
	switch baseType {
	case gocql.TypeList, gocql.TypeSet, gocql.TypeMap:
		// Handle collection types
		if collType, ok := typeInfo.(gocql.CollectionType); ok {
			switch collType.Type() {
			case gocql.TypeList:
				return fmt.Sprintf("list<%s>", formatGocqlTypeInfo(collType.Elem))
			case gocql.TypeSet:
				return fmt.Sprintf("set<%s>", formatGocqlTypeInfo(collType.Elem))
			case gocql.TypeMap:
				return fmt.Sprintf("map<%s, %s>", formatGocqlTypeInfo(collType.Key), formatGocqlTypeInfo(collType.Elem))
			}
		}
	case gocql.TypeTuple:
		// Handle tuple types
		if tupleType, ok := typeInfo.(gocql.TupleTypeInfo); ok {
			var elements []string
			for _, elem := range tupleType.Elems {
				elements = append(elements, formatGocqlTypeInfo(elem))
			}
			return fmt.Sprintf("tuple<%s>", strings.Join(elements, ", "))
		}
	case gocql.TypeUDT:
		// Handle nested UDT types
		if udtType, ok := typeInfo.(gocql.UDTTypeInfo); ok {
			if udtType.Keyspace != "" {
				return fmt.Sprintf("%s.%s", udtType.Keyspace, udtType.Name)
			}
			return udtType.Name
		}
	default:
		// Handle native types
		return gocqlTypeToString(baseType)
	}

	return gocqlTypeToString(baseType)
}

// gocqlTypeToString converts a gocql.Type to its string representation
func gocqlTypeToString(t gocql.Type) string {
	switch t {
	case gocql.TypeCustom:
		return "custom"
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
	case gocql.TypeText:
		return "text"
	case gocql.TypeTimestamp:
		return "timestamp"
	case gocql.TypeUUID:
		return "uuid"
	case gocql.TypeVarchar:
		return "varchar"
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
	case gocql.TypeList:
		return "list"
	case gocql.TypeMap:
		return "map"
	case gocql.TypeSet:
		return "set"
	case gocql.TypeTuple:
		return "tuple"
	case gocql.TypeUDT:
		return "udt"
	default:
		return "unknown"
	}
}