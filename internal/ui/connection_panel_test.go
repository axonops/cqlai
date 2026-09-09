package ui

import (
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectionRowsShowEveryFact(t *testing.T) {
	rows := connectionRows(db.ConnectionInfo{
		Version:  "5.0.9",
		Protocol: 5,
		Username: "cassandra",
		Host:     "10.0.0.4:9042",
	})

	joined := strings.Join(rows, "\n")
	for _, want := range []string{"Cassandra", "5.0.9", "Protocol", "v5", "User", "cassandra", "Host", "10.0.0.4:9042", "Encryption"} {
		assert.Contains(t, joined, want)
	}
}

// TestConnectionRowsLineUp: the values are in a column, so the panel reads as a
// table rather than as ragged text.
func TestConnectionRowsLineUp(t *testing.T) {
	rows := connectionRows(db.ConnectionInfo{Version: "5.0.9", Protocol: 5, Username: "u", Host: "h"})

	// The value starts after the run of spaces that follows the label.
	var at []int
	for _, row := range rows {
		i := strings.Index(row, "  ")
		for i < len(row) && row[i] == ' ' {
			i++
		}
		at = append(at, i)
	}
	for i := range at {
		assert.Equal(t, at[0], at[i], "row %d starts its value in a different column: %q", i, rows[i])
	}
}

// TestUnencryptedSaysSo rather than leaving the row blank, which reads as a
// missing fact instead of an answer.
func TestUnencryptedSaysSo(t *testing.T) {
	assert.Equal(t, "none", encryptionSummary(db.ConnectionInfo{}))
}

// TestUnverifiedEncryptionIsCalledOut is the point of showing this at all. TLS
// with certificate checking turned off is encrypted and still open to anything
// in the middle of the connection, so the panel has to say which it is.
func TestUnverifiedEncryptionIsCalledOut(t *testing.T) {
	verified := encryptionSummary(db.ConnectionInfo{Encrypted: true, HostVerified: true})
	assert.Contains(t, verified, "TLS")
	assert.Contains(t, verified, "certificate verified")
	assert.NotContains(t, verified, "NOT")

	unverified := encryptionSummary(db.ConnectionInfo{Encrypted: true})
	assert.Contains(t, unverified, "NOT verified")
}

func TestEncryptionExtrasAreListed(t *testing.T) {
	summary := encryptionSummary(db.ConnectionInfo{
		Encrypted:         true,
		HostVerified:      true,
		CustomCA:          true,
		ClientCertificate: true,
	})

	assert.Contains(t, summary, "own CA")
	assert.Contains(t, summary, "client certificate")
}

// TestMissingFactsSayUnknown: an empty value would leave a row that looks
// truncated.
func TestMissingFactsSayUnknown(t *testing.T) {
	rows := connectionRows(db.ConnectionInfo{})

	for _, row := range rows {
		assert.NotRegexp(t, `\s+$`, row, "row has no value at all: %q", row)
	}
	assert.Contains(t, strings.Join(rows, "\n"), "unknown")
}

// TestConnectionIsClickable: it replaced three fields that were not, so the
// facts they carried have to be reachable.
func TestConnectionIsClickable(t *testing.T) {
	m := testStatusBar()

	var found bool
	for _, seg := range placeSegments(m.segments(), statusTestWidth) {
		if seg.setting == settingConnection {
			found = true
			setting, _, ok := m.settingAt(statusTestWidth, (seg.start+seg.end)/2)
			require.True(t, ok)
			assert.Equal(t, settingConnection, setting)
		}
		assert.NotEqual(t, "User: ", seg.label, "the user moved into the panel")
		assert.NotEqual(t, "Host: ", seg.label, "the host moved into the panel")
		assert.NotEqual(t, "v", seg.label, "the version moved into the panel")
	}
	require.True(t, found, "there is no Connection field on the status line")
}

// TestConnectionPanelPicksNothing: its rows are facts, so clicking one closes
// the panel and changes nothing.
func TestConnectionPanelPicksNothing(t *testing.T) {
	assert.Equal(t, "", settingCommand(settingConnection, "Cassandra   5.0.9"))
}

// TestConnectionPanelNeedsASession: with nothing connected there is nothing to
// report, and an empty box would be worse than none.
func TestConnectionPanelNeedsASession(t *testing.T) {
	m := chooserModel()
	m.openConnectionPanel(4)

	assert.False(t, m.chooser.active)
}
