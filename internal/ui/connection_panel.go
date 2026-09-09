package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/axonops/cqlai/internal/db"
)

// The panel behind Connection on the status line.
//
// It opens through the same machinery as the settings lists: same anchoring
// above the field, same geometry, same dismissal. The difference is that its
// rows are facts rather than values, so nothing is marked as current and
// clicking a row only closes the panel - settingCommand has nothing to return
// for Connection, so applySettingChoice does nothing else.

// connectionRows lays the connection facts out as one row per fact, with the
// values lined up.
func connectionRows(info db.ConnectionInfo) []string {
	type row struct{ label, value string }

	rows := []row{
		{"Cassandra", info.Version},
		{"Protocol", protocolVersion(info.Protocol)},
		{"User", info.Username},
		{"Host", info.Host},
		{"Encryption", encryptionSummary(info)},
	}

	width := 0
	for _, r := range rows {
		if w := lipgloss.Width(r.label); w > width {
			width = w
		}
	}

	out := make([]string, 0, len(rows))
	for _, r := range rows {
		value := r.value
		if value == "" {
			value = "unknown"
		}
		out = append(out, r.label+strings.Repeat(" ", width-lipgloss.Width(r.label)+2)+value)
	}
	return out
}

// protocolVersion renders the native protocol version the way the servers and
// cqlsh both write it.
func protocolVersion(v int) string {
	if v <= 0 {
		return ""
	}
	return fmt.Sprintf("v%d", v)
}

// encryptionSummary says whether the connection is encrypted and what is
// actually being checked.
//
// "TLS" on its own is not the useful fact. A connection with certificate
// verification turned off is encrypted and still wide open to anything sitting
// in the middle of it, so that is said plainly rather than left to be inferred
// from the absence of a note.
func encryptionSummary(info db.ConnectionInfo) string {
	if !info.Encrypted {
		return "none"
	}

	parts := []string{"TLS"}
	if info.HostVerified {
		parts = append(parts, "certificate verified")
	} else {
		parts = append(parts, "certificate NOT verified")
	}
	if info.CustomCA {
		parts = append(parts, "own CA")
	}
	if info.ClientCertificate {
		parts = append(parts, "client certificate")
	}
	return strings.Join(parts, ", ")
}

// openConnectionPanel shows the connection facts above the status line.
func (m *MainModel) openConnectionPanel(anchorX int) {
	if m.session == nil {
		return
	}
	// No current value: these are facts, so nothing in the panel is marked and
	// nothing can be applied.
	m.openSettingChooser(settingConnection, anchorX, connectionRows(m.session.ConnectionInfo()), "")
}
