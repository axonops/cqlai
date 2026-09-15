package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTheChatShowsWhatClaudeThought.
//
// The models think before they answer. Without this the answer arrives after a
// pause with nothing to show for it, and the thinking - which is often the
// part worth reading - is thrown away.
func TestTheChatShowsWhatClaudeThought(t *testing.T) {
	m := helpModel()
	m.aiConversationMessages = []AIMessage{
		{Role: "user", Content: "what are the available keyspaces?"},
		{Role: "assistant", Content: "The keyspaces are in system_schema.keyspaces.", Type: "reasoning"},
		{Role: "assistant", Content: "There are three: shop, trial and system."},
	}

	m.rebuildAIConversation()
	drawn := stripAnsi(m.aiConversationHistory)

	assert.Contains(t, drawn, "Thinking:")
	assert.Contains(t, drawn, "system_schema.keyspaces")
	assert.Contains(t, drawn, "There are three")

	// The reasoning is not labelled as the answer.
	thinking := indexOfLine(drawn, "Thinking:")
	answer := indexOfLine(drawn, "Assistant:")
	assert.Less(t, thinking, answer, "the thinking comes before the answer it led to")
}

// indexOfLine is where a line holding this text is, or -1.
func indexOfLine(text, wanted string) int {
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, wanted) {
			return i
		}
	}
	return -1
}
