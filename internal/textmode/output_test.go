package textmode

import (
	"bytes"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/db"
)

func newTestPrinter(format string, pageSize int) *printer {
	return newPrinter(Options{
		Format:   format,
		PageSize: pageSize,
		FieldSep: ",",
	})
}

func TestPrintResult_String(t *testing.T) {
	p := newTestPrinter("ascii", 100)
	var buf bytes.Buffer
	exit, err := p.printResult(&buf, "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exit {
		t.Fatal("unexpected exit signal")
	}
	if got := strings.TrimSpace(buf.String()); got != "hello world" {
		t.Errorf("got %q, want \"hello world\"", got)
	}
}

func TestPrintResult_EmptyString(t *testing.T) {
	p := newTestPrinter("ascii", 100)
	var buf bytes.Buffer
	exit, err := p.printResult(&buf, "")
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestPrintResult_Nil(t *testing.T) {
	p := newTestPrinter("ascii", 100)
	var buf bytes.Buffer
	exit, err := p.printResult(&buf, nil)
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
}

func TestPrintResult_QueryResult_ASCII(t *testing.T) {
	p := newTestPrinter("ascii", 100)
	var buf bytes.Buffer
	r := db.QueryResult{
		Data: [][]string{
			{"name", "age"},
			{"Alice", "30"},
			{"Bob", "25"},
		},
	}
	exit, err := p.printResult(&buf, r)
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
	out := buf.String()
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected 'Alice' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "(2 rows)") {
		t.Errorf("expected '(2 rows)' in output, got:\n%s", out)
	}
}

func TestPrintResult_QueryResult_JSON(t *testing.T) {
	p := newTestPrinter("json", 100)
	var buf bytes.Buffer
	r := db.QueryResult{
		Data: [][]string{
			{"name"},
			{"Alice"},
		},
	}
	exit, err := p.printResult(&buf, r)
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
	out := buf.String()
	if !strings.Contains(out, `"columns"`) {
		t.Errorf("expected JSON with 'columns', got:\n%s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected 'Alice' in JSON output, got:\n%s", out)
	}
}

func TestPrintResult_QueryResult_CSV(t *testing.T) {
	p := newTestPrinter("csv", 100)
	var buf bytes.Buffer
	r := db.QueryResult{
		Data: [][]string{
			{"name", "age"},
			{"Alice", "30"},
		},
	}
	exit, err := p.printResult(&buf, r)
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
	out := buf.String()
	if !strings.Contains(out, "name,age") {
		t.Errorf("expected CSV header 'name,age', got:\n%s", out)
	}
	if !strings.Contains(out, "Alice,30") {
		t.Errorf("expected CSV row 'Alice,30', got:\n%s", out)
	}
}

func TestPrintResult_SliceOfSlices(t *testing.T) {
	p := newTestPrinter("ascii", 100)
	var buf bytes.Buffer
	data := [][]string{
		{"col1", "col2"},
		{"a", "b"},
	}
	exit, err := p.printResult(&buf, data)
	if err != nil || exit {
		t.Fatalf("unexpected error=%v exit=%v", err, exit)
	}
	out := buf.String()
	if !strings.Contains(out, "col1") {
		t.Errorf("expected table output containing 'col1', got:\n%s", out)
	}
}

func TestStripTableFooter(t *testing.T) {
	input := "+---+\n| a |\n+---+\n\n(2 rows)\n"
	got := stripTableFooter(input, 2)
	if strings.Contains(got, "(2 rows)") {
		t.Errorf("stripTableFooter should have removed row count, got:\n%s", got)
	}
	if strings.HasSuffix(got, "+---+\n") {
		// If bottom border was removed, good.
		// (It's ok if it was removed)
	}
}
