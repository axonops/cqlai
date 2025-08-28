package ai

import (
	"fmt"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
)

// AI is the main orchestrator for AI features
type AI struct {
	cache    *db.SchemaCache
	resolver *Resolver
	session  *db.Session
	config   *Config
}

// Config holds AI configuration
type Config struct {
	Provider    string // openai, anthropic, local
	Model       string
	APIKey      string
	Temperature float64
	MaxTokens   int
	PrivacyMode string // schema_only, allow_samples, full
	CacheTTL    time.Duration
}

// NewAI creates a new AI orchestrator
func NewAI(session *db.Session, config *Config) (*AI, error) {
	cache := db.NewSchemaCache(session)

	// Initialize schema cache
	if err := cache.Refresh(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema cache: %w", err)
	}

	resolver := NewResolver(cache)

	return &AI{
		cache:    cache,
		resolver: resolver,
		session:  session,
		config:   config,
	}, nil
}

// NewAIWithCache creates a new AI orchestrator using the session's existing cache
func NewAIWithCache(session *db.Session, config *Config) (*AI, error) {
	// Use the cache that was already populated when the session was created
	cache := session.GetSchemaCache()
	if cache == nil {
		// Fallback to creating a new cache if session doesn't have one
		cache = db.NewSchemaCache(session)
		if err := cache.Refresh(); err != nil {
			return nil, fmt.Errorf("failed to initialize schema cache: %w", err)
		}
	}

	resolver := NewResolver(cache)

	return &AI{
		cache:    cache,
		resolver: resolver,
		session:  session,
		config:   config,
	}, nil
}

// GetSchemaContext returns schema information formatted for LLM context
func (ai *AI) GetSchemaContext(limit int) (string, error) {
	ai.cache.Mu.RLock()
	defer ai.cache.Mu.RUnlock()

	var context strings.Builder
	context.WriteString("Available Cassandra Schema:\n\n")

	// List keyspaces
	fmt.Fprintf(&context, "Keyspaces (%d):\n", len(ai.cache.Keyspaces))
	for _, ks := range ai.cache.Keyspaces {
		fmt.Fprintf(&context, "  - %s\n", ks)
	}
	context.WriteString("\n")

	// List tables with columns
	tableCount := 0
	for keyspace, tables := range ai.cache.Tables {
		for _, table := range tables {
			if limit > 0 && tableCount >= limit {
				fmt.Fprintf(&context, "\n... and %d more tables\n", ai.cache.CountTotalTables()-tableCount)
				return context.String(), nil
			}

			fmt.Fprintf(&context, "Table: %s.%s\n", keyspace, table.TableName)

			// Add columns
			if columns, ok := ai.cache.Columns[keyspace][table.TableName]; ok {
				context.WriteString("  Columns:\n")
				for _, col := range columns {
					kind := ""
					if col.Kind == "partition_key" || col.Kind == "clustering" {
						kind = fmt.Sprintf(" (%s)", col.Kind)
					}
					fmt.Fprintf(&context, "    - %s %s%s\n", col.Name, col.DataType, kind)
				}
			}
			context.WriteString("\n")
			tableCount++
		}
	}

	return context.String(), nil
}

// GetSchemaStats returns statistics about the cached schema
func (ai *AI) GetSchemaStats() map[string]any {
	ai.cache.Mu.RLock()
	defer ai.cache.Mu.RUnlock()

	return map[string]any{
		"keyspaces":    ai.cache.Keyspaces,
		"table_count":  ai.cache.CountTotalTables(),
		"last_refresh": ai.cache.LastRefresh,
	}
}

// RefreshSchema manually refreshes the schema cache
func (ai *AI) RefreshSchema() error {
	return ai.cache.Refresh()
}

// GetCache returns the schema cache
func (ai *AI) GetCache() *db.SchemaCache {
	return ai.cache
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
