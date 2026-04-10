package steps_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/ui"
	"github.com/cucumber/godog"
)

// sslFlagsWorld holds scenario-scoped state for SSL flag BDD scenarios.
// Each scenario gets a fresh instance via the Before hook.
type sslFlagsWorld struct {
	// baseCfg is the config.Config as it would be loaded from a config file.
	baseCfg *config.Config

	// connOptions represents the ConnectionOptions built from parsed CLI flags.
	connOptions ui.ConnectionOptions

	// resolvedSSL is the SSLConfig after the CLI-override merge has been applied.
	resolvedSSL *config.SSLConfig

	// ptrResult holds the *bool produced by a pointer-building step.
	ptrResult *bool
}

// applySSLOverrides replicates the merge logic that exists in both
// internal/ui/model.go (NewMainModelWithConnectionOptions) and
// internal/batch/executor.go (NewExecutor).
//
// This is the observable contract under test: given a base SSLConfig and a
// ConnectionOptions with optional pointer overrides, what does the final
// SSLConfig look like?
func applySSLOverrides(base *config.SSLConfig, opts ui.ConnectionOptions) *config.SSLConfig {
	// Work on a shallow copy so tests don't mutate the background state.
	var out config.SSLConfig
	if base != nil {
		out = *base
	}

	if opts.SSLHostVerification != nil || opts.SSLInsecureSkipVerify != nil {
		if opts.SSLHostVerification != nil {
			out.HostVerification = *opts.SSLHostVerification
		}
		if opts.SSLInsecureSkipVerify != nil {
			out.InsecureSkipVerify = *opts.SSLInsecureSkipVerify
		}
	}

	return &out
}

// buildSSLHostVerificationPtr replicates the pointer-building logic from main.go
// lines 264-268 without touching pflag or os.Exit.
//
// flagChanged  — whether --no-ssl-host-verification was explicitly passed on the CLI.
// flagValue    — the boolean value the flag holds (true = no host verification).
// envValue     — the raw value of CQLAI_NO_SSL_HOST_VERIFICATION ("" if unset).
//
// Returns nil when neither the flag nor the env var was active, matching the
// "nil = use config file value" semantics in ConnectionOptions.SSLHostVerification.
func buildSSLHostVerificationPtr(flagChanged bool, flagValue bool, envValue string) *bool {
	effectiveValue := flagValue
	if !flagChanged && envValue != "" {
		effectiveValue = envValue == "true" || envValue == "1"
	}

	if flagChanged || envValue != "" {
		v := !effectiveValue // --no-ssl-host-verification inverts the flag value
		return &v
	}

	return nil
}

// buildSSLInsecureSkipVerifyPtr replicates the pointer-building logic from main.go
// lines 269-273.
func buildSSLInsecureSkipVerifyPtr(flagChanged bool, flagValue bool, envValue string) *bool {
	effectiveValue := flagValue
	if !flagChanged && envValue != "" {
		effectiveValue = envValue == "true" || envValue == "1"
	}

	if flagChanged || envValue != "" {
		v := effectiveValue
		return &v
	}

	return nil
}

// ---------------------------------------------------------------------------
// Background step
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) iHaveABaseSSLConfigWithHostVerificationEnabledAndInsecureSkipVerifyDisabled() {
	w.baseCfg = &config.Config{
		SSL: &config.SSLConfig{
			Enabled:            true,
			HostVerification:   true,
			InsecureSkipVerify: false,
		},
	}
}

// ---------------------------------------------------------------------------
// "Given the config file has …" steps
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) theConfigFileHasHostVerificationSetTo(value string) error {
	if w.baseCfg.SSL == nil {
		w.baseCfg.SSL = &config.SSLConfig{}
	}

	switch value {
	case "true":
		w.baseCfg.SSL.HostVerification = true
	case "false":
		w.baseCfg.SSL.HostVerification = false
	default:
		return fmt.Errorf("unexpected host verification value %q: want \"true\" or \"false\"", value)
	}

	return nil
}

func (w *sslFlagsWorld) theConfigFileHasInsecureSkipVerifySetTo(value string) error {
	if w.baseCfg.SSL == nil {
		w.baseCfg.SSL = &config.SSLConfig{}
	}

	switch value {
	case "true":
		w.baseCfg.SSL.InsecureSkipVerify = true
	case "false":
		w.baseCfg.SSL.InsecureSkipVerify = false
	default:
		return fmt.Errorf("unexpected insecure skip verify value %q: want \"true\" or \"false\"", value)
	}

	return nil
}

func (w *sslFlagsWorld) theConfigFileHasNoSSLSection() {
	w.baseCfg.SSL = nil
}

// ---------------------------------------------------------------------------
// "When the connection option … pointer is set to …" steps
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) theConnectionOptionSSLHostVerificationPointerIsSetTo(value string) error {
	switch value {
	case "true":
		v := true
		w.connOptions.SSLHostVerification = &v
	case "false":
		v := false
		w.connOptions.SSLHostVerification = &v
	default:
		return fmt.Errorf("unexpected SSLHostVerification pointer value %q: want \"true\" or \"false\"", value)
	}

	w.resolvedSSL = applySSLOverrides(w.baseCfg.SSL, w.connOptions)

	return nil
}

func (w *sslFlagsWorld) theConnectionOptionSSLInsecureSkipVerifyPointerIsSetTo(value string) error {
	switch value {
	case "true":
		v := true
		w.connOptions.SSLInsecureSkipVerify = &v
	case "false":
		v := false
		w.connOptions.SSLInsecureSkipVerify = &v
	default:
		return fmt.Errorf("unexpected SSLInsecureSkipVerify pointer value %q: want \"true\" or \"false\"", value)
	}

	w.resolvedSSL = applySSLOverrides(w.baseCfg.SSL, w.connOptions)

	return nil
}

func (w *sslFlagsWorld) noSSLHostVerificationConnectionOptionIsSet() {
	w.connOptions.SSLHostVerification = nil
	w.resolvedSSL = applySSLOverrides(w.baseCfg.SSL, w.connOptions)
}

func (w *sslFlagsWorld) noSSLInsecureSkipVerifyConnectionOptionIsSet() {
	w.connOptions.SSLInsecureSkipVerify = nil
	w.resolvedSSL = applySSLOverrides(w.baseCfg.SSL, w.connOptions)
}

// ---------------------------------------------------------------------------
// "Then the resolved SSL config has …" steps
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) theResolvedSSLConfigHasHostVerification(expected string) error {
	if w.resolvedSSL == nil {
		return fmt.Errorf("resolvedSSL is nil; the 'When' step did not run correctly")
	}

	want := expected == "true"
	got := w.resolvedSSL.HostVerification

	if got != want {
		return fmt.Errorf("SSL HostVerification: got %v, want %v", got, want)
	}

	return nil
}

func (w *sslFlagsWorld) theResolvedSSLConfigHasInsecureSkipVerify(expected string) error {
	if w.resolvedSSL == nil {
		return fmt.Errorf("resolvedSSL is nil; the 'When' step did not run correctly")
	}

	want := expected == "true"
	got := w.resolvedSSL.InsecureSkipVerify

	if got != want {
		return fmt.Errorf("SSL InsecureSkipVerify: got %v, want %v", got, want)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Env-var pointer-building steps
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) theEnvVarIsSetTo(varName, value string) error {
	return os.Setenv(varName, value)
}

func (w *sslFlagsWorld) theEnvVarIsNotSet(varName string) error {
	return os.Unsetenv(varName)
}

func (w *sslFlagsWorld) buildSSLHostVerificationPtrReturnsANonNilPointer() error {
	envVal := os.Getenv("CQLAI_NO_SSL_HOST_VERIFICATION")
	flagChanged := false // no CLI flag in env-var-only scenarios
	w.ptrResult = buildSSLHostVerificationPtr(flagChanged, false, envVal)

	if w.ptrResult == nil {
		return fmt.Errorf("expected a non-nil *bool from buildSSLHostVerificationPtr, got nil")
	}

	return nil
}

func (w *sslFlagsWorld) buildSSLInsecureSkipVerifyPtrReturnsANonNilPointer() error {
	envVal := os.Getenv("CQLAI_SSL_INSECURE_SKIP_VERIFY")
	flagChanged := false
	w.ptrResult = buildSSLInsecureSkipVerifyPtr(flagChanged, false, envVal)

	if w.ptrResult == nil {
		return fmt.Errorf("expected a non-nil *bool from buildSSLInsecureSkipVerifyPtr, got nil")
	}

	return nil
}

func (w *sslFlagsWorld) buildSSLHostVerificationPtrReturnsANilPointer() error {
	envVal := os.Getenv("CQLAI_NO_SSL_HOST_VERIFICATION")
	flagChanged := false
	w.ptrResult = buildSSLHostVerificationPtr(flagChanged, false, envVal)

	if w.ptrResult != nil {
		return fmt.Errorf("expected nil *bool from buildSSLHostVerificationPtr, got %v", *w.ptrResult)
	}

	return nil
}

func (w *sslFlagsWorld) buildSSLInsecureSkipVerifyPtrReturnsANilPointer() error {
	envVal := os.Getenv("CQLAI_SSL_INSECURE_SKIP_VERIFY")
	flagChanged := false
	w.ptrResult = buildSSLInsecureSkipVerifyPtr(flagChanged, false, envVal)

	if w.ptrResult != nil {
		return fmt.Errorf("expected nil *bool from buildSSLInsecureSkipVerifyPtr, got %v", *w.ptrResult)
	}

	return nil
}

func (w *sslFlagsWorld) thePointedToHostVerificationValueIs(expected string) error {
	if w.ptrResult == nil {
		return fmt.Errorf("ptrResult is nil; the preceding 'Then' step must have failed")
	}

	want := expected == "true"
	got := *w.ptrResult

	if got != want {
		return fmt.Errorf("host verification pointer value: got %v, want %v", got, want)
	}

	return nil
}

func (w *sslFlagsWorld) thePointedToInsecureSkipVerifyValueIs(expected string) error {
	if w.ptrResult == nil {
		return fmt.Errorf("ptrResult is nil; the preceding 'Then' step must have failed")
	}

	want := expected == "true"
	got := *w.ptrResult

	if got != want {
		return fmt.Errorf("insecure skip verify pointer value: got %v, want %v", got, want)
	}

	return nil
}

// ---------------------------------------------------------------------------
// CLI-flag-wins-over-env-var steps
// ---------------------------------------------------------------------------

func (w *sslFlagsWorld) theCLIFlagNoSSLHostVerificationIsExplicitlySetTo(flagValueStr string) error {
	flagValue := flagValueStr == "true"
	envVal := os.Getenv("CQLAI_NO_SSL_HOST_VERIFICATION")

	// flagChanged=true simulates --no-ssl-host-verification being explicitly passed.
	w.ptrResult = buildSSLHostVerificationPtr(true, flagValue, envVal)

	return nil
}

func (w *sslFlagsWorld) theCLIFlagSSLInsecureSkipVerifyIsExplicitlySetTo(flagValueStr string) error {
	flagValue := flagValueStr == "true"
	envVal := os.Getenv("CQLAI_SSL_INSECURE_SKIP_VERIFY")

	// flagChanged=true simulates --ssl-insecure-skip-verify being explicitly passed.
	w.ptrResult = buildSSLInsecureSkipVerifyPtr(true, flagValue, envVal)

	return nil
}

func (w *sslFlagsWorld) theResolvedHostVerificationOverridePointerValueIs(expected string) error {
	if w.ptrResult == nil {
		return fmt.Errorf("ptrResult is nil; expected a non-nil override pointer")
	}

	want := expected == "true"
	got := *w.ptrResult

	if got != want {
		return fmt.Errorf("host verification override pointer value: got %v, want %v", got, want)
	}

	return nil
}

func (w *sslFlagsWorld) theResolvedInsecureSkipVerifyOverridePointerValueIs(expected string) error {
	if w.ptrResult == nil {
		return fmt.Errorf("ptrResult is nil; expected a non-nil override pointer")
	}

	want := expected == "true"
	got := *w.ptrResult

	if got != want {
		return fmt.Errorf("insecure skip verify override pointer value: got %v, want %v", got, want)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Step registration
// ---------------------------------------------------------------------------

func initSSLFlagsScenario(ctx *godog.ScenarioContext) {
	w := &sslFlagsWorld{}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Reset world state and clear env vars before each scenario so tests are
		// fully isolated — no scenario can pollute the next one.
		*w = sslFlagsWorld{}
		_ = os.Unsetenv("CQLAI_NO_SSL_HOST_VERIFICATION")
		_ = os.Unsetenv("CQLAI_SSL_INSECURE_SKIP_VERIFY")
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Ensure env vars are cleaned up even if the scenario fails mid-way.
		_ = os.Unsetenv("CQLAI_NO_SSL_HOST_VERIFICATION")
		_ = os.Unsetenv("CQLAI_SSL_INSECURE_SKIP_VERIFY")
		return ctx, nil
	})

	// Background
	ctx.Step(`^I have a base SSL config with host verification enabled and insecure skip verify disabled$`,
		w.iHaveABaseSSLConfigWithHostVerificationEnabledAndInsecureSkipVerifyDisabled)

	// Given — config file state
	ctx.Step(`^the config file has host verification set to (true|false)$`,
		w.theConfigFileHasHostVerificationSetTo)
	ctx.Step(`^the config file has insecure skip verify set to (true|false)$`,
		w.theConfigFileHasInsecureSkipVerifySetTo)
	ctx.Step(`^the config file has no SSL section$`,
		w.theConfigFileHasNoSSLSection)

	// When — connection option pointer mutations
	ctx.Step(`^the connection option SSLHostVerification pointer is set to (true|false)$`,
		w.theConnectionOptionSSLHostVerificationPointerIsSetTo)
	ctx.Step(`^the connection option SSLInsecureSkipVerify pointer is set to (true|false)$`,
		w.theConnectionOptionSSLInsecureSkipVerifyPointerIsSetTo)
	ctx.Step(`^no SSLHostVerification connection option is set$`,
		w.noSSLHostVerificationConnectionOptionIsSet)
	ctx.Step(`^no SSLInsecureSkipVerify connection option is set$`,
		w.noSSLInsecureSkipVerifyConnectionOptionIsSet)

	// When — env var manipulation
	ctx.Step(`^the env var (CQLAI_NO_SSL_HOST_VERIFICATION|CQLAI_SSL_INSECURE_SKIP_VERIFY) is set to "([^"]*)"$`,
		w.theEnvVarIsSetTo)
	ctx.Step(`^the env var (CQLAI_NO_SSL_HOST_VERIFICATION|CQLAI_SSL_INSECURE_SKIP_VERIFY) is not set$`,
		w.theEnvVarIsNotSet)

	// When — CLI flag explicitly set (wins over env var)
	ctx.Step(`^the CLI flag no-ssl-host-verification is explicitly set to "(true|false)"$`,
		w.theCLIFlagNoSSLHostVerificationIsExplicitlySetTo)
	ctx.Step(`^the CLI flag ssl-insecure-skip-verify is explicitly set to "(true|false)"$`,
		w.theCLIFlagSSLInsecureSkipVerifyIsExplicitlySetTo)

	// Then — resolved SSL config assertions
	ctx.Step(`^the resolved SSL config has host verification (true|false)$`,
		w.theResolvedSSLConfigHasHostVerification)
	ctx.Step(`^the resolved SSL config has insecure skip verify (true|false)$`,
		w.theResolvedSSLConfigHasInsecureSkipVerify)

	// Then — pointer-building assertions
	ctx.Step(`^buildSSLHostVerificationPtr returns a non-nil pointer$`,
		w.buildSSLHostVerificationPtrReturnsANonNilPointer)
	ctx.Step(`^buildSSLInsecureSkipVerifyPtr returns a non-nil pointer$`,
		w.buildSSLInsecureSkipVerifyPtrReturnsANonNilPointer)
	ctx.Step(`^buildSSLHostVerificationPtr returns a nil pointer$`,
		w.buildSSLHostVerificationPtrReturnsANilPointer)
	ctx.Step(`^buildSSLInsecureSkipVerifyPtr returns a nil pointer$`,
		w.buildSSLInsecureSkipVerifyPtrReturnsANilPointer)
	ctx.Step(`^the pointed-to host verification value is (true|false)$`,
		w.thePointedToHostVerificationValueIs)
	ctx.Step(`^the pointed-to insecure skip verify value is (true|false)$`,
		w.thePointedToInsecureSkipVerifyValueIs)

	// Then — CLI flag wins-over-env assertions
	ctx.Step(`^the resolved host verification override pointer value is (true|false)$`,
		w.theResolvedHostVerificationOverridePointerValueIs)
	ctx.Step(`^the resolved insecure skip verify override pointer value is (true|false)$`,
		w.theResolvedInsecureSkipVerifyOverridePointerValueIs)
}

// ---------------------------------------------------------------------------
// Test runner
// ---------------------------------------------------------------------------

func TestSSLFlags_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name: "ssl-flags",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			initSSLFlagsScenario(ctx)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/ssl-flags.feature"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("BDD scenarios failed — see output above for details")
	}
}
