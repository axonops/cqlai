package router

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestACSVFieldIsTheColumnsJSONValue: a string as it is, a timestamp as the
// shell shows it, NULL as NULLVAL, and anything else as its JSON.
func TestACSVFieldIsTheColumnsJSONValue(t *testing.T) {
	fields, err := jsonFields(`{"id": 1, "\"MixedCase\"": "02134", "ts": "2024-01-02 03:04:05.678Z", "tags": ["a", "b, c"], "n": null, "ok": true}`)
	require.NoError(t, err)
	require.Len(t, fields, 6)

	types := []string{"int", "text", "timestamp", "list<text>", "int", "boolean"}
	var got []string
	for i, f := range fields {
		got = append(got, csvField(f, types[i], "null"))
	}
	assert.Equal(t, []string{"1", "02134", "2024-01-02 03:04:05.678+0000", `["a", "b, c"]`, "null", "true"}, got)

	_, err = jsonFields("not json")
	assert.Error(t, err)
}

// TestACSVRecordGoesInAsStrings, for Cassandra to read as each column's type,
// with NULLVAL as NULL and a quoted name kept in its quotes.
func TestACSVRecordGoesInAsStrings(t *testing.T) {
	doc, err := jsonRow([]string{"id", ` "MixedCase"`, "zip", "tags"}, []string{"1", "kept", "02134", "null"}, "null")
	require.NoError(t, err)
	var back map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(doc), &back))
	assert.Equal(t, map[string]interface{}{"id": "1", `"MixedCase"`: "kept", "zip": "02134", "tags": nil}, back)
}

func TestTableNamesAreReadAsCassandraStoresThem(t *testing.T) {
	for in, want := range map[string]string{"Orders": "orders", `"Orders"`: "Orders", `"a""b"`: `a"b`} {
		assert.Equal(t, want, storedName(in), in)
	}
	ks, table, ok := cutName(`"my.ks".t`)
	assert.True(t, ok)
	assert.Equal(t, `"my.ks"`, ks)
	assert.Equal(t, "t", table)
}
