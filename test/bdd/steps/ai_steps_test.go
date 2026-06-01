package steps_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/ai"
	"github.com/cucumber/godog"
)

// aiWorld holds per-scenario state for AI-command-parser BDD scenarios.
type aiWorld struct {
	tool    ai.ToolName
	arg     string
	parsed  bool
	calledP bool // sentinel — guards "tool/arg" assertions if parse not called
}

func (w *aiWorld) iParseTheAIResponse(doc *godog.DocString) error {
	// Doc strings come in with the leading/trailing whitespace pattern Gherkin
	// stripped to the indentation level. Trim to be defensive.
	tool, arg, ok := ai.ParseCommand(strings.TrimSpace(doc.Content))
	w.tool = tool
	w.arg = arg
	w.parsed = ok
	w.calledP = true
	return nil
}

func (w *aiWorld) aiResponseRecognisedAsCommand() error {
	if !w.calledP {
		return fmt.Errorf("ParseCommand was not invoked in this scenario")
	}
	if !w.parsed {
		return fmt.Errorf("expected parsed=true, got false")
	}
	return nil
}

func (w *aiWorld) aiResponseNotRecognised() error {
	if !w.calledP {
		return fmt.Errorf("ParseCommand was not invoked in this scenario")
	}
	if w.parsed {
		return fmt.Errorf("expected parsed=false, got true (tool=%q arg=%q)", w.tool, w.arg)
	}
	return nil
}

func (w *aiWorld) aiCommandToolIs(expected string) error {
	if string(w.tool) != expected {
		return fmt.Errorf("tool: got %q, want %q", w.tool, expected)
	}
	return nil
}

func (w *aiWorld) aiCommandArgIs(expected string) error {
	if w.arg != expected {
		return fmt.Errorf("arg: got %q, want %q", w.arg, expected)
	}
	return nil
}

func initAIScenario(ctx *godog.ScenarioContext) {
	w := &aiWorld{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*w = aiWorld{}
		return ctx, nil
	})

	ctx.Step(`^I parse the AI response:$`, w.iParseTheAIResponse)
	ctx.Step(`^the AI response is recognised as a command$`, w.aiResponseRecognisedAsCommand)
	ctx.Step(`^the AI response is not recognised as a command$`, w.aiResponseNotRecognised)
	ctx.Step(`^the AI command tool is "([^"]*)"$`, w.aiCommandToolIs)
	ctx.Step(`^the AI command arg is "([^"]*)"$`, w.aiCommandArgIs)
}

func TestAICommandParser_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "ai-command-parser",
		ScenarioInitializer: initAIScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/ai-command-parser.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
