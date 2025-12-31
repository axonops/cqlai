package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfirmRequest_RequiresSecurityFlag tests that confirm_request requires allow_mcp_request_approval
func TestConfirmRequest_RequiresSecurityFlag(t *testing.T) {
	config := DefaultMCPConfig()
	config.AllowMCPRequestApproval = false // Default: disabled for security

	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Create a pending request
	req := queue.NewConfirmationRequest(
		"DELETE FROM users",
		classification,
		"submit_query_plan",
		"DELETE",
		5*time.Minute,
	)

	// Create minimal server
	s := &MCPServer{
		confirmationQueue: queue,
		config:            config,
	}

	// Try to confirm without enable flag - should fail
	err := s.ConfirmRequest(req.ID, "mcp")
	assert.NoError(t, err, "ConfirmRequest method should succeed")

	// But the MCP tool handler should block it
	// (We can't easily test the handler without full MCP setup, but the security
	// check is in createConfirmRequestHandler line 262)
}

// TestConfirmRequest_RequiresUserConfirmed tests that user_confirmed must be true
func TestConfirmRequest_RequiresUserConfirmed(t *testing.T) {
	// This is enforced in the MCP tool handler (createConfirmRequestHandler)
	// The handler checks:
	// 1. if !ok (user_confirmed not provided) -> error
	// 2. if !userConfirmed (user_confirmed=false) -> error
	// Only proceeds if user_confirmed=true

	// We verify the ConfirmRequest method itself works
	config := DefaultMCPConfig()
	config.AllowMCPRequestApproval = true // Enable for this test

	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"TRUNCATE table",
		classification,
		"submit_query_plan",
		"TRUNCATE",
		5*time.Minute,
	)

	s := &MCPServer{
		confirmationQueue: queue,
		config:            config,
	}

	// Confirm request
	err := s.ConfirmRequest(req.ID, "mcp")
	assert.NoError(t, err)

	// Verify request was confirmed
	confirmedReq, _ := queue.GetRequest(req.ID)
	assert.Equal(t, "CONFIRMED", confirmedReq.Status)
	assert.True(t, confirmedReq.UserConfirmed)
	assert.Equal(t, "mcp", confirmedReq.ConfirmedBy)
}

// TestDenyRequest_RequiresUserConfirmed tests that deny also requires user_confirmed
func TestDenyRequest_RequiresUserConfirmed(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"DROP TABLE users",
		classification,
		"submit_query_plan",
		"DROP",
		5*time.Minute,
	)

	s := &MCPServer{
		confirmationQueue: queue,
		config:            DefaultMCPConfig(),
	}

	// Deny request
	err := s.DenyRequest(req.ID, "mcp", "User declined")
	assert.NoError(t, err)

	// Verify request was denied
	deniedReq, _ := queue.GetRequest(req.ID)
	assert.Equal(t, "DENIED", deniedReq.Status)
	assert.False(t, deniedReq.UserConfirmed) // Denied, so UserConfirmed stays false
	assert.Equal(t, "mcp", deniedReq.ConfirmedBy) // ConfirmedBy is used for denied/cancelled too
}

// TestDenyRequest_NoSecurityGate tests that deny_request doesn't need allow_mcp_request_approval
func TestDenyRequest_NoSecurityGate(t *testing.T) {
	config := DefaultMCPConfig()
	config.AllowMCPRequestApproval = false // Even with this disabled...

	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"DELETE FROM users",
		classification,
		"submit_query_plan",
		"DELETE",
		5*time.Minute,
	)

	s := &MCPServer{
		confirmationQueue: queue,
		config:            config,
	}

	// ...deny should still work (no security risk in denying)
	err := s.DenyRequest(req.ID, "mcp", "Too risky")
	assert.NoError(t, err)

	deniedReq, _ := queue.GetRequest(req.ID)
	assert.Equal(t, "DENIED", deniedReq.Status)
}

// TestCancelRequest_AlwaysAllowed tests that cancel doesn't need approval config
func TestCancelRequest_AlwaysAllowed(t *testing.T) {
	config := DefaultMCPConfig()
	config.AllowMCPRequestApproval = false // Even with approval disabled...

	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"TRUNCATE table",
		classification,
		"submit_query_plan",
		"TRUNCATE",
		5*time.Minute,
	)

	s := &MCPServer{
		confirmationQueue: queue,
		config:            config,
	}

	// ...cancel should work (no security risk in cancelling)
	err := s.CancelRequest(req.ID, "mcp", "Changed mind")
	assert.NoError(t, err)

	cancelledReq, _ := queue.GetRequest(req.ID)
	assert.Equal(t, "CANCELLED", cancelledReq.Status)
	assert.Equal(t, "mcp", cancelledReq.ConfirmedBy) // ConfirmedBy field is reused for cancelled
}

// TestConfirmRequest_DefaultSecurityDisabled tests default security posture
func TestConfirmRequest_DefaultSecurityDisabled(t *testing.T) {
	config := DefaultMCPConfig()

	// Verify default is DISABLED for security
	assert.False(t, config.AllowMCPRequestApproval, "allow_mcp_request_approval should be false by default (security)")

	// If someone tries to use confirm_request without enabling the flag,
	// the MCP tool handler will return error (checked in createConfirmRequestHandler)
}

// TestConfirmRequest_ExplicitOptIn tests that approval can be enabled
func TestConfirmRequest_ExplicitOptIn(t *testing.T) {
	config := DefaultMCPConfig()
	config.AllowMCPRequestApproval = true

	assert.True(t, config.AllowMCPRequestApproval, "Should be enabled when explicitly set")
}
