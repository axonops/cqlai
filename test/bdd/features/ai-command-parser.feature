# Behaviour of ai.ParseCommand. The parser turns an LLM response into a
# typed tool invocation (or rejects it). Pure in-process function — no
# persisted state (Rule 4), parallel-safe (Rule 6).
#
# JSON payloads are passed via Gherkin doc strings to avoid double-escaping
# braces and quotes inside a quoted step argument.

Feature: AI command JSON parsing
  As a developer extending cqlai with new AI tools
  I want the parser to extract well-formed tool calls and reject malformed ones
  So that bad LLM output never reaches the executor

  # ---------------------------------------------------------------------------
  # Happy path — every tool
  # ---------------------------------------------------------------------------

  Scenario: a fuzzy_search request is parsed
    When I parse the AI response:
      """
      {"tool": "fuzzy_search", "params": {"query": "users"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "fuzzy_search"
    And the AI command arg is "users"

  Scenario: a get_schema request is parsed into keyspace.table form
    When I parse the AI response:
      """
      {"tool": "get_schema", "params": {"keyspace": "ks1", "table": "users"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "get_schema"
    And the AI command arg is "ks1.users"

  Scenario: a list_keyspaces request is parsed with an empty arg
    When I parse the AI response:
      """
      {"tool": "list_keyspaces", "params": {}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "list_keyspaces"
    And the AI command arg is ""

  Scenario: a list_tables request is parsed with the keyspace as arg
    When I parse the AI response:
      """
      {"tool": "list_tables", "params": {"keyspace": "ks1"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "list_tables"
    And the AI command arg is "ks1"

  Scenario: a user_selection request is parsed into type:value1,value2 form
    When I parse the AI response:
      """
      {"tool": "user_selection", "params": {"type": "keyspace", "options": ["a", "b", "c"]}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "user_selection"
    And the AI command arg is "keyspace:a,b,c"

  Scenario: a not_enough_info request carries the message
    When I parse the AI response:
      """
      {"tool": "not_enough_info", "params": {"message": "need keyspace"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "not_enough_info"
    And the AI command arg is "need keyspace"

  Scenario: a not_relevant request carries the message
    When I parse the AI response:
      """
      {"tool": "not_relevant", "params": {"message": "off topic"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "not_relevant"
    And the AI command arg is "off topic"

  # ---------------------------------------------------------------------------
  # Tolerant of surrounding prose
  # ---------------------------------------------------------------------------

  Scenario: a JSON object embedded in prose is still extracted
    When I parse the AI response:
      """
      Sure, here is the call:
      {"tool": "list_keyspaces", "params": {}}
      Hope that helps!
      """
    Then the AI response is recognised as a command
    And the AI command tool is "list_keyspaces"

  Scenario: nested braces inside a JSON string do not confuse the extractor
    When I parse the AI response:
      """
      {"tool": "fuzzy_search", "params": {"query": "weird }} value"}}
      """
    Then the AI response is recognised as a command
    And the AI command tool is "fuzzy_search"
    And the AI command arg is "weird }} value"

  # ---------------------------------------------------------------------------
  # Failure modes
  # ---------------------------------------------------------------------------

  Scenario: a response containing no JSON object is rejected
    When I parse the AI response:
      """
      Sorry, I cannot help with that.
      """
    Then the AI response is not recognised as a command

  Scenario: malformed JSON is rejected
    When I parse the AI response:
      """
      {"tool": "fuzzy_search", "params":
      """
    Then the AI response is not recognised as a command

  Scenario: an unknown tool name is rejected
    When I parse the AI response:
      """
      {"tool": "delete_keyspace", "params": {"keyspace": "ks1"}}
      """
    Then the AI response is not recognised as a command

  Scenario: fuzzy_search without a query parameter is rejected
    When I parse the AI response:
      """
      {"tool": "fuzzy_search", "params": {}}
      """
    Then the AI response is not recognised as a command

  Scenario: get_schema missing the table parameter is rejected
    When I parse the AI response:
      """
      {"tool": "get_schema", "params": {"keyspace": "ks1"}}
      """
    Then the AI response is not recognised as a command

  Scenario: user_selection with no options is rejected
    When I parse the AI response:
      """
      {"tool": "user_selection", "params": {"type": "keyspace", "options": []}}
      """
    Then the AI response is not recognised as a command
