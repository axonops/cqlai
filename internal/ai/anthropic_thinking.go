package ai

import (
	"context"
	"strings"
	"sync"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/axonops/cqlai/internal/logger"
)

// Asking Claude to show its reasoning.
//
// The current models think before they answer whether or not they are asked
// to, and by default the thinking comes back empty - the answer appears after
// a pause with nothing to show for it. Asking for it summarized is what fills
// the CHAT tab in while it works.
//
// Not every model takes the parameter: the ones before adaptive thinking
// refuse it outright. Which ones those are is Anthropic's list rather than
// ours, and a copy of it here would go stale the way every other copied list
// in this project has. The refusal is the answer instead - a model that says
// no is asked once, and not asked again.

// thinkingRefused is the models that will not be asked again.
var thinkingRefused sync.Map

// askAnthropic sends a request, asking for the reasoning where the model takes
// it and going again without it where it does not.
func askAnthropic(ctx context.Context, client *anthropic.Client, params anthropic.MessageNewParams) (*anthropic.Message, error) {
	model := string(params.Model)
	if _, refused := thinkingRefused.Load(model); refused {
		return client.Messages.New(ctx, params)
	}

	thinking := params
	thinking.Thinking = anthropic.ThinkingConfigParamUnion{
		OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{
			Display: anthropic.ThinkingConfigAdaptiveDisplaySummarized,
		},
	}

	response, err := client.Messages.New(ctx, thinking)
	if err == nil || !refusesThinking(err) {
		return response, err
	}

	logger.DebugfToFile("Anthropic", "%s does not take thinking, asking without it: %v", model, err)
	thinkingRefused.Store(model, true)
	return client.Messages.New(ctx, params)
}

// refusesThinking reports whether the request was turned down for asking for
// the reasoning, rather than for any of the other reasons a request is turned
// down - a bad key, a model that is not there, a rate limit.
func refusesThinking(err error) bool {
	said := strings.ToLower(err.Error())
	return strings.Contains(said, "thinking") &&
		(strings.Contains(said, "invalid_request") || strings.Contains(said, "400"))
}

// reasoningOf is the reasoning Claude showed, gathered from the response.
func reasoningOf(response *anthropic.Message) string {
	var shown []string
	for _, content := range response.Content {
		if content.Type == "thinking" && strings.TrimSpace(content.Thinking) != "" {
			shown = append(shown, strings.TrimSpace(content.Thinking))
		}
	}
	return strings.Join(shown, "\n\n")
}
