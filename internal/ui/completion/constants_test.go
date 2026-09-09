package completion

import (
	"testing"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
)

// TestCompletionOffersOnlyLevelsThatCanBeSet: Tab suggested SERIAL and
// LOCAL_SERIAL while CONSISTENCY rejected them, so accepting the suggestion
// printed an error. Both completion paths now read the same list as the code
// that applies a level.
func TestCompletionOffersOnlyLevelsThatCanBeSet(t *testing.T) {
	assert.Equal(t, db.ConsistencyLevels(), ConsistencyLevels)

	sce := &SimpleCompletionEngine{}
	assert.Equal(t, db.ConsistencyLevels(), sce.getConsistencyLevels())

	for _, level := range ConsistencyLevels {
		assert.NoError(t, (&db.Session{}).SetConsistency(level),
			"%s is completed but cannot be set", level)
	}
}
