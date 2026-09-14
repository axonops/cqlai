package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// What completion offers for the parts of a statement that are lists rather
// than structure. The structure - where in a statement each list belongs - is
// the other half, and is not what these are about.

// TestTheTypesIncludeTheOnesAddedSince.
//
// vector is the point of vector<float, 3>, and the word was not there.
func TestTheTypesIncludeTheOnesAddedSince(t *testing.T) {
	assert.Contains(t, CQLDataTypes, "vector")
	assert.Contains(t, CQLDataTypes, "duration")
}

// TestTheOperatorsIncludeTheNegatedForms and BETWEEN, which arrived in 5.0.
func TestTheOperatorsIncludeTheNegatedForms(t *testing.T) {
	for _, wanted := range []string{"BETWEEN", "NOT IN", "NOT CONTAINS", "NOT CONTAINS KEY"} {
		assert.Contains(t, ComparisonOperators, wanted)
	}
}

// TestEveryFunctionFamilyIsOffered.
//
// Five aggregates and some time functions were offered; Cassandra has scalar
// maths, collection functions and masking as well.
func TestEveryFunctionFamilyIsOffered(t *testing.T) {
	ce := &CompletionEngine{}
	offered := map[string]bool{}
	for _, suggestion := range ce.getFunctionSuggestions() {
		offered[suggestion] = true
	}

	for _, wanted := range []string{
		"COUNT(", "AVG(", // aggregates
		"ABS(", "ROUND(", // scalar maths
		"MAP_KEYS(", "COLLECTION_SUM(", // collections
		"MASK_HASH(", "MASK_INNER(", // masking
		"TO_UNIX_TIMESTAMP(", "MAX_WRITETIME(", // conversions, and what a cell carries
		"CAST(", "token(",
	} {
		assert.True(t, offered[wanted], "%s should be offered", wanted)
	}

	// The ones that take nothing are written with both brackets, so pressing
	// Tab finishes them.
	for _, complete := range []string{"now()", "uuid()", "CURRENT_TIMESTAMP()"} {
		assert.True(t, offered[complete], "%s should be offered as it is written", complete)
	}
}

// TestNoFunctionIsOfferedTwice, since the families overlap in what they are
// about but not in what they contain.
func TestNoFunctionIsOfferedTwice(t *testing.T) {
	ce := &CompletionEngine{}
	seen := map[string]bool{}

	for _, suggestion := range ce.getFunctionSuggestions() {
		assert.False(t, seen[suggestion], "%s is offered twice", suggestion)
		seen[suggestion] = true
	}
}
