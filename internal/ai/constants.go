package ai

import "strings"

// Provider represents an AI provider type
type Provider string

// AI Provider constants
const (
	ProviderOpenAI     Provider = "openai"
	ProviderAnthropic  Provider = "anthropic"
	ProviderGemini     Provider = "gemini"
	ProviderOllama     Provider = "ollama"
	ProviderOpenRouter Provider = "openrouter"
	ProviderMock       Provider = "mock"
)

// anthropicMaxTokens is the room an answer is given.
//
// It was 1024, which is a page: a plan with a long CQL statement in it, or an
// answer about a schema, was cut off partway through and came back as a reply
// that would not parse.
const anthropicMaxTokens = 16000

// Default model names for each provider
const (
	DefaultOpenAIModel = "gpt-4-turbo-preview"

	// DefaultAnthropicModel: claude-3-sonnet-20240229 was the default, and
	// that model has been retired - the provider answered every question with
	// a 404 until a model was named in the file.
	DefaultAnthropicModel = "claude-opus-5"

	// DefaultGeminiModel: gemini-pro was the default and has been retired,
	// the same way claude-3-sonnet was.
	DefaultGeminiModel = "gemini-3.8-flash"

	DefaultOllamaModel     = "llama3"
	DefaultOpenRouterModel = "openai/gpt-4-turbo-preview"
)

// JSON parsing markers
const (
	JSONStartMarker = "```json"
	JSONEndMarker   = "```"
	CodeStartMarker = "```"
	CodeEndMarker   = "```"
)

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

// IsValid reports whether a tool of this name is defined. It reads the
// definitions rather than a list of its own, so a tool cannot be defined and
// then refused as unknown, or the other way round.
func (t ToolName) IsValid() bool {
	_, ok := toolDefinition(t)
	return ok
}

// Tool names for AI function calling
const (
	ToolFuzzySearch     ToolName = "fuzzy_search"
	ToolGetSchema       ToolName = "get_schema"
	ToolListKeyspaces   ToolName = "list_keyspaces"
	ToolListTables      ToolName = "list_tables"
	ToolSubmitQueryPlan ToolName = "submit_query_plan"
	ToolUserSelection   ToolName = "user_selection"
	ToolNotEnoughInfo   ToolName = "not_enough_info"
	ToolNotRelevant     ToolName = "not_relevant"
	ToolInfo            ToolName = "info" // For informational responses

	// Offered to MCP clients only.
	ToolConnectionInfo ToolName = "connection_info"
	ToolDescribe       ToolName = "describe"
	ToolQuery          ToolName = "query"
	ToolTraceQuery     ToolName = "trace_query"
)

// Environment variable names
const (
	EnvOpenAIKey     = "OPENAI_API_KEY"
	EnvAnthropicKey  = "ANTHROPIC_API_KEY"
	EnvGeminiKey     = "GEMINI_API_KEY"
	EnvOllamaHost    = "OLLAMA_HOST"
	EnvOpenRouterKey = "OPENROUTER_API_KEY"
	EnvAIProvider    = "AI_PROVIDER"
	EnvAIModel       = "AI_MODEL"
)

// ErrInvalidProvider is what a provider that cannot be talked to is told.
const ErrInvalidProvider = "unsupported AI provider: %s"

// System message prefixes
const (
	SystemMessagePrefix = "System: "
	UserMessagePrefix   = "User: "
	AssistantPrefix     = "Assistant: "
)
