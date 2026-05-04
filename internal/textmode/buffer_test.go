package textmode

import (
	"testing"
)

func TestInputBuffer_SingleStatement(t *testing.T) {
	buf := &InputBuffer{}
	buf.Add("SELECT * FROM system.local;")
	if !buf.IsComplete() {
		t.Error("expected IsComplete() true for single terminated statement")
	}
}

func TestInputBuffer_MultiLineStatement(t *testing.T) {
	buf := &InputBuffer{}
	buf.Add("SELECT *")
	if buf.IsComplete() {
		t.Error("expected IsComplete() false for incomplete first line")
	}
	buf.Add("FROM system.local;")
	if !buf.IsComplete() {
		t.Error("expected IsComplete() true after adding terminating line")
	}
}

func TestInputBuffer_EmptyIsNotComplete(t *testing.T) {
	buf := &InputBuffer{}
	if buf.IsComplete() {
		t.Error("expected IsComplete() false for empty buffer")
	}
}

func TestInputBuffer_Reset(t *testing.T) {
	buf := &InputBuffer{}
	buf.Add("SELECT 1;")
	buf.Reset()
	if !buf.IsEmpty() {
		t.Error("expected IsEmpty() after Reset()")
	}
}

func TestInputBuffer_Text(t *testing.T) {
	buf := &InputBuffer{}
	buf.Add("line1")
	buf.Add("line2")
	want := "line1\nline2"
	if got := buf.Text(); got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestIsMeta(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"HELP", true},
		{"help", true},
		{"EXIT;", true},
		{"QUIT", true},
		{"DESCRIBE keyspaces", true},
		{"SELECT * FROM t;", false},
		{"", false},
		{"CONSISTENCY LOCAL_QUORUM", true},
		// Word boundary: "SHOWING" must NOT match "SHOW"
		{"SHOWING tables", false},
		// Word boundary: bare SHOW should match
		{"SHOW version", true},
		{"SHOW", true},
		// DESCRIPTION must NOT match DESC (no space after)
		// actually "DESCRIPTION" starts with "DESC" — but "DESC" is in the list
		// and "DESCRIPTION" != "DESC" and "DESCRIPTION" does not start with "DESC "
		{"DESCRIPTION", false},
	}
	for _, tc := range cases {
		got := IsMeta(tc.input)
		if got != tc.want {
			t.Errorf("IsMeta(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseUseKeyspace(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"USE myks;", "myks"},
		{"use myks;", "myks"},
		{`USE "MyKeyspace";`, "MyKeyspace"},
		{`USE "MyKeyspace"`, "MyKeyspace"},
		{"USE myks", "myks"},
		{"SELECT * FROM t;", ""},
		{"", ""},
		// Extra spaces
		{"  USE   myks  ;  ", "myks"},
	}
	for _, tc := range cases {
		got := parseUseKeyspace(tc.input)
		if got != tc.want {
			t.Errorf("parseUseKeyspace(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestContinuationPrompt(t *testing.T) {
	// With keyspace "system": primary = "cqlai:system> " (14 chars)
	// continuation should be 10 spaces + "... "
	primary := buildPrompt("system")
	cont := continuationPrompt("system")
	if len([]rune(primary)) != len([]rune(cont)) {
		t.Errorf("primary len=%d cont len=%d, should be equal", len([]rune(primary)), len([]rune(cont)))
	}
	if !endsWith(cont, "... ") {
		t.Errorf("continuationPrompt should end with '... ', got %q", cont)
	}
}

func TestBuildPrompt(t *testing.T) {
	if got := buildPrompt(""); got != "cqlai> " {
		t.Errorf("buildPrompt(\"\") = %q, want \"cqlai> \"", got)
	}
	if got := buildPrompt("ks"); got != "cqlai:ks> " {
		t.Errorf("buildPrompt(\"ks\") = %q, want \"cqlai:ks> \"", got)
	}
}

func endsWith(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
