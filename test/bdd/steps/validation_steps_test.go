package steps_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/axonops/cqlai/internal/validation"
	"github.com/cucumber/godog"
)

// validationWorld holds per-scenario state for command-validation BDD scenarios.
type validationWorld struct {
	validateErr error
	dangerous   bool
}

func (w *validationWorld) iValidateTheCommand(cmd string) error {
	w.validateErr = validation.ValidateCommandSyntax(cmd)
	return nil
}

func (w *validationWorld) iValidateAWhitespaceOnlyCommand() error {
	// Picked deliberately to mix spaces, tabs and newlines — the validator
	// must treat the whole thing as an empty no-op.
	w.validateErr = validation.ValidateCommandSyntax("   \t\n  ")
	return nil
}

func (w *validationWorld) theCommandIsAccepted() error {
	if w.validateErr != nil {
		return fmt.Errorf("expected accepted, got error: %v", w.validateErr)
	}
	return nil
}

func (w *validationWorld) theCommandIsRejectedWithErrorContaining(needle string) error {
	if w.validateErr == nil {
		return fmt.Errorf("expected rejection containing %q, got nil error", needle)
	}
	if !strings.Contains(w.validateErr.Error(), needle) {
		return fmt.Errorf("error %q does not contain %q", w.validateErr.Error(), needle)
	}
	return nil
}

func (w *validationWorld) theErrorExplainsCommandIsNotRecognised() error {
	if w.validateErr == nil {
		return fmt.Errorf("expected error, got nil")
	}
	msg := strings.ToLower(w.validateErr.Error())
	if !strings.Contains(msg, "not a recognized") {
		return fmt.Errorf("error %q does not explain that the command is unrecognised", w.validateErr.Error())
	}
	return nil
}

func (w *validationWorld) iCheckWhetherIsDangerous(cmd string) error {
	w.dangerous = validation.IsDangerousCommand(cmd)
	return nil
}

func (w *validationWorld) theCommandIsMarkedDangerous() error {
	if !w.dangerous {
		return fmt.Errorf("expected dangerous=true, got false")
	}
	return nil
}

func (w *validationWorld) theCommandIsNotMarkedDangerous() error {
	if w.dangerous {
		return fmt.Errorf("expected dangerous=false, got true")
	}
	return nil
}

func initValidationScenario(ctx *godog.ScenarioContext) {
	w := &validationWorld{}
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		*w = validationWorld{}
		return ctx, nil
	})

	ctx.Step(`^I validate the cqlai command "([^"]*)"$`, w.iValidateTheCommand)
	ctx.Step(`^I validate the whitespace-only cqlai command$`, w.iValidateAWhitespaceOnlyCommand)
	ctx.Step(`^the cqlai command is accepted$`, w.theCommandIsAccepted)
	ctx.Step(`^the cqlai command is rejected with an error containing "([^"]*)"$`, w.theCommandIsRejectedWithErrorContaining)
	ctx.Step(`^the cqlai error explains the command is not recognised$`, w.theErrorExplainsCommandIsNotRecognised)
	ctx.Step(`^I check whether the cqlai command "([^"]*)" is dangerous$`, w.iCheckWhetherIsDangerous)
	ctx.Step(`^the cqlai command is marked dangerous$`, w.theCommandIsMarkedDangerous)
	ctx.Step(`^the cqlai command is not marked dangerous$`, w.theCommandIsNotMarkedDangerous)
}

func TestCommandValidation_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "command-validation",
		ScenarioInitializer: initValidationScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/command-validation.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
