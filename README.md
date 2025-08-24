# CQLAI - AI-Powered Cassandra CQL Shell

![cqlai banner](https://user-images.githubusercontent.com/1253874/234567890-12345678-1234-1234-1234-123456789012.png)

**CQLAI** is a fast, portable, and AI-enhanced interactive terminal for Cassandra (CQL), built in Go. It provides a modern, user-friendly alternative to `cqlsh` with an advanced terminal UI, natural language query generation, client-side command parsing, and enhanced productivity features.

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
    - Simple configuration via `cqlai.json` in current directory or `~/.cqlai.json`.
    - Support for SSL/TLS connections with certificate authentication.
- **AI-Powered Query Generation:** 
    - Natural language to CQL conversion using AI providers (OpenAI, Anthropic, Gemini).
    - Schema-aware query generation with automatic context.
    - Safe preview and confirmation before execution.
    - Support for complex operations including DDL and DML.
- **Single Binary:** Distributed as a single, static binary with no external dependencies. Fast startup and small footprint.

## Installation

You can install `cqlai` in several ways:

#### From Pre-compiled Binaries

Download the appropriate binary for your OS and architecture from the [**Releases**](https://github.com/axonops/cqlai/releases) page.

#### Using `go install`

```bash
go install github.com/axonops/cqlai/cmd/cqlai@latest
```

#### From Source

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
go build -o cqlai cmd/cqlai/main.go
```

#### Using Docker

```bash
# Build the image
docker build -t cqlai .

# Run the container
docker run -it --rm --name cqlai-session cqlai --host your-cassandra-host
```

## Usage

Connect to a Cassandra host:
```bash
cqlai --host 127.0.0.1 --port 9042 --user cassandra --password cassandra
```

Or use a configuration file:
```bash
# Create configuration from example
cp cqlai.json.example cqlai.json
# Edit cqlai.json with your settings, then run:
cqlai
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
  PAGING 50;
  EXPAND ON;  -- Vertical output mode
  SOURCE 'script.cql';  -- Execute CQL script
  ```
- **AI-Powered Query Generation:**
  ```sql
  .ai show all users with age greater than 25
  .ai create a table for storing product inventory
  .ai delete orders older than 1 year
  ```
- **Keyboard Shortcuts:**
  - `↑`/`↓`: Navigate command history.
  - `Tab`: Autocomplete commands and table/keyspace names.
  - `Ctrl+C`: Clear input or exit (press twice to exit).
  - `Ctrl+R`: Search command history.
  - `F1`: Switch between history and table view.
  - `F2`: Toggle column data types in table view.
  - `Alt+←/→`: Scroll table horizontally.
  - `Alt+↑/↓`: Scroll viewport vertically.

## Configuration

`cqlai` can be configured via a JSON file. The application looks for `cqlai.json` in the current directory first, then `~/.cqlai.json` in your home directory.

**Example `cqlai.json`:**
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  "keyspace": "",
  "username": "cassandra",
  "password": "cassandra",
  "requireConfirmation": true,
  "ssl": {
    "enabled": false,
    "certPath": "/path/to/client-cert.pem",
    "keyPath": "/path/to/client-key.pem",
    "caPath": "/path/to/ca-cert.pem",
    "hostVerification": true,
    "insecureSkipVerify": false
  },
  "ai": {
    "provider": "openai",
    "openai": {
      "apiKey": "sk-...",
      "model": "gpt-4-turbo-preview"
    }
  }
}
```

Settings are prioritized in the following order: **Command-line flags > Configuration file**.

## AI-Powered Query Generation

CQLAI includes built-in AI capabilities to convert natural language into CQL queries. Simply prefix your request with `.ai`:

### Examples

```sql
-- Simple queries
.ai show all users
.ai find products with price less than 100
.ai count orders from last month

-- Complex operations
.ai create a table for storing customer feedback with id, customer_id, rating, and comment
.ai update user status to inactive where last_login is older than 90 days
.ai delete all expired sessions

-- Schema exploration
.ai what tables are in this keyspace
.ai describe the structure of the users table
```

### How It Works

1. **Natural Language Input**: Type `.ai` followed by your request in plain English
2. **Schema Context**: CQLAI automatically extracts your current schema to provide context
3. **Query Generation**: The AI generates a structured query plan
4. **Preview & Confirm**: Review the generated CQL before execution
5. **Execute or Edit**: Choose to execute, edit, or cancel the query

### Supported AI Providers

Configure your preferred AI provider in `cqlai.json`:

- **OpenAI** (GPT-4, GPT-3.5)
- **Anthropic** (Claude 3)
- **Google Gemini**
- **Mock** (default, for testing without API keys)

### Safety Features

- **Read-only by default**: AI prefers SELECT queries unless explicitly asked to modify
- **Dangerous operation warnings**: DROP, DELETE, TRUNCATE operations show warnings
- **Confirmation required**: Destructive operations require additional confirmation
- **Schema validation**: Queries are validated against your current schema

## Development

To work on `cqlai`, you'll need Go (≥ 1.22) and ANTLR v4.

#### Setup

```bash
# Clone the repository
git clone https://github.com/axonops/cqlai.git
cd cqlai

# Install dependencies
go mod download
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