package db

import (
	"encoding/json"
	"math/big"
	"net"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONValueWritesWhatCassandraHasAndJSONDoesNot.
//
// Left to itself json.Marshal writes a UUID as sixteen numbers and a blob as
// base64, neither of which is what the value is anywhere else in cqlai.
func TestJSONValueWritesWhatCassandraHasAndJSONDoesNot(t *testing.T) {
	id, err := gocql.ParseUUID("d2177dd0-eaa2-11de-a572-001b779c76e3")
	require.NoError(t, err)

	for name, this := range map[string]struct {
		value interface{}
		want  string
	}{
		"uuid":      {id, `"d2177dd0-eaa2-11de-a572-001b779c76e3"`},
		"blob":      {[]byte{0xca, 0xfe}, `"0xcafe"`},
		"timestamp": {time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC), `"2026-09-11T10:00:00Z"`},
		"inet":      {net.ParseIP("10.0.0.1"), `"10.0.0.1"`},
		"varint":    {big.NewInt(12345678901234), `"12345678901234"`},
		"int":       {7, `7`},
		"boolean":   {true, `true`},
		"text":      {"alice", `"alice"`},
		"null":      {nil, `null`},
	} {
		encoded, err := json.Marshal(JSONValue(this.value))
		require.NoError(t, err, name)
		assert.Equal(t, this.want, string(encoded), name)
	}
}

// TestJSONValueWalksIntoCollections, which is where the awkward values usually
// are: a list of UUIDs is a list of sixteen-number arrays otherwise.
func TestJSONValueWalksIntoCollections(t *testing.T) {
	id, err := gocql.ParseUUID("d2177dd0-eaa2-11de-a572-001b779c76e3")
	require.NoError(t, err)

	encoded, err := json.Marshal(JSONValue([]gocql.UUID{id}))
	require.NoError(t, err)
	assert.JSONEq(t, `["d2177dd0-eaa2-11de-a572-001b779c76e3"]`, string(encoded))

	encoded, err = json.Marshal(JSONValue(map[string][]byte{"key": {0x01}}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"key":"0x01"}`, string(encoded))

	// A map whose keys are not strings - map<uuid, text> - cannot be written at
	// all without this: the keys become what they read as everywhere else.
	encoded, err = json.Marshal(JSONValue(map[gocql.UUID]string{id: "alice"}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"d2177dd0-eaa2-11de-a572-001b779c76e3":"alice"}`, string(encoded))

	// A user-defined type arrives as a map of its fields.
	encoded, err = json.Marshal(JSONValue(map[string]interface{}{
		"street": "1 High Street",
		"zip":    12345,
	}))
	require.NoError(t, err)
	assert.JSONEq(t, `{"street":"1 High Street","zip":12345}`, string(encoded))
}
