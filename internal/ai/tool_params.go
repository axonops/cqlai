package ai

import (
	"encoding/json"
	"fmt"
)

// ToolParams is the interface that all tool parameter structs must implement
type ToolParams interface {
	Validate() error
}

// FuzzySearchParams represents parameters for fuzzy search tool
type FuzzySearchParams struct {
	Query string `json:"query"`
}

// ListKeyspacesParams represents parameters for list keyspaces tool
type ListKeyspacesParams struct{}

// ListTablesParams represents parameters for list tables tool
type ListTablesParams struct {
	Keyspace string `json:"keyspace"`
}

// UserSelectionParams represents parameters for user selection
type UserSelectionParams struct {
	Type    string   `json:"type"`
	Options []string `json:"options"`
}

// GetSchemaParams represents parameters for get schema tool
type GetSchemaParams struct {
	Keyspace string `json:"keyspace"`
	Table    string `json:"table"`
}

// InfoMessageParams represents parameters for info messages
type InfoMessageParams struct {
	Message string `json:"message"`
}

// InfoResponseParams represents parameters for informational responses
type InfoResponseParams struct {
	ResponseType string         `json:"response_type"` // "text" or "schema_info"
	Title        string         `json:"title,omitempty"`
	Content      string         `json:"content"`               // Text content for display
	SchemaInfo   map[string]any `json:"schema_info,omitempty"` // Optional structured data
	Confidence   float64        `json:"confidence"`
}

// SubmitQueryPlanParams represents the final query plan to execute
type SubmitQueryPlanParams struct {
	Operation      string            `json:"operation"`
	Keyspace       string            `json:"keyspace,omitempty"`
	Table          string            `json:"table,omitempty"`
	Columns        []string          `json:"columns,omitempty"`
	Values         map[string]any    `json:"values,omitempty"`
	ValueTypes     map[string]string `json:"value_types,omitempty"` // Phase 0: Type hints for values
	Where          []WhereClause     `json:"where,omitempty"`

	// Phase 1: USING clauses
	UsingTTL       int   `json:"using_ttl,omitempty"`       // TTL in seconds
	UsingTimestamp int64 `json:"using_timestamp,omitempty"` // Timestamp in microseconds

	// Phase 1: SELECT modifiers
	Distinct       bool `json:"distinct,omitempty"`     // SELECT DISTINCT
	SelectJSON     bool `json:"select_json,omitempty"`  // SELECT JSON
	PerPartitionLimit int `json:"per_partition_limit,omitempty"` // PER PARTITION LIMIT

	OrderBy        []OrderClause     `json:"order_by,omitempty"`
	Limit          int               `json:"limit,omitempty"`
	AllowFiltering bool              `json:"allow_filtering,omitempty"`
	Schema         map[string]string `json:"schema,omitempty"`
	Options        map[string]any    `json:"options,omitempty"`
	Confidence     float64           `json:"confidence"`
	Warning        string            `json:"warning,omitempty"`
	ReadOnly       bool              `json:"read_only"`
}

func (p FuzzySearchParams) Validate() error {
	if p.Query == "" {
		return fmt.Errorf("query is required")
	}
	return nil
}

func (p GetSchemaParams) Validate() error {
	if p.Keyspace == "" {
		return fmt.Errorf("keyspace is required")
	}
	if p.Table == "" {
		return fmt.Errorf("table is required")
	}
	return nil
}

func (p ListKeyspacesParams) Validate() error {
	return nil
}

func (p ListTablesParams) Validate() error {
	if p.Keyspace == "" {
		return fmt.Errorf("keyspace is required")
	}
	return nil
}

func (p UserSelectionParams) Validate() error {
	if p.Type == "" {
		return fmt.Errorf("selection type is required")
	}
	if len(p.Options) == 0 {
		return fmt.Errorf("at least one option is required")
	}
	return nil
}

func (p InfoMessageParams) Validate() error {
	if p.Message == "" {
		return fmt.Errorf("message is required")
	}
	return nil
}

func (p *InfoResponseParams) Validate() error {
	if p.Content == "" {
		return fmt.Errorf("content is required")
	}
	if p.ResponseType == "" {
		p.ResponseType = "text" // Default to text
	}
	return nil
}

func (p SubmitQueryPlanParams) Validate() error {
	if p.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	// Validate based on operation type
	switch p.Operation {
	case "SELECT", "UPDATE", "DELETE":
		if p.Table == "" {
			return fmt.Errorf("table is required for %s operation", p.Operation)
		}
	case "INSERT":
		if p.Table == "" {
			return fmt.Errorf("table is required for INSERT operation")
		}
		if len(p.Values) == 0 {
			return fmt.Errorf("values are required for INSERT operation")
		}
	}

	return nil
}

// ToQueryPlan converts the params to a QueryPlan
func (p SubmitQueryPlanParams) ToQueryPlan() *AIResult {
	return &AIResult{
		Operation:      p.Operation,
		Keyspace:       p.Keyspace,
		Table:          p.Table,
		Columns:        p.Columns,
		Values:         p.Values,
		ValueTypes:     p.ValueTypes, // Phase 0: Pass through type hints
		Where:          p.Where,
		UsingTTL:          p.UsingTTL,          // Phase 1
		UsingTimestamp:    p.UsingTimestamp,    // Phase 1
		Distinct:          p.Distinct,          // Phase 1
		SelectJSON:        p.SelectJSON,        // Phase 1
		OrderBy:           p.OrderBy,
		Limit:             p.Limit,
		PerPartitionLimit: p.PerPartitionLimit, // Phase 1
		AllowFiltering: p.AllowFiltering,
		Schema:         p.Schema,
		Options:        p.Options,
		Confidence:     p.Confidence,
		Warning:        p.Warning,
		ReadOnly:       p.ReadOnly,
	}
}

// ParseToolParams parses raw parameters into the appropriate typed struct
func ParseToolParams(toolName ToolName, rawParams json.RawMessage) (ToolParams, error) {
	switch toolName {
	case ToolFuzzySearch:
		var params FuzzySearchParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid fuzzy search parameters: %w", err)
		}
		return params, nil

	case ToolGetSchema:
		var params GetSchemaParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid get schema parameters: %w", err)
		}
		return params, nil

	case ToolListKeyspaces:
		return ListKeyspacesParams{}, nil

	case ToolListTables:
		var params ListTablesParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid list tables parameters: %w", err)
		}
		return params, nil

	case ToolUserSelection:
		var params UserSelectionParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid user selection parameters: %w", err)
		}
		return params, nil

	case ToolNotEnoughInfo:
		var params InfoMessageParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid info message parameters: %w", err)
		}
		return params, nil

	case ToolNotRelevant:
		var params InfoMessageParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			// not_relevant might not have a message
			return InfoMessageParams{}, nil
		}
		return params, nil

	case ToolSubmitQueryPlan:
		var params SubmitQueryPlanParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid query plan parameters: %w", err)
		}
		return params, nil

	case ToolInfo:
		var params InfoResponseParams
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid info response parameters: %w", err)
		}
		return &params, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// ParseToolParamsFromMap parses a map into the appropriate typed struct
func ParseToolParamsFromMap(toolName ToolName, args map[string]any) (ToolParams, error) {
	// Convert map to JSON then parse
	jsonBytes, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}
	return ParseToolParams(toolName, jsonBytes)
}
