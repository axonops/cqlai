package ai

import "strings"

// Provider represents an AI provider type
type Provider string

// AI Provider constants
const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGemini    Provider = "gemini"
	ProviderOllama    Provider = "ollama"
	ProviderMock      Provider = "mock"
)

// Default model names for each provider
const (
	DefaultOpenAIModel    = "gpt-4-turbo-preview"
	DefaultAnthropicModel = "claude-3-sonnet-20240229"
	DefaultGeminiModel    = "gemini-pro"
	DefaultOllamaModel    = "llama3"
)

// JSON parsing markers
const (
	JSONStartMarker = "```json"
	JSONEndMarker   = "```"
	CodeStartMarker = "```"
	CodeEndMarker   = "```"
)

// ToolName represents a type-safe tool name
type ToolName string

// String returns the string representation of the tool name
func (t ToolName) String() string {
	return string(t)
}

// ParseToolName converts a string to ToolName, returns empty string if invalid
func ParseToolName(s string) ToolName {
	// Normalize to lowercase
	s = strings.ToLower(strings.TrimSpace(s))
	
	// Check if it's a valid tool name
	toolName := ToolName(s)
	if toolName.IsValid() {
		return toolName
	}
	
	return ""
}

// IsValid checks if this is a valid tool name
func (t ToolName) IsValid() bool {
	switch t {
	case ToolFuzzySearch, ToolGetSchema, ToolGetTableInfo,
		ToolListKeyspaces, ToolListTables, ToolSubmitQueryPlan,
		ToolUserSelection, ToolNotEnoughInfo, ToolNotRelevant:
		return true
	}
	return false
}

// Tool names for AI function calling
const (
	ToolFuzzySearch     ToolName = "fuzzy_search"
	ToolGetSchema       ToolName = "get_schema"
	ToolGetTableInfo    ToolName = "get_table_info"
	ToolListKeyspaces   ToolName = "list_keyspaces"
	ToolListTables      ToolName = "list_tables"
	ToolSubmitQueryPlan ToolName = "submit_query_plan"
	ToolUserSelection   ToolName = "user_selection"
	ToolNotEnoughInfo   ToolName = "not_enough_info"
	ToolNotRelevant     ToolName = "not_relevant"
)

// Environment variable names
const (
	EnvOpenAIKey    = "OPENAI_API_KEY"
	EnvAnthropicKey = "ANTHROPIC_API_KEY"
	EnvGeminiKey    = "GEMINI_API_KEY"
	EnvOllamaHost   = "OLLAMA_HOST"
	EnvAIProvider   = "AI_PROVIDER"
	EnvAIModel      = "AI_MODEL"
)

// Error messages
const (
	ErrNoAPIKey          = "API key is required for %s"
	ErrUnsupportedMethod = "%s client does not support %s"
	ErrInvalidProvider   = "unsupported AI provider: %s"
	ErrJSONParsing       = "failed to parse JSON response"
	ErrNoToolCalls       = "no tool calls in response"
)

// System message prefixes
const (
	SystemMessagePrefix = "System: "
	UserMessagePrefix   = "User: "
	AssistantPrefix     = "Assistant: "
)
