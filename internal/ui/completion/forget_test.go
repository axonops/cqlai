package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestForgetDropsEveryNameRead: the names were read once and kept, so after a
// DROP, Tab went on offering what was gone.
func TestForgetDropsEveryNameRead(t *testing.T) {
	ce := NewCompletionEngine(nil, nil)
	ce.cache.keyspaces = []string{"coc"}
	ce.cache.tables["coc"] = []string{"talks"}
	ce.cache.columns["coc.talks"] = []string{"id"}

	ce.Forget()

	assert.Empty(t, ce.getKeyspaceNames())
	assert.Empty(t, ce.getTablesForKeyspace("coc"))
	assert.Empty(t, ce.cache.columns)
}
