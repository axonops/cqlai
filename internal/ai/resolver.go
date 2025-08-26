package ai

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// TableCandidate represents a potential table match
type TableCandidate struct {
	Keyspace  string
	Table     string
	Score     float64
	MatchType string   // exact, fuzzy, semantic
	Columns   []string // Sample columns for context
}

// Resolver handles fuzzy table name resolution
type Resolver struct {
	cache *db.SchemaCache
}

// NewResolver creates a new table resolver
func NewResolver(cache *db.SchemaCache) *Resolver {
	return &Resolver{
		cache: cache,
	}
}

// FindTables searches for tables matching a query
func (r *Resolver) FindTables(query string, limit int) []TableCandidate {
	r.cache.Mu.RLock()
	defer r.cache.Mu.RUnlock()
	
	query = strings.ToLower(strings.TrimSpace(query))
	candidates := []TableCandidate{}
	
	// Debug: Finding tables for query
	
	// Check all tables
	for keyspace, tables := range r.cache.Tables {
		for _, table := range tables {
			tableLower := strings.ToLower(table.TableName)
			score := 0.0
			matchType := ""
			
			// Exact match
			if tableLower == query {
				score = 1.0
				matchType = "exact"
			} else if strings.Contains(tableLower, query) || strings.Contains(query, tableLower) {
				// Substring match
				score = 0.8
				matchType = "substring"
			} else {
				// Fuzzy string similarity
				score = r.calculateSimilarity(query, tableLower)
				if score > 0.5 {
					matchType = "fuzzy"
				}
				
				// Token overlap - simple token matching without patterns
				tableKey := fmt.Sprintf("%s.%s", keyspace, table.TableName)
				if tokens, ok := r.cache.SearchIndex.TableTokens[tableKey]; ok {
					tokenScore := r.calculateTokenOverlap(query, tokens)
					if tokenScore > score {
						score = tokenScore
						matchType = "token"
					}
				}
			}
			
			if score > 0.3 { // Minimum threshold
				// Get sample columns
				sampleColumns := []string{}
				if columns, ok := r.cache.Columns[keyspace][table.TableName]; ok {
					for i, col := range columns {
						if i < 5 { // Limit to 5 columns
							sampleColumns = append(sampleColumns, col.Name)
						}
					}
				}
				
				candidates = append(candidates, TableCandidate{
					Keyspace:  keyspace,
					Table:     table.TableName,
					Score:     score,
					MatchType: matchType,
					Columns:   sampleColumns,
				})
			}
		}
	}
	
	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	
	// Apply limit
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	
	// Debug: Found candidates
	
	return candidates
}

// calculateSimilarity calculates string similarity using fuzzy search
func (r *Resolver) calculateSimilarity(s1, s2 string) float64 {
	// Use Levenshtein distance for similarity
	distance := fuzzy.LevenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	
	if maxLen == 0 {
		return 1.0
	}
	
	// Convert distance to similarity score (0-1)
	similarity := 1.0 - (float64(distance) / maxLen)
	return similarity
}

// calculateTokenOverlap calculates the overlap between query and table tokens
func (r *Resolver) calculateTokenOverlap(query string, tokens []string) float64 {
	queryLower := strings.ToLower(query)
	matchCount := 0
	
	for _, token := range tokens {
		tokenLower := strings.ToLower(token)
		if queryLower == tokenLower {
			return 0.9 // High score for exact token match
		}
		if strings.Contains(queryLower, tokenLower) || strings.Contains(tokenLower, queryLower) {
			matchCount++
		}
	}
	
	if matchCount > 0 {
		return 0.6 + (0.3 * float64(matchCount) / float64(len(tokens)))
	}
	
	return 0.0
}


// FindTablesWithFuzzy uses the fuzzy search library for better matching
func (r *Resolver) FindTablesWithFuzzy(query string, limit int) []TableCandidate {
	r.cache.Mu.RLock()
	defer r.cache.Mu.RUnlock()
	
	query = strings.ToLower(strings.TrimSpace(query))
	candidates := []TableCandidate{}
	
	// Debug: log the query and cache state
	logger.DebugfToFile("Resolver", "FindTablesWithFuzzy: query='%s', keyspaces=%d, tables=%d",
		query, len(r.cache.Keyspaces), len(r.cache.Tables))
	
	// Collect all table names for fuzzy searching
	tableList := []string{}
	tableMap := make(map[string]db.CachedTableInfo)
	
	// Also check if query matches keyspaces
	for _, keyspace := range r.cache.Keyspaces {
		// Add keyspace name to search list
		tableList = append(tableList, keyspace)
	}
	
	for keyspace, tables := range r.cache.Tables {
		logger.DebugfToFile("Resolver", "Keyspace %s has %d tables", keyspace, len(tables))
		for _, table := range tables {
			tableKey := fmt.Sprintf("%s.%s", keyspace, table.TableName)
			tableList = append(tableList, table.TableName)
			tableList = append(tableList, tableKey) // Also search with keyspace prefix
			tableMap[tableKey] = table
			tableMap[table.TableName] = table
			
			// If query matches keyspace, add all tables from that keyspace
			if fuzzy.MatchNormalized(query, strings.ToLower(keyspace)) {
				logger.DebugfToFile("Resolver", "Keyspace %s matches query %s", keyspace, query)
			}
		}
	}
	
	// Use fuzzy.RankFind for ranking matches
	matches := fuzzy.RankFindNormalized(query, tableList)
	logger.DebugfToFile("Resolver", "Fuzzy search found %d matches", len(matches))
	
	// Check if any keyspace matches - if so, return all tables from that keyspace
	matchedKeyspaces := make(map[string]bool)
	for _, match := range matches {
		// Check if the match is a keyspace
		for _, ks := range r.cache.Keyspaces {
			if match.Target == ks && match.Distance > 30 {
				matchedKeyspaces[ks] = true
				logger.DebugfToFile("Resolver", "Matched keyspace: %s (score: %d)", ks, match.Distance)
			}
		}
	}
	
	// If keyspaces matched, add all their tables as candidates
	if len(matchedKeyspaces) > 0 {
		for keyspace := range matchedKeyspaces {
			if tables, ok := r.cache.Tables[keyspace]; ok {
				for _, table := range tables {
					// Get sample columns
					sampleColumns := []string{}
					if columns, ok := r.cache.Columns[keyspace][table.TableName]; ok {
						for i, col := range columns {
							if i < 5 { // Limit to 5 columns
								sampleColumns = append(sampleColumns, col.Name)
							}
						}
					}
					
					candidates = append(candidates, TableCandidate{
						Keyspace:  keyspace,
						Table:     table.TableName,
						Score:     0.9, // High score for keyspace match
						MatchType: "keyspace",
						Columns:   sampleColumns,
					})
				}
			}
		}
		logger.DebugfToFile("Resolver", "Added %d tables from matched keyspaces", len(candidates))
	}
	
	// Also add direct table matches
	for _, match := range matches {
		if match.Distance > 30 { // Minimum threshold (30 out of 100)
			var tableInfo db.CachedTableInfo
			var found bool
			
			// Try to find the table info
			if info, ok := tableMap[match.Target]; ok {
				tableInfo = info
				found = true
			}
			
			if found {
				// Check if we already added this table from keyspace match
				alreadyAdded := false
				for _, c := range candidates {
					if c.Keyspace == tableInfo.KeyspaceName && c.Table == tableInfo.TableName {
						alreadyAdded = true
						break
					}
				}
				
				if !alreadyAdded {
					// Get sample columns
					sampleColumns := []string{}
					if columns, ok := r.cache.Columns[tableInfo.KeyspaceName][tableInfo.TableName]; ok {
						for i, col := range columns {
							if i < 5 { // Limit to 5 columns
								sampleColumns = append(sampleColumns, col.Name)
							}
						}
					}
					
					candidates = append(candidates, TableCandidate{
						Keyspace:  tableInfo.KeyspaceName,
						Table:     tableInfo.TableName,
						Score:     float64(match.Distance) / 100.0, // Normalize to 0-1
						MatchType: "fuzzy",
						Columns:   sampleColumns,
					})
				}
			}
		}
	}
	
	// Apply limit
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	
	return candidates
}