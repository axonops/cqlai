package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseName reads a name as CQL does. The shell trimmed quotes off the ends
// instead, so "MyKeyspace"."TableA" became MyKeyspace"."TableA.
func TestParseName(t *testing.T) {
	for typed, want := range map[string]Name{
		`users`:                  {Table: "users"},
		`Users`:                  {Table: "users"},
		`"Users"`:                {Table: "Users"},
		`k."_staging"`:           {Keyspace: "k", Table: "_staging", Qualified: true},
		`"MyKeyspace"."TableA"`:  {Keyspace: "MyKeyspace", Table: "TableA", Qualified: true},
		`MyKeyspace.TableA`:      {Keyspace: "mykeyspace", Table: "tablea", Qualified: true},
		`"MyKeyspace"."TableA";`: {Keyspace: "MyKeyspace", Table: "TableA", Qualified: true},
		`"a""b"`:                 {Table: `a"b`},
	} {
		got, err := ParseName(typed)
		require.NoError(t, err, typed)
		assert.Equal(t, want, got, typed)
	}

	for _, typed := range []string{``, `a.`, `a b`, `'users'`, `"open`} {
		_, err := ParseName(typed)
		assert.Error(t, err, typed)
	}
}

// TestNameString is the stored name, without quotes.
func TestNameString(t *testing.T) {
	assert.Equal(t, "MyKeyspace.TableA", Name{Keyspace: "MyKeyspace", Table: "TableA", Qualified: true}.String())
	assert.Equal(t, "users", Name{Table: "users"}.String())
}
