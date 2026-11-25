package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowCompletesUser(t *testing.T) {
	sce := &SimpleCompletionEngine{}
	assert.Contains(t, sce.getShowCompletions([]string{"SHOW"}, true), "USER")
	assert.Equal(t, []string{"USER"}, sce.getShowCompletions([]string{"SHOW", "us"}, false))
}
