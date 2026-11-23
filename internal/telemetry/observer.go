package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/axonops/cqlai/internal/telemetry"

// Observer turns the driver's record of each request into a span. It is set
// on the cluster as both its QueryObserver and its BatchObserver, and so sees
// every request the session makes, cqlai's own schema queries included.
//
// The driver reports a request once it is over, so the span is made after the
// fact with the times the driver took. A paged query is one request, and one
// span, per page; a retry is a span of its own.
//
// The statement text and its bound values are not recorded. cqlai runs CQL as
// typed, and that can carry a password (CREATE ROLE ... WITH PASSWORD) or the
// data being written. For the same reason a failed request records the
// Cassandra error code and not the message, which can quote a value.
type Observer struct {
	tracer trace.Tracer
}

// NewObserver returns an observer that makes its spans with tp.
func NewObserver(tp trace.TracerProvider) *Observer {
	return &Observer{tracer: tp.Tracer(instrumentationName)}
}

// ObserveQuery implements gocql.QueryObserver.
func (o *Observer) ObserveQuery(ctx context.Context, q gocql.ObservedQuery) {
	op := operationName(q.Statement)
	attrs := []attribute.KeyValue{
		semconv.DBOperationName(op),
		attribute.Int("db.response.returned_rows", q.Rows),
	}
	o.record(ctx, op, q.Keyspace, q.Host, q.Start, q.End, q.Attempt, q.Err, attrs)
}

// ObserveBatch implements gocql.BatchObserver.
func (o *Observer) ObserveBatch(ctx context.Context, b gocql.ObservedBatch) {
	attrs := []attribute.KeyValue{
		semconv.DBOperationName("BATCH"),
		semconv.DBOperationBatchSize(len(b.Statements)),
	}
	o.record(ctx, "BATCH", b.Keyspace, b.Host, b.Start, b.End, b.Attempt, b.Err, attrs)
}

func (o *Observer) record(ctx context.Context, op, keyspace string, host *gocql.HostInfo,
	start, end time.Time, attempt int, err error, attrs []attribute.KeyValue) {
	if ctx == nil {
		ctx = context.Background()
	}

	attrs = append(attrs, semconv.DBSystemNameCassandra)
	name := op
	if keyspace != "" {
		attrs = append(attrs, semconv.DBNamespace(keyspace))
		name = op + " " + keyspace
	}
	if host != nil {
		if addr := host.ConnectAddress(); addr != nil && !addr.IsUnspecified() {
			attrs = append(attrs, semconv.ServerAddress(addr.String()))
		}
		if port := host.Port(); port > 0 {
			attrs = append(attrs, semconv.ServerPort(port))
		}
	}
	if attempt > 0 {
		attrs = append(attrs, attribute.Int("cassandra.request.attempt", attempt))
	}

	_, span := o.tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithTimestamp(start),
		trace.WithAttributes(attrs...),
	)
	if err != nil {
		code := errorType(err)
		span.SetAttributes(semconv.ErrorTypeKey.String(code))
		var reqErr gocql.RequestError
		if errors.As(err, &reqErr) {
			span.SetAttributes(semconv.DBResponseStatusCode(code))
		}
		span.SetStatus(codes.Error, code)
	}
	span.End(trace.WithTimestamp(end))
}

// operationName is the statement's first keyword, which is as much of the
// statement as a span name can carry without the risk of carrying a value.
func operationName(stmt string) string {
	fields := strings.Fields(stmt)
	if len(fields) == 0 {
		return "CQL"
	}
	word := strings.ToUpper(strings.TrimRight(fields[0], ";"))
	for _, r := range word {
		if (r < 'A' || r > 'Z') && r != '_' {
			return "CQL"
		}
	}
	if word == "" {
		return "CQL"
	}
	return word
}

// errorType is the Cassandra error code, as a hex string the way the semantic
// conventions ask for it, or the Go type of an error that has no code.
func errorType(err error) string {
	var reqErr gocql.RequestError
	if errors.As(err, &reqErr) {
		return fmt.Sprintf("0x%04X", reqErr.Code())
	}
	return fmt.Sprintf("%T", err)
}
