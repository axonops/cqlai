package router

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SHOW USER is answered by cqlai, not sent to Cassandra, which has no such
// statement.
func TestShowUserIsAShellCommand(t *testing.T) {
	p := &CommandParser{}
	for _, command := range []string{"SHOW USER", "show user", "SHOW USER;"} {
		assert.True(t, p.isMetaShowCommand(strings.ToUpper(command)), command)
	}
}
