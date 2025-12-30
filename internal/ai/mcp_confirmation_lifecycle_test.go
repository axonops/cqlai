package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCancelRequest tests cancelling a confirmation request
func TestCancelRequest(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"DELETE FROM users WHERE id = 1",
		classification,
		"submit_query_plan",
		"",
		5*time.Minute,
	)

	// Cancel the request
	err := queue.CancelRequest(req.ID, "testuser", "Testing cancellation")
	assert.NoError(t, err)

	// Verify status changed to CANCELLED
	retrieved, _ := queue.GetRequest(req.ID)
	assert.Equal(t, "CANCELLED", retrieved.Status)
	assert.False(t, retrieved.UserConfirmed)
	assert.Equal(t, "testuser", retrieved.ConfirmedBy)
	assert.False(t, retrieved.ConfirmedAt.IsZero())
}

// TestCancelRequest_NonExistent tests cancelling non-existent request
func TestCancelRequest_NonExistent(t *testing.T) {
	queue := NewConfirmationQueue()

	err := queue.CancelRequest("req_999", "testuser", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestCancelRequest_AnyState tests cancel works in any state
func TestCancelRequest_AnyState(t *testing.T) {
	states := []string{"PENDING", "CONFIRMED", "DENIED", "TIMEOUT"}

	for _, initialState := range states {
		t.Run(initialState, func(t *testing.T) {
			queue := NewConfirmationQueue()
			classification := QueryClassification{IsDangerous: true}

			req := queue.NewConfirmationRequest(
				"DELETE FROM users",
				classification,
				"submit_query_plan",
				"",
				5*time.Minute,
			)

			// Set to initial state
			queue.requests[req.ID].Status = initialState

			// Cancel should work regardless of state
			err := queue.CancelRequest(req.ID, "testuser", "cancel test")
			assert.NoError(t, err)

			// Verify now CANCELLED
			retrieved, _ := queue.GetRequest(req.ID)
			assert.Equal(t, "CANCELLED", retrieved.Status)
		})
	}
}

// TestGetApprovedConfirmations tests getting approved requests
func TestGetApprovedConfirmations(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Create 3 requests
	req1 := queue.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)
	_ = queue.NewConfirmationRequest("DROP TABLE logs", classification, "tool", "", 5*time.Minute)
	req3 := queue.NewConfirmationRequest("TRUNCATE events", classification, "tool", "", 5*time.Minute)

	// Confirm 2 of them
	queue.ConfirmRequest(req1.ID, "user1")
	queue.ConfirmRequest(req3.ID, "user2")

	// Get approved
	approved := queue.GetApprovedConfirmations()

	assert.Equal(t, 2, len(approved))
	ids := []string{approved[0].ID, approved[1].ID}
	assert.Contains(t, ids, req1.ID)
	assert.Contains(t, ids, req3.ID)
}

// TestGetDeniedConfirmations tests getting denied requests
func TestGetDeniedConfirmations(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Create 3 requests
	req1 := queue.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)
	req2 := queue.NewConfirmationRequest("DROP TABLE logs", classification, "tool", "", 5*time.Minute)
	_ = queue.NewConfirmationRequest("TRUNCATE events", classification, "tool", "", 5*time.Minute)

	// Deny 2 of them
	queue.DenyRequest(req1.ID, "user1", "too dangerous")
	queue.DenyRequest(req2.ID, "user2", "not needed")

	// Get denied
	denied := queue.GetDeniedConfirmations()

	assert.Equal(t, 2, len(denied))
	ids := []string{denied[0].ID, denied[1].ID}
	assert.Contains(t, ids, req1.ID)
	assert.Contains(t, ids, req2.ID)
}

// TestGetCancelledConfirmations tests getting cancelled requests
func TestGetCancelledConfirmations(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Create 3 requests
	req1 := queue.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)
	_ = queue.NewConfirmationRequest("DROP TABLE logs", classification, "tool", "", 5*time.Minute)
	req3 := queue.NewConfirmationRequest("TRUNCATE events", classification, "tool", "", 5*time.Minute)

	// Cancel 2 of them
	queue.CancelRequest(req1.ID, "claude", "user changed mind")
	queue.CancelRequest(req3.ID, "claude", "not needed")

	// Get cancelled
	cancelled := queue.GetCancelledConfirmations()

	assert.Equal(t, 2, len(cancelled))
	ids := []string{cancelled[0].ID, cancelled[1].ID}
	assert.Contains(t, ids, req1.ID)
	assert.Contains(t, ids, req3.ID)
}

// TestMixedStates tests multiple requests in different states
func TestMixedStates(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Create 5 requests
	req1 := queue.NewConfirmationRequest("SELECT", classification, "tool", "", 5*time.Minute)
	req2 := queue.NewConfirmationRequest("INSERT", classification, "tool", "", 5*time.Minute)
	req3 := queue.NewConfirmationRequest("DELETE", classification, "tool", "", 5*time.Minute)
	_ = queue.NewConfirmationRequest("DROP", classification, "tool", "", 5*time.Minute)
	req5 := queue.NewConfirmationRequest("TRUNCATE", classification, "tool", "", 5*time.Minute)

	// Set different states
	queue.ConfirmRequest(req1.ID, "user")       // CONFIRMED
	queue.DenyRequest(req2.ID, "user", "")      // DENIED
	queue.CancelRequest(req3.ID, "claude", "")  // CANCELLED
	// req4 stays PENDING
	queue.requests[req5.ID].Status = "TIMEOUT" // Manually set TIMEOUT

	// Verify counts
	assert.Equal(t, 1, len(queue.GetPendingConfirmations()), "Should have 1 pending")
	assert.Equal(t, 1, len(queue.GetApprovedConfirmations()), "Should have 1 approved")
	assert.Equal(t, 1, len(queue.GetDeniedConfirmations()), "Should have 1 denied")
	assert.Equal(t, 1, len(queue.GetCancelledConfirmations()), "Should have 1 cancelled")

	// Total should be 5
	assert.Equal(t, 5, queue.Size())
}

// TestWaitForConfirmation_Cancelled tests waiting on cancelled request
func TestWaitForConfirmation_Cancelled(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	req := queue.NewConfirmationRequest(
		"DELETE FROM users",
		classification,
		"submit_query_plan",
		"",
		5*time.Minute,
	)

	// Cancel in background
	go func() {
		time.Sleep(100 * time.Millisecond)
		queue.CancelRequest(req.ID, "claude", "test")
	}()

	// Wait for confirmation - should return false with cancelled error
	confirmed, err := queue.WaitForConfirmation(req.ID, 50*time.Millisecond)

	assert.False(t, confirmed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
}

// TestGetConfirmationsByState_EmptyQueue tests getting confirmations from empty queue
func TestGetConfirmationsByState_EmptyQueue(t *testing.T) {
	queue := NewConfirmationQueue()

	assert.Equal(t, 0, len(queue.GetPendingConfirmations()))
	assert.Equal(t, 0, len(queue.GetApprovedConfirmations()))
	assert.Equal(t, 0, len(queue.GetDeniedConfirmations()))
	assert.Equal(t, 0, len(queue.GetCancelledConfirmations()))
}

// TestConfirmationLifecycle_Complete tests full lifecycle: create → confirm/deny/cancel
func TestConfirmationLifecycle_Complete(t *testing.T) {
	queue := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	// Test PENDING → CONFIRMED
	req1 := queue.NewConfirmationRequest("SELECT", classification, "tool", "", 5*time.Minute)
	assert.Equal(t, "PENDING", req1.Status)
	queue.ConfirmRequest(req1.ID, "user")
	r1, _ := queue.GetRequest(req1.ID)
	assert.Equal(t, "CONFIRMED", r1.Status)

	// Test PENDING → DENIED
	req2 := queue.NewConfirmationRequest("INSERT", classification, "tool", "", 5*time.Minute)
	assert.Equal(t, "PENDING", req2.Status)
	queue.DenyRequest(req2.ID, "user", "reason")
	r2, _ := queue.GetRequest(req2.ID)
	assert.Equal(t, "DENIED", r2.Status)

	// Test PENDING → CANCELLED
	req3 := queue.NewConfirmationRequest("DELETE", classification, "tool", "", 5*time.Minute)
	assert.Equal(t, "PENDING", req3.Status)
	queue.CancelRequest(req3.ID, "claude", "reason")
	r3, _ := queue.GetRequest(req3.ID)
	assert.Equal(t, "CANCELLED", r3.Status)

	// Test PENDING → TIMEOUT
	req4 := queue.NewConfirmationRequest("DROP", classification, "tool", "", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	queue.CleanupExpired()
	r4, _ := queue.GetRequest(req4.ID)
	assert.Equal(t, "TIMEOUT", r4.Status)

	// Test CONFIRMED → CANCELLED (cancel after confirmation)
	req5 := queue.NewConfirmationRequest("TRUNCATE", classification, "tool", "", 5*time.Minute)
	queue.ConfirmRequest(req5.ID, "user")
	queue.CancelRequest(req5.ID, "claude", "changed mind")
	r5, _ := queue.GetRequest(req5.ID)
	assert.Equal(t, "CANCELLED", r5.Status)
}
