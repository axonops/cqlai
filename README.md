<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">

  # CQLAI - Modern Cassandra® CQL Shell

  [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/axonops/cqlai)](https://golang.org/)
  [![GitHub Issues](https://img.shields.io/github/issues/axonops/cqlai)](https://github.com/axonops/cqlai/issues)
  [![GitHub Discussions](https://img.shields.io/github/discussions/axonops/cqlai)](https://github.com/axonops/cqlai/discussions)
  [![GitHub Stars](https://img.shields.io/github/stars/axonops/cqlai)](https://github.com/axonops/cqlai/stargazers)
</div>

**CQLAI** is a fast, portable interactive terminal for Cassandra (CQL), built in Go. It provides a modern, user-friendly alternative to `cqlsh` with an advanced terminal UI, client-side command parsing, and enhanced productivity features.

**AI features are completely optional** - CQLAI works perfectly as a standalone CQL shell without any AI configuration or API keys.

<div align="center">
  <video src="https://github.com/user-attachments/assets/334bd302-3152-4f48-9d2d-ed617e8d86d3" controls width="100%" style="max-width: 800px;">
    Your browser does not support the video tag.
  </video>
</div>

<div align="center">

### 🎁 100% Free & Open Source
**No hidden costs • No premium tiers • No license keys**

Community-driven development with full transparency

</div>

The original cqlsh command in the [Apache Cassandra](https://cassandra.apache.org/) project is written in Python which requires Python to be installed on the system. cqlai is compiled to a single executable binary, requiring no external dependencies. This project provides binaries for the following platforms:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


It is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for the beautiful terminal UI. A big shout out to the cassandra gocql driver team for implementing the latest Cassandra functionalities [gocql](https://github.com/apache/cassandra-gocql-driver)

---

## 📑 Table of Contents

- [📊 Project Status](#-project-status)
- [✨ Features](#-features)
- [🔧 Installation](#-installation)
- [📚 Usage](#-usage)
  - [Interactive Mode](#interactive-mode)
  - [Command-Line Options](#command-line-options)
  - [Batch Mode Examples](#batch-mode-examples)
  - [Basic Commands](#basic-commands)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
  - [Tab Completion](#tab-completion)
- [⚙️ Available Commands](#️-available-commands)
  - [CQL Commands](#cql-commands)
  - [Meta-Commands](#meta-commands)
  - [AI Commands](#ai-commands)
- [🛠️ Configuration](#️-configuration)
  - [Configuration Precedence](#configuration-precedence)
  - [CQLSHRC Compatibility](#cqlshrc-compatibility)
  - [CQLAI JSON Configuration](#cqlai-json-configuration)
  - [AI Provider Configuration](#ai-provider-configuration)
    - [OpenAI](#openai-gpt-4--gpt-35)
    - [Anthropic](#anthropic-claude-3)
    - [Google Gemini](#google-gemini)
    - [Synthetic](#synthetic-multiple-open-source-models)
    - [Ollama](#ollama-local-models)
    - [OpenRouter](#openrouter-multiple-models)
    - [Mock Provider](#mock-provider-for-testing)
- [🤖 AI-Powered Query Generation](#-ai-powered-query-generation)
- [🔌 MCP Server (Model Context Protocol)](#-mcp-server-model-context-protocol)
- [📦 Apache Parquet Support](#-apache-parquet-support)
- [⚠️ Known Limitations](#️-known-limitations)
- [🔨 Development](#-development)
- [🏗️ Technology Stack](#️-technology-stack)
- [🙏 Acknowledgements](#-acknowledgements)
- [💬 Community & Support](#-community--support)
- [📝 License](#-license)
- [⚖️ Legal Notices](#️-legal-notices)

---

## 📊 Project Status

**CQLAI is production-ready** and actively used in development, testing, and production environments with Cassandra clusters. The tool provides a complete, stable alternative to `cqlsh` with enhanced features and performance.

### What Works
- All core CQL operations and queries
- Complete meta-command support (`DESCRIBE`, `SHOW`, `CONSISTENCY`, etc.)
- Client-side command parsing (lightweight, no ANTLR dependency)
- Data import/export with `COPY TO/FROM` (CSV and Parquet formats)
- SSL/TLS connections and authentication
- User-Defined Types (UDTs) and complex data types
- Batch mode for scripting and automation
- Apache Parquet format support for efficient data interchange
- Tab completion for CQL keywords, tables, columns, and keyspaces
- **Optional**: AI-powered query generation ([OpenAI](https://openai.com/), [Anthropic](https://www.anthropic.com/), [Google Gemini](https://ai.google.dev/), [Synthetic](https://synthetic.new/))
- **Optional**: MCP server for Claude Code/Desktop integration (20 tools, complete Cassandra administration)

### Coming Soon
- Enhanced AI context awareness
- BATCH operation support
- Additional performance optimizations

We encourage you to **try CQLAI today** and help shape its development! Your feedback and contributions are invaluable in making this the best CQL shell for the Cassandra community. Please [report issues](https://github.com/axonops/cqlai/issues) or [contribute](https://github.com/axonops/cqlai/pulls).

---

## ✨ Features

- **Interactive CQL Shell:** Execute any CQL query that your Cassandra cluster supports.
- **Rich Terminal UI:**
    - A multi-layer, full-screen terminal application with alternate screen buffer (preserves terminal history).
    - Virtualized, scrollable table for results with automatic data loading, preventing memory overload from large queries.
    - Advanced navigation modes with vim-style keyboard shortcuts.
    - Full mouse support including wheel scrolling and text selection.
    - Sticky footer/status bar showing connection details, query latency, and session status (consistency, tracing).
    - Modal overlays for history, help, and command completion.
- **Apache Parquet Support:**
    - High-performance columnar data format for analytics and machine learning workflows.
    - Export Cassandra tables to Parquet files with `COPY TO` command.
    - Import Parquet files into Cassandra with automatic schema inference.
    - Partitioned datasets with Hive-style directory structures.
    - TimeUUID / timestamp virtual columns for intelligent time-based partitioning.
    - Support for all Cassandra data types including UDTs, collections, and vectors.
- **Optional AI-Powered Query Generation:**
    - Natural language to CQL conversion using AI providers ([OpenAI](https://openai.com/), [Anthropic](https://www.anthropic.com/), [Google Gemini](https://ai.google.dev/), [Synthetic](https://synthetic.new/)).
    - Schema-aware query generation with automatic context.
    - Safe preview and confirmation before execution.
    - Support for complex operations including DDL and DML.
    - **Requires API key configuration** - not needed for core functionality.
- **Configuration:**
    - Simple configuration via `cqlai.json` in current directory or `~/.cqlai.json`.
    - Support for SSL/TLS connections with certificate authentication.
- **Single Binary:** Distributed as a single, static binary with no external dependencies. Fast startup and small footprint.

## 🔧 Installation

You can install `cqlai` in several ways. For detailed instructions including package managers (APT, YUM) and Docker, see the [Installation Guide](docs/INSTALLATION.md).

### Pre-compiled Binaries

Download the appropriate binary for your OS and architecture from the [**Releases**](https://github.com/axonops/cqlai/releases) page.


### Using Go

```bash
go install github.com/axonops/cqlai/cmd/cqlai@latest
```

### From Source

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
go build -o cqlai cmd/cqlai/main.go
```

### Using Docker

```bash
# Build the image
docker build -t cqlai .

# Run the container
docker run -it --rm --name cqlai-session cqlai --host your-cassandra-host
```

## 📚 Usage

### Interactive Mode

Connect to a Cassandra host:
```bash
# With password on command line (not recommended - visible in ps)
cqlai --host 127.0.0.1 --port 9042 --username cassandra --password cassandra

# With password prompt (secure - password hidden)
cqlai --host 127.0.0.1 --port 9042 -u cassandra
# Password: [hidden input]

# Using environment variable (secure for scripts/containers)
export CQLAI_PASSWORD=cassandra
cqlai --host 127.0.0.1 -u cassandra
```

Or use a configuration file:
```bash
# Create configuration from example
cp cqlai.json.example cqlai.json
# Edit cqlai.json with your settings, then run:
cqlai
```

### Command-Line Options

```bash
cqlai [options]
```

#### Connection Options
| Option | Short | Description |
|--------|-------|-------------|
| `--host <host>` | | Cassandra host (overrides config) |
| `--port <port>` | | Cassandra port (overrides config) |
| `--keyspace <keyspace>` | `-k` | Default keyspace (overrides config) |
| `--username <username>` | `-u` | Username for authentication |
| `--password <password>` | `-p` | Password for authentication* |
| `--no-confirm` | | Disable confirmation prompts |
| `--connect-timeout <seconds>` | | Connection timeout (default: 10) |
| `--request-timeout <seconds>` | | Request timeout (default: 10) |
| `--debug` | | Enable debug logging |

*\*Note: Password can be provided in three ways:*
1. *Command line with `-p` (not recommended - visible in process list)*
2. *Interactive prompt when `-u` is used without `-p` (recommended)*
3. *Environment variable `CQLAI_PASSWORD` (good for automation)*

#### Batch Mode Options
| Option | Short | Description |
|--------|-------|-------------|
| `--execute <statement>` | `-e` | Execute CQL statement and exit |
| `--file <file>` | `-f` | Execute CQL from file and exit |
| `--format <format>` | | Output format: ascii, json, csv, table |
| `--no-header` | | Don't output column headers (CSV) |
| `--field-separator <sep>` | | Field separator for CSV (default: ,) |
| `--page-size <n>` | | Rows per batch (default: 100) |

#### General Options
| Option | Short | Description |
|--------|-------|-------------|
| `--config-file <path>` | | Path to config file (overrides default locations) |
| `--help` | `-h` | Show help message |
| `--version` | `-v` | Print version and exit |

### Batch Mode Examples

Execute CQL statements non-interactively (compatible with cqlsh):

```bash
# Execute a single statement
cqlai -e "SELECT * FROM system_schema.keyspaces;"

# Execute from a file
cqlai -f script.cql

# Pipe input
echo "SELECT * FROM users;" | cqlai

# Control output format
cqlai -e "SELECT * FROM users;" --format json
cqlai -e "SELECT * FROM users;" --format csv --no-header

# Control pagination size
cqlai -e "SELECT * FROM large_table;" --page-size 50
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
  .ai What keyspaces are there?
  .ai What columns does the users table have?
  .ai create a table for storing product inventory
  .ai delete orders older than 1 year from the orders table
  ```

### Keyboard Shortcuts

#### Navigation & Control
| Shortcut | Action | macOS Alternative |
|----------|--------|-------------------|
| `↑`/`↓` | Navigate command history | Same |
| `Ctrl+P`/`Ctrl+N` | Previous/Next in command history | Same |
| `Alt+N` | Move to next line in history | `Option+N` |
| `Tab` | Autocomplete commands and table/keyspace names | Same |
| `Ctrl+C` | Clear input / Cancel pagination / Cancel operation (twice to exit) | `⌘+C` or `Ctrl+C` |
| `Ctrl+D` | Exit application | `⌘+D` or `Ctrl+D` |
| `Ctrl+R` | Search command history | `⌘+R` or `Ctrl+R` |
| `Esc` | Toggle navigation mode / Cancel pagination / Close modals | Same |
| `Enter` | Execute command / Load next page (during pagination) | Same |

#### Text Editing
| Shortcut | Action | macOS Alternative |
|----------|--------|-------------------|
| `Ctrl+A` | Jump to beginning of line | Same |
| `Ctrl+E` | Jump to end of line | Same |
| `Ctrl+Left`/`Ctrl+Right` | Jump by word (or 20 chars) | Same |
| `PgUp`/`PgDn` (in input) | Page left/right in long queries | `Fn+↑`/`Fn+↓` |
| `Ctrl+K` | Cut from cursor to end of line | Same |
| `Ctrl+U` | Cut from beginning to cursor | Same |
| `Ctrl+W` | Cut word backward | Same |
| `Alt+D` | Delete word forward | `Option+D` |
| `Ctrl+Y` | Paste previously cut text | Same |

#### View Switching
| Shortcut | Action |
|----------|--------|
| `F2` | Switch to query/history view |
| `F3` | Switch to table view |
| `F4` | Switch to trace view (when tracing enabled) |
| `F5` | Switch to AI conversation view |
| `F6` | Toggle column data types in table headers |

#### Scrolling & Table Navigation
| Shortcut | Action | macOS Alternative |
|----------|--------|-------------------|
| `PgUp`/`PgDn` | Scroll viewport by page / Load more data when available | `Fn+↑`/`Fn+↓` |
| `Space` | Load next page when more data available | Same |
| `Enter` (empty input) | Load next page when more data available | Same |
| `Alt+↑`/`Alt+↓` | Scroll viewport by single row (respects row boundaries) | `Option+↑`/`Option+↓` |
| `Alt+←`/`Alt+→` | Scroll table horizontally (wide tables) | `Option+←`/`Option+→` |
| `↑`/`↓` | Navigate table rows (when in navigation mode) | Same |

#### Navigation Mode (Table/Trace Views)
Press `Esc` to toggle navigation mode when viewing tables or traces.

| Shortcut | Action in Navigation Mode |
|----------|---------------------------|
| `j` / `k` | Scroll down/up by single line |
| `d` / `u` | Scroll down/up by half page |
| `g` / `G` | Jump to top/bottom of results |
| `<` / `>` | Scroll left/right by 10 columns |
| `{` / `}` | Scroll left/right by 50 columns |
| `0` / `$` | Jump to first/last column |
| `Esc` | Exit navigation mode / Cancel pagination if active |

#### Mouse Support
| Action | Function |
|--------|----------|
| Mouse Wheel | Scroll vertically with automatic data loading |
| Alt+Mouse Wheel | Scroll horizontally in tables |
| Shift+Mouse Wheel | Scroll horizontally (alternative) |
| Ctrl+Mouse Wheel | Scroll horizontally (alternative) |
| Shift+Click+Drag | Select text for copying |
| Ctrl+Shift+C | Copy selected text to clipboard |
| Middle Click | Paste from selection buffer (Linux/Unix) |

**Note for macOS Users:**
- Most `Ctrl` shortcuts work as-is on macOS, but you can also use `⌘` (Command) key as an alternative
- `Alt` key is labeled as `Option` on Mac keyboards
- Function keys (F1-F6) may require holding `Fn` key depending on your Mac settings

### Tab Completion

CQLAI provides intelligent, context-aware tab completion to speed up your workflow. Press `Tab` at any point to see available completions.

#### What Can Be Completed

**CQL Keywords & Commands:**
- All CQL keywords: `SELECT`, `INSERT`, `CREATE`, `ALTER`, `DROP`, etc.
- Meta-commands: `DESCRIBE`, `CONSISTENCY`, `COPY`, `SHOW`, etc.
- Data types: `TEXT`, `INT`, `UUID`, `TIMESTAMP`, etc.
- Consistency levels: `ONE`, `QUORUM`, `ALL`, `LOCAL_QUORUM`, etc.

**Schema Objects:**
- Keyspace names
- Table names (within current keyspace)
- Column names (when context allows)
- User-defined type names
- Function and aggregate names
- Index names

**Context-Aware Completions:**
```sql
-- After SELECT, suggests column names and keywords
SELECT <Tab>           -- Shows: *, column names, DISTINCT, JSON, etc.

-- After FROM, suggests table names
SELECT * FROM <Tab>    -- Shows: available tables in current keyspace

-- After USE, suggests keyspace names
USE <Tab>              -- Shows: available keyspaces

-- After DESCRIBE, suggests object types
DESCRIBE <Tab>         -- Shows: KEYSPACE, TABLE, TYPE, etc.

-- After consistency command
CONSISTENCY <Tab>      -- Shows: ONE, QUORUM, ALL, etc.
```

**File Path Completion:**
```sql
-- For commands that accept file paths
SOURCE '<Tab>          -- Shows: files in current directory
SOURCE '/path/<Tab>    -- Shows: files in /path/
```

#### Completion Behavior

- **Case Insensitive:** Type `sel<Tab>` to get `SELECT`
- **Partial Matching:** Type part of a word and press Tab
- **Multiple Matches:** When multiple completions are available:
  - First Tab: Shows inline completion if unique
  - Second Tab: Shows all available options in a modal
- **Smart Filtering:** Completions are filtered based on current context
- **Escape to Cancel:** Press `Esc` to close the completion modal

#### Examples

```sql
-- Complete table name
SELECT * FROM us<Tab>
-- Completes to: SELECT * FROM users

-- Complete consistency level
CONSISTENCY LOC<Tab>
-- Shows: LOCAL_ONE, LOCAL_QUORUM, LOCAL_SERIAL

-- Complete column names after SELECT
SELECT id, na<Tab> FROM users
-- Completes to: SELECT id, name FROM users

-- Complete file paths for SOURCE command
SOURCE 'sche<Tab>
-- Completes to: SOURCE 'schema.cql'

-- Complete COPY command options
COPY users TO 'file.csv' WITH <Tab>
-- Shows: HEADER, DELIMITER, NULLVAL, PAGESIZE, etc.

-- Show all tables when multiple exist
SELECT * FROM <Tab>
-- Shows modal with: users, orders, products, etc.
```

#### Tips for Effective Use

1. **Use Tab liberally:** The completion system is smart and context-aware
2. **Type minimum characters:** Often 2-3 characters are enough to get unique completion
3. **Use for discovery:** Press Tab on empty input to see what's available
4. **File paths:** Remember to include quotes for file path completion
5. **Navigate completions:** Use arrow keys to select from multiple options

## ⚙️ Available Commands

CQLAI supports all standard CQL commands plus additional meta-commands for enhanced functionality.

### CQL Commands
Execute any valid CQL statement supported by your Cassandra cluster:
- DDL: `CREATE`, `ALTER`, `DROP` (KEYSPACE, TABLE, INDEX, etc.)
- DML: `SELECT`, `INSERT`, `UPDATE`, `DELETE`
- DCL: `GRANT`, `REVOKE`
- Other: `USE`, `TRUNCATE`, `BEGIN BATCH`, etc.

### Meta-Commands

Meta-commands provide additional functionality beyond standard CQL:

#### Session Management
- **CONSISTENCY** `<level>` - Set consistency level (ONE, QUORUM, ALL, etc.)
  ```sql
  CONSISTENCY QUORUM
  CONSISTENCY LOCAL_ONE
  ```

- **PAGING** `<size>` | OFF - Set result paging size
  ```sql
  PAGING 1000
  PAGING OFF
  ```

- **TRACING** ON | OFF - Enable/disable query tracing
  ```sql
  TRACING ON
  SELECT * FROM users;
  TRACING OFF
  ```

- **OUTPUT** [FORMAT] - Set output format
  ```sql
  OUTPUT          -- Show current format
  OUTPUT TABLE    -- Table format (default)
  OUTPUT JSON     -- JSON format
  OUTPUT EXPAND   -- Expanded vertical format
  OUTPUT ASCII    -- ASCII table format
  ```

#### Schema Description
- **DESCRIBE** - Show schema information
  ```sql
  DESCRIBE KEYSPACES                    -- List all keyspaces
  DESCRIBE KEYSPACE <name>              -- Show keyspace definition
  DESCRIBE TABLES                       -- List tables in current keyspace
  DESCRIBE TABLE <name>                 -- Show table structure
  DESCRIBE TYPES                        -- List user-defined types
  DESCRIBE TYPE <name>                  -- Show UDT definition
  DESCRIBE FUNCTIONS                    -- List user functions
  DESCRIBE FUNCTION <name>              -- Show function definition
  DESCRIBE AGGREGATES                   -- List user aggregates
  DESCRIBE AGGREGATE <name>             -- Show aggregate definition
  DESCRIBE MATERIALIZED VIEWS           -- List materialized views
  DESCRIBE MATERIALIZED VIEW <name>     -- Show view definition
  DESCRIBE INDEX <name>                 -- Show index definition
  DESCRIBE CLUSTER                      -- Show cluster information
  DESC <keyspace>.<table>               -- Shorthand for table description
  ```

#### Data Export/Import
- **COPY TO** - Export table data to CSV or Parquet file
  ```sql
  -- Basic export to CSV
  COPY users TO 'users.csv'

  -- Export to Parquet format (auto-detected by extension)
  COPY users TO 'users.parquet'

  -- Export to Parquet with explicit format and compression
  COPY users TO 'data.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY'

  -- Export specific columns
  COPY users (id, name, email) TO 'users_partial.csv'

  -- Export with options
  COPY users TO 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|'

  -- Export to stdout
  COPY users TO STDOUT WITH HEADER = TRUE

  -- Available options:
  -- FORMAT = 'CSV'/'PARQUET' -- Output format (default: CSV, auto-detected)
  -- HEADER = TRUE/FALSE      -- Include column headers (CSV only)
  -- DELIMITER = ','          -- Field delimiter (CSV only)
  -- NULLVAL = 'NULL'        -- String to use for NULL values
  -- PAGESIZE = 1000         -- Rows per page for large exports
  -- COMPRESSION = 'SNAPPY'  -- For Parquet: SNAPPY, GZIP, ZSTD, LZ4, NONE
  -- CHUNKSIZE = 10000       -- Rows per chunk for Parquet
  ```

- **COPY FROM** - Import CSV or Parquet data into table
  ```sql
  -- Basic import from CSV file
  COPY users FROM 'users.csv'

  -- Import from Parquet file (auto-detected)
  COPY users FROM 'users.parquet'

  -- Import from Parquet with explicit format
  COPY users FROM 'data.parquet' WITH FORMAT='PARQUET'

  -- Import with header row (CSV)
  COPY users FROM 'users.csv' WITH HEADER = TRUE

  -- Import specific columns
  COPY users (id, name, email) FROM 'users_partial.csv'

  -- Import from stdin
  COPY users FROM STDIN

  -- Import with custom options
  COPY users FROM 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|' AND NULLVAL = 'N/A'

  -- Available options:
  -- HEADER = TRUE/FALSE      -- First row contains column names
  -- DELIMITER = ','          -- Field delimiter
  -- NULLVAL = 'NULL'        -- String representing NULL values
  -- MAXROWS = -1            -- Maximum rows to import (-1 = unlimited)
  -- SKIPROWS = 0            -- Number of initial rows to skip
  -- MAXPARSEERRORS = -1     -- Max parsing errors allowed (-1 = unlimited)
  -- MAXINSERTERRORS = 1000  -- Max insert errors allowed
  -- MAXBATCHSIZE = 20       -- Max rows per batch insert
  -- MINBATCHSIZE = 2        -- Min rows per batch insert
  -- CHUNKSIZE = 5000        -- Rows between progress updates
  -- ENCODING = 'UTF8'       -- File encoding
  -- QUOTE = '"'             -- Quote character for strings
  ```

- **CAPTURE** - Capture query output to file (continuous recording)
  ```sql
  CAPTURE 'output.txt'          -- Start capturing to text file
  CAPTURE JSON 'output.json'    -- Capture as JSON
  CAPTURE CSV 'output.csv'      -- Capture as CSV
  SELECT * FROM users;
  CAPTURE OFF                   -- Stop capturing
  ```

- **SAVE** - Save displayed query results to file (without re-executing)
  ```sql
  -- First run a query
  SELECT * FROM users WHERE status = 'active';

  -- Then save the displayed results in various formats:
  SAVE                           -- Interactive dialog (choose format & filename)
  SAVE 'users.csv'               -- Save to CSV (format auto-detected)
  SAVE 'users.json'              -- Save to JSON (format auto-detected)
  SAVE 'users.txt' ASCII         -- Save as ASCII table
  SAVE 'data.csv' CSV            -- Explicitly specify format

  -- Key differences from CAPTURE:
  -- - SAVE exports the currently displayed results
  -- - No need to re-run the query
  -- - Preserves exact data shown in terminal
  -- - Works with paginated results (saves only loaded pages)
  ```

#### Information Display
- **SHOW** - Display session information
  ```sql
  SHOW VERSION          -- Show Cassandra version
  SHOW HOST            -- Show current connection details
  SHOW SESSION         -- Show all session settings
  ```

- **EXPAND** ON | OFF - Toggle expanded output mode
  ```sql
  EXPAND ON            -- Vertical output (one field per line)
  SELECT * FROM users WHERE id = 1;
  EXPAND OFF           -- Normal table output
  ```

#### Script Execution
- **SOURCE** - Execute CQL scripts from file
  ```sql
  SOURCE 'schema.cql'           -- Execute script
  SOURCE '/path/to/script.cql'  -- Absolute path
  ```

#### Help
- **HELP** - Display command help
  ```sql
  HELP                 -- Show all commands
  HELP DESCRIBE        -- Help for specific command
  HELP CONSISTENCY     -- Help for consistency levels
  ```

### AI Commands
- **.ai** `<natural language query>` - Generate CQL from natural language
  ```sql
  .ai show all users with active status
  .ai create a table for storing user sessions
  .ai find orders placed in the last 30 days
  ```

## 🛠️ Configuration

CQLAI supports multiple configuration methods for maximum flexibility and compatibility with existing Cassandra setups.

### Configuration Precedence

Configuration sources are loaded in the following order (later sources override earlier ones):

1. **CQLSHRC files** (for compatibility with existing cqlsh setups)
   - `~/.cassandra/cqlshrc` (standard location)
   - `~/.cqlshrc` (alternative location)
   - `$CQLSH_RC` (if environment variable is set)

2. **CQLAI JSON configuration files**
   - `./cqlai.json` (current directory)
   - `~/.cqlai.json` (user home directory)
   - `~/.config/cqlai/config.json` (XDG config directory)

3. **Environment variables**
   - `CQLAI_HOST`, `CQLAI_PORT`, `CQLAI_KEYSPACE`, etc.
   - `CASSANDRA_HOST`, `CASSANDRA_PORT` (for compatibility)

4. **Command-line flags** (highest priority)
   - `--host`, `--port`, `--keyspace`, `--username`, `--password`, etc.

### CQLSHRC Compatibility

CQLAI can read standard CQLSHRC files used by the traditional `cqlsh` tool, making migration seamless.

**Supported CQLSHRC sections:**
- `[connection]` - hostname, port, ssl settings
- `[authentication]` - keyspace, credentials file path
- `[auth_provider]` - authentication module and username
- `[ssl]` - SSL/TLS certificate configuration

**Example CQLSHRC file:**
```ini
; ~/.cassandra/cqlshrc
[connection]
hostname = cassandra.example.com
port = 9042
ssl = true

[authentication]
keyspace = my_keyspace
credentials = ~/.cassandra/credentials

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client-key.pem
usercert = ~/certs/client-cert.pem
validate = true
```

See [CQLSHRC_SUPPORT.md](docs/CQLSHRC_SUPPORT.md) for complete CQLSHRC compatibility details.

### CQLAI JSON Configuration

For advanced features and AI configuration, CQLAI uses its own JSON format:

**Example `cqlai.json`:**
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  "keyspace": "",
  "username": "cassandra",
  "password": "cassandra",
  "requireConfirmation": true,
  "consistency": "LOCAL_ONE",
  "pageSize": 100,
  "maxMemoryMB": 10,
  "connectTimeout": 10,
  "requestTimeout": 10,
  "debug": false,
  "historyFile": "~/.cqlai/history",
  "aiHistoryFile": "~/.cqlai/ai_history",
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
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

**Note:** You can also use the `url` field to override the API endpoint for OpenAI-compatible APIs:
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "your-api-key",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

### AI Provider Configuration

**Note:** AI features are completely optional. CQLAI works as a full-featured CQL shell without any AI configuration.

To enable AI-powered query generation, configure your preferred provider in the `ai` section of your `cqlai.json` file.

#### OpenAI (GPT-4 & GPT-3.5)

Use OpenAI for high-quality, general-purpose query generation. Requires an OpenAI API key.

- **Get API Key:** [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **Recommended Models:**
  - `gpt-4-turbo-preview` (default, recommended for best results)
  - `gpt-3.5-turbo` (faster, more cost-effective)

**Configuration:**
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

#### Anthropic (Claude 3)

Use Anthropic for powerful, context-aware models. Ideal for complex queries and reasoning. Requires an Anthropic API key.

- **Get API Key:** [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)
- **Recommended Models:**
  - `claude-3-opus-20240229` (most powerful)
  - `claude-3-sonnet-20240229` (default, balanced performance)
  - `claude-3-haiku-20240307` (fastest)

**Configuration:**
```json
{
  "ai": {
    "provider": "anthropic",
    "apiKey": "sk-ant-...",
    "model": "claude-3-sonnet-20240229"
  }
}
```

#### Google Gemini

Use Google Gemini for a fast and capable model from Google. Requires a Google AI Studio API key.

- **Get API Key:** [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)
- **Recommended Model:**
  - `gemini-pro` (default)

**Configuration:**
```json
{
  "ai": {
    "provider": "gemini",
    "apiKey": "...",
    "model": "gemini-pro"
  }
}
```

#### Synthetic (Multiple Open-Source Models)

Use Synthetic to access a vast selection of open-source AI models at very reasonable prices. Synthetic provides an OpenAI-compatible API that makes it easy to work with various open-source models.

- **Get Started:** [synthetic.new](https://synthetic.new/)
- **API Documentation:** [dev.synthetic.new/docs](https://dev.synthetic.new/docs)
- **Recommended Model:**
  - `hf:Qwen/Qwen3-235B-A22B-Instruct-2507` (recommended, though we haven't extensively tested all models)
- **Available Models:** See [Always-On Models](https://dev.synthetic.new/docs/api/models#always-on-models)

**Configuration:**
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "your-synthetic-api-key",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

**Key Benefits:**
- Access to a wide variety of open-source models
- Cost-effective pricing
- OpenAI-compatible API for easy integration
- No vendor lock-in

**Notes:**
- Synthetic presents an OpenAI-compatible interface, so you use the `openai` provider in your configuration
- The `url` field overrides the default OpenAI endpoint to point to Synthetic
- API key is required - obtain one from [synthetic.new](https://synthetic.new/)

#### Ollama (Local Models)

Use Ollama for running AI models locally or connecting to OpenAI-compatible APIs. Ollama allows you to run powerful language models on your own hardware without sending data to external services.

- **Get Started:** [ollama.ai](https://ollama.ai)
- **Recommended Models:**
  - `llama3.2` (Meta's Llama 3.2)
  - `codellama` (Code-specialized Llama)
  - `mistral` (Mistral AI's model)
  - `qwen2.5-coder` (Alibaba's code model)

**Configuration:**
```json
{
  "ai": {
    "provider": "ollama",
    "model": "llama3.2",
    "url": "http://localhost:11434/v1"
  }
}
```

**Environment Variables:**
- `OLLAMA_URL` - Custom Ollama server URL (default: `http://localhost:11434/v1`)
- `OLLAMA_MODEL` - Model to use

**Notes:**
- No API key required for local Ollama installations
- Supports custom URLs for remote Ollama servers or OpenAI-compatible endpoints
- The `url` field can be set at the top level (`ai.url`) or provider-specific (`ai.ollama.url`)

#### OpenRouter (Multiple Models)

Use OpenRouter to access multiple AI models through a single API. OpenRouter provides access to various models from different providers.

- **Get API Key:** [openrouter.ai/keys](https://openrouter.ai/keys)
- **Available Models:** See [openrouter.ai/models](https://openrouter.ai/models)

**Configuration:**
```json
{
  "ai": {
    "provider": "openrouter",
    "apiKey": "sk-or-...",
    "model": "anthropic/claude-3-sonnet",
    "url": "https://openrouter.ai/api/v1"
  }
}
```

**Environment Variables:**
- `OPENROUTER_API_KEY` - OpenRouter API key
- `OPENROUTER_MODEL` - Model to use
- `OPENROUTER_URL` - Custom OpenRouter URL (default: `https://openrouter.ai/api/v1`)

#### Mock Provider (for Testing)

The `mock` provider is the default and requires no API key. It's useful for testing the AI workflow or for users who don't need real AI capabilities. It generates simple, predictable queries based on keywords.

**Configuration:**
```json
{
  "ai": {
    "provider": "mock"
  }
}
```

#### Using Environment Variables for API Keys and URLs

For better security, you can provide API keys and custom URLs via environment variables instead of writing them in the configuration file.

**API Keys:**
- **OpenAI:** `OPENAI_API_KEY`
- **Anthropic:** `ANTHROPIC_API_KEY`
- **Google Gemini:** `GEMINI_API_KEY`
- **OpenRouter:** `OPENROUTER_API_KEY`

**Custom URLs:**
- **Ollama:** `OLLAMA_URL` (default: `http://localhost:11434/v1`)
- **OpenRouter:** `OPENROUTER_URL` (default: `https://openrouter.ai/api/v1`)

If an environment variable is set, it will be used even if a value is present in `cqlai.json`.

**Configuration Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Cassandra host address |
| `port` | number | `9042` | Cassandra port |
| `keyspace` | string | `""` | Default keyspace to use |
| `username` | string | `""` | Authentication username |
| `password` | string | `""` | Authentication password |
| `requireConfirmation` | boolean | `true` | Require confirmation for dangerous commands |
| `consistency` | string | `LOCAL_ONE` | Default consistency level (ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE) |
| `pageSize` | number | `100` | Number of rows per page |
| `maxMemoryMB` | number | `10` | Maximum memory for query results in MB |
| `connectTimeout` | number | `10` | Connection timeout in seconds |
| `requestTimeout` | number | `10` | Request timeout in seconds |
| `historyFile` | string | `~/.cqlai/history` | Path to CQL command history file (supports `~` expansion) |
| `aiHistoryFile` | string | `~/.cqlai/ai_history` | Path to AI command history file (supports `~` expansion) |
| `debug` | boolean | `false` | Enable debug logging |

### Configuration File Locations

CQLAI searches for configuration files in the following locations:

**CQLSHRC files:**
1. `$CQLSH_RC` (if environment variable is set)
2. `~/.cassandra/cqlshrc` (standard cqlsh location)
3. `~/.cqlshrc` (alternative location)

**CQLAI JSON files:**
1. `./cqlai.json` (current working directory)
2. `~/.cqlai.json` (user home directory)
3. `~/.config/cqlai/config.json` (XDG config directory on Linux/macOS)

### Environment Variables

Common environment variables:
- `CQLAI_HOST` or `CASSANDRA_HOST` - Cassandra host
- `CQLAI_PORT` or `CASSANDRA_PORT` - Cassandra port
- `CQLAI_KEYSPACE` - Default keyspace
- `CQLAI_USERNAME` - Authentication username
- `CQLAI_PASSWORD` - Authentication password
- `CQLAI_PAGE_SIZE` - Batch mode pagination size (default: 100)
- `CQLSH_RC` - Path to custom CQLSHRC file

### Migration from cqlsh

If you're migrating from `cqlsh`, CQLAI will automatically read your existing `~/.cassandra/cqlshrc` file. No changes are needed to start using CQLAI with your existing Cassandra configuration.

## 🤖 AI-Powered Query Generation

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

- **[OpenAI](https://openai.com/)** (GPT-4, GPT-3.5)
- **[Anthropic](https://www.anthropic.com/)** (Claude 3)
- **[Google Gemini](https://ai.google.dev/)**
- **[Synthetic](https://synthetic.new/)** (Multiple open-source models)
- **[Ollama](https://ollama.ai/)** (Local models or OpenAI-compatible APIs)
- **[OpenRouter](https://openrouter.ai/)** (Access to multiple models)
- **Mock** (default, for testing without API keys)

### Safety Features

- **Read-only by default**: AI prefers SELECT queries unless explicitly asked to modify
- **Dangerous operation warnings**: DROP, DELETE, TRUNCATE operations show warnings
- **Confirmation required**: Destructive operations require additional confirmation
- **Schema validation**: Queries are validated against your current schema

---

## 🔌 MCP Server (Model Context Protocol)

CQLAI includes a built-in MCP server that enables AI assistants like **Claude Code** and **Claude Desktop** to interact with Cassandra databases. The MCP server provides a comprehensive API for querying, schema management, and database administration with robust security controls.

### Key Features

- **20 MCP Tools**: Complete Cassandra administration via AI
- **37 CQL Operations**: SELECT, INSERT, UPDATE, DELETE, ALTER, CREATE INDEX, GRANT, and more
- **Permission System**: 6-category classification (DQL, DML, DDL, DCL, SESSION, FILE)
- **Confirmation Workflows**: Dangerous queries require user approval
- **Query History**: Complete audit trail with lifecycle event logging
- **Trace Analysis**: Performance analysis via Cassandra traces
- **Security Controls**: Multiple layers including permission modes, confirmation requirements, and approval gates

### Quick Start

```bash
# In CQLAI, start MCP server
.mcp start --dba_mode

# Or with custom configuration
.mcp start --config-file ~/.cqlai/mcp_config.json
```

Then configure Claude Code/Desktop to connect to the MCP server.

### Security Model

The MCP server implements defense-in-depth security:

1. **Permission Modes**: readonly, readwrite, or dba
2. **Confirmation Requirements**: Dangerous queries require approval
3. **User Confirmation Flag**: Tools require explicit user consent
4. **Request Approval Gate**: Optional MCP tool approval (disabled by default)
5. **Runtime Lockdown**: Prevent configuration changes
6. **Complete Audit Trail**: All operations logged to history file

**Default**: Secure (readonly mode, MCP approval disabled)

### Documentation

**Complete MCP documentation**: See [MCP.md](MCP.md) for:
- All 20 MCP tools with examples
- Configuration reference (JSON and CLI)
- Security best practices
- Confirmation workflows
- History file format
- Troubleshooting guide

---

## 📦 Apache Parquet Support

CQLAI provides comprehensive support for Apache Parquet format, making it ideal for data analytics workflows and integration with modern data ecosystems.

### Key Benefits

- **Efficient Storage**: Columnar format with excellent compression (50-80% smaller than CSV)
- **Fast Analytics**: Optimized for analytical queries in Spark, Presto, and other engines
- **Type Preservation**: Maintains Cassandra data types including collections and UDTs
- **Machine Learning Ready**: Direct compatibility with pandas, PyArrow, and ML frameworks
- **Streaming Support**: Memory-efficient streaming for large datasets

### Quick Examples

```sql
-- Export to Parquet (auto-detected by extension)
COPY users TO 'users.parquet';

-- Export with compression
COPY events TO 'events.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='ZSTD';

-- Import from Parquet
COPY users FROM 'users.parquet';

-- Capture query results in Parquet format
CAPTURE 'results.parquet' FORMAT='PARQUET';
SELECT * FROM large_table WHERE condition = true;
CAPTURE OFF;
```

### Supported Features

- All Cassandra primitive types (int, text, timestamp, uuid, etc.)
- Collection types (list, set, map)
- User-Defined Types (UDTs)
- Frozen collections
- Vector types for ML workloads (Cassandra 5.0+)
- Multiple compression algorithms (Snappy, GZIP, ZSTD, LZ4)

For detailed documentation, see [Parquet Support Guide](docs/PARQUET.md).

## ⚠️ Known Limitations

### JSON Output (CAPTURE JSON and --format json)

When outputting data as JSON, there are some limitations due to how the underlying gocql driver handles dynamic typing:

#### NULL Values
- **Issue**: NULL values in primitive columns (int, boolean, text, etc.) appear as zero values (`0`, `false`, `""`) instead of `null`
- **Cause**: The gocql driver returns zero values for NULLs when scanning into dynamic types (`interface{}`)
- **Workaround**: Use `SELECT JSON` queries which return proper JSON from Cassandra server-side

#### User-Defined Types (UDTs)
- **Issue**: UDT columns appear as empty objects `{}` in JSON output
- **Cause**: The gocql driver cannot properly unmarshal UDTs without compile-time knowledge of their structure
- **Workaround**: Use `SELECT JSON` queries for proper UDT serialization

#### Example
```sql
-- Regular SELECT (has limitations)
SELECT * FROM users;  
-- Returns: {"id": 1, "age": 0, "active": false}  -- age and active might be NULL

-- Using SELECT JSON (preserves types correctly)
SELECT JSON * FROM users;
-- Returns: {"id": 1, "age": null, "active": null}  -- NULLs properly represented
```

**Note**: Complex types (lists, sets, maps, vectors) are properly preserved in JSON output.

## 🔨 Development

To work on `cqlai`, you'll need Go (≥ 1.24).

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


## 🏗️ Technology Stack

- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUI Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Cassandra Driver:** [gocql](https://github.com/gocql/gocql)

## 🙏 Acknowledgements

CQLAI builds upon the foundation laid by several open-source projects, particularly Apache Cassandra. We extend our sincere gratitude to the Apache Cassandra community for their outstanding work and contributions to the field of distributed databases.

Apache Cassandra is a free and open-source, distributed, wide-column store, NoSQL database management system designed to handle large amounts of data across many commodity servers, providing high availability with no single point of failure.

### Apache Cassandra Resources

- **Official Website**: [cassandra.apache.org](https://cassandra.apache.org/)
- **Source Code**: Available on [GitHub](https://github.com/apache/cassandra) or the Apache Git repository at `gitbox.apache.org/repos/asf/cassandra.git`
- **Documentation**: Comprehensive guides and references available at the [Apache Cassandra website](https://cassandra.apache.org/)

CQLAI incorporates and extends functionality from various Cassandra tools and utilities, enhancing them to provide a modern, efficient terminal experience for Cassandra developers and DBAs.

We encourage users to explore and contribute to the main Apache Cassandra project, as well as to provide feedback and suggestions for CQLAI through our [GitHub discussions](https://github.com/axonops/cqlai/discussions) and [issues](https://github.com/axonops/cqlai/issues) pages.

## 💬 Community & Support

### Get Involved
- 💡 **Share Ideas**: Visit our [GitHub Discussions](https://github.com/axonops/cqlai/discussions) to propose new features
- 🐛 **Report Issues**: Found a bug? [Open an issue](https://github.com/axonops/cqlai/issues/new/choose)
- 🤝 **Contribute**: We welcome pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines
- ⭐ **Star Us**: If you find CQLAI useful, please star our repository!

### Stay Connected
- 🌐 **Website**: [axonops.com](https://axonops.com)
- 📧 **Contact**: Visit our website for support options

## 📝 License

This project is licensed under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.

Third-party dependency licenses are available in the [THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/) directory. To regenerate license attributions, run `make licenses`.

## ⚖️ Legal Notices

*This project may contain trademarks or logos for projects, products, or services. Any use of third-party trademarks or logos are subject to those third-party's policies.*

- **AxonOps** is a registered trademark of AxonOps Limited.
- **Apache**, **Apache Cassandra**, **Cassandra**, **Apache Spark**, **Spark**, **Apache TinkerPop**, **TinkerPop**, **Apache Kafka** and **Kafka** are either registered trademarks or trademarks of the Apache Software Foundation or its subsidiaries in Canada, the United States and/or other countries.
- **DataStax** is a registered trademark of DataStax, Inc. and its subsidiaries in the United States and/or other countries.

---

<div align="center">
  <p>Made with ❤️ by the <a href="https://axonops.com">AxonOps</a> Team</p>
</div>
