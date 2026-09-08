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

// bindValue converts a single Parquet value for the destination CQL type.
func bindValue(val interface{}, cqlType string) interface{} {
	if val == nil {
		return nil
	}

	s, isString := val.(string)
	if !isString {
		return val
	}

	// gocql will not bind a plain string to a uuid column, and Parquet has no
	// UUID type of its own so these arrive as text.
	switch strings.ToLower(strings.TrimSpace(cqlType)) {
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
