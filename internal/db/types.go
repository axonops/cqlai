package db

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// CQLTypeHandler provides standardized handling for all Cassandra/CQL data types
type CQLTypeHandler struct {
	// Configuration options
	TimeFormat      string // Format for time display (default RFC3339)
	HexPrefix       string // Prefix for hex values (default "0x")
	NullString      string // String to display for null values (default "null")
	CollectionLimit int    // Max items to display in collections (0 = unlimited)
	TruncateStrings int    // Max length for strings (0 = no truncation)
}

// NewCQLTypeHandler creates a new type handler with default settings
func NewCQLTypeHandler() *CQLTypeHandler {
	return &CQLTypeHandler{
		TimeFormat:      time.RFC3339,
		HexPrefix:       "0x",
		NullString:      "null",
		CollectionLimit: 0,
		TruncateStrings: 0,
	}
}

// FormatValue formats any CQL value for display
func (h *CQLTypeHandler) FormatValue(val interface{}, typeInfo gocql.TypeInfo) string {
	if val == nil {
		return h.NullString
	}

	// If we have type info, use it for better formatting
	if typeInfo != nil {
		return h.formatWithTypeInfo(val, typeInfo)
	}

	// Otherwise use runtime type detection
	return h.formatByType(val)
}

// formatWithTypeInfo formats a value using CQL type information
func (h *CQLTypeHandler) formatWithTypeInfo(val interface{}, typeInfo gocql.TypeInfo) string {
	switch typeInfo.Type() {
	// String types
	case gocql.TypeVarchar, gocql.TypeText, gocql.TypeAscii:
		return h.formatString(val)

	// Numeric types
	case gocql.TypeInt:
		return h.formatInt32(val)
	case gocql.TypeBigInt, gocql.TypeCounter:
		return h.formatInt64(val)
	case gocql.TypeSmallInt:
		return h.formatInt16(val)
	case gocql.TypeTinyInt:
		return h.formatInt8(val)
	case gocql.TypeFloat:
		return h.formatFloat32(val)
	case gocql.TypeDouble:
		return h.formatFloat64(val)
	case gocql.TypeDecimal:
		return h.formatDecimal(val)
	case gocql.TypeVarint:
		return h.formatVarint(val)

	// Boolean
	case gocql.TypeBoolean:
		return h.formatBool(val)

	// UUID types
	case gocql.TypeUUID, gocql.TypeTimeUUID:
		return h.formatUUID(val)

	// Time types
	case gocql.TypeTimestamp:
		return h.formatTimestamp(val)
	case gocql.TypeDate:
		return h.formatDate(val)
	case gocql.TypeTime:
		return h.formatTime(val)
	case gocql.TypeDuration:
		return h.formatDuration(val)

	// Binary
	case gocql.TypeBlob:
		return h.formatBlob(val)

	// Network
	case gocql.TypeInet:
		return h.formatInet(val)

	// Collections
	case gocql.TypeList:
		return h.formatList(val)
	case gocql.TypeSet:
		return h.formatSet(val)
	case gocql.TypeMap:
		return h.formatMap(val)

	// Complex types
	case gocql.TypeUDT:
		return h.formatUDT(val)
	case gocql.TypeTuple:
		return h.formatTuple(val)
	case gocql.TypeCustom:
		// For custom types (including vectors), always use formatVector
		// which will handle both vectors and other custom types appropriately
		return h.formatVector(val)

	default:
		return h.formatByType(val)
	}
}

// formatByType formats a value based on its runtime type
func (h *CQLTypeHandler) formatByType(val interface{}) string {
	if val == nil {
		return h.NullString
	}

	switch v := val.(type) {
	// String types
	case string:
		return h.truncateString(v)
	case []byte:
		return h.formatBytes(v)

	// Numeric types
	case int:
		return fmt.Sprintf("%d", v)
	case int8:
		return fmt.Sprintf("%d", v)
	case int16:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint8:
		return fmt.Sprintf("%d", v)
	case uint16:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%g", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case *big.Int:
		return v.String()
	case *big.Float:
		return v.String()

	// Boolean
	case bool:
		return fmt.Sprintf("%v", v)

	// UUID types
	case gocql.UUID:
		return v.String()
	case *gocql.UUID:
		if v != nil {
			return v.String()
		}
		return h.NullString

	// Time types
	case time.Time:
		if v.IsZero() {
			return h.NullString
		}
		return v.Format(h.TimeFormat)
	case *time.Time:
		if v != nil && !v.IsZero() {
			return v.Format(h.TimeFormat)
		}
		return h.NullString
	case time.Duration:
		return v.String()

	// Network types
	case net.IP:
		return v.String()
	case *net.IP:
		if v != nil {
			return v.String()
		}
		return h.NullString

	// Collection types
	case map[string]interface{}:
		return h.formatGenericMap(v)
	case []interface{}:
		return h.formatGenericList(v)
	case []map[string]interface{}:
		// Format list of maps (e.g., list<frozen<udt>>)
		// Use formatGenericMap which will quote strings inside the maps
		if len(v) == 0 {
			return "[]"
		}
		items := make([]string, 0, len(v))
		for i, m := range v {
			if h.CollectionLimit > 0 && i >= h.CollectionLimit {
				items = append(items, "...")
				break
			}
			items = append(items, h.formatGenericMap(m))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case []string:
		return h.formatStringList(v)
	case []int:
		return h.formatIntList(v)
	case []int32:
		return h.formatInt32List(v)
	case []int64:
		return h.formatInt64List(v)
	case []float32:
		return h.formatFloat32List(v)
	case []float64:
		return h.formatFloat64List(v)
	case []gocql.UUID:
		return h.formatUUIDList(v)
	case []bool:
		return h.formatBoolList(v)

	// Default fallback
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Individual type formatters

func (h *CQLTypeHandler) formatString(val interface{}) string {
	if s, ok := val.(string); ok {
		return h.truncateString(s)
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) truncateString(s string) string {
	if h.TruncateStrings > 0 && len(s) > h.TruncateStrings {
		return s[:h.TruncateStrings] + "..."
	}
	return s
}

func (h *CQLTypeHandler) formatInt8(val interface{}) string {
	switch v := val.(type) {
	case int8:
		return fmt.Sprintf("%d", v)
	case *int8:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatInt16(val interface{}) string {
	switch v := val.(type) {
	case int16:
		return fmt.Sprintf("%d", v)
	case *int16:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatInt32(val interface{}) string {
	switch v := val.(type) {
	case int32:
		return fmt.Sprintf("%d", v)
	case int:
		return fmt.Sprintf("%d", v)
	case *int32:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatInt64(val interface{}) string {
	switch v := val.(type) {
	case int64:
		return fmt.Sprintf("%d", v)
	case *int64:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatFloat32(val interface{}) string {
	switch v := val.(type) {
	case float32:
		return fmt.Sprintf("%g", v)
	case *float32:
		if v != nil {
			return fmt.Sprintf("%g", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatFloat64(val interface{}) string {
	switch v := val.(type) {
	case float64:
		return fmt.Sprintf("%g", v)
	case *float64:
		if v != nil {
			return fmt.Sprintf("%g", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatDecimal(val interface{}) string {
	// Decimal is typically returned as *inf.Dec or string
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatVarint(val interface{}) string {
	switch v := val.(type) {
	case *big.Int:
		return v.String()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatBool(val interface{}) string {
	switch v := val.(type) {
	case bool:
		return fmt.Sprintf("%v", v)
	case *bool:
		if v != nil {
			return fmt.Sprintf("%v", *v)
		}
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatUUID(val interface{}) string {
	switch v := val.(type) {
	case gocql.UUID:
		return v.String()
	case *gocql.UUID:
		if v != nil {
			return v.String()
		}
		return h.NullString
	case string:
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatTimestamp(val interface{}) string {
	switch v := val.(type) {
	case time.Time:
		if v.IsZero() {
			return h.NullString
		}
		return v.Format(h.TimeFormat)
	case *time.Time:
		if v != nil && !v.IsZero() {
			return v.Format(h.TimeFormat)
		}
		return h.NullString
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatDate(val interface{}) string {
	switch v := val.(type) {
	case time.Time:
		if v.IsZero() {
			return h.NullString
		}
		return v.Format("2006-01-02")
	case string:
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatTime(val interface{}) string {
	switch v := val.(type) {
	case time.Duration:
		return v.String()
	case int64:
		// Time is stored as nanoseconds since midnight
		d := time.Duration(v)
		return d.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatDuration(val interface{}) string {
	switch v := val.(type) {
	case time.Duration:
		return v.String()
	case gocql.Duration:
		return fmt.Sprintf("%dmo%dd%dns", v.Months, v.Days, v.Nanoseconds)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatBlob(val interface{}) string {
	switch v := val.(type) {
	case []byte:
		return h.formatBytes(v)
	case string:
		// Sometimes blob is returned as hex string
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatBytes(b []byte) string {
	if len(b) == 0 {
		return h.HexPrefix
	}
	return h.HexPrefix + hex.EncodeToString(b)
}

func (h *CQLTypeHandler) formatInet(val interface{}) string {
	switch v := val.(type) {
	case net.IP:
		return v.String()
	case *net.IP:
		if v != nil {
			return v.String()
		}
		return h.NullString
	case string:
		return v
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatList(val interface{}) string {
	switch v := val.(type) {
	case []interface{}:
		return h.formatGenericList(v)
	case []string:
		return h.formatStringList(v)
	default:
		return h.formatByType(val)
	}
}

func (h *CQLTypeHandler) formatSet(val interface{}) string {
	// Sets are typically returned as slices
	return h.formatList(val)
}

func (h *CQLTypeHandler) formatMap(val interface{}) string {
	switch v := val.(type) {
	case map[string]interface{}:
		return h.formatGenericMap(v)
	case map[string]string:
		return h.formatStringMap(v)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (h *CQLTypeHandler) formatUDT(val interface{}) string {
	// UDTs are typically returned as maps
	if m, ok := val.(map[string]interface{}); ok {
		return h.formatGenericMap(m)
	}
	return fmt.Sprintf("%v", val)
}

func (h *CQLTypeHandler) formatTuple(val interface{}) string {
	// Tuples are typically returned as slices
	if l, ok := val.([]interface{}); ok {
		return "(" + h.formatListItems(l) + ")"
	}
	return fmt.Sprintf("%v", val)
}

// formatVector formats a vector value with proper comma separation
// This handles both vector types and other custom types from Cassandra
func (h *CQLTypeHandler) formatVector(val interface{}) string {
	
	// Vectors can come through as various slice types
	switch v := val.(type) {
	case []float32:
		return h.formatFloat32List(v)
	case []float64:
		return h.formatFloat64List(v)
	case []interface{}:
		// Sometimes vectors come as generic interface slices
		return h.formatGenericList(v)
	case string:
		// Sometimes the value is already a string
		if len(v) > 2 && v[0] == '[' && v[len(v)-1] == ']' {
			inner := v[1 : len(v)-1]
			if inner != "" && !strings.Contains(inner, ",") {
				parts := strings.Fields(inner)
				if len(parts) > 0 {
					allNumbers := true
					for _, part := range parts {
						if _, err := strconv.ParseFloat(part, 64); err != nil {
							allNumbers = false
							break
						}
					}
					if allNumbers {
						return "[" + strings.Join(parts, ", ") + "]"
					}
				}
			}
		}
		return v
	default:
		// For any other type, try to parse the string representation
		str := fmt.Sprintf("%v", val)
		// Check if it looks like a vector/array formatted without commas
		if len(str) > 2 && str[0] == '[' && str[len(str)-1] == ']' {
			// Extract content between brackets
			inner := str[1 : len(str)-1]
			if inner != "" && !strings.Contains(inner, ",") {
				// Split by spaces and rejoin with commas
				parts := strings.Fields(inner)
				if len(parts) > 0 {
					// Verify all parts are numbers (characteristic of vectors)
					allNumbers := true
					for _, part := range parts {
						if _, err := strconv.ParseFloat(part, 64); err != nil {
							allNumbers = false
							break
						}
					}
					if allNumbers {
						return "[" + strings.Join(parts, ", ") + "]"
					}
				}
			}
		}
		return str
	}
}

// Collection formatters

// formatValueInCollection formats a value that appears inside a collection or UDT
// This adds quotes to strings for proper CQL representation
func (h *CQLTypeHandler) formatValueInCollection(val interface{}) string {
	if val == nil {
		return h.NullString
	}

	switch v := val.(type) {
	case string:
		// Quote strings inside collections/UDTs
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case map[string]interface{}:
		return h.formatGenericMap(v)
	case []interface{}:
		return h.formatGenericListWithQuotes(v)
	case []map[string]interface{}:
		// Format list of maps (e.g., list<frozen<udt>>)
		// Use formatGenericMap which will quote strings inside the maps
		if len(v) == 0 {
			return "[]"
		}
		items := make([]string, 0, len(v))
		for i, m := range v {
			if h.CollectionLimit > 0 && i >= h.CollectionLimit {
				items = append(items, "...")
				break
			}
			items = append(items, h.formatGenericMap(m))
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		// For other types, use the regular formatting
		return h.formatByType(val)
	}
}

func (h *CQLTypeHandler) formatGenericListWithQuotes(l []interface{}) string {
	if len(l) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, h.formatValueInCollection(v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatGenericMap(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	pairs := make([]string, 0, len(m))
	count := 0
	for k, v := range m {
		if h.CollectionLimit > 0 && count >= h.CollectionLimit {
			pairs = append(pairs, "...")
			break
		}
		// Format values with quotes for strings inside UDTs
		formattedValue := h.formatValueInCollection(v)
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, formattedValue))
		count++
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (h *CQLTypeHandler) formatStringMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	
	pairs := make([]string, 0, len(m))
	count := 0
	for k, v := range m {
		if h.CollectionLimit > 0 && count >= h.CollectionLimit {
			pairs = append(pairs, "...")
			break
		}
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, v))
		count++
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (h *CQLTypeHandler) formatGenericList(l []interface{}) string {
	if len(l) == 0 {
		return "[]"
	}
	return "[" + h.formatListItems(l) + "]"
}

func (h *CQLTypeHandler) formatListItems(l []interface{}) string {
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, h.formatByType(v))
	}
	return strings.Join(items, ", ")
}

func (h *CQLTypeHandler) formatStringList(l []string) string {
	if len(l) == 0 {
		return "[]"
	}
	
	limit := len(l)
	if h.CollectionLimit > 0 && limit > h.CollectionLimit {
		limit = h.CollectionLimit
	}
	
	result := "[" + strings.Join(l[:limit], ", ")
	if limit < len(l) {
		result += ", ..."
	}
	return result + "]"
}

func (h *CQLTypeHandler) formatIntList(l []int) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%d", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatInt32List(l []int32) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%d", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatInt64List(l []int64) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%d", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatFloat32List(l []float32) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%g", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatFloat64List(l []float64) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%g", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatUUIDList(l []gocql.UUID) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, v.String())
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func (h *CQLTypeHandler) formatBoolList(l []bool) string {
	if len(l) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(l))
	for i, v := range l {
		if h.CollectionLimit > 0 && i >= h.CollectionLimit {
			items = append(items, "...")
			break
		}
		items = append(items, fmt.Sprintf("%v", v))
	}
	return "[" + strings.Join(items, ", ") + "]"
}