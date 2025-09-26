# SAVE Command Documentation

## Overview

The `SAVE` command in CQLAI allows you to export the currently displayed query results to a file without re-executing the query. This is particularly useful when you have already run a query and want to export the results in different formats, or when working with large result sets that take time to execute.

## Key Features

- **No Re-execution**: Saves the exact results currently displayed in your terminal
- **Multiple Formats**: Supports CSV, JSON, and ASCII table formats
- **Interactive Mode**: Simple dialog for choosing format and filename
- **Auto-detection**: Automatically detects format from file extension
- **Smart JSON Handling**: Preserves JSON structure when OUTPUT JSON mode is used

## Syntax

```sql
SAVE [filename] [format]
```

### Parameters

- **filename** (optional): Path to the output file. If omitted, opens interactive dialog.
- **format** (optional): Output format (CSV, JSON, ASCII). If omitted, auto-detected from extension.

## Usage Examples

### Interactive Mode

```sql
-- Run a query first
SELECT * FROM users WHERE created_at > '2024-01-01';

-- Open interactive save dialog
SAVE
-- Dialog will prompt for:
-- 1. Format selection (CSV/JSON/ASCII)
-- 2. Filename input
```

### Direct Save with Auto-detection

```sql
-- Save to CSV (detected from .csv extension)
SAVE 'users.csv'

-- Save to JSON (detected from .json extension)
SAVE 'users.json'

-- Save to text file as ASCII table
SAVE 'users.txt' ASCII
```

### Explicit Format Specification

```sql
-- Force CSV format regardless of extension
SAVE 'output.data' CSV

-- Force JSON format
SAVE 'output.data' JSON

-- Force ASCII table format
SAVE 'output.data' ASCII
```

## Format Details

### CSV Format
- Includes headers by default
- Properly escapes special characters
- Removes ANSI color codes
- Strips key indicators (PK), (C) from column names

Example output:
```csv
id,name,email,created_at
1,John Doe,john@example.com,2024-01-15
2,Jane Smith,jane@example.com,2024-01-16
```

### JSON Format
- Creates an array of objects
- Each row becomes a JSON object with column names as keys
- Smart detection: If OUTPUT JSON was used, preserves original JSON structure
- Pretty-printed for readability

Example output:
```json
[
  {
    "id": "1",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-15"
  },
  {
    "id": "2",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "created_at": "2024-01-16"
  }
]
```

### ASCII Format
- Formatted table with borders
- Column alignment preserved
- Suitable for documentation or reports

Example output:
```
+----+------------+------------------+------------+
| id | name       | email            | created_at |
+----+------------+------------------+------------+
| 1  | John Doe   | john@example.com | 2024-01-15 |
| 2  | Jane Smith | jane@example.com | 2024-01-16 |
+----+------------+------------------+------------+
```

## Differences from CAPTURE

| Feature | SAVE | CAPTURE |
|---------|------|---------|
| **Purpose** | Export displayed results | Record future query outputs |
| **Timing** | After query execution | Before query execution |
| **Re-execution** | No | Yes (records all subsequent queries) |
| **Interactive** | Yes (dialog available) | No |
| **Use Case** | Export existing results | Logging session activity |

### When to Use SAVE

- After running a complex query you don't want to re-execute
- When you need results in multiple formats
- For exporting paginated results (saves only loaded pages)
- Quick export of query results for sharing

### When to Use CAPTURE

- Recording an entire session
- Automated scripting with predictable output
- Continuous logging of multiple queries
- When you need to capture errors and warnings

## Working with Paginated Results

When `AUTOFETCH OFF` is set and you have paginated results:

```sql
-- Set manual pagination
AUTOFETCH OFF;

-- Run query (loads first page)
SELECT * FROM large_table;

-- Load more pages as needed
-- (Press PgDn or Space)

-- Save only the loaded pages
SAVE 'partial_results.csv'
```

**Note**: SAVE only exports the data that has been loaded and displayed. If you need all pages, either:
1. Set `AUTOFETCH ON` before running the query, or
2. Manually page through all results before saving

## File Path Handling

- **Relative paths**: Saved relative to current working directory
- **Absolute paths**: Use full path starting with `/` (Unix) or drive letter (Windows)
- **Home directory**: Use `~/` for home directory expansion
- **Spaces in paths**: Enclose in single quotes

Examples:
```sql
SAVE 'output.csv'                    -- Current directory
SAVE '/tmp/results.json'             -- Absolute path
SAVE '~/Documents/data.csv'          -- Home directory
SAVE '/path with spaces/file.csv'    -- Path with spaces
```

## Error Handling

Common error scenarios:

```sql
-- No query results to save
SAVE 'output.csv'
-- Error: No data to export

-- Invalid path
SAVE '/invalid/path/file.csv'
-- Error: Failed to create file

-- Permission denied
SAVE '/root/protected.csv'
-- Error: Permission denied
```

## Tips and Best Practices

1. **Check results first**: Always verify the displayed data before saving
2. **Use meaningful filenames**: Include date or query context in filename
3. **Choose appropriate format**:
   - CSV for Excel/spreadsheet import
   - JSON for API integration or NoSQL databases
   - ASCII for documentation or plain text reports
4. **Handle large results**: For very large result sets, consider using `COPY TO` instead
5. **Preserve JSON structure**: When using `OUTPUT JSON`, save as JSON to maintain structure

## Integration with OUTPUT Modes

The SAVE command respects the current OUTPUT mode:

```sql
-- Default TABLE mode
SELECT * FROM users;
SAVE 'users.csv'  -- Saves structured data

-- JSON output mode
OUTPUT JSON;
SELECT * FROM users;
SAVE 'users.json'  -- Preserves JSON format without re-encoding

-- ASCII mode
OUTPUT ASCII;
SELECT * FROM users;
SAVE 'users.txt' ASCII  -- Saves formatted ASCII table
```

## Keyboard Shortcuts

While not directly related to SAVE, these shortcuts help navigate results before saving:

- **PgUp/PgDn**: Page through results
- **Space**: Load next page (when AUTOFETCH OFF)
- **Esc**: Cancel paging operation
- **Alt+←/→**: Scroll horizontally for wide tables

## See Also

- [CAPTURE Command](./CAPTURE_COMMAND.md) - For continuous output recording
- [COPY TO Command](./COPY_COMMAND.md) - For direct table export with full control
- [OUTPUT Command](./OUTPUT_COMMAND.md) - For changing display format