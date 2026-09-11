# Contributing to CQLAI

Thanks for taking a look. Pull requests are welcome.

## Getting set up

CQLAI builds with Go 1.24 or newer and has no other build dependencies.

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
make build        # writes bin/cqlai
```

A Cassandra to talk to is useful but not needed for most work: the tests do not
require one.

## Before you open a pull request

```bash
make test         # the unit tests
make lint         # golangci-lint, if it is installed
make check        # format, lint and test together
```

All three should be clean. CI runs the same checks plus the integration tests
against Cassandra 2.1, 3.0, 3.11, 4.0, 4.1 and 5.0.

## How changes are made here

- An issue first, saying what is wrong or what is missing, and why.
- A branch off `main`, named for the issue: `2026-09-11-123-short-description`.
- Tests with the change. A bug fix comes with a test that fails without it.
- Documentation with the change: the README, its translations, and the in-app
  `HELP` text are part of the change, not a follow-up. `internal/ui/docs_test.go`
  checks some of this for you - it fails if the views, their keys or the FILE
  menu entries are missing from the docs.
- A pull request that says what changed and why, and closes its issue.

## Things worth knowing

- The TUI is Bubble Tea v2. Rendering and hit-testing must come from one
  description of the layout, or a click lands on the row above the one you
  pointed at.
- Meta-commands are parsed by hand in `internal/router/command_parser.go`. There
  is no parser generator, and ANTLR is not coming back.
- Widths are measured in columns (`lipgloss.Width`), not bytes: box-drawing
  characters are three bytes each.

## Reporting a bug

Say what you did, what happened, and what you expected. The version helps
(`cqlai --version`), as does the terminal you are using when the problem is a
drawing one.
