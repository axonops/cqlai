package ai

import (
	"context"

	"github.com/axonops/cqlai/internal/config"
)

// The providers cqlai can be pointed at, and everything that differs between
// them, in one table.
//
// It was in six places: the list PREFERENCES offers, the switch that reads a
// provider's block out of cqlai.json, the switch that decides whether one is
// configured, the switch that builds its client, the switch that continues a
// conversation with it, and the switch that asks it to explain something. A
// provider added to some of those and not the others is a provider that is
// offered and then fails, and nothing in the build says so.
//
// Add a provider here and every one of those follows. Add it anywhere else and
// TestEveryProviderIsReachable fails.

// api is the shape of the requests a provider takes.
type api int

const (
	// noAPI is a provider nothing here knows how to talk to. It can still be
	// named in cqlai.json and still have a block of its own - it just has
	// nowhere to send a question.
	noAPI api = iota

	// anthropicAPI is Anthropic's own API, through its SDK.
	anthropicAPI

	// openAIAPI is the OpenAI API, which most providers speak. They differ
	// only in the base URL.
	//
	// OpenAI, OpenRouter, Gemini and Ollama all answer chat completions with
	// tool calls at /chat/completions under their own base URL, so they are
	// one conversation with four addresses.
	openAIAPI
)

// needs is what has to be set before a provider can answer anything.
type needs int

const (
	needsKey     needs = iota // an API key
	needsURL                  // somewhere to reach it, and no key
	needsNothing              // neither, because it does not leave the process
)

// provider is one row of the table.
type provider struct {
	name Provider

	// section is the provider's own block in cqlai.json, which overrides the
	// general apiKey/model/url above it. nil where a provider has no block.
	section func(*config.AIConfig) *config.AIProviderConfig

	// field is the same block named as PREFERENCES addresses it: the field on
	// config.AIConfig, which the window turns into a path like "AI.OpenAI".
	// Empty wherever section is nil.
	field string

	model string // the model used when none is configured
	url   string // the base URL used when none is configured

	speaks api
	needs  needs

	// continues carries a conversation with this provider forward a round.
	// nil wherever speaks is noAPI.
	continues func(*AIConversation, context.Context, string) (*AIResult, *InteractionRequest, error)
}

// providerTable is the table, in the order PREFERENCES offers them.
//
// A function rather than a variable: the rows name the functions that continue
// a conversation, those reach Continue, and Continue reads the table back.
// Between variables that is an initialisation cycle; between functions it is
// just a lookup.
func providerTable() []provider {
	return []provider{
		{
			name:      ProviderOpenAI,
			section:   func(c *config.AIConfig) *config.AIProviderConfig { return c.OpenAI },
			field:     "OpenAI",
			model:     DefaultOpenAIModel,
			url:       openAiBaseURL,
			speaks:    openAIAPI,
			needs:     needsKey,
			continues: (*AIConversation).continueOpenAI,
		},
		{
			name:    ProviderAnthropic,
			section: func(c *config.AIConfig) *config.AIProviderConfig { return c.Anthropic },
			field:   "Anthropic",
			model:   DefaultAnthropicModel,
			// No URL: the Anthropic SDK knows where Anthropic is.
			speaks:    anthropicAPI,
			needs:     needsKey,
			continues: (*AIConversation).continueAnthropic,
		},
		{
			// Gemini had a block in cqlai.json, an environment variable, a
			// section in PREFERENCES and four sections of the README, and
			// nothing that could talk to it: choosing it failed on the first
			// question. Its OpenAI-compatible endpoint answers the same chat
			// completions with the same tool calls as the rest, so being
			// reachable costs it a URL.
			name:      ProviderGemini,
			section:   func(c *config.AIConfig) *config.AIProviderConfig { return c.Gemini },
			field:     "Gemini",
			model:     DefaultGeminiModel,
			url:       geminiBaseURL,
			speaks:    openAIAPI,
			needs:     needsKey,
			continues: (*AIConversation).continueOpenAI,
		},
		{
			name:      ProviderOllama,
			section:   func(c *config.AIConfig) *config.AIProviderConfig { return c.Ollama },
			field:     "Ollama",
			model:     DefaultOllamaModel,
			url:       ollamaBaseURL,
			speaks:    openAIAPI,
			needs:     needsURL, // it runs locally and takes no key
			continues: (*AIConversation).continueOpenAI,
		},
		{
			name:      ProviderOpenRouter,
			section:   func(c *config.AIConfig) *config.AIProviderConfig { return c.OpenRouter },
			field:     "OpenRouter",
			model:     DefaultOpenRouterModel,
			url:       openRouterBaseURL,
			speaks:    openAIAPI,
			needs:     needsKey,
			continues: (*AIConversation).continueOpenAI,
		},
		{
			// Mock answers nothing and needs nothing. It is how the AI
			// features are turned off on purpose, as opposed to left
			// unconfigured.
			name:   ProviderMock,
			speaks: noAPI,
			needs:  needsNothing,
		},
	}
}

// providerNamed finds a provider by the name in the configuration.
func providerNamed(name Provider) (provider, bool) {
	for _, p := range providerTable() {
		if p.name == name {
			return p, true
		}
	}
	return provider{}, false
}

// settings is a provider's own section of cqlai.json, or nil where it has none.
func (p provider) settings(cfg *config.AIConfig) *config.AIProviderConfig {
	if p.section == nil || cfg == nil {
		return nil
	}
	return p.section(cfg)
}

// Providers is every provider cqlai can be pointed at, in the order the
// PREFERENCES window offers them.
func Providers() []string {
	table := providerTable()
	names := make([]string, 0, len(table))
	for _, p := range table {
		names = append(names, string(p.name))
	}
	return names
}

// ProviderBlock is one provider's own block of settings in cqlai.json.
type ProviderBlock struct {
	// Name is the provider, as it is written in the configuration.
	Name string

	// Field is the block on config.AIConfig that holds its settings.
	Field string
}

// ProviderBlocks is the block each provider keeps its own key, model and URL
// in, for the PREFERENCES window to offer one section per provider.
//
// Providers without a block of their own - mock - are not here.
func ProviderBlocks() []ProviderBlock {
	var blocks []ProviderBlock
	for _, p := range providerTable() {
		if p.field == "" {
			continue
		}
		blocks = append(blocks, ProviderBlock{Name: string(p.name), Field: p.field})
	}
	return blocks
}
