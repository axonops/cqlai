package batch

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/session"
)

// offline is an executor with no cluster: what it runs is what the shell
// decides for itself.
func offline() (*Executor, *bytes.Buffer) {
	var out bytes.Buffer
	return &Executor{
		session:        &db.Session{},
		sessionManager: session.NewManager(&config.Config{}),
		options:        &Options{},
		writer:         &out,
	}, &out
}

// TestARejectedStatementIsAnError: batch mode exits 0 on what it returns nil
// for, so a statement the shell refuses has to come back as an error.
func TestARejectedStatementIsAnError(t *testing.T) {
	e, _ := offline()
	for _, stmt := range []string{
		"SELEC * FROM ks.t;",
		"CONSISTENCY BOGUS;",
		"OUTPUT NOSUCHFORMAT;",
	} {
		assert.Error(t, e.Execute(stmt), stmt)
	}
	assert.NoError(t, e.Execute("EXPAND ON;"))
}

// TestEachStatementLetsGoOfCtrlC: the Ctrl+C handler a statement sets up is
// gone when it is done. Left behind, one per statement, they piled up, and a
// signal after the first statement stopped nothing.
func TestEachStatementLetsGoOfCtrlC(t *testing.T) {
	e, _ := offline()
	require.NoError(t, e.Execute("EXPAND ON;"))
	time.Sleep(50 * time.Millisecond)
	before := runtime.NumGoroutine()

	for i := 0; i < 50; i++ {
		_ = e.Execute("EXPAND ON;")
	}
	time.Sleep(50 * time.Millisecond)
	assert.LessOrEqual(t, runtime.NumGoroutine(), before+2, "a goroutine left behind per statement")
}
