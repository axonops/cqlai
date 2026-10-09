package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAQuoteInAFunctionNameEndsNothing: DESCRIBE FUNCTION puts the name typed
// into the lookup as a CQL string, so a quote in it is doubled rather than
// closing the string and leaving the rest to be read as CQL.
func TestAQuoteInAFunctionNameEndsNothing(t *testing.T) {
	assert.Equal(t,
		"SELECT * FROM system_schema.functions WHERE function_name = 'f'' ALLOW FILTERING'",
		functionLookup("f' ALLOW FILTERING", ""))
	assert.Equal(t,
		"SELECT * FROM system_schema.functions WHERE keyspace_name = 'shop' AND function_name = 'o''brien'",
		functionLookup("o'brien", "shop"))
}
