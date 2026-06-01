# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- BDD test suites (godog) covering pure-logic packages: `validation` command
  syntax + dangerous-command guard, `batch` CQL statement splitter and shell
  command detection, `router` SAVE-command parser, `ai` JSON command parser,
  and `router` COPY WITH-clause option parser. 116 new scenarios.
- CI job `bdd-tests` running all godog suites on every push.
- Cassandra 5.0 integration step in `cassandra-tests` job — runs the Go
  `+build integration` suite (`go test -tags integration ./test/integration/...`)
  against the live Cassandra 5.0 service container.

### Changed

- Repo-wide `gofmt` pass — formatting only, no behavioural changes.
