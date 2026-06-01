package steps_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/router"
	"github.com/cucumber/godog"
)

// saveWorld holds per-scenario state for SAVE-command BDD scenarios.
type saveWorld struct {
	cmd *router.SaveCommand
	err error
}

func (w *saveWorld) parse(input string) {
	w.cmd, w.err = router.ParseSaveCommand(input)
}

func (w *saveWorld) iParseTheSaveCommand(input string) error {
	w.parse(input)
	return nil
}

func (w *saveWorld) iParseTheSaveCommandDocString(doc *godog.DocString) error {
	w.parse(strings.TrimSpace(doc.Content))
	return nil
}

func (w *saveWorld) saveParsedOK() error {
	if w.err != nil {
		return fmt.Errorf("expected success, got error: %v", w.err)
	}
	if w.cmd == nil {
		return fmt.Errorf("expected SaveCommand, got nil")
	}
	return nil
}

func (w *saveWorld) saveIsInteractive() error {
	if !w.cmd.Interactive {
		return fmt.Errorf("expected Interactive=true, got false")
	}
	return nil
}

func (w *saveWorld) saveHasFilename(expected string) error {
	if w.cmd.Filename != expected {
		return fmt.Errorf("filename: got %q, want %q", w.cmd.Filename, expected)
	}
	return nil
}

func (w *saveWorld) saveHasFormat(expected string) error {
	if w.cmd.Format != expected {
		return fmt.Errorf("format: got %q, want %q", w.cmd.Format, expected)
	}
	return nil
}

func (w *saveWorld) saveOptionEqualsBool(key, expected string) error {
	raw, ok := w.cmd.Options[key]
	if !ok {
		return fmt.Errorf("option %q not present in %#v", key, w.cmd.Options)
	}
	got, ok := raw.(bool)
	if !ok {
		return fmt.Errorf("option %q is not a bool (got %T = %v)", key, raw, raw)
	}
	want := expected == "true"
	if got != want {
		return fmt.Errorf("option %q: got %v, want %v", key, got, want)
	}
	return nil
}

func (w *saveWorld) saveOptionEqualsString(key, expected string) error {
	raw, ok := w.cmd.Options[key]
	if !ok {
		return fmt.Errorf("option %q not present in %#v", key, w.cmd.Options)
	}
	got, ok := raw.(string)
	if !ok {
		return fmt.Errorf("option %q is not a string (got %T = %v)", key, raw, raw)
	}
	if got != expected {
		return fmt.Errorf("option %q: got %q, want %q", key, got, expected)
	}
	return nil
}

func (w *saveWorld) saveRejectedWith(needle string) error {
	if w.err == nil {
		return fmt.Errorf("expected rejection containing %q, got nil error (cmd=%#v)", needle, w.cmd)
	}
	if !strings.Contains(w.err.Error(), needle) {
		return fmt.Errorf("error %q does not contain %q", w.err.Error(), needle)
	}
	return nil
}

func initSaveScenario(ctx *godog.ScenarioContext) {
	w := &saveWorld{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*w = saveWorld{}
		return ctx, nil
	})

	ctx.Step(`^I parse the SAVE command "([^"]*)"$`, w.iParseTheSaveCommand)
	ctx.Step(`^I parse the SAVE command:$`, w.iParseTheSaveCommandDocString)

	ctx.Step(`^the SAVE command is parsed successfully$`, w.saveParsedOK)
	ctx.Step(`^the SAVE command is interactive$`, w.saveIsInteractive)
	ctx.Step(`^the SAVE command has filename "([^"]*)"$`, w.saveHasFilename)
	ctx.Step(`^the SAVE command has format "([^"]*)"$`, w.saveHasFormat)
	ctx.Step(`^the SAVE command option "([^"]*)" equals boolean "(true|false)"$`, w.saveOptionEqualsBool)
	ctx.Step(`^the SAVE command option "([^"]*)" equals string "([^"]*)"$`, w.saveOptionEqualsString)
	ctx.Step(`^the SAVE command is rejected with an error containing "([^"]*)"$`, w.saveRejectedWith)
}

func TestSaveCommand_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "save-command",
		ScenarioInitializer: initSaveScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/save-command.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
