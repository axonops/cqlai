package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/validation"
)

// Auditor writes one line for every tool call: what was asked, on which
// connection, and what the server decided. Refusals are written too - a run of
// them is how an attempt to steer the model shows up.
//
// No value from the data is written: statements are logged with their strings
// and numbers replaced, and no row is ever logged.
type Auditor struct {
	mu   sync.Mutex
	file *os.File
}

// AuditOff is the path that turns the log off.
const AuditOff = "-"

// OpenAudit opens the log for appending, creating it readable only by its
// owner. The path AuditOff gives an Auditor that writes nothing.
func OpenAudit(path string) (*Auditor, error) {
	if path == AuditOff {
		return &Auditor{}, nil
	}
	// The default log is in ~/.cassandra, which may not be there yet, and
	// may still be where an older cqlai wrote it.
	if path == DefaultAuditLog() {
		if err := config.MoveIn(path, oldAuditLog()); err != nil {
			return nil, fmt.Errorf("cannot open the audit log %s: %w", path, err)
		}
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) // #nosec G304 - the path the user gave for the log
	if err != nil {
		return nil, fmt.Errorf("cannot open the audit log %s: %w", path, err)
	}
	return &Auditor{file: f}, nil
}

// Entry is one call.
type Entry struct {
	Time       time.Time `json:"time"`
	Connection string    `json:"connection,omitempty"`
	Tool       string    `json:"tool"`
	Statement  string    `json:"statement,omitempty"`
	Decision   string    `json:"decision"` // "allowed", "refused" or "failed"
	Reason     string    `json:"reason,omitempty"`
	Rows       int       `json:"rows,omitempty"`
	DurationMS int64     `json:"duration_ms"`
	Confirmed  *bool     `json:"confirmed,omitempty"`
}

// Log writes an entry. The statement is stripped of its values here, so no
// caller can forget to.
func (a *Auditor) Log(e Entry) {
	if a == nil || a.file == nil {
		return
	}
	if e.Statement != "" {
		// A driver error repeats the value it could not use.
		e.Reason = validation.WithoutStatementValues(e.Reason, e.Statement)
		e.Statement = validation.WithoutValues(e.Statement)
	}
	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, _ = a.file.Write(append(line, '\n'))
}

// Close closes the log.
func (a *Auditor) Close() error {
	if a == nil || a.file == nil {
		return nil
	}
	return a.file.Close()
}
