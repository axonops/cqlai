package ai

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// FuzzyMatch represents a fuzzy search match result
type FuzzyMatch struct {
	Keyspace string
	Table    string
	Score    float64
	Columns  []string
}

// SearchIndexManager manages the search index for fuzzy matching
// This is separate from schema cache to optimize performance
type SearchIndexManager struct {
	cache           *db.SchemaCache
	tableIndex      map[string]*TableSearchEntry // keyspace.table -> search entry
	lastIndexBuild  time.Time
	indexTTL        time.Duration
	mu              sync.RWMutex
	buildInProgress bool
}

// TableSearchEntry contains search metadata for a table
type TableSearchEntry struct {
	Keyspace      string
	Table         string
	Tokens        []string // Tokenized table name for fuzzy matching
	Columns       []string // Column names for enhanced matching
	LastAccessed  time.Time
}

// NewSearchIndexManager creates a new search index manager
func NewSearchIndexManager(cache *db.SchemaCache) *SearchIndexManager {
	return &SearchIndexManager{
		cache:      cache,
		tableIndex: make(map[string]*TableSearchEntry),
		indexTTL:   5 * time.Minute, // Cache for 5 minutes
	}
}

// BuildIndexIfNeeded builds the search index if it's expired or not built
func (sim *SearchIndexManager) BuildIndexIfNeeded() error {
	sim.mu.RLock()
	needsBuild := time.Since(sim.lastIndexBuild) > sim.indexTTL || len(sim.tableIndex) == 0
	inProgress := sim.buildInProgress
	sim.mu.RUnlock()

	if !needsBuild || inProgress {
		return nil
	}

	return sim.BuildIndex()
}

// BuildIndex rebuilds the entire search index from schema cache
func (sim *SearchIndexManager) BuildIndex() error {
	sim.mu.Lock()
	if sim.buildInProgress {
		sim.mu.Unlock()
		return nil // Another goroutine is already building
	}
	sim.buildInProgress = true
	sim.mu.Unlock()

	defer func() {
		sim.mu.Lock()
		sim.buildInProgress = false
		sim.mu.Unlock()
	}()

	logger.DebugfToFile("SearchIndexManager", "Building search index")
	startTime := time.Now()

	// Ensure schema cache is refreshed
	if err := sim.cache.RefreshIfNeeded(5 * time.Minute); err != nil {
		return fmt.Errorf("failed to refresh schema cache: %w", err)
	}

	newIndex := make(map[string]*TableSearchEntry)

	// Get cached data with read lock
	sim.cache.Mu.RLock()
	for keyspace, tables := range sim.cache.Tables {
		for _, table := range tables {
			key := fmt.Sprintf("%s.%s", keyspace, table.TableName)

			// Create search entry
			entry := &TableSearchEntry{
				Keyspace:     keyspace,
				Table:        table.TableName,
				Tokens:       tokenizeTableName(table.TableName),
				LastAccessed: time.Now(),
			}

			// Add column names if available
			if columns, ok := sim.cache.Columns[keyspace][table.TableName]; ok {
				for _, col := range columns {
					entry.Columns = append(entry.Columns, col.Name)
				}
			}

			newIndex[key] = entry
		}
	}
	sim.cache.Mu.RUnlock()

	// Update the index atomically
	sim.mu.Lock()
	sim.tableIndex = newIndex
	sim.lastIndexBuild = time.Now()
	sim.mu.Unlock()

	logger.DebugfToFile("SearchIndexManager", "Search index built in %v with %d entries",
		time.Since(startTime), len(newIndex))

	return nil
}

// FindTables performs fuzzy search for tables
func (sim *SearchIndexManager) FindTables(query string, limit int) []FuzzyMatch {
	// Ensure index is built
	if err := sim.BuildIndexIfNeeded(); err != nil {
		logger.DebugfToFile("SearchIndexManager", "Failed to build index: %v", err)
	}

	sim.mu.RLock()
	defer sim.mu.RUnlock()

	queryLower := strings.ToLower(query)
	queryTokens := tokenizeTableName(query)

	var matches []FuzzyMatch

	for _, entry := range sim.tableIndex {
		score := calculateFuzzyScore(queryLower, queryTokens, entry)

		if score > 0 {
			matches = append(matches, FuzzyMatch{
				Keyspace: entry.Keyspace,
				Table:    entry.Table,
				Score:    score,
				Columns:  entry.Columns,
			})
		}
	}

	// Sort by score (highest first)
	sortFuzzyMatches(matches)

	// Limit results
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	// Update last accessed time for matched entries
	go sim.updateLastAccessed(matches)

	return matches
}

// updateLastAccessed updates the last accessed time for matched entries
func (sim *SearchIndexManager) updateLastAccessed(matches []FuzzyMatch) {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	now := time.Now()
	for _, match := range matches {
		key := fmt.Sprintf("%s.%s", match.Keyspace, match.Table)
		if entry, ok := sim.tableIndex[key]; ok {
			entry.LastAccessed = now
		}
	}
}

// calculateFuzzyScore calculates the fuzzy match score
func calculateFuzzyScore(queryLower string, queryTokens []string, entry *TableSearchEntry) float64 {
	tableLower := strings.ToLower(entry.Table)

	// Exact match
	if tableLower == queryLower {
		return 100.0
	}

	// Prefix match
	if strings.HasPrefix(tableLower, queryLower) {
		return 90.0
	}

	// Contains match
	if strings.Contains(tableLower, queryLower) {
		return 70.0
	}

	// Token matching
	score := 0.0
	for _, queryToken := range queryTokens {
		for _, tableToken := range entry.Tokens {
			tableTokenLower := strings.ToLower(tableToken)

			switch {
			case tableTokenLower == queryToken:
				score += 50.0
			case strings.HasPrefix(tableTokenLower, queryToken):
				score += 30.0
			case strings.Contains(tableTokenLower, queryToken):
				score += 10.0
			}
		}
	}

	// Column name matching (lower weight)
	for _, col := range entry.Columns {
		colLower := strings.ToLower(col)
		if colLower == queryLower {
			score += 20.0
		} else if strings.Contains(colLower, queryLower) {
			score += 5.0
		}
	}

	// Normalize score
	if score > 100.0 {
		score = 100.0
	}

	return score
}

// tokenizeTableName splits a table name into searchable tokens
func tokenizeTableName(name string) []string {
	tokens := []string{strings.ToLower(name)}

	// Split by underscores
	parts := strings.Split(name, "_")
	for _, part := range parts {
		if part != "" && len(part) > 1 {
			tokens = append(tokens, strings.ToLower(part))
		}
	}

	// Split by hyphens
	parts = strings.Split(name, "-")
	for _, part := range parts {
		if part != "" && len(part) > 1 {
			tokens = append(tokens, strings.ToLower(part))
		}
	}

	// Simple camelCase splitting
	camelTokens := splitCamelCase(name)
	for _, token := range camelTokens {
		if len(token) > 1 {
			tokens = append(tokens, strings.ToLower(token))
		}
	}

	return uniqueTokens(tokens)
}

// splitCamelCase splits a camelCase string into tokens
func splitCamelCase(s string) []string {
	var tokens []string
	var current []rune

	for i, r := range s {
		if i > 0 && isUpper(r) && (i == len(s)-1 || !isUpper(rune(s[i+1]))) {
			if len(current) > 0 {
				tokens = append(tokens, string(current))
				current = []rune{}
			}
		}
		current = append(current, r)
	}

	if len(current) > 0 {
		tokens = append(tokens, string(current))
	}

	return tokens
}

// isUpper checks if a rune is uppercase
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// uniqueTokens returns unique tokens from a slice
func uniqueTokens(tokens []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, token := range tokens {
		if !seen[token] && token != "" {
			seen[token] = true
			unique = append(unique, token)
		}
	}

	return unique
}

// sortFuzzyMatches sorts matches by score (highest first)
func sortFuzzyMatches(matches []FuzzyMatch) {
	// Simple bubble sort for small datasets
	for i := 0; i < len(matches)-1; i++ {
		for j := 0; j < len(matches)-i-1; j++ {
			if matches[j].Score < matches[j+1].Score {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}
}

// GetStats returns statistics about the search index
func (sim *SearchIndexManager) GetStats() map[string]interface{} {
	sim.mu.RLock()
	defer sim.mu.RUnlock()

	return map[string]interface{}{
		"total_entries":     len(sim.tableIndex),
		"last_index_build":  sim.lastIndexBuild,
		"index_age_seconds": time.Since(sim.lastIndexBuild).Seconds(),
		"ttl_seconds":       sim.indexTTL.Seconds(),
	}
}

// InvalidateIndex marks the index as needing rebuild
func (sim *SearchIndexManager) InvalidateIndex() {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	sim.lastIndexBuild = time.Time{} // Zero time forces rebuild
	logger.DebugfToFile("SearchIndexManager", "Search index invalidated")
}