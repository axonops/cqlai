package ai

import (
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNothingConfiguredIsNotConfigured(t *testing.T) {
	assert.False(t, IsConfigured(nil))
	assert.False(t, IsConfigured(&config.AIConfig{}))
}

// TestAProviderWithoutItsKeyIsNotConfigured. Offering the AI view and then
// failing on the first question is worse than not offering it.
func TestAProviderWithoutItsKeyIsNotConfigured(t *testing.T) {
	for _, provider := range []string{"openai", "anthropic", "gemini", "openrouter"} {
		assert.False(t, IsConfigured(&config.AIConfig{Provider: provider}),
			"%s has no key", provider)
	}
}

func TestAProviderWithItsKeyIsConfigured(t *testing.T) {
	for _, provider := range []string{"openai", "anthropic", "gemini", "openrouter"} {
		assert.True(t, IsConfigured(&config.AIConfig{Provider: provider, APIKey: "k"}),
			"%s with the general key", provider)
	}

	// A provider's own key counts as well as the general one.
	assert.True(t, IsConfigured(&config.AIConfig{
		Provider:  "anthropic",
		Anthropic: &config.AIProviderConfig{APIKey: "k"},
	}))
}

// TestAKeyForTheWrongProviderDoesNotCount.
func TestAKeyForTheWrongProviderDoesNotCount(t *testing.T) {
	assert.False(t, IsConfigured(&config.AIConfig{
		Provider: "anthropic",
		OpenAI:   &config.AIProviderConfig{APIKey: "k"},
	}))
}

// TestOllamaNeedsAURLNotAKey: it runs locally and takes no key, so a key is
// not what tells you it is set up.
func TestOllamaNeedsAURLNotAKey(t *testing.T) {
	assert.False(t, IsConfigured(&config.AIConfig{Provider: "ollama"}))
	assert.False(t, IsConfigured(&config.AIConfig{Provider: "ollama", APIKey: "k"}))

	assert.True(t, IsConfigured(&config.AIConfig{Provider: "ollama", URL: "http://localhost:11434"}))
	assert.True(t, IsConfigured(&config.AIConfig{
		Provider: "ollama",
		Ollama:   &config.AIProviderConfig{URL: "http://localhost:11434"},
	}))
}

// TestMockCountsOnlyWhenItWasAskedFor.
//
// NewClient falls back to mock when no provider is set, which is a sensible
// default for a library but means asking it for a client cannot tell the two
// apart. Set on purpose, mock is a choice; arrived at, it is not.
func TestMockCountsOnlyWhenItWasAskedFor(t *testing.T) {
	assert.True(t, IsConfigured(&config.AIConfig{Provider: "mock"}))
	assert.False(t, IsConfigured(&config.AIConfig{Provider: ""}))
}

func TestAnUnknownProviderIsNotConfigured(t *testing.T) {
	assert.False(t, IsConfigured(&config.AIConfig{Provider: "telepathy", APIKey: "k"}))
}
