# CQLAI - AI-Powered Cassandra CQL TUI

<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">
</div>

**CQLAI** is a fast, portable, and AI-enhanced interactive terminal for Cassandra (CQL), built in Go. It provides a modern, user-friendly alternative to `cqlsh` with an advanced terminal UI, natural language query generation, client-side command parsing, and enhanced productivity features.

The original cqlsh command is written in Python which requires Python to be installed on the system. cqlai is compiled to a single execuable binary, requiring no external dependencies. This project provides binaries for the following platforms:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


It is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), and [Lip Gloss](https://github.com/charmbracelet/lipgloss) for the beautiful terminal UI. A big shout out to the cassandra gocql driver team for implementing the latest Cassandra functionalities [gocql](https://github.com/apache/cassandra-gocql-driver)


---

## Project Status

**CQLAI is production-ready** and actively used in development, testing, and production environments with Cassandra clusters. The tool provides a complete, stable alternative to `cqlsh` with enhanced features and performance.

### What Works
- All core CQL operations and queries
- Complete meta-command support (`DESCRIBE`, `SHOW`, `CONSISTENCY`, etc.)
- Client-side command parsing (lightweight, no ANTLR dependency)
- Data import/export with `COPY TO/FROM` (CSV and Parquet formats)
- AI-powered query generation (OpenAI, Anthropic, Gemini)
- SSL/TLS connections and authentication
- User-Defined Types (UDTs) and complex data types
- Batch mode for scripting and automation
- Apache Parquet format support for efficient data interchange
- Tab completion for CQL keywords, tables, columns, and keyspaces
- Small binary size (~43MB, 53% smaller than previous versions)

### Coming Soon
- Enhanced AI context awareness
- Cassandra MCP service
- Additional performance optimizations

We encourage you to **try CQLAI today** and help shape its development! Your feedback and contributions are invaluable in making this the best CQL shell for the Cassandra community. Please [report issues](https://github.com/axonops/cqlai/issues) or [contribute](https://github.com/axonops/cqlai/pulls).

---

## Features

- **Interactive CQL Shell:** Execute any CQL query that your Cassandra cluster supports.
- **Rich Terminal UI:**
    - A multi-layer, full-screen terminal application with alternate screen buffer (preserves terminal history).
    - Virtualized, scrollable table for results with automatic data loading, preventing memory overload from large queries.
    - Advanced navigation modes with vim-style keyboard shortcuts.
    - Full mouse support including wheel scrolling and text selection.
    - Sticky footer/status bar showing connection details, query latency, and session status (consistency, tracing).
    - Modal overlays for history, help, and command completion.
- **AI-Powered Query Generation:**
    - Natural language to CQL conversion using AI providers (OpenAI, Anthropic, Gemini).
    - Schema-aware query generation with automatic context.
    - Safe preview and confirmation before execution.
    - Support for complex operations including DDL and DML.
- **Apache Parquet Support:**
    - High-performance columnar data format for analytics and machine learning workflows.
    - Export Cassandra tables to Parquet files with `COPY TO` command.
    - Import Parquet files into Cassandra with automatic schema inference.
    - Partitioned datasets with Hive-style directory structures.
    - TimeUUID / timestamp virtual columns for intelligent time-based partitioning.
    - Support for all Cassandra data types including UDTs, collections, and vectors.
- **Configuration:**
    - Simple configuration via `cqlai.json` in current directory or `~/.cqlai.json`.
    - Support for SSL/TLS connections with certificate authentication.
- **Single Binary:** Distributed as a single, static binary with no external dependencies. Fast startup and small footprint.

## Installation

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

## Usage

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

## Available Commands

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

## Configuration

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
   - `/etc/cqlai/config.json` (system-wide)

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
  "pageSize": 100,
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
4. `/etc/cqlai/config.json` (system-wide configuration)

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

## Apache Parquet Support

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

## Known Limitations

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

## Development

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


## Technology Stack

- **Language:** Go
- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **TUI Components:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Styling:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Cassandra Driver:** [gocql](https://github.com/gocql/gocql)

## License

This project is licensed under the Apache 2.0 license. See the LICENSE file for details.

Third-party dependency licenses are available in the [THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/) directory. To regenerate license attributions, run `make licenses`.

---

<div align="center">
  <br>
  <p>Developed by</p>
  <img src="./assets/AxonOps-RGB-transparent-small.png" alt="AxonOps" width="200">
</div>