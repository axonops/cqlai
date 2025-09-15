package db

import (
	"fmt"
	"sync"

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

// UDTRegistry manages cached UDT definitions for all keyspaces
type UDTRegistry struct {
	definitions map[string]map[string]*UDTDefinition // keyspace -> udtname -> definition
	mu          sync.RWMutex
	session     *gocql.Session
}

// NewUDTRegistry creates a new UDT registry with the given session
func NewUDTRegistry(session *gocql.Session) *UDTRegistry {
	return &UDTRegistry{
		definitions: make(map[string]map[string]*UDTDefinition),
		session:     session,
	}
}

// LoadKeyspaceUDTs loads all UDT definitions for a given keyspace
func (r *UDTRegistry) LoadKeyspaceUDTs(keyspace string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Query system_schema.types for all UDTs in the keyspace
	query := `
		SELECT type_name, field_names, field_types
		FROM system_schema.types
		WHERE keyspace_name = ?`

	iter := r.session.Query(query, keyspace).Iter()
	defer iter.Close()

	// Initialize the keyspace map if it doesn't exist
	if r.definitions[keyspace] == nil {
		r.definitions[keyspace] = make(map[string]*UDTDefinition)
	}

	var typeName string
	var fieldNames []string
	var fieldTypes []string

	for iter.Scan(&typeName, &fieldNames, &fieldTypes) {
		// Validate that field names and types have the same length
		if len(fieldNames) != len(fieldTypes) {
			return fmt.Errorf("mismatched field names and types for UDT %s.%s", keyspace, typeName)
		}

		// Create the UDT definition
		udtDef := &UDTDefinition{
			Keyspace: keyspace,
			Name:     typeName,
			Fields:   make([]UDTField, len(fieldNames)),
		}

		// Parse each field's type information
		for i := range fieldNames {
			typeInfo, err := ParseCQLType(fieldTypes[i])
			if err != nil {
				return fmt.Errorf("failed to parse type for field %s in UDT %s.%s: %w",
					fieldNames[i], keyspace, typeName, err)
			}

			udtDef.Fields[i] = UDTField{
				Name:     fieldNames[i],
				TypeStr:  fieldTypes[i],
				TypeInfo: typeInfo,
			}
		}

		// Cache the definition
		r.definitions[keyspace][typeName] = udtDef
	}

	if err := iter.Close(); err != nil {
		return fmt.Errorf("failed to load UDTs for keyspace %s: %w", keyspace, err)
	}

	return nil
}

// GetUDTDefinition retrieves a cached UDT definition
func (r *UDTRegistry) GetUDTDefinition(keyspace, udtName string) (*UDTDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ks, ok := r.definitions[keyspace]; ok {
		if udt, ok := ks[udtName]; ok {
			return udt, nil
		}
	}

	return nil, fmt.Errorf("UDT %s.%s not found in registry", keyspace, udtName)
}

// GetUDTDefinitionOrLoad retrieves a UDT definition, loading it if necessary
func (r *UDTRegistry) GetUDTDefinitionOrLoad(keyspace, udtName string) (*UDTDefinition, error) {
	// Try to get from cache first
	if udt, err := r.GetUDTDefinition(keyspace, udtName); err == nil {
		return udt, nil
	}

	// Not in cache, try to load the keyspace UDTs
	if err := r.LoadKeyspaceUDTs(keyspace); err != nil {
		return nil, err
	}

	// Try again after loading
	return r.GetUDTDefinition(keyspace, udtName)
}

// Clear removes all cached UDT definitions
func (r *UDTRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.definitions = make(map[string]map[string]*UDTDefinition)
}

// ClearKeyspace removes cached UDT definitions for a specific keyspace
func (r *UDTRegistry) ClearKeyspace(keyspace string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.definitions, keyspace)
}

// GetAllUDTs returns all cached UDT definitions for a keyspace
func (r *UDTRegistry) GetAllUDTs(keyspace string) map[string]*UDTDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ks, ok := r.definitions[keyspace]; ok {
		// Return a copy to prevent external modification
		result := make(map[string]*UDTDefinition, len(ks))
		for k, v := range ks {
			result[k] = v
		}
		return result
	}

	return nil
}

// HasUDT checks if a UDT is cached
func (r *UDTRegistry) HasUDT(keyspace, udtName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ks, ok := r.definitions[keyspace]; ok {
		_, exists := ks[udtName]
		return exists
	}
	return false
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