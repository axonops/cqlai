package router

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

// copyOptionsWorld holds per-scenario state for COPY-options BDD scenarios.
//
// These steps live inside the `router` package (not test/bdd/steps) so they
// can call the unexported parseCopyOptions helper directly. The companion
// .feature file is still under test/bdd/features for discoverability.
type copyOptionsWorld struct {
	opts map[string]string
}

func (w *copyOptionsWorld) iParseCopyOptions(input string) error {
	w.opts = parseCopyOptions(input)
	return nil
}

func (w *copyOptionsWorld) iParseCopyOptionsDocString(doc *godog.DocString) error {
	w.opts = parseCopyOptions(doc.Content)
	return nil
}

func (w *copyOptionsWorld) copyOptionIs(key, expected string) error {
	got, ok := w.opts[key]
	if !ok {
		return fmt.Errorf("option %q not present in %#v", key, w.opts)
	}
	if got != expected {
		return fmt.Errorf("option %q: got %q, want %q", key, got, expected)
	}
	return nil
}

func (w *copyOptionsWorld) copyOptionIsDefaultDoubleQuote(key string) error {
	got, ok := w.opts[key]
	if !ok {
		return fmt.Errorf("option %q not present in %#v", key, w.opts)
	}
	if got != "\"" {
		return fmt.Errorf("option %q: got %q, want %q", key, got, "\"")
	}
	return nil
}

func initCopyOptionsScenario(ctx *godog.ScenarioContext) {
	w := &copyOptionsWorld{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*w = copyOptionsWorld{}
		return ctx, nil
	})

	ctx.Step(`^I parse the COPY options "([^"]*)"$`, w.iParseCopyOptions)
	ctx.Step(`^I parse the COPY options doc string:$`, w.iParseCopyOptionsDocString)
	ctx.Step(`^the COPY option "([^"]*)" is "([^"]*)"$`, w.copyOptionIs)
	ctx.Step(`^the COPY option "([^"]*)" is the default double-quote character$`, w.copyOptionIsDefaultDoubleQuote)
}

func TestCopyOptions_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "copy-options",
		ScenarioInitializer: initCopyOptionsScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../test/bdd/features/copy-options.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
