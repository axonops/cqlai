package ai

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfirmationQueue tests queue creation
func TestNewConfirmationQueue(t *testing.T) {
	q := NewConfirmationQueue()

	assert.NotNil(t, q)
	assert.Equal(t, 0, q.Size())
	assert.NotNil(t, q.requests)
}

// TestConfirmationQueue_NewRequest tests creating confirmation requests
func TestConfirmationQueue_NewRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{
		IsDangerous: true,
		Severity:    SeverityCritical,
		Operation:   "DELETE",
		Description: "Dangerous delete operation",
	}

	req := q.NewConfirmationRequest(
		"DELETE FROM users WHERE id = 1",
		classification,
		"submit_query_plan",
		"",
		5*time.Minute,
	)

	assert.NotNil(t, req)
	assert.Equal(t, "req_001", req.ID)
	assert.Equal(t, "DELETE FROM users WHERE id = 1", req.Query)
	assert.Equal(t, "PENDING", req.Status)
	assert.Equal(t, SeverityCritical, req.Classification.Severity)
	assert.False(t, req.UserConfirmed)
	assert.Equal(t, 1, q.Size())

	// Create second request - should get req_002
	req2 := q.NewConfirmationRequest("DROP TABLE test", classification, "tool", "", 5*time.Minute)
	assert.Equal(t, "req_002", req2.ID)
	assert.Equal(t, 2, q.Size())
}

// TestConfirmationQueue_ConfirmRequest tests confirming a request
func TestConfirmationQueue_ConfirmRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{
		IsDangerous: true,
		Severity:    SeverityHigh,
	}

	req := q.NewConfirmationRequest("DELETE FROM events", classification, "tool", "", 5*time.Minute)

	// Confirm the request
	err := q.ConfirmRequest(req.ID, "cassandra")
	require.NoError(t, err)

	// Verify status updated
	updated, err := q.GetRequest(req.ID)
	require.NoError(t, err)
	assert.Equal(t, "CONFIRMED", updated.Status)
	assert.True(t, updated.UserConfirmed)
	assert.Equal(t, "cassandra", updated.ConfirmedBy)
	assert.False(t, updated.ConfirmedAt.IsZero())
}

// TestConfirmationQueue_DenyRequest tests denying a request
func TestConfirmationQueue_DenyRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{
		IsDangerous: true,
		Severity:    SeverityHigh,
	}

	req := q.NewConfirmationRequest("DELETE FROM events", classification, "tool", "", 5*time.Minute)

	// Deny the request
	err := q.DenyRequest(req.ID, "cassandra", "Too risky")
	require.NoError(t, err)

	// Verify status updated
	updated, err := q.GetRequest(req.ID)
	require.NoError(t, err)
	assert.Equal(t, "DENIED", updated.Status)
	assert.False(t, updated.UserConfirmed)
	assert.Equal(t, "cassandra", updated.ConfirmedBy)
}

// TestConfirmationQueue_GetRequest tests retrieving requests
func TestConfirmationQueue_GetRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}

	req1 := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)

	// Get existing request
	got, err := q.GetRequest(req1.ID)
	require.NoError(t, err)
	assert.Equal(t, req1.ID, got.ID)
	assert.Equal(t, req1.Query, got.Query)

	// Get non-existent request
	_, err = q.GetRequest("req_999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestConfirmationQueue_GetPendingConfirmations tests listing pending requests
func TestConfirmationQueue_GetPendingConfirmations(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}

	// Create 3 requests
	req1 := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)
	req2 := q.NewConfirmationRequest("DROP TABLE test", classification, "tool", "", 5*time.Minute)
	req3 := q.NewConfirmationRequest("TRUNCATE events", classification, "tool", "", 5*time.Minute)

	// All should be pending
	pending := q.GetPendingConfirmations()
	assert.Equal(t, 3, len(pending))

	// Confirm one
	q.ConfirmRequest(req1.ID, "user")

	pending = q.GetPendingConfirmations()
	assert.Equal(t, 2, len(pending))

	// Deny one
	q.DenyRequest(req2.ID, "user", "too risky")

	pending = q.GetPendingConfirmations()
	assert.Equal(t, 1, len(pending))
	assert.Equal(t, req3.ID, pending[0].ID)
}

// TestConfirmationQueue_CleanupExpired tests expiring old requests
func TestConfirmationQueue_CleanupExpired(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}

	// Create request with very short timeout
	req1 := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 10*time.Millisecond)

	// Create request with long timeout
	req2 := q.NewConfirmationRequest("DROP TABLE test", classification, "tool", "", 1*time.Hour)

	// Wait for first to expire
	time.Sleep(50 * time.Millisecond)

	// Cleanup expired
	expired := q.CleanupExpired()
	assert.Equal(t, 1, expired)

	// Verify first request is timed out
	updated1, _ := q.GetRequest(req1.ID)
	assert.Equal(t, "TIMEOUT", updated1.Status)

	// Verify second request is still pending
	updated2, _ := q.GetRequest(req2.ID)
	assert.Equal(t, "PENDING", updated2.Status)
}

// TestConfirmationQueue_RemoveRequest tests removing requests
func TestConfirmationQueue_RemoveRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}
	req := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)

	assert.Equal(t, 1, q.Size())

	// Remove the request
	err := q.RemoveRequest(req.ID)
	require.NoError(t, err)

	assert.Equal(t, 0, q.Size())

	// Try to remove again - should error
	err = q.RemoveRequest(req.ID)
	assert.Error(t, err)
}

// TestConfirmationQueue_DoubleConfirm tests confirming already confirmed request
func TestConfirmationQueue_DoubleConfirm(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}
	req := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 5*time.Minute)

	// Confirm once
	err := q.ConfirmRequest(req.ID, "user1")
	require.NoError(t, err)

	// Try to confirm again - should error
	err = q.ConfirmRequest(req.ID, "user2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not pending")
}

// TestConfirmationQueue_ConfirmExpiredRequest tests confirming expired request
func TestConfirmationQueue_ConfirmExpiredRequest(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}

	// Create request with very short timeout
	req := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", 10*time.Millisecond)

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Try to confirm - should error
	err := q.ConfirmRequest(req.ID, "user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")

	// Verify status is TIMEOUT
	updated, _ := q.GetRequest(req.ID)
	assert.Equal(t, "TIMEOUT", updated.Status)
}

// TestConfirmationQueue_ConcurrentAccess tests thread safety
func TestConfirmationQueue_ConcurrentAccess(t *testing.T) {
	q := NewConfirmationQueue()
	classification := QueryClassification{IsDangerous: true}

	done := make(chan bool)

	// Goroutine 1: Create requests
	go func() {
		for i := 0; i < 50; i++ {
			q.NewConfirmationRequest(
				fmt.Sprintf("DELETE FROM table_%d", i),
				classification,
				"tool",
				"",
				5*time.Minute,
			)
		}
		done <- true
	}()

	// Goroutine 2: Confirm requests
	go func() {
		time.Sleep(10 * time.Millisecond) // Let some requests get created
		for i := 1; i <= 25; i++ {
			q.ConfirmRequest(fmt.Sprintf("req_%03d", i), "user")
		}
		done <- true
	}()

	// Goroutine 3: Get pending
	go func() {
		for i := 0; i < 20; i++ {
			q.GetPendingConfirmations()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Verify queue is consistent
	assert.Equal(t, 50, q.Size(), "Should have all 50 requests")

	pending := q.GetPendingConfirmations()
	assert.LessOrEqual(t, len(pending), 50, "Pending count should be reasonable")
}

// TestConfirmationQueue_Timeout tests timeout calculation
func TestConfirmationQueue_Timeout(t *testing.T) {
	q := NewConfirmationQueue()

	classification := QueryClassification{IsDangerous: true}

	timeout := 2 * time.Second
	req := q.NewConfirmationRequest("DELETE FROM users", classification, "tool", "", timeout)

	// Verify timeout is set correctly
	expectedTimeout := req.Timestamp.Add(timeout)
	assert.Equal(t, expectedTimeout, req.Timeout)

	// Verify not expired immediately
	assert.True(t, time.Now().Before(req.Timeout))
}
