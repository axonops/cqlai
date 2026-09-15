package ai

import (
	"errors"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
)

// The request cqlai sends Anthropic.
//
// It sent `temperature`, which the current models refuse: the chat answered
// every question with a 400 saying the parameter is deprecated. It asked for
// a model that has been retired. And it gave the answer a page to fit in.

// TestTheModelIsOneThatStillExists.
//
// claude-3-sonnet-20240229 was the default, and that model is retired.
func TestTheModelIsOneThatStillExists(t *testing.T) {
	assert.Equal(t, "claude-opus-5", DefaultAnthropicModel)
	assert.NotContains(t, DefaultAnthropicModel, "claude-3")
}

// TestAnAnswerHasRoomToBeWritten: a plan with a long CQL statement in it does
// not fit in a page, and what comes back cut off is a reply that will not
// parse.
func TestAnAnswerHasRoomToBeWritten(t *testing.T) {
	assert.GreaterOrEqual(t, anthropicMaxTokens, 8000)
}

// TestAModelIsAskedForItsReasoningOnce.
//
// The models that came before adaptive thinking refuse the parameter. Which
// ones those are is Anthropic's list rather than ours, so the refusal is the
// answer: a model that says no is asked once, and not asked again.
func TestAModelIsAskedForItsReasoningOnce(t *testing.T) {
	thinkingRefused.Delete("an-older-model")
	t.Cleanup(func() { thinkingRefused.Delete("an-older-model") })

	assert.True(t, refusesThinking(errors.New(
		`POST "https://api.anthropic.com/v1/messages": 400 Bad Request {"type":"error","error":`+
			`{"type":"invalid_request_error","message":"thinking is not supported for this model"}}`)))

	// Everything else a request is turned down for is not this.
	assert.False(t, refusesThinking(errors.New("401 Unauthorized: invalid x-api-key")))
	assert.False(t, refusesThinking(errors.New("404 Not Found: model not found")))
	assert.False(t, refusesThinking(errors.New("429 Too Many Requests")))
}

// TestTheReasoningIsWhatClaudeThought, gathered from the response and nothing
// else in it.
func TestTheReasoningIsWhatClaudeThought(t *testing.T) {
	response := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "thinking", Thinking: "  The keyspaces are in system_schema.  "},
			{Type: "text", Text: "Here they are."},
			{Type: "thinking", Thinking: "And then the answer."},
			{Type: "thinking", Thinking: "   "},
		},
	}

	assert.Equal(t, "The keyspaces are in system_schema.\n\nAnd then the answer.", reasoningOf(response))
	assert.Empty(t, reasoningOf(&anthropic.Message{Content: []anthropic.ContentBlockUnion{
		{Type: "text", Text: "Here they are."},
	}}))
}
