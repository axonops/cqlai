package router

import (
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestACSVFieldIsTheColumnsJSONValue: a string as it is, a timestamp as the
// shell shows it, NULL as NULLVAL, and anything else as its JSON.
func TestACSVFieldIsTheColumnsJSONValue(t *testing.T) {
	fields, err := jsonFields(`{"id": 1, "\"MixedCase\"": "02134", "ts": "2024-01-02 03:04:05.678Z", "tags": ["a", "b, c"], "n": null, "ok": true}`)
	require.NoError(t, err)
	require.Len(t, fields, 6)

	ts := gocql.NewNativeType(4, gocql.TypeTimestamp, "")
	text := gocql.NewNativeType(4, gocql.TypeText, "")
	types := []gocql.TypeInfo{nil, text, ts, nil, nil, nil}
	var got []string
	for i, f := range fields {
		got = append(got, csvField(f, types[i], "null"))
	}
	assert.Equal(t, []string{"1", "02134", "2024-01-02 03:04:05.678+0000", `["a", "b, c"]`, "null", "true"}, got)

	_, err = jsonFields("not json")
	assert.Error(t, err)
}

// TestACSVRecordGoesInAsStrings, for Cassandra to read as each column's type
// with fromJson, and NULLVAL as NULL.
func TestACSVRecordGoesInAsStrings(t *testing.T) {
	values, err := fromJSONValues([]string{"1", "kept", "02134", "null", `say "hi"`}, "null")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{`"1"`, `"kept"`, `"02134"`, "null", `"say \"hi\""`}, values)
	assert.Equal(t, `INSERT INTO ks.t (id, "MixedCase") VALUES (fromJson(?), fromJson(?))`,
		fromJSONInsert("ks.t", []string{"id", `"MixedCase"`}))
}

// TestBeforeJSONEachFieldIsItsColumnsType: Cassandra 2.1 reads no JSON, so
// each field is made the Go value of its column's type here.
func TestBeforeJSONEachFieldIsItsColumnsType(t *testing.T) {
	native := func(typ gocql.Type) gocql.TypeInfo { return gocql.NewNativeType(3, typ, "") }
	types := []gocql.TypeInfo{native(gocql.TypeInt), native(gocql.TypeText), native(gocql.TypeTimestamp),
		native(gocql.TypeBlob), native(gocql.TypeBoolean), native(gocql.TypeUUID)}
	values, err := typedValues([]string{"7", "02134", "2024-01-02 03:04:05.678+0000", "0xcafe", "true", "null"}, types, "null")
	require.NoError(t, err)
	assert.Equal(t, int32(7), values[0])
	assert.Equal(t, "02134", values[1])
	assert.Equal(t, int64(1704164645678), values[2].(time.Time).UnixMilli())
	assert.Equal(t, []byte{0xca, 0xfe}, values[3])
	assert.Equal(t, true, values[4])
	assert.Nil(t, values[5])

	_, err = typedValues([]string{"seven"}, types[:1], "null")
	assert.Error(t, err)
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
