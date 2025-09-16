# UDT (User-Defined Type) Support in CQLAI

## Overview

CQLAI now has full support for Cassandra User-Defined Types (UDTs), providing proper display and handling of complex nested data structures.

## Implementation Details

### The Challenge

The Apache Cassandra gocql driver v2 has a fundamental limitation: when scanning UDT columns without compile-time type information, it returns empty maps `{}` instead of the actual data. This is because gocql attempts to decode UDTs into Go maps but fails without knowing the exact structure at compile time.

### The Solution

We implemented a comprehensive UDT support system that:

1. **Bypasses gocql's automatic decoding** using a custom `RawBytes` type that implements the `UnmarshalCQL` interface
2. **Extracts raw binary data** directly from Cassandra for UDT columns
3. **Implements a binary protocol decoder** that properly decodes the raw bytes using schema information
4. **Loads UDT definitions** from Cassandra's system tables (`system_schema.types` and `system_schema.columns`)
5. **Recursively decodes** nested structures including UDTs within collections and collections within UDTs

## Components

### 1. Type Parser (`internal/db/type_parser.go`)
Parses CQL type strings like `frozen<address>` or `list<frozen<map<text, frozen<address>>>>` into structured type information.

### 2. UDT Schema Loader (`internal/db/udt_schema.go`)
- Loads UDT definitions from `system_schema.types`
- Caches definitions for performance
- Provides thread-safe access to UDT metadata

### 3. Binary Decoder (`internal/db/udt_decoder.go`)
- Decodes Cassandra's binary protocol format
- Handles all CQL native types
- Recursively decodes nested structures
- Properly handles null values and fields

### 4. Raw Bytes Scanner (`internal/db/raw_bytes.go`)
- Custom type that implements `UnmarshalCQL`
- Captures raw bytes from gocql instead of letting it decode
- Essential for bypassing gocql's broken UDT decoding

### 5. Query Result Enhancement
- Modified `ExecuteSelectQuery` to detect UDT columns
- Enhanced streaming result handler to use `RawBytes` for UDT columns
- Automatic detection of keyspace and table from queries

## Supported Features

✅ **Simple UDTs**: Display of basic user-defined types
✅ **Frozen UDTs**: Proper handling of frozen UDT columns
✅ **Nested UDTs**: UDTs containing other UDTs
✅ **Collections of UDTs**: Lists, sets, and maps containing UDT values
✅ **UDTs with Collections**: UDTs containing collection fields
✅ **Null Handling**: Proper display of null UDT fields
✅ **Multiple Keyspaces**: Supports cross-keyspace UDT references
✅ **COPY Commands**: Export/import UDT data as JSON in CSV files
✅ **Tab Completion**: Field-level completion for UDT columns

## Usage Examples

### Creating a UDT

```sql
CREATE TYPE address (
    street text,
    city text,
    state text,
    zip_code text,
    country text
);

CREATE TABLE users (
    user_id uuid PRIMARY KEY,
    name text,
    home_address frozen<address>,
    work_address frozen<address>
);
```

### Querying UDT Data

```sql
SELECT * FROM users;
```

Output in CQLAI:
```
┌──────────────────────────────────────┬──────────┬──────────────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────────────┐
│ user_id (PK)                         │ name     │ home_address                                                     │ work_address                                                      │
├──────────────────────────────────────┼──────────┼──────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────────────┤
│ 123e4567-e89b-12d3-a456-426614174000 │ John Doe │ {street: 123 Main St, city: New York, state: NY, zip_code:      │ {street: 456 Work Ave, city: New York, state: NY, zip_code:      │
│                                      │          │ 10001, country: USA}                                             │ 10002, country: USA}                                              │
└──────────────────────────────────────┴──────────┴──────────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────────────┘
```

### COPY TO with UDTs

```sql
COPY users TO 'users.csv' WITH HEADER=true;
```

UDT data is exported as JSON in the CSV file for easy re-import.

### DESCRIBE TYPE

```sql
DESCRIBE TYPE address;
```

Shows the full UDT definition with all fields and their types.

## Technical Notes

### Binary Protocol

UDT data in Cassandra's binary protocol is encoded as:
- 4 bytes: field length (int32)
- N bytes: field data
- Repeated for each field in definition order
- Null fields are indicated by -1 length

### Performance Considerations

- UDT schemas are cached per keyspace to avoid repeated queries
- Binary decoding is performed on-demand
- Raw bytes extraction has minimal overhead

## Known Limitations

1. **Batch Mode JSON Export**: Currently shows empty objects for UDTs in batch mode JSON output. Interactive mode works correctly.
2. **Field Ordering**: UDT fields may display in different orders depending on the internal map implementation.
3. **Large UDTs**: Very large UDTs with many fields may affect display formatting.

## Future Enhancements

- [ ] Improve batch mode JSON export for UDTs
- [ ] Add UDT field ordering preservation
- [ ] Support for UDT alterations and schema evolution
- [ ] Performance optimizations for large UDT collections