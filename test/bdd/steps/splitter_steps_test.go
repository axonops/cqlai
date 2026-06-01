package steps_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/batch"
	"github.com/cucumber/godog"
)

// splitterWorld holds per-scenario state for cql-splitter BDD scenarios.
type splitterWorld struct {
	statements []string
	splitErr   error
	incomplete bool

	shellResult bool
}

func (w *splitterWorld) splitInput(input string) {
	// We call SplitStatements (not SplitForNode) because we also want to
	// assert on the Incomplete flag for unterminated input.
	res, err := batch.SplitStatements(input)
	w.splitErr = err
	if res != nil {
		w.statements = res.GetStatementStrings()
		w.incomplete = res.Incomplete
	} else {
		w.statements = nil
		w.incomplete = false
	}
}

func (w *splitterWorld) iSplitTheCQLInputQuoted(input string) error {
	w.splitInput(input)
	return nil
}

func (w *splitterWorld) iSplitTheCQLInputDocString(doc *godog.DocString) error {
	w.splitInput(doc.Content)
	return nil
}

func (w *splitterWorld) iSplitWhitespaceOnlyInput() error {
	w.splitInput("   \t\n  ")
	return nil
}

func (w *splitterWorld) theSplitterYieldsNStatements(n int) error {
	if w.splitErr != nil {
		return fmt.Errorf("splitter returned error: %v", w.splitErr)
	}
	if len(w.statements) != n {
		return fmt.Errorf("expected %d statements, got %d: %#v", n, len(w.statements), w.statements)
	}
	return nil
}

func (w *splitterWorld) splitStatementIs(idx int, expected string) error {
	if idx < 1 || idx > len(w.statements) {
		return fmt.Errorf("statement index %d out of range (have %d)", idx, len(w.statements))
	}
	got := strings.TrimSpace(w.statements[idx-1])
	want := strings.TrimSpace(expected)
	if got != want {
		return fmt.Errorf("statement %d:\n  got:  %q\n  want: %q", idx, got, want)
	}
	return nil
}

func (w *splitterWorld) splitStatementContains(idx int, needle string) error {
	if idx < 1 || idx > len(w.statements) {
		return fmt.Errorf("statement index %d out of range (have %d)", idx, len(w.statements))
	}
	if !strings.Contains(w.statements[idx-1], needle) {
		return fmt.Errorf("statement %d (%q) does not contain %q", idx, w.statements[idx-1], needle)
	}
	return nil
}

func (w *splitterWorld) theSplitterMarksInputIncomplete() error {
	if w.splitErr != nil {
		return fmt.Errorf("splitter returned unexpected error: %v", w.splitErr)
	}
	if !w.incomplete {
		return fmt.Errorf("expected Incomplete=true, got false (statements=%#v)", w.statements)
	}
	return nil
}

func (w *splitterWorld) iCheckShellCommand(cmd string) error {
	w.shellResult = batch.IsShellCommand(cmd)
	return nil
}

func (w *splitterWorld) theShellCommandResultIs(expected string) error {
	want := expected == "true"
	if w.shellResult != want {
		return fmt.Errorf("IsShellCommand: got %v, want %v", w.shellResult, want)
	}
	return nil
}

func initSplitterScenario(ctx *godog.ScenarioContext) {
	w := &splitterWorld{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*w = splitterWorld{}
		return ctx, nil
	})

	ctx.Step(`^I split the CQL input "([^"]*)"$`, w.iSplitTheCQLInputQuoted)
	ctx.Step(`^I split the CQL input:$`, w.iSplitTheCQLInputDocString)
	ctx.Step(`^I split the whitespace-only CQL input$`, w.iSplitWhitespaceOnlyInput)

	ctx.Step(`^the splitter yields (\d+) statements?$`, w.theSplitterYieldsNStatements)
	ctx.Step(`^split statement (\d+) is "([^"]*)"$`, w.splitStatementIs)
	ctx.Step(`^split statement (\d+) contains "([^"]*)"$`, w.splitStatementContains)
	ctx.Step(`^the splitter marks the input as incomplete$`, w.theSplitterMarksInputIncomplete)

	ctx.Step(`^I check whether "([^"]*)" is a cqlai shell command$`, w.iCheckShellCommand)
	ctx.Step(`^the shell command result is (true|false)$`, w.theShellCommandResultIs)
}

func TestCQLSplitter_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "cql-splitter",
		ScenarioInitializer: initSplitterScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/cql-splitter.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
