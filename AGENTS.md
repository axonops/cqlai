# Repository Guidelines

## Project Structure & Module Organization
- `cmd/cqlai/main.go` is the CLI entry point for the `cqlai` binary.
- `internal/` holds core application packages (parsing, routing, UI, storage, etc.).
- `assets/` contains branding and UI assets used by docs and release artifacts.
- `docs/` includes user guides and feature documentation (Parquet, AI config, batch mode).
- `test/` hosts test suites; `test/parquet/` for Parquet I/O unit tests and `test/integration/` for Cassandra-backed integration tests.
- `scripts/` and `bin/` hold helper scripts and build outputs.
- `cqlai.json.example` shows configuration defaults.

## Build, Test, and Development Commands
- `make build`: build the binary into `bin/`.
- `make run`: build and launch locally.
- `make test`: run all Go tests with race detection and coverage output.
- `make test-coverage`: generate `coverage.html` from `coverage.out`.
- `make lint`: run `golangci-lint` if available, otherwise `go vet`.
- `make fmt`: format with `go fmt` and `goimports` (if installed).
- `make grammar`: regenerate ANTLR grammar sources in `internal/parser/grammar/`.

## Coding Style & Naming Conventions
- Go standard formatting is required; use `make fmt` before pushing.
- Follow idiomatic Go naming (mixedCaps for exported identifiers, lowerCamel for locals).
- Keep package names short and lowercase; avoid underscores unless required by tools.

## Testing Guidelines
- Unit tests use Go’s `testing` package and live alongside code as `*_test.go`.
- Parquet unit tests: `go test -v ./test/parquet/...` or `./test/parquet/run_tests.sh`.
- Integration tests require a running Cassandra instance; see `test/integration/README.md`.
- No explicit coverage threshold is documented; aim to cover new logic.

## Commit & Pull Request Guidelines
- Commit messages follow a conventional style seen in history (`feat:`, `fix:`, `perf:`, `chore:`).
- PRs should include a clear description, linked issues when applicable, and test notes.
- For UI or output changes, add terminal screenshots or recordings when it improves review clarity.

## Configuration & Security
- Local config is read from `cqlai.json` or `~/.cqlai.json`.
- Do not commit API keys or secrets; use `cqlai.json.example` as a template.
