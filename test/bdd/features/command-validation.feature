# Behaviour of the CQL command syntax validator and the dangerous-command
# guard in internal/validation. These are pure in-process function checks
# against a string parser. Nothing is persisted, so no teardown is required
# (Rule 4). Scenarios are read-only and parallel-safe (Rule 6).

Feature: CQL command syntax validation
  As a cqlai operator
  I want the shell to reject unknown commands and to flag destructive ones
  So that I get a clear error before any bad request reaches the cluster

  # ---------------------------------------------------------------------------
  # Happy path — recognised CQL keywords
  # ---------------------------------------------------------------------------

  Scenario Outline: recognised CQL commands are accepted
    When I validate the cqlai command "<command>"
    Then the cqlai command is accepted

    Examples:
      | command                            |
      | SELECT * FROM t                    |
      | select * from t                    |
      | INSERT INTO t (a) VALUES (1)       |
      | UPDATE t SET a = 1 WHERE b = 2     |
      | DELETE FROM t WHERE b = 2          |
      | CREATE TABLE t (a int PRIMARY KEY) |
      | ALTER TABLE t ADD c int            |
      | DROP TABLE t                       |
      | TRUNCATE t                         |
      | USE ks                             |
      | BEGIN BATCH                        |
      | APPLY BATCH                        |
      | LIST USERS                         |

  # ---------------------------------------------------------------------------
  # Happy path — recognised meta-commands
  # ---------------------------------------------------------------------------

  Scenario Outline: recognised meta-commands are accepted
    When I validate the cqlai command "<command>"
    Then the cqlai command is accepted

    Examples:
      | command             |
      | DESCRIBE KEYSPACES  |
      | DESC TABLE t        |
      | CONSISTENCY ONE     |
      | OUTPUT JSON         |
      | PAGING 100          |
      | AUTOFETCH ON        |
      | TRACING ON          |
      | SOURCE 'file.cql'   |
      | COPY t TO 'out.csv' |
      | SHOW VERSION        |
      | EXPAND ON           |
      | CAPTURE 'out.txt'   |
      | HELP                |
      | SAVE TO 'out.csv'   |

  Scenario: keywords are matched case-insensitively
    When I validate the cqlai command "Select id From t"
    Then the cqlai command is accepted

  Scenario: a trailing semicolon is accepted
    When I validate the cqlai command "SELECT * FROM t;"
    Then the cqlai command is accepted

  Scenario: an empty command is accepted as a no-op
    When I validate the cqlai command ""
    Then the cqlai command is accepted

  Scenario: whitespace-only input is accepted as a no-op
    When I validate the whitespace-only cqlai command
    Then the cqlai command is accepted

  # ---------------------------------------------------------------------------
  # Failure modes — unrecognised commands
  # ---------------------------------------------------------------------------

  Scenario: an unknown leading keyword is rejected and names the offender
    When I validate the cqlai command "FROBNICATE t"
    Then the cqlai command is rejected with an error containing "FROBNICATE"
    And the cqlai error explains the command is not recognised

  Scenario: a meta-command must have a word boundary after the keyword
    # "DESCRIBES" must NOT be accepted just because it starts with "DESCRIBE".
    When I validate the cqlai command "DESCRIBES KEYSPACES"
    Then the cqlai command is rejected with an error containing "DESCRIBES"

  Scenario: a CQL keyword prefix without a boundary is rejected
    # "SELECTOR" must NOT be accepted just because it starts with "SELECT".
    When I validate the cqlai command "SELECTOR x"
    Then the cqlai command is rejected with an error containing "SELECTOR"

  # ---------------------------------------------------------------------------
  # Dangerous-command guard
  # ---------------------------------------------------------------------------

  Scenario Outline: destructive commands are flagged as dangerous
    When I check whether the cqlai command "<command>" is dangerous
    Then the cqlai command is marked dangerous

    Examples:
      | command                |
      | DROP TABLE t           |
      | drop table t           |
      | DELETE FROM t          |
      | TRUNCATE t             |
      | ALTER TABLE t ADD c    |
      | REVOKE ALL ON x FROM y |

  Scenario Outline: non-destructive commands are not flagged
    When I check whether the cqlai command "<command>" is dangerous
    Then the cqlai command is not marked dangerous

    Examples:
      | command                      |
      | SELECT * FROM t              |
      | INSERT INTO t (a) VALUES (1) |
      | UPDATE t SET a=1 WHERE b=2   |
      | DESCRIBE KEYSPACES           |
      | USE ks                       |

  Scenario: word boundary protects against prefix matches on dangerous keywords
    # "DROPDOWN" must NOT be flagged as DROP.
    When I check whether the cqlai command "DROPDOWN x" is dangerous
    Then the cqlai command is not marked dangerous
