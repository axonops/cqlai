# Behaviour of the CQL statement splitter (internal/batch). This is the
# tokeniser that powers batch execution: it must split on semicolons but
# preserve them inside string literals, quoted identifiers, comments, and
# BEGIN/APPLY BATCH blocks. Pure in-process function — no persisted state
# (Rule 4), parallel-safe (Rule 6).
#
# Note on input: complex CQL with embedded quotes is passed via Gherkin doc
# strings (the """...""" blocks) so we don't have to escape quote characters
# inside a quoted step argument.

Feature: CQL statement splitting
  As a developer running a batch script through cqlai
  I want my CQL to be split into the right statements
  So that semicolons inside strings, comments or BATCH blocks do not break the run

  # ---------------------------------------------------------------------------
  # Basic splitting
  # ---------------------------------------------------------------------------

  Scenario: a single statement with a trailing semicolon
    When I split the CQL input "SELECT * FROM t1;"
    Then the splitter yields 1 statement
    And split statement 1 is "SELECT * FROM t1;"

  Scenario: multiple statements separated by semicolons
    When I split the CQL input "SELECT * FROM t1; SELECT * FROM t2;"
    Then the splitter yields 2 statements
    And split statement 1 is "SELECT * FROM t1;"
    And split statement 2 is "SELECT * FROM t2;"

  Scenario: a single statement without a trailing semicolon is kept whole
    When I split the CQL input "SELECT * FROM t1"
    Then the splitter yields 1 statement
    And split statement 1 is "SELECT * FROM t1"

  # ---------------------------------------------------------------------------
  # Empty / whitespace
  # ---------------------------------------------------------------------------

  Scenario: empty input yields no statements
    When I split the CQL input ""
    Then the splitter yields 0 statements

  Scenario: whitespace-only input yields no statements
    When I split the whitespace-only CQL input
    Then the splitter yields 0 statements

  # ---------------------------------------------------------------------------
  # Embedded characters that must not split
  # ---------------------------------------------------------------------------

  Scenario: a semicolon inside a single-quoted string does not split
    When I split the CQL input:
      """
      INSERT INTO t (a) VALUES ('hello; world');
      """
    Then the splitter yields 1 statement
    And split statement 1 contains "'hello; world'"

  Scenario: a semicolon inside a quoted identifier does not split
    When I split the CQL input:
      """
      SELECT * FROM "weird;name";
      """
    Then the splitter yields 1 statement

  Scenario: a semicolon inside a line comment is ignored
    When I split the CQL input:
      """
      -- a; b; c
      SELECT 1;
      """
    Then the splitter yields 1 statement

  Scenario: a BEGIN/APPLY BATCH block is treated as one statement
    When I split the CQL input:
      """
      BEGIN BATCH
        INSERT INTO t (a) VALUES (1);
        INSERT INTO t (a) VALUES (2);
      APPLY BATCH;
      """
    Then the splitter yields 1 statement
    And split statement 1 contains "BEGIN BATCH"
    And split statement 1 contains "APPLY BATCH"

  # ---------------------------------------------------------------------------
  # Failure / edge inputs
  # ---------------------------------------------------------------------------

  Scenario: an unclosed string literal marks the input as incomplete
    When I split the CQL input "INSERT INTO t (a) VALUES ('oops"
    Then the splitter marks the input as incomplete

  Scenario: an unclosed BATCH block marks the input as incomplete
    When I split the CQL input:
      """
      BEGIN BATCH
        INSERT INTO t (a) VALUES (1);
      """
    Then the splitter marks the input as incomplete

  Scenario: an unterminated block comment marks the input as incomplete
    When I split the CQL input:
      """
      SELECT * FROM t1 /* unclosed
      """
    Then the splitter marks the input as incomplete

  # ---------------------------------------------------------------------------
  # Shell-command detection (commandsEndWithNewline)
  # ---------------------------------------------------------------------------

  Scenario Outline: shell commands are recognised as newline-terminated
    When I check whether "<command>" is a cqlai shell command
    Then the shell command result is <expected>

    Examples:
      | command  | expected |
      | exit     | true     |
      | quit     | true     |
      | clear    | true     |
      | help     | true     |
      | describe | true     |
      | DESCRIBE | true     |
      | source   | true     |
      | capture  | true     |
      | select   | false    |
      | insert   | false    |
      | nonsense | false    |
