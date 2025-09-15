package db

import (
	"fmt"
	"strings"
)

// CQLTypeInfo represents parsed CQL type information
type CQLTypeInfo struct {
	BaseType   string         // "text", "int", "list", "map", "udt", etc.
	Frozen     bool           // Whether the type is frozen
	Parameters []*CQLTypeInfo // For collections/tuples - element types
	UDTName    string         // For UDT types - the name of the UDT
	Keyspace   string         // For UDT types - optional keyspace qualifier
}

// ParseCQLType parses a CQL type string into structured type information
func ParseCQLType(typeStr string) (*CQLTypeInfo, error) {
	typeStr = strings.TrimSpace(typeStr)
	if typeStr == "" {
		return nil, fmt.Errorf("empty type string")
	}

	parser := &typeParser{
		input: typeStr,
		pos:   0,
	}
	return parser.parse()
}

type typeParser struct {
	input string
	pos   int
}

func (p *typeParser) parse() (*CQLTypeInfo, error) {
	return p.parseType()
}

func (p *typeParser) parseType() (*CQLTypeInfo, error) {
	// Check for frozen modifier
	if p.consumeKeyword("frozen") {
		if !p.consume('<') {
			return nil, fmt.Errorf("expected '<' after 'frozen' at position %d", p.pos)
		}
		typeInfo, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if !p.consume('>') {
			return nil, fmt.Errorf("expected '>' to close 'frozen' at position %d", p.pos)
		}
		typeInfo.Frozen = true
		return typeInfo, nil
	}

	// Parse the base type
	typeName := p.parseIdentifier()
	if typeName == "" {
		return nil, fmt.Errorf("expected type name at position %d", p.pos)
	}

	typeInfo := &CQLTypeInfo{
		BaseType: strings.ToLower(typeName),
	}

	// Check for parameterized types
	switch typeInfo.BaseType {
	case "list", "set":
		if !p.consume('<') {
			return nil, fmt.Errorf("expected '<' after '%s' at position %d", typeInfo.BaseType, p.pos)
		}
		elementType, err := p.parseType()
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s element type: %w", typeInfo.BaseType, err)
		}
		if !p.consume('>') {
			return nil, fmt.Errorf("expected '>' to close '%s' at position %d", typeInfo.BaseType, p.pos)
		}
		typeInfo.Parameters = []*CQLTypeInfo{elementType}

	case "map":
		if !p.consume('<') {
			return nil, fmt.Errorf("expected '<' after 'map' at position %d", p.pos)
		}
		keyType, err := p.parseType()
		if err != nil {
			return nil, fmt.Errorf("failed to parse map key type: %w", err)
		}
		if !p.consume(',') {
			return nil, fmt.Errorf("expected ',' between map key and value types at position %d", p.pos)
		}
		valueType, err := p.parseType()
		if err != nil {
			return nil, fmt.Errorf("failed to parse map value type: %w", err)
		}
		if !p.consume('>') {
			return nil, fmt.Errorf("expected '>' to close 'map' at position %d", p.pos)
		}
		typeInfo.Parameters = []*CQLTypeInfo{keyType, valueType}

	case "tuple":
		if !p.consume('<') {
			return nil, fmt.Errorf("expected '<' after 'tuple' at position %d", p.pos)
		}
		var elements []*CQLTypeInfo
		for {
			elementType, err := p.parseType()
			if err != nil {
				return nil, fmt.Errorf("failed to parse tuple element: %w", err)
			}
			elements = append(elements, elementType)

			if p.consume('>') {
				break
			}
			if !p.consume(',') {
				return nil, fmt.Errorf("expected ',' or '>' in tuple at position %d", p.pos)
			}
		}
		typeInfo.Parameters = elements

	default:
		// Check if this might be a UDT (could be qualified with keyspace)
		if p.consume('.') {
			// This was a keyspace name, now get the actual UDT name
			typeInfo.Keyspace = typeName
			udtName := p.parseIdentifier()
			if udtName == "" {
				return nil, fmt.Errorf("expected UDT name after keyspace at position %d", p.pos)
			}
			typeInfo.BaseType = "udt"
			typeInfo.UDTName = udtName
		} else if !isPrimitiveType(typeInfo.BaseType) {
			// If it's not a known primitive type, treat it as a UDT
			typeInfo.UDTName = typeName
			typeInfo.BaseType = "udt"
		}
	}

	return typeInfo, nil
}

func (p *typeParser) parseIdentifier() string {
	p.skipWhitespace()
	start := p.pos
	for p.pos < len(p.input) {
		ch := p.input[p.pos]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *typeParser) consume(ch byte) bool {
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == ch {
		p.pos++
		return true
	}
	return false
}

func (p *typeParser) consumeKeyword(keyword string) bool {
	p.skipWhitespace()
	if p.pos+len(keyword) <= len(p.input) {
		if strings.EqualFold(p.input[p.pos:p.pos+len(keyword)], keyword) {
			// Make sure it's not part of a larger identifier
			if p.pos+len(keyword) < len(p.input) {
				next := p.input[p.pos+len(keyword)]
				if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') ||
					(next >= '0' && next <= '9') || next == '_' {
					return false
				}
			}
			p.pos += len(keyword)
			return true
		}
	}
	return false
}

func (p *typeParser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' ||
		p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

// isPrimitiveType checks if a type is a known CQL primitive type
func isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"ascii":     true,
		"bigint":    true,
		"blob":      true,
		"boolean":   true,
		"counter":   true,
		"date":      true,
		"decimal":   true,
		"double":    true,
		"duration":  true,
		"float":     true,
		"inet":      true,
		"int":       true,
		"smallint":  true,
		"text":      true,
		"time":      true,
		"timestamp": true,
		"timeuuid":  true,
		"tinyint":   true,
		"uuid":      true,
		"varchar":   true,
		"varint":    true,
	}
	return primitives[typeName]
}

// String returns a string representation of the TypeInfo
func (t *CQLTypeInfo) String() string {
	var result strings.Builder

	if t.Frozen {
		result.WriteString("frozen<")
	}

	switch t.BaseType {
	case "list", "set":
		result.WriteString(t.BaseType)
		result.WriteString("<")
		if len(t.Parameters) > 0 {
			result.WriteString(t.Parameters[0].String())
		}
		result.WriteString(">")
	case "map":
		result.WriteString("map<")
		if len(t.Parameters) > 1 {
			result.WriteString(t.Parameters[0].String())
			result.WriteString(", ")
			result.WriteString(t.Parameters[1].String())
		}
		result.WriteString(">")
	case "tuple":
		result.WriteString("tuple<")
		for i, param := range t.Parameters {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(param.String())
		}
		result.WriteString(">")
	case "udt":
		if t.Keyspace != "" {
			result.WriteString(t.Keyspace)
			result.WriteString(".")
		}
		result.WriteString(t.UDTName)
	default:
		result.WriteString(t.BaseType)
	}

	if t.Frozen {
		result.WriteString(">")
	}

	return result.String()
}