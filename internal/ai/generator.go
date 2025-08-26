package ai

import (
	"fmt"

	"github.com/axonops/cqlai/internal/db"
)

// Global AI instance
var globalAI *AI

// Initialize sets up the global AI instance
func Initialize(session *db.Session) error {
	config := &Config{
		Provider:    "local", // Start with local processing
		PrivacyMode: "schema_only",
		Temperature: 0.3,
		MaxTokens:   2000,
	}
	
	var err error
	globalAI, err = NewAIWithCache(session, config)
	if err != nil {
		return fmt.Errorf("failed to initialize AI: %w", err)
	}
	
	// Router's AI handler will be initialized separately
	
	return nil
}

// InitializeLocalAI ensures the local AI cache is initialized
func InitializeLocalAI(session *db.Session) error {
	if globalAI == nil {
		return Initialize(session)
	}
	return nil
}

// NOTE: QueryPlan is already defined in ai.go

// GetAI returns the global AI instance
func GetAI() *AI {
	return globalAI
}

// GetGlobalAI returns the global AI instance (alias for GetAI)
func GetGlobalAI() *AI {
	return globalAI
}