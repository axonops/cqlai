package completion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/cqlai/internal/logger"
)

// TestCompletionDoesNotLogWhenDebugDisabled guards against the completion engine
// writing a debug log on every keystroke regardless of the --debug flag, which
// used to grow an unbounded cqlai_debug.log in the working directory.
func TestCompletionDoesNotLogWhenDebugDisabled(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "cqlai_debug.log")
	t.Setenv("CQLAI_DEBUG_LOG_PATH", logPath)

	logger.SetDebugEnabled(false)
	t.Cleanup(func() { logger.SetDebugEnabled(false) })

	ce := NewCompletionEngine(nil, nil)
	ce.Complete("SEL")
	ce.Complete("SELECT * FROM ")
	ce.Complete("")

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("completion wrote %s with debug disabled (stat err: %v)", logPath, err)
	}
}

// TestCompletionLogsWhenDebugEnabled is the other half: the debug output must
// still be there when it is asked for.
func TestCompletionLogsWhenDebugEnabled(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "cqlai_debug.log")
	t.Setenv("CQLAI_DEBUG_LOG_PATH", logPath)

	logger.SetDebugEnabled(true)
	t.Cleanup(func() { logger.SetDebugEnabled(false) })

	ce := NewCompletionEngine(nil, nil)
	ce.Complete("SEL")

	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("completion wrote no debug log with debug enabled: %v", err)
	}
}
