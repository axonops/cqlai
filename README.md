# cqlai - A modern Cassandra CQL Shell

![cqlai banner](https://user-images.githubusercontent.com/1253874/234567890-12345678-1234-1234-1234-123456789012.png)

**cqlai** is a fast, portable, and feature-rich interactive terminal for Cassandra (CQL), built in Go. It provides a modern, user-friendly alternative to `cqlsh` with an advanced terminal UI, client-side command parsing, and enhanced productivity features.

It is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for the terminal UI, and uses [ANTLR](https://www.antlr.org/) for robust meta-command parsing.

---

## Features

- **Interactive CQL Shell:** Execute any CQL query that your Cassandra cluster supports.
- **Rich Terminal UI:**
    - A multi-layer, full-screen terminal application.
    - Virtualized, scrollable table for results, preventing memory overload from large queries.
    - Zebra-striped rows, dimmed `null` values, and proper column alignment for readability.
    - Sticky footer/status bar showing connection details, query latency, and session status (consistency, tracing).
    - Modal overlays for history, help, and command completion.
- **Client-Side Meta-Commands:** A powerful set of `cqlsh`-compatible commands parsed by a real grammar (ANTLR):
    - `DESCRIBE` (keyspaces, tables, types, functions, etc.).
    - `COPY ... TO/FROM` for high-performance CSV import and export.
    - `SOURCE 'file.cql'` to execute scripts.
    - `CONSISTENCY`, `PAGING`, `TRACING` to manage session settings.
    - `SHOW` to view current session details.
- **Advanced Autocompletion:** Context-aware completion for keywords, table/keyspace names, and more.
- **Configuration:**
    - Simple configuration via `~/.cqlai.yml`, environment variables, or command-line flags.
    - Support for connection profiles to easily switch between clusters.
- **Single Binary:** Distributed as a single, static binary with no external dependencies. Fast startup and small footprint.

## Installation

You can install `cqlai` in several ways:

#### From Pre-compiled Binaries

Download the appropriate binary for your OS and architecture from the [**Releases**](https://github.com/your-repo/cqlai/releases) page.

#### Using `go install`

```bash
go install github.com/your-repo/cqlai/cmd/cqlai@latest
```

#### From Source

```bash
git clone https://github.com/your-repo/cqlai.git
cd cqlai
make build
```
The binary will be available at `./bin/cqlai`.

#### Using Docker

```bash
# Build the image
make docker-build

# Run the container
docker run -it --rm --name cqlai-session cqlai:latest --host your-cassandra-host
```

## Usage

Connect to a Cassandra host:
```bash
cqlai --host 127.0.0.1 --port 9042 --user cassandra --password cassandra
```

Or use a configuration profile:
```bash
cqlai --profile my-cluster
```

### Basic Commands

- **Execute CQL:** Type any CQL statement and press Enter.
- **Meta-Commands:**
  ```sql
  DESCRIBE KEYSPACES;
  USE my_keyspace;
  DESCRIBE TABLES;
  CONSISTENCY QUORUM;
  TRACING ON;
  COPY users (id, name, email) TO 'users.csv' WITH HEADER = TRUE;
  ```
- **Keyboard Shortcuts:**
  - `↑`/`↓`: Navigate command history.
  - `Ctrl+C`: Exit the application.
  - `F2`: Toggle query tracing.
  - `F3`: Cycle through consistency levels.
  - `F4`: Set page size.

## Configuration

`cqlai` can be configured via a YAML file located at `~/.cqlai.yml`.

**Example `~/.cqlai.yml`:**
```yaml
# Default connection settings (can be overridden by flags)
contact_points: ["127.0.0.1"]
port: 9042
username: "cassandra"
# password: "your_password" # Use env var or flag for safety
keyspace: "system"
local_dc: "dc1"

# Session defaults
consistency: "LOCAL_QUORUM"
page_size: 100
timeout: "10s"

# TLS settings
tls:
  enabled: false
  ca_path: "/path/to/ca.pem"
  cert_path: "/path/to/cert.pem"
  key_path: "/path/to/key.pem"
  insecure_skip_verify: false

# Connection profiles
profiles:
  prod-cluster:
    contact_points: ["prod-node1.example.com", "prod-node2.example.com"]
    username: "prod_user"
    local_dc: "us-east-1"
    tls:
      enabled: true
      ca_path: "/path/to/prod-ca.pem"
  staging-cluster:
    contact_points: ["staging-node.example.com"]
    username: "staging_user"
```

Settings are prioritized in the following order: **Flags > Environment Variables > YAML Configuration**.

## Development

To work on `cqlai`, you'll need Go (≥ 1.22) and ANTLR v4.

#### Setup

```bash
# Clone the repository
git clone https://github.com/your-repo/cqlai.git
cd cqlai

# Install dependencies
make deps
```

#### Building

```bash
# Build a standard binary
make build

# Build a development binary with race detection
make build-dev
```

#### Running Tests & Linter

```bash
# Run all tests
make test

# Run tests with coverage report
make test-coverage

# Run the linter
make lint

# Run all checks (format, lint, test)
make check
```

#### ANTLR Grammar

The meta-command grammar is defined in `internal/parser/grammar/`. If you modify the `.g4` files, you must regenerate the Go parser files.

```bash
# Install the antlr4 tool if you haven't already
go install github.com/antlr4-go/antlr/v4/cmd/antlr4@latest

# Regenerate grammar files
make grammar
```

## Technology Stack

- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUI Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Cassandra Driver:** [gocql](https://github.com/gocql/gocql)
- **Parser Generator:** [ANTLR v4](https://www.antlr.org/)

## License

This project is licensed under the [MIT License](LICENSE).