package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTheTableIsTheOneTheQueryReads, whatever comes before FROM or around it.
func TestTheTableIsTheOneTheQueryReads(t *testing.T) {
	for query, want := range map[string][2]string{
		"SELECT * FROM ks.users":                {"ks", "users"},
		"   SELECT * FROM ks.users":             {"ks", "users"},
		"SELECT * FROM users WHERE id = 1":      {"", "users"},
		"SELECT fromage FROM ks.cheese":         {"ks", "cheese"},
		"SELECT * FROM\nks.users":               {"ks", "users"},
		"SELECT * FROM \"MyKs\".\"MyTable\"":    {"MyKs", "MyTable"},
		"SELECT * FROM KS.Users":                {"ks", "users"},
		"SELECT 'FROM x' AS s FROM ks.t":        {"ks", "t"},
		"/* FROM other.t */ SELECT * FROM ks.t": {"ks", "t"},
		"INSERT INTO ks.t (id) VALUES (1)":      {"", ""},
		"not cql at all":                        {"", ""},
	} {
		ks, table := extractTableName(query)
		assert.Equal(t, want, [2]string{ks, table}, query)
	}
}

// TestALetterThatGrowsInUpperCaseDoesNotCrash: "ɐ" is two bytes, and three
// in upper case, which moved the cut past the end of the query.
func TestALetterThatGrowsInUpperCaseDoesNotCrash(t *testing.T) {
	assert.NotPanics(t, func() { extractTableName("SELECT ɐɐɐɐɐɐɐɐɐɐ FROM t") })
	assert.NotPanics(t, func() { extractTableName("SELECT 'ɐɐɐɐɐɐɐɐɐɐ' FROM t") })
}
