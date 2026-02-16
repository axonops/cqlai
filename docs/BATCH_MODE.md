# Batch Mode Documentation

CQLAI supports batch execution mode compatible with `cqlsh`, allowing you to execute CQL commands non-interactively with automatic pagination and multiple output formats.

## Overview

Batch mode is activated when:
- You use the `-e` flag to execute a command directly
- You use the `-f` flag to execute commands from a file
- You pipe input to `cqlai` (stdin is not a terminal)

In batch mode, CQLAI automatically iterates through all result pages without manual intervention, similar to `cqlsh`.

## Command-Line Options

### Execute CQL Directly (`-e`)

Execute a single CQL statement and exit:

```bash
cqlai -e "SELECT * FROM system.local;"
```

### Execute from File (`-f`)

Execute CQL statements from a file:

```bash
cqlai -f script.cql
```

The file can contain multiple statements separated by semicolons.

### Pipe Input

Execute CQL from stdin:

```bash
echo "SELECT * FROM system.local;" | cqlai

# Or from a here-document
cqlai <<EOF
USE my_keyspace;
SELECT * FROM my_table;
EOF
```

## Output Formats

Use the `--format` flag to specify the output format:

### ASCII Table (Default)

```bash
cqlai -e "SELECT * FROM table;" --format ascii
```

Produces output similar to `cqlsh`:

```
 column1 | column2 | column3
---------+---------+---------
 value1  | value2  | value3
 value4  | value5  | value6

(2 rows)
```

### JSON

```bash
cqlai -e "SELECT * FROM table;" --format json
```

Produces JSON array output:

```json
[
  {
    "column1": "value1",
    "column2": "value2",
    "column3": "value3"
  },
  {
    "column1": "value4",
    "column2": "value5",
    "column3": "value6"
  }
]
```

### CSV

```bash
cqlai -e "SELECT * FROM table;" --format csv
```

Produces CSV output:

```csv
column1,column2,column3
value1,value2,value3
value4,value5,value6
```

#### CSV Options

- `--no-header`: Omit the header row
- `--field-separator`: Use a custom field separator (default: comma)

```bash
# Without headers
cqlai -e "SELECT * FROM table;" --format csv --no-header

# With semicolon separator
cqlai -e "SELECT * FROM table;" --format csv --field-separator ";"

# Tab-separated values
cqlai -e "SELECT * FROM table;" --format csv --field-separator $'\t'
```

## Automatic Pagination

In batch mode, CQLAI automatically fetches all pages of results:

```bash
# This will fetch ALL rows automatically
cqlai -e "SELECT * FROM large_table;" > all_data.csv
```

The pagination happens transparently:
- Results are streamed to output as they're fetched
- Memory usage is optimized by processing in batches
- You can interrupt with Ctrl+C at any time

## Connection Options

All standard connection options work in batch mode:

```bash
cqlai --host cassandra.example.com \
      --port 9042 \
      --username myuser \
      --password mypass \
      --keyspace mykeyspace \
      -e "SELECT * FROM mytable;"

# With SSL and consistency level
cqlai --host cassandra.example.com \
      --ssl \
      --consistency QUORUM \
      -e "SELECT * FROM mytable;"

# Using positional arguments (cqlsh compatible)
cqlai cassandra.example.com 9042 \
      -u myuser \
      -e "SELECT * FROM mytable;"
```

## Disabling Confirmation Prompts

By default, CQLAI requires confirmation for destructive commands (DROP, DELETE, TRUNCATE). For automated scripts and batch operations, you can disable these prompts:

### Using Command-Line Flag

```bash
cqlai --no-confirm -e "TRUNCATE my_table;"
cqlai --no-confirm -f cleanup_script.cql
```

### Using Environment Variable

```bash
export CQLAI_NO_CONFIRM=true
cqlai -e "DROP TABLE old_data;"
```

Or inline:

```bash
CQLAI_NO_CONFIRM=true cqlai -f maintenance.cql
```

### Using Configuration File

In your `cqlai.json`:

```json
{
  "requireConfirmation": false
}
```

**Warning**: Disabling confirmation prompts removes a safety check against accidental data loss. Use with caution, especially in production environments.

## Examples

### Export to CSV

```bash
# Export table to CSV
cqlai -e "SELECT * FROM products;" --format csv > products.csv

# Export without headers
cqlai -e "SELECT * FROM products;" --format csv --no-header > products_data.csv
```

### Export to JSON

```bash
# Export for processing with jq
cqlai -e "SELECT * FROM users;" --format json | jq '.[] | select(.age > 25)'
```

### Execute Multiple Statements

Create a script file `queries.cql`:

```sql
USE my_keyspace;
SELECT COUNT(*) FROM users;
SELECT * FROM products WHERE price < 100;
```

Execute it:

```bash
cqlai -f queries.cql
```

### Pipeline Processing

```bash
# Count rows
echo "SELECT * FROM large_table;" | cqlai --format csv | wc -l

# Extract specific column
echo "SELECT email FROM users;" | cqlai --format csv --no-header | sort | uniq
```

### Automated Reports

```bash
#!/bin/bash
# Daily report script

DATE=$(date +%Y-%m-%d)

# Generate report
cqlai -e "
  SELECT date, metric, value 
  FROM metrics 
  WHERE date = '$DATE'
" --format csv > "report_$DATE.csv"
```

## Differences from Interactive Mode

In batch mode:
- No interactive UI
- No command history
- No autocompletion
- Automatic pagination (no manual page control)
- Results stream directly to stdout
- Errors go to stderr
- Exit code indicates success (0) or failure (non-zero)

## Compatibility with cqlsh

CQLAI's batch mode is designed to be compatible with `cqlsh`:

```bash
# These commands work similarly in both tools
cqlsh -e "SELECT * FROM system.local"
cqlai -e "SELECT * FROM system.local;"

# Pipe compatibility
echo "SELECT * FROM table;" | cqlsh
echo "SELECT * FROM table;" | cqlai
```

Main differences:
- CQLAI supports additional output formats (JSON)
- CQLAI requires semicolons for CQL statements (like in interactive cqlsh)
- CQLAI uses Go's gocql driver instead of Python's cassandra-driver