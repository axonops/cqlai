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

	p, known := providerNamed(Provider(cfg.Provider))
	if !known {
		// No provider at all, or one nothing here knows how to talk to.
		return false
	}

	switch p.needs {
	case needsNothing:
		return true
	case needsURL:
		return providerURL(cfg, p.settings(cfg)) != ""
	case needsKey:
		return providerKey(cfg, p.settings(cfg)) != ""
	}
	return false
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
