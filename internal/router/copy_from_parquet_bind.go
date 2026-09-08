package router

import (
	"strings"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// buildInsertTemplate returns an INSERT whose values are placeholders.
//
// The table and column names are identifiers rather than data - they come from
// the COPY command the user typed, not from the Parquet file - so they are the
// only things interpolated. Every value travels as a bind parameter.
func buildInsertTemplate(table string, columns []string) string {
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	return "INSERT INTO " + table +
		" (" + strings.Join(columns, ", ") + ")" +
		" VALUES (" + strings.Join(placeholders, ", ") + ")"
}

// bindValuesForInsert converts a row's values into what gocql should bind.
//
// Most values go through untouched: gocql marshals Go strings, numbers, bools,
// byte slices, times, slices and maps against whatever the destination column
// is. The exceptions are the CQL types gocql will not accept a plain string
// for, which is what this handles.
//
// columnTypes holds the destination table's real CQL types, read from
// system_schema.columns. Where a column is missing from it - no keyspace, or a
// table we could not read - the value is passed through and gocql decides.
func bindValuesForInsert(columns []string, values []interface{}, columnTypes map[string]string) []interface{} {
	bound := make([]interface{}, len(values))
	for i, val := range values {
		var cqlType string
		if i < len(columns) {
			cqlType = columnTypes[columns[i]]
		}
		bound[i] = bindValue(val, cqlType)
	}
	return bound
}

// cqlPrimitives are the CQL types gocql marshals from a plain Go value,
// including from an untyped nil.
var cqlPrimitives = map[string]bool{
	"ascii": true, "bigint": true, "blob": true, "boolean": true,
	"counter": true, "date": true, "decimal": true, "double": true,
	"duration": true, "float": true, "inet": true, "int": true,
	"smallint": true, "text": true, "time": true, "timestamp": true,
	"timeuuid": true, "tinyint": true, "uuid": true, "varchar": true,
	"varint": true,
}

// normaliseCQLType lowercases a type and unwraps frozen<...>, so frozen<profile>
// and profile are treated alike.
func normaliseCQLType(cqlType string) string {
	t := strings.ToLower(strings.TrimSpace(cqlType))
	for strings.HasPrefix(t, "frozen<") && strings.HasSuffix(t, ">") {
		t = strings.TrimSpace(t[len("frozen<") : len(t)-1])
	}
	return t
}

// isUserDefinedType reports whether a column type is a UDT, meaning neither a
// primitive nor one of the built-in containers.
func isUserDefinedType(cqlType string) bool {
	t := normaliseCQLType(cqlType)
	if t == "" || cqlPrimitives[t] {
		return false
	}
	for _, container := range []string{"list<", "set<", "map<", "tuple<", "vector<"} {
		if strings.HasPrefix(t, container) {
			return false
		}
	}
	return true
}

// bindValue converts a single Parquet value for the destination CQL type.
func bindValue(val interface{}, cqlType string) interface{} {
	if val == nil {
		// gocql marshals an untyped nil for every primitive and container, but
		// refuses it for a UDT. Binding a nil or empty map there is not a
		// substitute: that writes a non-null UDT whose fields are all zero.
		// UnsetValue leaves the column unwritten, which is a real null.
		if isUserDefinedType(cqlType) {
			return gocql.UnsetValue
		}
		return nil
	}

	s, isString := val.(string)
	if !isString {
		return val
	}

	// gocql will not bind a plain string to a uuid column, and Parquet has no
	// UUID type of its own so these arrive as text.
	switch normaliseCQLType(cqlType) {
	case "uuid", "timeuuid":
		parsed, err := gocql.ParseUUID(strings.TrimSpace(s))
		if err != nil {
			// Leave it alone and let the insert fail with Cassandra's own
			// message, rather than silently substituting something else.
			return val
		}
		return parsed
	}

	return val
}
