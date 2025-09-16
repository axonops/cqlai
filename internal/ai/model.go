package ai

import (
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/axonops/cqlai/internal/db"
	"github.com/openai/openai-go"
)

// ToolName represents a type-safe tool name
type ToolName string

// InteractionRequest represents a request for user interaction
type InteractionRequest struct {
	Type             string   // "selection" or "info"
	SelectionType    string   // For selection: type of item to select
	SelectionOptions []string // For selection: available options
	InfoMessage      string   // For info: message to show user
	ConversationID   string   // ID of the conversation to resume
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	// Success case - command executed and returned data
	Success bool
	Data    string

	// User interaction needed
	NeedsUserSelection bool
	SelectionType      string
	SelectionOptions   []string

	NeedsMoreInfo bool
	InfoMessage   string

	NotRelevant bool

	// Query plan submission (for submit_query_plan tool)
	QueryPlan *AIResult

	// Info response submission (for info tool)
	InfoResponse *InfoResponseParams

	// Error case
	Error error
}

// AIConversation represents a single AI conversation session
type AIConversation struct {
	ID              string
	Provider        string
	Model           string
	APIKey          string
	OriginalRequest string
	SchemaContext   string
	CreatedAt       time.Time
	LastActivity    time.Time

	// Conversation state
	Messages     []ConversationMessage
	CurrentRound int
	MaxRounds    int

	// Provider-specific clients and message history
	anthropicClient   *anthropic.Client
	anthropicMessages []anthropic.MessageParam // Track actual Anthropic message format
	openaiClient      *openai.Client
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// AIResult represents a structured plan for CQL generation
type AIResult struct {
	Operation      string         `json:"operation"` // SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP
	Keyspace       string         `json:"keyspace,omitempty"`
	Table          string         `json:"table,omitempty"`
	Columns        []string       `json:"columns,omitempty"`
	Values         map[string]any `json:"values,omitempty"`
	Where          []WhereClause  `json:"where,omitempty"`
	OrderBy        []OrderClause  `json:"order_by,omitempty"`
	Limit          int            `json:"limit,omitempty"`
	AllowFiltering bool           `json:"allow_filtering,omitempty"`

	// For DDL operations
	Schema  map[string]string `json:"schema,omitempty"`  // Column definitions for CREATE TABLE
	Options map[string]any    `json:"options,omitempty"` // Table/keyspace options

	// Metadata
	Confidence float64 `json:"confidence"`
	Warning    string  `json:"warning,omitempty"`
	ReadOnly   bool    `json:"read_only"`

	// For informational responses (when Operation == "INFO")
	InfoContent string `json:"info_content,omitempty"`
	InfoTitle   string `json:"info_title,omitempty"`
}

// WhereClause represents a WHERE condition
type WhereClause struct {
	Column   string `json:"column"`
	Operator string `json:"operator"` // =, <, >, <=, >=, IN, CONTAINS
	Value    any    `json:"value"`
}

// OrderClause represents ORDER BY
type OrderClause struct {
	Column string `json:"column"`
	Order  string `json:"order"` // ASC or DESC
}

// PlanValidator validates a query plan against schema
type PlanValidator struct {
	Schema any // Will be *db.SchemaCatalog when we integrate
}

// Resolver handles fuzzy table name resolution
type Resolver struct {
	cache        *db.SchemaCache
	searchIndex  *SearchIndexManager
}
