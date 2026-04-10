Feature: SSL CLI flag override behaviour
  As a cqlai operator
  I want CLI flags and environment variables to control SSL verification settings
  So that I can connect to Cassandra clusters with non-standard TLS configurations
  without modifying the shared config file

  Background:
    Given I have a base SSL config with host verification enabled and insecure skip verify disabled

  # ---------------------------------------------------------------------------
  # --no-ssl-host-verification flag
  # ---------------------------------------------------------------------------

  Scenario: no-ssl-host-verification flag disables host verification regardless of config
    Given the config file has host verification set to true
    When the connection option SSLHostVerification pointer is set to false
    Then the resolved SSL config has host verification false

  Scenario: no-ssl-host-verification flag overrides a config that already had verification disabled
    Given the config file has host verification set to false
    When the connection option SSLHostVerification pointer is set to false
    Then the resolved SSL config has host verification false

  Scenario: no-ssl-host-verification flag can re-enable verification when config had it disabled
    Given the config file has host verification set to false
    When the connection option SSLHostVerification pointer is set to true
    Then the resolved SSL config has host verification true

  # ---------------------------------------------------------------------------
  # --ssl-insecure-skip-verify flag
  # ---------------------------------------------------------------------------

  Scenario: ssl-insecure-skip-verify flag enables insecure skip verify regardless of config
    Given the config file has insecure skip verify set to false
    When the connection option SSLInsecureSkipVerify pointer is set to true
    Then the resolved SSL config has insecure skip verify true

  Scenario: ssl-insecure-skip-verify flag overrides a config that already had it enabled
    Given the config file has insecure skip verify set to true
    When the connection option SSLInsecureSkipVerify pointer is set to true
    Then the resolved SSL config has insecure skip verify true

  Scenario: ssl-insecure-skip-verify pointer set to false overrides a permissive config
    Given the config file has insecure skip verify set to true
    When the connection option SSLInsecureSkipVerify pointer is set to false
    Then the resolved SSL config has insecure skip verify false

  # ---------------------------------------------------------------------------
  # Neither flag set — config file wins
  # ---------------------------------------------------------------------------

  Scenario: config file host verification is preserved when no CLI flag is set
    Given the config file has host verification set to true
    When no SSLHostVerification connection option is set
    Then the resolved SSL config has host verification true

  Scenario: config file insecure skip verify is preserved when no CLI flag is set
    Given the config file has insecure skip verify set to true
    When no SSLInsecureSkipVerify connection option is set
    Then the resolved SSL config has insecure skip verify true

  # ---------------------------------------------------------------------------
  # SSL config block is created on demand when absent from config file
  # ---------------------------------------------------------------------------

  Scenario: SSL block is created when absent from config and host verification is overridden
    Given the config file has no SSL section
    When the connection option SSLHostVerification pointer is set to false
    Then the resolved SSL config has host verification false

  Scenario: SSL block is created when absent from config and insecure skip verify is overridden
    Given the config file has no SSL section
    When the connection option SSLInsecureSkipVerify pointer is set to true
    Then the resolved SSL config has insecure skip verify true

  # ---------------------------------------------------------------------------
  # Environment variable behaviour (pointer-building contract from main.go)
  # ---------------------------------------------------------------------------

  Scenario: CQLAI_NO_SSL_HOST_VERIFICATION env var produces a non-nil host verification pointer
    When the env var CQLAI_NO_SSL_HOST_VERIFICATION is set to "true"
    Then buildSSLHostVerificationPtr returns a non-nil pointer
    And the pointed-to host verification value is false

  Scenario: CQLAI_NO_SSL_HOST_VERIFICATION env var set to "false" produces a non-nil pointer with value true
    When the env var CQLAI_NO_SSL_HOST_VERIFICATION is set to "false"
    Then buildSSLHostVerificationPtr returns a non-nil pointer
    And the pointed-to host verification value is true

  Scenario: CQLAI_SSL_INSECURE_SKIP_VERIFY env var produces a non-nil insecure skip verify pointer
    When the env var CQLAI_SSL_INSECURE_SKIP_VERIFY is set to "true"
    Then buildSSLInsecureSkipVerifyPtr returns a non-nil pointer
    And the pointed-to insecure skip verify value is true

  Scenario: CQLAI_SSL_INSECURE_SKIP_VERIFY env var set to "false" produces a non-nil pointer with value false
    When the env var CQLAI_SSL_INSECURE_SKIP_VERIFY is set to "false"
    Then buildSSLInsecureSkipVerifyPtr returns a non-nil pointer
    And the pointed-to insecure skip verify value is false

  Scenario: absent CQLAI_NO_SSL_HOST_VERIFICATION env var produces a nil pointer
    When the env var CQLAI_NO_SSL_HOST_VERIFICATION is not set
    Then buildSSLHostVerificationPtr returns a nil pointer

  Scenario: absent CQLAI_SSL_INSECURE_SKIP_VERIFY env var produces a nil pointer
    When the env var CQLAI_SSL_INSECURE_SKIP_VERIFY is not set
    Then buildSSLInsecureSkipVerifyPtr returns a nil pointer

  # ---------------------------------------------------------------------------
  # CLI flag takes precedence over env var (pointer-building contract)
  # ---------------------------------------------------------------------------

  Scenario Outline: CLI flag value wins over env var for no-ssl-host-verification
    Given the env var CQLAI_NO_SSL_HOST_VERIFICATION is set to "<env_value>"
    When the CLI flag no-ssl-host-verification is explicitly set to "<flag_value>"
    Then the resolved host verification override pointer value is <expected_host_verification>

    Examples:
      | env_value | flag_value | expected_host_verification |
      | false     | true       | false                      |
      | true      | false      | true                       |

  Scenario Outline: CLI flag value wins over env var for ssl-insecure-skip-verify
    Given the env var CQLAI_SSL_INSECURE_SKIP_VERIFY is set to "<env_value>"
    When the CLI flag ssl-insecure-skip-verify is explicitly set to "<flag_value>"
    Then the resolved insecure skip verify override pointer value is <expected_skip_verify>

    Examples:
      | env_value | flag_value | expected_skip_verify |
      | false     | true       | true                 |
      | true      | false      | false                |
