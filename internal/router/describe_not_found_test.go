package router

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNotFound takes an error as well as text. describeIdentifier asserted text
// of what describeTable gave it, which is an error for a table that is not
// there, and DESCRIBE k.nosuch panicked.
func TestNotFound(t *testing.T) {
	assert.True(t, notFound(errors.New("table 'nosuch' not found in keyspace 'k'")))
	assert.True(t, notFound("keyspace 'x' not found"))
	assert.False(t, notFound("CREATE TABLE k.t (id int PRIMARY KEY);"))
	assert.False(t, notFound(errors.New("no keyspace selected")))
	assert.False(t, notFound(nil))
}
