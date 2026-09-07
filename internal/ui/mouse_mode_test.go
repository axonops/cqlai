package ui

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs f and returns whatever it printed to stdout.
//
// The mouse mode helpers write escape sequences straight to stdout, which is the
// only thing worth asserting about them.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	orig := os.Stdout
	os.Stdout = w
	f()
	os.Stdout = orig

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(out)
}

// TestInitEnablesAlternateScrollNotMouseReporting guards the fix for the mouse
// selection bug. Turning on mouse reporting (DECSET 1000) makes the terminal
// hand us every click, which kills its own text selection and right-click paste.
// Init must ask for alternate scroll mode instead.
func TestInitEnablesAlternateScrollNotMouseReporting(t *testing.T) {
	m := &MainModel{}

	out := captureStdout(t, func() {
		m.Init()
	})

	if !strings.Contains(out, "\x1b[?1007h") {
		t.Errorf("Init did not enable alternate scroll mode (DECSET 1007); wrote %q", out)
	}

	for _, banned := range []string{"\x1b[?1000h", "\x1b[?1002h", "\x1b[?1003h", "\x1b[?1006h"} {
		if strings.Contains(out, banned) {
			t.Errorf("Init enabled mouse reporting %q, which breaks terminal text selection and right-click paste", banned)
		}
	}
}

func TestDisableAlternateScroll(t *testing.T) {
	out := captureStdout(t, DisableAlternateScroll)

	if !strings.Contains(out, "\x1b[?1007l") {
		t.Errorf("DisableAlternateScroll wrote %q, want it to reset DECSET 1007", out)
	}
}

// TestViewportOwnsArrows covers which views treat a bare Up/Down as scrolling.
// Alternate scroll mode delivers wheel spins as Up/Down key presses, so this is
// also the answer to "what does the wheel do here".
func TestViewportOwnsArrows(t *testing.T) {
	tests := []struct {
		name     string
		model    MainModel
		expected bool
	}{
		{
			name:     "table view with results scrolls",
			model:    MainModel{viewMode: "table", hasTable: true},
			expected: true,
		},
		{
			name:     "table view without results recalls history",
			model:    MainModel{viewMode: "table"},
			expected: false,
		},
		{
			name:     "trace view with a trace scrolls",
			model:    MainModel{viewMode: "trace", hasTrace: true},
			expected: true,
		},
		{
			name:     "trace view without a trace recalls history",
			model:    MainModel{viewMode: "trace"},
			expected: false,
		},
		{
			name:     "normal view recalls history",
			model:    MainModel{viewMode: "history"},
			expected: false,
		},
		{
			// handleAIConversationInput takes Up/Down before handleUpArrow sees
			// them and scrolls the conversation itself.
			name:     "ai view is handled elsewhere",
			model:    MainModel{viewMode: "ai", aiConversationActive: true},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.viewportOwnsArrows(); got != tt.expected {
				t.Errorf("viewportOwnsArrows() = %v, want %v", got, tt.expected)
			}
		})
	}
}
