package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/logger"
)

// ConvertDBConfigToAIConfig resolves the configuration down to the four
// things a request needs: which provider, its key, its model and its URL.
//
// A provider's own block in cqlai.json overrides the general settings above
// it, and what is still unset comes from the provider's row in the table.
func ConvertDBConfigToAIConfig(dbConfig *config.AIConfig) *AIConfig {
	if dbConfig == nil {
		// Mock, so that nothing built from this reaches a provider by
		// accident when there is no configuration at all.
		return &AIConfig{Provider: string(ProviderMock)}
	}

	settings := &AIConfig{
		Provider: dbConfig.Provider,
		APIKey:   dbConfig.APIKey,
		Model:    dbConfig.Model,
		URL:      dbConfig.URL,
	}
	if settings.Provider == "" {
		settings.Provider = string(ProviderMock)
	}

	p, known := providerNamed(Provider(settings.Provider))
	if !known {
		logger.DebugfToFile("AI", "No such provider: %s", settings.Provider)
		return settings
	}

	if block := p.settings(dbConfig); block != nil {
		if block.APIKey != "" {
			settings.APIKey = block.APIKey
		}
		if block.Model != "" {
			settings.Model = block.Model
		}
		if block.URL != "" {
			settings.URL = block.URL
		}
	}
	if settings.Model == "" {
		settings.Model = p.model
	}
	if settings.URL == "" {
		settings.URL = p.url
	}

	logger.DebugfToFile("AI", "AI settings: provider=%s, key=%v, model=%s, url=%s",
		settings.Provider, settings.APIKey != "", settings.Model, settings.URL)

	return settings
}

// AIConfig holds configuration for AI providers
type AIConfig struct {
	Provider string // "gemini", "openai", "anthropic", "ollama", "openrouter", "mock"
	APIKey   string
	Model    string // Optional model override
	URL      string // For providers that support custom URLs (Ollama, OpenRouter)
}

// extractJSON attempts to extract JSON from a text response
func extractJSON(text string) string {
	// First priority: Look for JSON between ```json and ``` markers
	startMarker := JSONStartMarker
	endMarker := JSONEndMarker
	startIdx := strings.Index(text, startMarker)
	if startIdx != -1 {
		startIdx += len(startMarker)
		endIdx := strings.Index(text[startIdx:], endMarker)
		if endIdx != -1 {
			candidate := strings.TrimSpace(text[startIdx : startIdx+endIdx])
			// Validate it's actual JSON
			if json.Valid([]byte(candidate)) {
				return candidate
			}
		}
	}

	// Second priority: Extract balanced JSON object
	startIdx = strings.Index(text, "{")
	if startIdx != -1 {
		// Find the matching closing brace by counting braces
		candidate := extractBalancedJSON(text[startIdx:])
		if candidate != "" && json.Valid([]byte(candidate)) {
			return candidate
		}
	}

	return ""
}

// extractBalancedJSON extracts a balanced JSON object from text starting with {
func extractBalancedJSON(text string) string {
	if len(text) == 0 || text[0] != '{' {
		return ""
	}

	depth := 0
	inString := false
	escaped := false

	for i, c := range text {
		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[:i+1]
			}
		}
	}

	return ""
}

// FormatPlanAsJSON returns a pretty-printed JSON representation of the plan
func FormatPlanAsJSON(plan *AIResult) string {
	b, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting plan: %v", err)
	}
	return string(b)
}
