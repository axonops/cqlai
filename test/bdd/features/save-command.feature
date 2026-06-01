# Behaviour of router.ParseSaveCommand. The parser turns a user-typed SAVE
# command into a SaveCommand value or returns a syntax error. Pure
# in-process function — no persisted state (Rule 4), parallel-safe (Rule 6).

Feature: SAVE command parsing
  As a cqlai user
  I want SAVE to accept several shorthand forms and reject garbage input
  So that I can export the current result without surprises

  # ---------------------------------------------------------------------------
  # Happy path
  # ---------------------------------------------------------------------------

  Scenario: bare SAVE opens the interactive save modal
    When I parse the SAVE command "SAVE"
    Then the SAVE command is parsed successfully
    And the SAVE command is interactive

  Scenario: bare SAVE is case-insensitive
    When I parse the SAVE command "save"
    Then the SAVE command is parsed successfully
    And the SAVE command is interactive

  Scenario: SAVE TO 'filename' without AS infers the format from the extension
    When I parse the SAVE command "SAVE TO 'out.csv'"
    Then the SAVE command is parsed successfully
    And the SAVE command has filename "out.csv"
    And the SAVE command has format "CSV"

  Scenario Outline: format is inferred from the file extension
    When I parse the SAVE command "SAVE TO '<filename>'"
    Then the SAVE command is parsed successfully
    And the SAVE command has format "<format>"

    Examples:
      | filename     | format |
      | data.csv     | CSV    |
      | data.json    | JSON   |
      | data.txt     | ASCII  |
      | data.text    | ASCII  |
      | nameless     | CSV    |

  Scenario: SAVE TO with explicit AS overrides the extension
    When I parse the SAVE command "SAVE TO 'out.dat' AS JSON"
    Then the SAVE command is parsed successfully
    And the SAVE command has filename "out.dat"
    And the SAVE command has format "JSON"

  Scenario: TXT and TEXT are normalised to ASCII
    When I parse the SAVE command "SAVE TO 'out.dat' AS TEXT"
    Then the SAVE command is parsed successfully
    And the SAVE command has format "ASCII"

  Scenario: double-quoted filenames are accepted
    When I parse the SAVE command:
      """
      SAVE TO "out file.csv"
      """
    Then the SAVE command is parsed successfully
    And the SAVE command has filename "out file.csv"

  Scenario: unquoted filenames are accepted up to a keyword boundary
    When I parse the SAVE command "SAVE TO out.json AS JSON"
    Then the SAVE command is parsed successfully
    And the SAVE command has filename "out.json"
    And the SAVE command has format "JSON"

  # ---------------------------------------------------------------------------
  # WITH options
  # ---------------------------------------------------------------------------

  Scenario: WITH parses boolean options as booleans
    When I parse the SAVE command "SAVE TO 'out.csv' WITH header=false"
    Then the SAVE command is parsed successfully
    And the SAVE command option "header" equals boolean "false"

  Scenario: WITH parses multiple comma-separated options
    When I parse the SAVE command "SAVE TO 'out.json' AS JSON WITH pretty=true, indent=2"
    Then the SAVE command is parsed successfully
    And the SAVE command option "pretty" equals boolean "true"
    And the SAVE command option "indent" equals string "2"

  Scenario: WITH strips surrounding single quotes from values
    When I parse the SAVE command "SAVE TO 'out.csv' WITH delimiter='|'"
    Then the SAVE command is parsed successfully
    And the SAVE command option "delimiter" equals string "|"

  # ---------------------------------------------------------------------------
  # Failure modes
  # ---------------------------------------------------------------------------

  Scenario: a SAVE command that does not start with TO is rejected
    When I parse the SAVE command "SAVE FROM 'out.csv'"
    Then the SAVE command is rejected with an error containing "invalid SAVE syntax"

  Scenario: an unsupported AS format is rejected
    When I parse the SAVE command "SAVE TO 'out.dat' AS XML"
    Then the SAVE command is rejected with an error containing "unsupported format"

  Scenario: an unclosed quoted filename is rejected
    When I parse the SAVE command "SAVE TO 'never-closed"
    Then the SAVE command is rejected with an error containing "invalid filename"
