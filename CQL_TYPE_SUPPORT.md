# Standardized Cassandra Data Type Handling

## Overview

cqlai now includes a comprehensive, standardized mechanism for handling all Cassandra/CQL data types through the `CQLTypeHandler` in `internal/db/types.go`.

## Architecture

### Type Handler (`CQLTypeHandler`)
Located in `cqlai/internal/db/types.go`

```go
type CQLTypeHandler struct {
    TimeFormat      string // Format for time display (default RFC3339)
    HexPrefix       string // Prefix for hex values (default "0x")
    NullString      string // String to display for null values (default "null")
    CollectionLimit int    // Max items to display in collections (0 = unlimited)
    TruncateStrings int    // Max length for strings (0 = no truncation)
}
```

### Two-Tier Type Resolution

1. **Type-Info Based**: When CQL type information is available from gocql
2. **Runtime Type Detection**: Fallback using Go's type system

## Supported CQL Data Types

### Native Types (Fully Supported ✅)

| CQL Type | Go Type | Display Format | Example |
|----------|---------|----------------|---------|
| **ascii** | string | Plain text | `"hello"` |
| **bigint** | int64 | Decimal | `9223372036854775807` |
| **blob** | []byte | Hexadecimal | `0x48656c6c6f` |
| **boolean** | bool | true/false | `true` |
| **counter** | int64 | Decimal | `42` |
| **date** | time.Time | YYYY-MM-DD | `2024-01-15` |
| **decimal** | *inf.Dec/string | Decimal string | `123.456` |
| **double** | float64 | Scientific notation | `3.14159` |
| **duration** | gocql.Duration | Human readable | `2mo5d100ns` |
| **float** | float32 | Scientific notation | `3.14` |
| **inet** | net.IP | IP address | `192.168.1.1` |
| **int** | int32 | Decimal | `2147483647` |
| **smallint** | int16 | Decimal | `32767` |
| **text** | string | Plain text | `"Hello World"` |
| **time** | time.Duration | Duration string | `13h45m30s` |
| **timestamp** | time.Time | RFC3339 | `2024-01-15T14:30:00Z` |
| **timeuuid** | gocql.UUID | UUID string | `550e8400-e29b-41d4-a716-446655440000` |
| **tinyint** | int8 | Decimal | `127` |
| **uuid** | gocql.UUID | UUID string | `123e4567-e89b-12d3-a456-426614174000` |
| **varchar** | string | Plain text | `"text value"` |
| **varint** | *big.Int | Decimal string | `123456789012345678901234567890` |

### Collection Types (Fully Supported ✅)

| CQL Type | Go Type | Display Format | Example |
|----------|---------|----------------|---------|
| **list<T>** | []T | Square brackets | `[1, 2, 3]` |
| **set<T>** | []T | Square brackets | `[a, b, c]` |
| **map<K,V>** | map[K]V | Curly braces | `{key1: val1, key2: val2}` |

### Complex Types (Fully Supported ✅)

| CQL Type | Go Type | Display Format | Example |
|----------|---------|----------------|---------|
| **tuple<T1,T2,...>** | []interface{} | Parentheses | `(1, "text", true)` |
| **UDT** | map[string]interface{} | Like maps | `{field1: val1, field2: val2}` |
| **frozen<T>** | Same as T | Same as T | Transparent |

### Special Types (Supported ✅)

| CQL Type | Go Type | Display Format | Notes |
|----------|---------|----------------|-------|
| **vector<float, n>** | []float32 | Square brackets | `[0.1, 0.2, 0.3]` - SAI vector search |
| **custom** | interface{} | String representation | User-defined types |

## Type-Specific Features

### Null Handling
- All types properly handle null values
- Configurable null display string (default: `"null"`)
- Zero values for time types shown as null

### Collection Features
- **Truncation**: Optional limit on collection items displayed
- **Nested Collections**: Properly formatted nested maps/lists
- **Type Preservation**: Maintains type information for elements

### Binary Data
- Blob types displayed as hexadecimal with configurable prefix
- Empty blobs shown as just the prefix (e.g., `0x`)

### Time Types
- Configurable time format (default RFC3339)
- Date type shows only date portion
- Duration types use human-readable format
- Zero times displayed as null

### Numeric Types
- Scientific notation for floats when appropriate
- Full precision for decimal and varint types
- Proper handling of counter columns

## Usage in cqlai

### SELECT Queries
```go
// In visitor_stubs.go
typeHandler := db.NewCQLTypeHandler()
for _, rowMap := range rows {
    for i, col := range columns {
        if val, ok := rowMap[col.Name]; ok {
            // Use type info when available for best formatting
            row[i] = typeHandler.FormatValue(val, col.TypeInfo)
        }
    }
}
```

### Configuration Options
```go
handler := db.NewCQLTypeHandler()
handler.TimeFormat = "2006-01-02 15:04:05"  // Custom time format
handler.HexPrefix = ""                       // No prefix for hex
handler.NullString = "NULL"                  // SQL-style null
handler.CollectionLimit = 10                 // Show max 10 items
handler.TruncateStrings = 100               // Truncate long strings
```

## Benefits

1. **Consistency**: All types formatted uniformly across the application
2. **Completeness**: Handles all 30+ CQL types including vectors and UDTs
3. **Flexibility**: Configurable formatting options
4. **Performance**: Efficient type detection and formatting
5. **Maintainability**: Centralized type handling logic
6. **Extensibility**: Easy to add new types or formats

## Error Handling

- Never panics on unknown types
- Falls back to `fmt.Sprintf("%v", val)` for unrecognized types
- Handles nil pointers gracefully
- Manages type assertion failures safely

## Testing Coverage

The type handler covers:
- All native CQL types
- All collection types with various element types
- Nested collections (map of lists, list of maps, etc.)
- User-defined types
- Vector types for ML/AI workloads
- Null and zero value handling
- Edge cases (empty collections, zero times, etc.)

## Future Enhancements

Potential improvements:
1. JSON output format option
2. CSV-compatible formatting
3. Custom type registries for application-specific types
4. Locale-specific number formatting
5. Binary data encoding options (base64, etc.)
6. Collection depth limits for deeply nested structures