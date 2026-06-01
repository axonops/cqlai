# Behaviour of router.parseCopyOptions (the WITH-clause parser used by
# COPY TO / COPY FROM). Pure in-process function — no persisted state
# (Rule 4), parallel-safe (Rule 6).
#
# Because parseCopyOptions is unexported, the step definitions for this
# feature live inside the internal/router package (see
# internal/router/copy_options_bdd_test.go) rather than under
# test/bdd/steps/. They still share this feature file.

Feature: COPY WITH-clause option parsing
  As a cqlai user importing or exporting CSV
  I want WITH options to be parsed with the right types and defaults
  So that my data round-trips faithfully and bad input is not silently swallowed

  # ---------------------------------------------------------------------------
  # Defaults
  # ---------------------------------------------------------------------------

  Scenario: an empty option string returns the documented defaults
    When I parse the COPY options ""
    Then the COPY option "HEADER" is "false"
    And the COPY option "NULLVAL" is "null"
    And the COPY option "DELIMITER" is ","
    And the COPY option "QUOTE" is the default double-quote character
    And the COPY option "ENCODING" is "utf8"
    And the COPY option "PAGESIZE" is "1000"
    And the COPY option "CHUNKSIZE" is "5000"
    And the COPY option "MAXROWS" is "-1"
    And the COPY option "SKIPROWS" is "0"

  # ---------------------------------------------------------------------------
  # Single option overrides
  # ---------------------------------------------------------------------------

  Scenario: an unquoted bare value overrides a default
    When I parse the COPY options "HEADER=true"
    Then the COPY option "HEADER" is "true"
    And the COPY option "DELIMITER" is ","

  Scenario: a single-quoted value is unquoted
    When I parse the COPY options "DELIMITER='|'"
    Then the COPY option "DELIMITER" is "|"

  Scenario: a double-quoted value is unquoted
    # Use a doc string so the embedded double-quotes don't collide with the
    # step's own argument delimiters.
    When I parse the COPY options doc string:
      """
      QUOTE="'"
      """
    Then the COPY option "QUOTE" is "'"

  Scenario: keys are upper-cased regardless of how the user typed them
    When I parse the COPY options "header=true"
    Then the COPY option "HEADER" is "true"

  # ---------------------------------------------------------------------------
  # Multiple options
  # ---------------------------------------------------------------------------

  Scenario: multiple comma-separated options are all applied
    When I parse the COPY options "HEADER=true, DELIMITER='|', NULLVAL=NIL"
    Then the COPY option "HEADER" is "true"
    And the COPY option "DELIMITER" is "|"
    And the COPY option "NULLVAL" is "NIL"

  Scenario: multiple AND-separated options are all applied
    # cqlsh accepts the SQL-style "WITH a=1 AND b=2" form.
    When I parse the COPY options "HEADER=true AND DELIMITER=';'"
    Then the COPY option "HEADER" is "true"
    And the COPY option "DELIMITER" is ";"

  # ---------------------------------------------------------------------------
  # Edge / failure-shaped inputs
  # ---------------------------------------------------------------------------

  Scenario: an unknown option key is still recorded (no rejection)
    # parseCopyOptions does not validate option names — it is the caller's
    # job to ignore unknown keys. Pin that contract here.
    When I parse the COPY options "BOGUS=42"
    Then the COPY option "BOGUS" is "42"
    And the COPY option "HEADER" is "false"

  Scenario: malformed pairs without an "=" are silently ignored
    When I parse the COPY options "HEADER, =true, DELIMITER='|'"
    Then the COPY option "DELIMITER" is "|"
    And the COPY option "HEADER" is "false"
