package ai

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
)

// TestANameRememberedByItsSoundIsFound: vowels left out, letters left out, a
// letter wrong.
func TestANameRememberedByItsSoundIsFound(t *testing.T) {
	for _, c := range []struct{ query, name, why string }{
		{"hyto", "hayato", matchSoundsLike},
		{"hyto", "hayato2", matchSoundsLike},
		{"ordrs", "orders", matchSoundsLike},
		{"cstmr", "customers", matchSoundsLike},
		{"hayto", "hayato", matchSoundsLike},
		{"oders", "orders", matchLetters},
		{"ivces", "invoices", matchLetters},
		{"ordefs", "orders", matchSpelling},
		{"invoicse", "invoices", matchSoundsLike},
		{"orders", "orders", matchExact},
		{"ord", "orders", matchPrefix},
		{"rder", "orders", matchContains},
		{"HYTO", "hayato", matchSoundsLike},
	} {
		score, why := nameMatch(c.query, c.name)
		assert.Greater(t, score, 0.0, "%s should find %s", c.query, c.name)
		assert.Equal(t, c.why, why, "%s and %s", c.query, c.name)
	}
}

// TestUnlikeNamesAreNotFound, and a search too short to sound like anything
// finds only what it spells.
func TestUnlikeNamesAreNotFound(t *testing.T) {
	for _, c := range []struct{ query, name string }{
		{"hyto", "keyspace_1"},
		{"hyto", "metrics_index"},
		{"hyto", "test"},
		{"hyto", "heat"}, // with y as a vowel, both would be "ht"
		{"ordrs", "coordinator_scan_latency"}, // its letters are in there, among nineteen others
		{"orders", "customers"},
		{"ht", "hat"},
		{"xy", "hayato"},
	} {
		score, _ := nameMatch(c.query, c.name)
		assert.Zero(t, score, "%s should not find %s", c.query, c.name)
	}
}

// TestSpelledMatchesRankFirst: a near-miss never ranks above a name spelled
// like the search.
func TestSpelledMatchesRankFirst(t *testing.T) {
	contains, _ := nameMatch("order", "order_items")
	sounds, _ := nameMatch("order", "ordr")
	letters, _ := nameMatch("order", "o_r_d_e_r_s")
	assert.Greater(t, contains, sounds)
	assert.Greater(t, sounds, letters)
}

// TestFuzzySearchFindsKeyspacesByTheirSound, the question that found the gap:
// keyspaces that sound like "hyto".
func TestFuzzySearchFindsKeyspacesByTheirSound(t *testing.T) {
	cache := &db.SchemaCache{
		Keyspaces: []string{"hayato", "hayato2", "keyspace_1", "metrics_index", "secret_hayato"},
		Tables: map[string][]db.CachedTableInfo{
			"hayato": {{TableInfo: db.TableInfo{KeyspaceName: "hayato", TableName: "orders"}}},
		},
		Columns:     map[string]map[string][]db.ColumnInfo{},
		LastRefresh: time.Now(),
	}
	a := &AI{cache: cache, resolver: NewResolver(cache)}
	p := mcpPolicy(t, &config.MCPConfig{Deny: []string{"secret_hayato"}})

	result := executeCommandFor(p, a, ToolFuzzySearch, "hyto")
	if !assert.NoError(t, result.Error) {
		return
	}
	assert.Contains(t, result.Data, "Keyspaces matching 'hyto'")
	assert.Contains(t, result.Data, "- hayato (sounds like it)")
	assert.Contains(t, result.Data, "- hayato2 (sounds like it)")
	assert.NotContains(t, result.Data, "keyspace_1")
	assert.NotContains(t, result.Data, "secret_hayato", "a hidden keyspace is not found by its sound either")

	result = executeCommandFor(p, a, ToolFuzzySearch, "ordrs")
	assert.Contains(t, result.Data, "hayato.orders (sounds like it")
	assert.True(t, strings.HasPrefix(result.Data, "Found 1 tables"), result.Data)
}
