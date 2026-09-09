package ai

import "github.com/axonops/cqlai/internal/config"

// IsConfigured reports whether there is an AI provider that can actually answer.
//
// Without one, NewClient falls back to the mock provider, which returns canned
// replies. That is the right default for a library - it never fails to build a
// client - but it means the UI cannot tell "AI is set up" from "AI is not set
// up" by asking for a client, and so offered an AI view backed by a stub.
//
// The mock provider counts only when it was asked for by name. Set on purpose
// it is a choice; arrived at because nothing was configured it is not.
//
// Keys reach the config from cqlai.json or from the environment - config
// merges both before this sees it - so there is nothing to read here beyond
// what has already been loaded.
func IsConfigured(cfg *config.AIConfig) bool {
	if cfg == nil {
		return false
	}

	switch Provider(cfg.Provider) {
	case ProviderMock:
		return true

	case ProviderOllama:
		// Ollama runs locally and takes no key, so a URL is what it needs.
		return providerURL(cfg, cfg.Ollama) != ""

	case ProviderOpenAI:
		return providerKey(cfg, cfg.OpenAI) != ""
	case ProviderAnthropic:
		return providerKey(cfg, cfg.Anthropic) != ""
	case ProviderGemini:
		return providerKey(cfg, cfg.Gemini) != ""
	case ProviderOpenRouter:
		return providerKey(cfg, cfg.OpenRouter) != ""

	default:
		// No provider at all, or one nothing here knows how to talk to.
		return false
	}
}

// providerKey is the key for a provider: its own, or the general one it falls
// back to. The same order NewClient resolves them in.
func providerKey(cfg *config.AIConfig, provider *config.AIProviderConfig) string {
	if provider != nil && provider.APIKey != "" {
		return provider.APIKey
	}
	return cfg.APIKey
}

// providerURL is the URL for a provider, resolved the same way.
func providerURL(cfg *config.AIConfig, provider *config.AIProviderConfig) string {
	if provider != nil && provider.URL != "" {
		return provider.URL
	}
	return cfg.URL
}
