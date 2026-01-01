package ai

import (
	"fmt"
	"sync"
	"time"
)

// ConfirmationQueue manages pending confirmation requests for dangerous queries
type ConfirmationQueue struct {
	requests map[string]*ConfirmationRequest
	mu       sync.RWMutex
	nextID   int
}

// NewConfirmationQueue creates a new confirmation queue
func NewConfirmationQueue() *ConfirmationQueue {
	return &ConfirmationQueue{
		requests: make(map[string]*ConfirmationRequest),
		nextID:   1,
	}
}

// NewConfirmationRequest creates a new confirmation request for a dangerous query.
// The request is added to the queue and waits for user approval.
func (q *ConfirmationQueue) NewConfirmationRequest(query string, classification QueryClassification, tool, toolOp string, timeout time.Duration) *ConfirmationRequest {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Generate unique ID
	id := fmt.Sprintf("req_%03d", q.nextID)
	q.nextID++

	now := time.Now()
	req := &ConfirmationRequest{
		ID:             id,
		Timestamp:      now,
		Timeout:        now.Add(timeout),
		Query:          query,
		Classification: classification,
		Tool:           tool,
		ToolOperation:  toolOp,
		Status:         "PENDING",
		UserConfirmed:  false,
	}

	q.requests[id] = req

	return req
}

// ConfirmRequest marks a request as confirmed by the user
func (q *ConfirmationQueue) ConfirmRequest(requestID, confirmedBy string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	req, exists := q.requests[requestID]
	if !exists {
		return fmt.Errorf("confirmation request %s not found", requestID)
	}

	if req.Status != "PENDING" {
		return fmt.Errorf("request %s is not pending (status: %s)", requestID, req.Status)
	}

	// Check if timed out
	if time.Now().After(req.Timeout) {
		req.Status = "TIMEOUT"
		return fmt.Errorf("request %s has expired", requestID)
	}

	req.Status = "CONFIRMED"
	req.UserConfirmed = true
	req.ConfirmedBy = confirmedBy
	req.ConfirmedAt = time.Now()

	return nil
}

// DenyRequest marks a request as denied by the user
func (q *ConfirmationQueue) DenyRequest(requestID, deniedBy, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	req, exists := q.requests[requestID]
	if !exists {
		return fmt.Errorf("confirmation request %s not found", requestID)
	}

	if req.Status != "PENDING" {
		return fmt.Errorf("request %s is not pending (status: %s)", requestID, req.Status)
	}

	req.Status = "DENIED"
	req.UserConfirmed = false
	req.ConfirmedBy = deniedBy
	req.ConfirmedAt = time.Now()

	return nil
}

// CancelRequest marks a request as cancelled
func (q *ConfirmationQueue) CancelRequest(requestID, cancelledBy, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	req, exists := q.requests[requestID]
	if !exists {
		return fmt.Errorf("confirmation request %s not found", requestID)
	}

	// Can cancel in any state (PENDING, CONFIRMED, DENIED, TIMEOUT)
	req.Status = "CANCELLED"
	req.UserConfirmed = false
	req.ConfirmedBy = cancelledBy
	req.ConfirmedAt = time.Now()

	return nil
}

// GetRequest retrieves a confirmation request by ID
func (q *ConfirmationQueue) GetRequest(requestID string) (*ConfirmationRequest, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	req, exists := q.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("request %s not found", requestID)
	}

	return req, nil
}

// GetPendingConfirmations returns all pending confirmation requests
func (q *ConfirmationQueue) GetPendingConfirmations() []*ConfirmationRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()

	pending := []*ConfirmationRequest{}
	for _, req := range q.requests {
		if req.Status == "PENDING" {
			pending = append(pending, req)
		}
	}

	return pending
}

// GetApprovedConfirmations returns all confirmed requests
func (q *ConfirmationQueue) GetApprovedConfirmations() []*ConfirmationRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()

	approved := []*ConfirmationRequest{}
	for _, req := range q.requests {
		if req.Status == "CONFIRMED" {
			approved = append(approved, req)
		}
	}

	return approved
}

// GetDeniedConfirmations returns all denied requests
func (q *ConfirmationQueue) GetDeniedConfirmations() []*ConfirmationRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()

	denied := []*ConfirmationRequest{}
	for _, req := range q.requests {
		if req.Status == "DENIED" {
			denied = append(denied, req)
		}
	}

	return denied
}

// GetCancelledConfirmations returns all cancelled requests
func (q *ConfirmationQueue) GetCancelledConfirmations() []*ConfirmationRequest {
	q.mu.RLock()
	defer q.mu.RUnlock()

	cancelled := []*ConfirmationRequest{}
	for _, req := range q.requests {
		if req.Status == "CANCELLED" {
			cancelled = append(cancelled, req)
		}
	}

	return cancelled
}

// CleanupExpired marks expired requests as timed out and returns their IDs
func (q *ConfirmationQueue) CleanupExpired() []string {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	var timedOutIDs []string

	for _, req := range q.requests {
		if req.Status == "PENDING" && now.After(req.Timeout) {
			req.Status = "TIMEOUT"
			timedOutIDs = append(timedOutIDs, req.ID)
		}
	}

	return timedOutIDs
}

// RemoveRequest removes a confirmation request from the queue
func (q *ConfirmationQueue) RemoveRequest(requestID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.requests[requestID]; !exists {
		return fmt.Errorf("request %s not found", requestID)
	}

	delete(q.requests, requestID)
	return nil
}

// Size returns the number of requests in the queue
func (q *ConfirmationQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.requests)
}

// WaitForConfirmation waits for a confirmation request to be confirmed or denied.
// Returns true if confirmed, false if denied or timed out.
// This is a blocking call that polls the request status.
func (q *ConfirmationQueue) WaitForConfirmation(requestID string, pollInterval time.Duration) (bool, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		req, err := q.GetRequest(requestID)
		if err != nil {
			return false, err
		}

		switch req.Status {
		case "CONFIRMED":
			return true, nil
		case "DENIED":
			return false, fmt.Errorf("user denied the operation")
		case "CANCELLED":
			return false, fmt.Errorf("confirmation request was cancelled")
		case "TIMEOUT":
			return false, fmt.Errorf("confirmation request timed out")
		case "PENDING":
			// Check if expired
			if time.Now().After(req.Timeout) {
				q.mu.Lock()
				req.Status = "TIMEOUT"
				q.mu.Unlock()
				return false, fmt.Errorf("confirmation request timed out")
			}
			// Wait for next poll
			<-ticker.C
		default:
			return false, fmt.Errorf("unknown status: %s", req.Status)
		}
	}
}
