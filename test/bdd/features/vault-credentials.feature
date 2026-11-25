# Behaviour of internal/vault: reading dynamic database credentials from a
# Vault or OpenBao server and revoking their lease. Each scenario runs against
# its own fake server and sets the VAULT_*/BAO_* variables it needs.

Feature: Database credentials from Vault or OpenBao
  As a cqlai user with a Vault or OpenBao login
  I want cqlai to fetch short-lived Cassandra credentials with --vault-path
  So that no static password is kept on disk or in shell history

  Background:
    Given a fake Vault server that issues user "v-token-reader-abc" with password "fake-password" and lease "database/creds/reader/lease1" at "database/creds/reader"

  # ---------------------------------------------------------------------------
  # Happy path
  # ---------------------------------------------------------------------------

  Scenario: credentials are read with the Vault address and token
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "database/creds/reader"
    Then the credentials have username "v-token-reader-abc" and password "fake-password"
    And the fake server saw the token "hvs.user-token"

  Scenario: OpenBao variables are used when set
    Given the environment points BAO_ADDR at the fake server with BAO_TOKEN "s.bao-token"
    When I read credentials from "database/creds/reader"
    Then the credentials have username "v-token-reader-abc" and password "fake-password"
    And the fake server saw the token "s.bao-token"

  Scenario: the token is read from ~/.vault-token when no token variable is set
    Given the environment points VAULT_ADDR at the fake server with no token variable
    And the file ~/.vault-token contains "hvs.file-token"
    When I read credentials from "database/creds/reader"
    Then the fake server saw the token "hvs.file-token"

  Scenario Outline: the path is accepted as the CLI writes it
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "<path>"
    Then the credentials have username "v-token-reader-abc" and password "fake-password"

    Examples:
      | path                       |
      | database/creds/reader      |
      | /database/creds/reader     |
      | v1/database/creds/reader/  |

  Scenario: the lease is revoked on exit
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "database/creds/reader"
    And cqlai exits
    Then the fake server was asked to revoke lease "database/creds/reader/lease1" once

  Scenario: a lease closed with its session is not revoked again on exit
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "database/creds/reader"
    And the lease is revoked
    And cqlai exits
    Then the fake server was asked to revoke lease "database/creds/reader/lease1" once

  # ---------------------------------------------------------------------------
  # Revocation the token is not allowed to do
  # ---------------------------------------------------------------------------

  Scenario: a token without permission to revoke exits quietly
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    And the fake server denies revocation
    When I read credentials from "database/creds/reader"
    And cqlai exits
    Then the fake server was asked to revoke lease "database/creds/reader/lease1" once
    And no error is reported

  # ---------------------------------------------------------------------------
  # Invalid input and failures
  # ---------------------------------------------------------------------------

  Scenario: no address set
    Given the environment has no Vault or OpenBao address
    When I read credentials from "database/creds/reader"
    Then reading fails with an error containing "set VAULT_ADDR or BAO_ADDR"

  Scenario: no token anywhere
    Given the environment points VAULT_ADDR at the fake server with no token variable
    When I read credentials from "database/creds/reader"
    Then reading fails with an error containing "set VAULT_TOKEN or BAO_TOKEN"

  Scenario: an unknown path
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "database/creds/nosuchrole"
    Then reading fails with an error containing "no such path"

  Scenario: a token not allowed to read the path
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    And the fake server denies reading
    When I read credentials from "database/creds/reader"
    Then reading fails with an error containing "permission denied"
    And the error does not contain "hvs.user-token"

  Scenario: a path that is not a database credential
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    And the fake server answers "database/creds/reader" without a password
    When I read credentials from "database/creds/reader"
    Then reading fails with an error containing "has no username and password"

  Scenario Outline: malformed paths are rejected before any request
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I read credentials from "<path>"
    Then reading fails with an error containing "<message>"
    And the fake server received no read request

    Examples:
      | path                        | message                   |
      |                             | the Vault path is empty   |
      | database/../sys/creds       | is not a valid path       |
      | database//creds/reader      | is not a valid path       |
      | database/creds/reader?x=1   | cannot have a query       |

  # ---------------------------------------------------------------------------
  # Opening a session
  # ---------------------------------------------------------------------------

  Scenario: a session that cannot connect revokes the lease it was given
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    When I open a session to an unreachable cluster with Vault path "database/creds/reader"
    Then opening the session fails
    And the fake server received 1 read request
    And the fake server was asked to revoke lease "database/creds/reader/lease1" once

  Scenario: a Vault path only in the config file is not used for another connection
    Given the environment points VAULT_ADDR at the fake server with VAULT_TOKEN "hvs.user-token"
    And a config file with vaultPath "database/creds/reader"
    When I open a session to an unreachable cluster with no Vault path
    Then opening the session fails
    And the fake server received 0 read requests
