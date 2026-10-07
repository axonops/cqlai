package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/router"
	"github.com/axonops/cqlai/internal/session"
)

// TestDescribeLeavesTheShellsHandlerAlone: the MCP server's describe does
// not build the router's meta-command handler. Built from here, it was bound
// to the server's session, and the shell's TRACING ON turned on tracing for
// the model rather than the user.
func TestDescribeLeavesTheShellsHandlerAlone(t *testing.T) {
	require.Nil(t, router.GetMetaHandler(), "nothing in this test binary has built it")

	describe := Describer(nil, session.NewManager(&config.Config{}))
	_ = describe("DESCRIBE")

	assert.Nil(t, router.GetMetaHandler(), "and describe did not build it")
}
