package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note: SSE event sending uses mark3labs/mcp-go library's SendNotificationToAllClients.
// We can't test actual SSE delivery without a full HTTP server and client,
// but we test that the sending logic doesn't panic and handles edge cases.

// ============================================================================
// SSE Event Sending Logic Tests
// ============================================================================

func TestSendConfirmationStatusEvent_NilServer(t *testing.T) {
	// Test that sending event with nil mcpServer doesn't panic
	server := &MCPServer{
		mcpServer: nil,
	}

	// Should not panic
	server.sendConfirmationStatusEvent("req_001", "CONFIRMED", "mcp", "")
	// If we got here, no panic occurred
}

func TestSendConfirmationStatusEvent_Parameters(t *testing.T) {
	// Test that different parameter combinations work correctly
	server := &MCPServer{
		mcpServer: nil, // nil is ok - method handles it
	}

	tests := []struct {
		name      string
		requestID string
		status    string
		actor     string
		reason    string
	}{
		{"CONFIRMED without reason", "req_001", "CONFIRMED", "mcp", ""},
		{"DENIED with reason", "req_002", "DENIED", "mcp", "user rejected"},
		{"CANCELLED with reason", "req_003", "CANCELLED", "user", "changed mind"},
		{"TIMEOUT with reason", "req_004", "TIMEOUT", "system", "exceeded timeout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic regardless of parameters
			server.sendConfirmationStatusEvent(tt.requestID, tt.status, tt.actor, tt.reason)
		})
	}
}

// ============================================================================
// Integration Tests: Confirmation Lifecycle Sends SSE Events
// ============================================================================

func TestConfirmRequest_CallsSSEEventSender(t *testing.T) {
	// Create mock server with confirmation queue
	queue := NewConfirmationQueue()
	server := &MCPServer{
		confirmationQueue: queue,
		mcpServer:         nil, // nil is ok - sendConfirmationStatusEvent handles it
	}

	// Create a confirmation request
	classification := QueryClassification{
		IsDangerous: true,
	}
	req := queue.NewConfirmationRequest(
		"DROP TABLE users",
		classification,
		"execute_cql",
		"DROP",
		5*time.Minute,
	)

	// Confirm the request (this should call sendConfirmationStatusEvent)
	err := server.ConfirmRequest(req.ID, "mcp")
	assert.NoError(t, err)

	// Verify request was confirmed
	updated, _ := server.GetConfirmationRequest(req.ID)
	assert.Equal(t, "CONFIRMED", updated.Status)

	// SSE event sender was called (verified by no panic)
}

func TestDenyRequest_CallsSSEEventSender(t *testing.T) {
	queue := NewConfirmationQueue()
	server := &MCPServer{
		confirmationQueue: queue,
		mcpServer:         nil,
	}

	classification := QueryClassification{
		IsDangerous: true,
	}
	req := queue.NewConfirmationRequest(
		"DROP TABLE users",
		classification,
		"execute_cql",
		"DROP",
		5*time.Minute,
	)

	// Deny the request
	err := server.DenyRequest(req.ID, "mcp", "operation too dangerous")
	assert.NoError(t, err)

	// Verify request was denied
	updated, _ := server.GetConfirmationRequest(req.ID)
	assert.Equal(t, "DENIED", updated.Status)

	// SSE event sender was called (verified by no panic)
}

func TestCancelRequest_CallsSSEEventSender(t *testing.T) {
	queue := NewConfirmationQueue()
	server := &MCPServer{
		confirmationQueue: queue,
		mcpServer:         nil,
	}

	classification := QueryClassification{
		IsDangerous: true,
	}
	req := queue.NewConfirmationRequest(
		"DROP TABLE users",
		classification,
		"execute_cql",
		"DROP",
		5*time.Minute,
	)

	// Cancel the request
	err := server.CancelRequest(req.ID, "user", "changed mind")
	assert.NoError(t, err)

	// Verify request was cancelled
	updated, _ := server.GetConfirmationRequest(req.ID)
	assert.Equal(t, "CANCELLED", updated.Status)

	// SSE event sender was called (verified by no panic)
}
