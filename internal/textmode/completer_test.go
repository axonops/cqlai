package textmode

import (
	"testing"
)

// mockCompleter is a minimal Completer implementation for tests.
type mockCompleter struct {
	suggestions []string
}

func (m *mockCompleter) Complete(_ string) []string {
	return m.suggestions
}

func TestReadlineCompleter_Do_NoSuggestions(t *testing.T) {
	rc := NewReadlineCompleter(&mockCompleter{suggestions: nil})
	newLine, length := rc.Do([]rune("SEL"), 3)
	if len(newLine) != 0 {
		t.Errorf("expected no completions, got %v", newLine)
	}
	if length != 0 {
		t.Errorf("expected length 0, got %d", length)
	}
}

func TestReadlineCompleter_Do_WithSuggestions(t *testing.T) {
	// Engine returns ["SELECT", "SHOW"] for partial "SE".
	// Only "SELECT" starts with "SE"; "SHOW" does not and is filtered out.
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"SELECT", "SHOW"}})
	line := []rune("SE")
	newLine, length := rc.Do(line, 2)

	// length should be 2 (the word "SE" being completed)
	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}
	// Only SELECT starts with "SE", so exactly one completion.
	if len(newLine) != 1 {
		t.Errorf("expected 1 completion (SELECT), got %d: %v", len(newLine), newLine)
	}
	// The returned suffix must be "LECT" (not the full "SELECT").
	if len(newLine) == 1 {
		got := string(newLine[0])
		if got != "LECT" {
			t.Errorf("expected suffix \"LECT\", got %q", got)
		}
	}
}

func TestReadlineCompleter_Do_SuffixOnly(t *testing.T) {
	// Typing "SEL" with suggestion "SELECT" — suffix must be "ECT".
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"SELECT"}})
	line := []rune("SEL")
	newLine, length := rc.Do(line, 3)

	if length != 3 {
		t.Errorf("expected length 3, got %d", length)
	}
	if len(newLine) != 1 {
		t.Fatalf("expected 1 completion, got %d", len(newLine))
	}
	got := string(newLine[0])
	if got != "ECT" {
		t.Errorf("expected suffix \"ECT\", got %q", got)
	}
}

func TestReadlineCompleter_Do_ExactMatch_EmptySuffix(t *testing.T) {
	// Typing the full keyword "SELECT" — suffix should be empty (already complete).
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"SELECT"}})
	line := []rune("SELECT")
	newLine, length := rc.Do(line, 6)

	if length != 6 {
		t.Errorf("expected length 6, got %d", length)
	}
	if len(newLine) != 1 {
		t.Fatalf("expected 1 completion, got %d", len(newLine))
	}
	if len(newLine[0]) != 0 {
		t.Errorf("expected empty suffix for exact match, got %q", string(newLine[0]))
	}
}

func TestReadlineCompleter_Do_CaseInsensitive(t *testing.T) {
	// Lower-case input "sel" should match "SELECT".
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"SELECT"}})
	line := []rune("sel")
	newLine, length := rc.Do(line, 3)

	if length != 3 {
		t.Errorf("expected length 3, got %d", length)
	}
	if len(newLine) != 1 {
		t.Fatalf("expected 1 completion, got %d", len(newLine))
	}
	// Suffix comes from the original suggestion "SELECT", so suffix is "ECT".
	got := string(newLine[0])
	if got != "ECT" {
		t.Errorf("expected suffix \"ECT\", got %q", got)
	}
}

func TestReadlineCompleter_Do_WithBufferPrefix(t *testing.T) {
	// Buffer already has "SELECT *\nFROM " and user types "sy"
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"system"}})
	rc.SetBufferPrefix("SELECT *\nFROM ")

	line := []rune("sy")
	newLine, length := rc.Do(line, 2)

	if length != 2 {
		t.Errorf("expected length 2, got %d", length)
	}
	if len(newLine) == 0 {
		t.Fatal("expected at least one completion")
	}
	// Suffix of "system" after "sy" is "stem".
	got := string(newLine[0])
	if got != "stem" {
		t.Errorf("expected suffix \"stem\", got %q", got)
	}
}

func TestReadlineCompleter_Do_CursorMidLine(t *testing.T) {
	// line has more content than cursor pos — only up to pos matters
	rc := NewReadlineCompleter(&mockCompleter{suggestions: []string{"SELECT"}})
	line := []rune("SELECTxyz")
	// cursor is at position 6 — "SELECT"
	newLine, length := rc.Do(line, 6)

	if length != 6 {
		t.Errorf("expected length 6 (word 'SELECT'), got %d", length)
	}
	// Exact match at cursor pos 6 — empty suffix.
	if len(newLine) != 1 {
		t.Fatalf("expected 1 completion, got %d", len(newLine))
	}
	if len(newLine[0]) != 0 {
		t.Errorf("expected empty suffix for exact match, got %q", string(newLine[0]))
	}
}
