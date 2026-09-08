package router

import (
	"strings"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostile values a malicious Parquet file could carry. Each one changes the
// meaning of the statement if it is pasted into CQL text instead of bound.
var hostileValues = map[string]string{
	"closes the string literal":   "'; DROP TABLE users; --",
	"doubled quote is not enough": "'' OR 1=1; --",
	"ends the statement":          "x'); DROP KEYSPACE app; --",
	"comment terminator":          "x' -- ",
	"newline and a second stmt":   "x'\nAPPLY BATCH;\nDROP TABLE users;",
	"collection punctuation":      "[1, 2]}); DROP TABLE users; --",
	"looks like a UDT":            "{a: 1}); DROP TABLE users; --",
	"backslash":                   `x\'; DROP TABLE users; --`,
}

// TestInsertTemplateCarriesNoData is the core guarantee: the statement text is
// built from the table and column names only, so nothing from the file can
// reach it.
func TestInsertTemplateCarriesNoData(t *testing.T) {
	query := buildInsertTemplate("app.users", []string{"id", "name", "note"})

	assert.Equal(t, "INSERT INTO app.users (id, name, note) VALUES (?, ?, ?)", query)

	// One placeholder per column, and no literal quoting anywhere.
	assert.Equal(t, 3, strings.Count(query, "?"))
	assert.NotContains(t, query, "'")
}

// TestHostileValuesAreBoundNotFormatted checks the hostile strings above reach
// gocql as parameters, byte for byte, and never touch the statement.
func TestHostileValuesAreBoundNotFormatted(t *testing.T) {
	for name, hostile := range hostileValues {
		t.Run(name, func(t *testing.T) {
			columns := []string{"id", "note"}
			values := []interface{}{int32(1), hostile}

			query := buildInsertTemplate("users", columns)
			bound := bindValuesForInsert(columns, values, map[string]string{
				"id":   "int",
				"note": "text",
			})

			// The statement is unchanged by the payload.
			assert.Equal(t, "INSERT INTO users (id, note) VALUES (?, ?)", query)
			assert.NotContains(t, query, "DROP")

			// The value arrives intact - not escaped, not mangled, not truncated.
			require.Len(t, bound, 2)
			assert.Equal(t, hostile, bound[1],
				"a bound value must reach the driver exactly as it was read")
		})
	}
}

// TestHostileUDTFieldNamesAreBound covers the specific hole in issue #83: UDT
// field names came from the file and were written into the statement raw.
func TestHostileUDTFieldNamesAreBound(t *testing.T) {
	udt := map[string]interface{}{
		"name":                         "alice",
		"x: 1}); DROP TABLE users; --": "surprise",
	}

	columns := []string{"id", "profile"}
	query := buildInsertTemplate("users", columns)
	bound := bindValuesForInsert(columns, []interface{}{int32(1), udt},
		map[string]string{"id": "int", "profile": "profile_type"})

	assert.NotContains(t, query, "DROP")
	assert.NotContains(t, query, "profile_type")

	// The whole map is one parameter; its keys are never statement text.
	require.Len(t, bound, 2)
	assert.Equal(t, udt, bound[1])
}

// TestBindValueConvertsUUIDs covers the one conversion that has to happen:
// Parquet has no UUID type, so these arrive as strings, and gocql will not bind
// a plain string to a uuid column.
func TestBindValueConvertsUUIDs(t *testing.T) {
	const raw = "550e8400-e29b-41d4-a716-446655440000"

	for _, cqlType := range []string{"uuid", "timeuuid", "UUID", " TimeUUID "} {
		t.Run(cqlType, func(t *testing.T) {
			got := bindValue(raw, cqlType)

			parsed, ok := got.(gocql.UUID)
			require.True(t, ok, "expected a gocql.UUID, got %T", got)
			assert.Equal(t, raw, parsed.String())
		})
	}
}

func TestBindValueLeavesEverythingElseAlone(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		cqlType string
	}{
		{"text is not touched", "hello", "text"},
		{"a uuid-shaped string in a text column stays a string", "550e8400-e29b-41d4-a716-446655440000", "text"},
		{"numbers pass through", int64(42), "bigint"},
		{"bools pass through", true, "boolean"},
		{"blobs pass through", []byte{0xde, 0xad}, "blob"},
		{"lists pass through", []interface{}{"a", "b"}, "list<text>"},
		{"unknown column type passes through", "hello", ""},
		{"a malformed uuid is left for Cassandra to reject", "not-a-uuid", "uuid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.val, bindValue(tt.val, tt.cqlType))
		})
	}
}

func TestBindValueKeepsNil(t *testing.T) {
	assert.Nil(t, bindValue(nil, "text"))
	assert.Nil(t, bindValue(nil, "uuid"))
}
