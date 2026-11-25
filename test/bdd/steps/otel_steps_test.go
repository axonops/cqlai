package steps_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/telemetry"
	"github.com/cucumber/godog"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// otelWorld holds per-scenario state for OpenTelemetry tracing scenarios.
type otelWorld struct {
	recorder *tracetest.SpanRecorder
	observer *telemetry.Observer
	startErr error
	cleanup  []func()
}

// cassandraError stands in for an error frame from Cassandra: the driver's
// own types keep their code in unexported fields.
type cassandraError struct {
	code    int
	message string
}

func (e cassandraError) Code() int       { return e.code }
func (e cassandraError) Message() string { return e.message }
func (e cassandraError) Error() string   { return e.message }

func (w *otelWorld) setEnv(name, value string) {
	old, had := os.LookupEnv(name)
	w.cleanup = append(w.cleanup, func() {
		if had {
			_ = os.Setenv(name, old)
		} else {
			_ = os.Unsetenv(name)
		}
	})
	_ = os.Setenv(name, value)
}

func (w *otelWorld) tracingHasNotBeenStarted() error {
	return telemetry.Shutdown(context.Background())
}

func (w *otelWorld) theOTLPProtocolIs(protocol string) error {
	w.setEnv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	w.setEnv("OTEL_EXPORTER_OTLP_PROTOCOL", protocol)
	// Nothing listens here, and nothing is sent: the exporters do not
	// connect until there is a span to send.
	w.setEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:1")
	return nil
}

func (w *otelWorld) tracingIsStarted() error {
	w.cleanup = append(w.cleanup, func() { _ = telemetry.Shutdown(context.Background()) })
	w.startErr = telemetry.Start(context.Background())
	return nil
}

func (w *otelWorld) tracingIsStopped() error {
	return telemetry.Shutdown(context.Background())
}

func (w *otelWorld) tracingIsRunning() error {
	if w.startErr != nil {
		return fmt.Errorf("start failed: %w", w.startErr)
	}
	if !telemetry.Started() || telemetry.TracerProvider() == nil {
		return errors.New("tracing is not running")
	}
	return nil
}

func (w *otelWorld) startingFailsWith(want string) error {
	if w.startErr == nil {
		return errors.New("expected start to fail, it did not")
	}
	if !strings.Contains(w.startErr.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.startErr, want)
	}
	return nil
}

func (w *otelWorld) noTracerProvider() error {
	if telemetry.Started() || telemetry.TracerProvider() != nil {
		return errors.New("a tracer provider is set up")
	}
	return nil
}

func (w *otelWorld) stoppingDoesNothing() error {
	return telemetry.Shutdown(context.Background())
}

func (w *otelWorld) aSpanRecorder() error {
	w.recorder = tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(w.recorder))
	w.cleanup = append(w.cleanup, func() { _ = tp.Shutdown(context.Background()) })
	w.observer = telemetry.NewObserver(tp)
	return nil
}

func testHost() *gocql.HostInfo {
	h, err := gocql.NewHostInfoFromAddrPort(net.ParseIP("10.0.0.5"), 9042)
	if err != nil {
		panic(err)
	}
	return h
}

func (w *otelWorld) observeQuery(q gocql.ObservedQuery) {
	q.Start = time.Now().Add(-5 * time.Millisecond)
	q.End = time.Now()
	w.observer.ObserveQuery(context.Background(), q)
}

func (w *otelWorld) queryRuns(stmt, keyspace string, rows int) error {
	w.observeQuery(gocql.ObservedQuery{Statement: stmt, Keyspace: keyspace, Rows: rows, Host: testHost()})
	return nil
}

func (w *otelWorld) queryRunsWithNoHost(stmt, keyspace string) error {
	w.observeQuery(gocql.ObservedQuery{Statement: stmt, Keyspace: keyspace})
	return nil
}

func (w *otelWorld) attemptOfQueryRuns(attempt int, stmt, keyspace string) error {
	w.observeQuery(gocql.ObservedQuery{Statement: stmt, Keyspace: keyspace, Attempt: attempt, Host: testHost()})
	return nil
}

func (w *otelWorld) queryFailsInCassandra(code int, message string) error {
	w.observeQuery(gocql.ObservedQuery{
		Statement: "INSERT INTO t (id) VALUES ('hunter2')",
		Keyspace:  "shop",
		Host:      testHost(),
		Err:       cassandraError{code: code, message: message},
	})
	return nil
}

func (w *otelWorld) queryFailsBeforeCassandra() error {
	w.observeQuery(gocql.ObservedQuery{Statement: "SELECT * FROM t", Err: errors.New("no connections")})
	return nil
}

func (w *otelWorld) batchRuns(n int, keyspace string) error {
	w.observer.ObserveBatch(context.Background(), gocql.ObservedBatch{
		Keyspace:   keyspace,
		Statements: make([]string, n),
		Start:      time.Now().Add(-5 * time.Millisecond),
		End:        time.Now(),
		Host:       testHost(),
	})
	return nil
}

func (w *otelWorld) span() (sdktrace.ReadOnlySpan, error) {
	spans := w.recorder.Ended()
	if len(spans) != 1 {
		return nil, fmt.Errorf("expected 1 span, got %d", len(spans))
	}
	return spans[0], nil
}

func (w *otelWorld) oneSpanNamed(name string) error {
	s, err := w.span()
	if err != nil {
		return err
	}
	if s.Name() != name {
		return fmt.Errorf("span named %q, want %q", s.Name(), name)
	}
	return nil
}

func (w *otelWorld) spanIsClient() error {
	s, err := w.span()
	if err != nil {
		return err
	}
	if s.SpanKind() != trace.SpanKindClient {
		return fmt.Errorf("span kind %v, want client", s.SpanKind())
	}
	return nil
}

func (w *otelWorld) attribute(key string) (string, bool, error) {
	s, err := w.span()
	if err != nil {
		return "", false, err
	}
	for _, kv := range s.Attributes() {
		if string(kv.Key) == key {
			return kv.Value.String(), true, nil
		}
	}
	return "", false, nil
}

func (w *otelWorld) spanHasAttribute(key, want string) error {
	got, ok, err := w.attribute(key)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("span has no attribute %q", key)
	}
	if got != want {
		return fmt.Errorf("attribute %q is %q, want %q", key, got, want)
	}
	return nil
}

func (w *otelWorld) spanHasNoAttribute(key string) error {
	got, ok, err := w.attribute(key)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("span has attribute %q = %q", key, got)
	}
	return nil
}

func (w *otelWorld) noAttributeContains(text string) error {
	for _, s := range w.recorder.Ended() {
		if strings.Contains(s.Name(), text) || strings.Contains(s.Status().Description, text) {
			return fmt.Errorf("span %q carries %q", s.Name(), text)
		}
		for _, kv := range s.Attributes() {
			if strings.Contains(kv.Value.String(), text) {
				return fmt.Errorf("attribute %q carries %q", kv.Key, text)
			}
		}
	}
	return nil
}

func (w *otelWorld) spanHasErrorStatus() error {
	s, err := w.span()
	if err != nil {
		return err
	}
	if s.Status().Code != codes.Error {
		return fmt.Errorf("status %v, want error", s.Status().Code)
	}
	return nil
}

func (w *otelWorld) spanHasNoErrorStatus() error {
	s, err := w.span()
	if err != nil {
		return err
	}
	if s.Status().Code == codes.Error {
		return errors.New("span has an error status")
	}
	return nil
}

func initOTelScenario(ctx *godog.ScenarioContext) {
	w := &otelWorld{}

	ctx.After(func(c context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		for i := len(w.cleanup) - 1; i >= 0; i-- {
			w.cleanup[i]()
		}
		return c, nil
	})

	ctx.Step(`^OpenTelemetry tracing has not been started$`, w.tracingHasNotBeenStarted)
	ctx.Step(`^the OTLP protocol is "([^"]*)"$`, w.theOTLPProtocolIs)
	ctx.Step(`^OpenTelemetry tracing is started$`, w.tracingIsStarted)
	ctx.Step(`^tracing is stopped$`, w.tracingIsStopped)
	ctx.Step(`^tracing is running$`, w.tracingIsRunning)
	ctx.Step(`^starting tracing fails with "([^"]*)"$`, w.startingFailsWith)
	ctx.Step(`^no tracer provider is set up$`, w.noTracerProvider)
	ctx.Step(`^stopping tracing does nothing$`, w.stoppingDoesNothing)
	ctx.Step(`^a span recorder$`, w.aSpanRecorder)
	ctx.Step(`^the query "([^"]*)" runs in keyspace "([^"]*)" and returns (\d+) rows$`, w.queryRuns)
	ctx.Step(`^the query "([^"]*)" runs in keyspace "([^"]*)" with no host$`, w.queryRunsWithNoHost)
	ctx.Step(`^attempt (\d+) of the query "([^"]*)" runs in keyspace "([^"]*)"$`, w.attemptOfQueryRuns)
	ctx.Step(`^a query fails in Cassandra with code (\d+) and message "([^"]*)"$`, w.queryFailsInCassandra)
	ctx.Step(`^a query fails before reaching Cassandra$`, w.queryFailsBeforeCassandra)
	ctx.Step(`^a batch of (\d+) statements runs in keyspace "([^"]*)"$`, w.batchRuns)
	ctx.Step(`^one span is recorded named "([^"]*)"$`, w.oneSpanNamed)
	ctx.Step(`^the span is a client span$`, w.spanIsClient)
	ctx.Step(`^the span has attribute "([^"]*)" set to "([^"]*)"$`, w.spanHasAttribute)
	ctx.Step(`^the span has no attribute "([^"]*)"$`, w.spanHasNoAttribute)
	ctx.Step(`^no span attribute contains "([^"]*)"$`, w.noAttributeContains)
	ctx.Step(`^the span has an error status$`, w.spanHasErrorStatus)
	ctx.Step(`^the span has no error status$`, w.spanHasNoErrorStatus)
}

func TestOTelTracing_BDD(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "otel-tracing",
		ScenarioInitializer: initOTelScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features/otel-tracing.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
