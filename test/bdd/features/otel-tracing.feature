# OpenTelemetry spans for the CQL requests cqlai makes.
#
# The observer is driven with the driver's record of a request, the same
# record gocql hands it after a real one, and the spans go to an in-memory
# recorder, so no Cassandra and no collector are needed.
#
# Starting and stopping tracing changes process-wide state, the global tracer
# provider. Scenarios run one at a time (godog's default) and every scenario
# that starts tracing stops it in the After hook.
Feature: OpenTelemetry tracing

  Scenario: Tracing is off until it is asked for
    Given OpenTelemetry tracing has not been started
    Then no tracer provider is set up
    And stopping tracing does nothing

  Scenario: Tracing starts with the default OTLP protocol and stops again
    Given the OTLP protocol is ""
    When OpenTelemetry tracing is started
    Then tracing is running
    When tracing is stopped
    Then no tracer provider is set up

  Scenario Outline: Tracing starts with each supported OTLP protocol
    Given the OTLP protocol is "<protocol>"
    When OpenTelemetry tracing is started
    Then tracing is running

    Examples:
      | protocol      |
      | grpc          |
      | http/protobuf |

  Scenario Outline: Tracing refuses an OTLP protocol it cannot send
    Given the OTLP protocol is "<protocol>"
    When OpenTelemetry tracing is started
    Then starting tracing fails with "unsupported OTLP protocol"
    And no tracer provider is set up

    Examples:
      | protocol  |
      | http/json |
      | carrier   |

  Scenario: A query becomes a client span named after its operation and keyspace
    Given a span recorder
    When the query "SELECT * FROM users WHERE id = 1" runs in keyspace "shop" and returns 3 rows
    Then one span is recorded named "SELECT shop"
    And the span is a client span
    And the span has attribute "db.system.name" set to "cassandra"
    And the span has attribute "db.namespace" set to "shop"
    And the span has attribute "db.operation.name" set to "SELECT"
    And the span has attribute "db.response.returned_rows" set to "3"
    And the span has attribute "server.address" set to "10.0.0.5"
    And the span has attribute "server.port" set to "9042"
    And the span has no error status

  Scenario: A query outside a keyspace is named by its operation alone
    Given a span recorder
    When the query "describe keyspaces" runs in keyspace "" and returns 0 rows
    Then one span is recorded named "DESCRIBE"
    And the span has no attribute "db.namespace"

  Scenario: A batch becomes one span with its size
    Given a span recorder
    When a batch of 4 statements runs in keyspace "shop"
    Then one span is recorded named "BATCH shop"
    And the span has attribute "db.operation.batch.size" set to "4"

  Scenario: The statement text and its values are never recorded
    Given a span recorder
    When the query "CREATE ROLE alice WITH PASSWORD = 'hunter2' AND LOGIN = true" runs in keyspace "" and returns 0 rows
    Then one span is recorded named "CREATE"
    And no span attribute contains "hunter2"

  Scenario: A request Cassandra rejects records its error code, not its message
    Given a span recorder
    When a query fails in Cassandra with code 8704 and message "Invalid STRING constant (hunter2)"
    Then the span has an error status
    And the span has attribute "error.type" set to "0x2200"
    And the span has attribute "db.response.status_code" set to "0x2200"
    And no span attribute contains "hunter2"

  Scenario: A request that fails without a Cassandra error records the error type
    Given a span recorder
    When a query fails before reaching Cassandra
    Then the span has an error status
    And the span has attribute "error.type" set to "*errors.errorString"
    And the span has no attribute "db.response.status_code"

  Scenario Outline: A statement with no usable first keyword is named CQL
    Given a span recorder
    When the query "<statement>" runs in keyspace "" and returns 0 rows
    Then one span is recorded named "CQL"

    Examples:
      | statement |
      |           |
      | 'quoted'  |
      | 123       |

  Scenario: A retried request records which attempt it was
    Given a span recorder
    When attempt 2 of the query "SELECT * FROM t" runs in keyspace "shop"
    Then the span has attribute "cassandra.request.attempt" set to "2"

  Scenario: A request with no host recorded still makes a span
    Given a span recorder
    When the query "SELECT * FROM t" runs in keyspace "shop" with no host
    Then one span is recorded named "SELECT shop"
    And the span has no attribute "server.address"
