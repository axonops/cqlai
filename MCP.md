# CQLAI MCP Server Documentation

## Overview

CQLAI includes a Model Context Protocol (MCP) server that enables AI assistants like Claude Code and Claude Desktop to interact with Apache Cassandra databases. The MCP server provides a comprehensive API for querying, managing schemas, and administering Cassandra clusters with built-in permission controls and confirmation workflows for dangerous operations.

**Current Status**: The MCP server uses HTTP transport for communication with KSUID-based API key authentication and defense-in-depth security.

---

## Features

### Core Capabilities
- **Complete CQL Coverage**: 37 supported operations including SELECT, INSERT, UPDATE, DELETE, ALTER, CREATE INDEX, CREATE ROLE, and more
- **Permission System**: 6-category classification (DQL, DML, DDL, DCL, SESSION, FILE) for all 76 Cassandra operations
- **Confirmation Workflows**: Dangerous queries require user approval with configurable policies
- **Query History**: Complete audit trail with lifecycle event logging
- **Trace Analysis**: Retrieve Cassandra trace data for performance analysis
- **Runtime Configuration**: Dynamic permission updates via MCP tools
- **Shell Commands**: Support for SHOW, CONSISTENCY, PAGING, TRACING, COPY, SOURCE

### Security Features
- **Multiple Permission Modes**: readonly, readwrite, dba
- **Fine-Grained Control**: Skip confirmation on specific operation categories
- **Confirmation Lifecycle**: PENDING → CONFIRMED/DENIED/CANCELLED/TIMEOUT states
- **Request Approval Gate**: Optional allow_mcp_request_approval flag (disabled by default)
- **User Confirmation Requirement**: Tools require user_confirmed=true boolean
- **Runtime Lockdown**: Disable runtime permission changes via config

---

## Quick Start

### 1. Start MCP Server from CQLAI

```bash
# Connect to Cassandra
./cqlai -h 127.0.0.1 -p 9042 -u cassandra

# Start MCP server in DBA mode
.mcp start --dba_mode

# Or start with custom config
.mcp start --config-file ~/.cqlai/mcp_config.json
```

### 2. Configure Claude Code/Desktop

**HTTP Transport Configuration:**

Create or edit `.mcp.json` in your project or global config:

```json
{
  "mcpServers": {
    "cqlai": {
      "url": "http://127.0.0.1:8888/mcp",
      "headers": {
        "X-API-Key": "${MCP_API_KEY}"
      }
    }
  }
}
```

**Set your API key** (see "Generate API Key" section below):
```bash
export MCP_API_KEY="2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
```

**Or use OS keychain** (recommended):
```bash
# macOS
export MCP_API_KEY=$(security find-generic-password -w -s "cqlai_mcp_api_key" -a "$(whoami)" 2>/dev/null)
```

### 3. Check Status

```bash
# In CQLAI
.mcp status

# Via MCP
Use the get_mcp_status tool
```

---

## Configuration

### Configuration Methods

1. **CLI Flags** (`.mcp start` command)
2. **JSON Config File** (`--config-file path/to/config.json`)
3. **Runtime Updates** (via `update_mcp_permissions` tool, if enabled)

### Configuration File Format

```json
{
  "http_host": "127.0.0.1",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30,
  "allowed_origins": ["http://localhost"],
  "ip_allowlist": ["127.0.0.1"],
  "audit_http_headers": ["X-Forwarded-For", "User-Agent"],

  "log_level": "info",
  "log_file": "~/.cqlai/cqlai_mcp.log",

  "history_file": "~/.cqlai/cqlai_mcp_history",
  "history_max_size_mb": 10,
  "history_max_rotations": 5,
  "history_rotation_interval_seconds": 60,

  "mode": "readwrite",
  "confirm_queries": ["dml"],
  "confirmation_timeout_seconds": 300,

  "disable_runtime_permission_changes": false,
  "allow_mcp_request_approval": false
}
```

### Environment Variable Expansion

**ALL configuration fields support environment variables:**

```json
{
  "http_host": "${MCP_HOST:-127.0.0.1}",
  "api_key": "${MCP_API_KEY}",
  "ip_allowlist": ["${OFFICE_SUBNET}", "${VPN_GATEWAY}"],
  "allowed_origins": ["${ALLOWED_ORIGIN}"],
  "required_headers": {
    "${PROXY_HEADER}": "${PROXY_VALUE}"
  },
  "log_level": "${LOG_LEVEL:-info}"
}
```

**Syntax:**
- `${VAR}` - Required: fails if VAR not set
- `${VAR:-default}` - Optional: uses default if VAR not set
- Explicit values still work: `"api_key": "2ABCDEFGHIJKLMNOPQRSTUVWXYZa"`

**Environment variables are OPTIONAL** - use them for secrets/environment-specific values, use explicit values for static config.

**Benefits:**
- ✅ Keep secrets out of config files
- ✅ Environment-specific configuration (dev/staging/prod)
- ✅ CI/CD integration (inject secrets at runtime)
- ✅ Team collaboration (share config, not secrets)

**Example:**
```bash
# Set environment variables
export MCP_API_KEY="2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
export OFFICE_SUBNET="10.0.1.0/24"
export MCP_HOST="192.168.1.100"

# Start CQLAI - config file uses ${VAR} syntax
cqlai --mcpconfig ~/.cqlai/.mcp.json
```

**CLI flags also support env vars** (inside CQLAI console):
```bash
# Use single quotes to prevent shell expansion!
.mcp start --api-key='${MCP_API_KEY}' --ip-allowlist='${OFFICE_SUBNET}'
```

### Permission Modes

#### Preset Modes

**Readonly Mode** (`--readonly_mode`):
- **Allowed**: SELECT, DESCRIBE, LIST, SHOW (DQL operations)
- **Blocked**: All modifications (INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, etc.)
- **Use Case**: Safe read-only access for queries and analysis

**Readwrite Mode** (`--readwrite_mode`):
- **Allowed**: DQL + DML (INSERT, UPDATE, DELETE) + FILE operations
- **Blocked**: DDL (schema changes) and DCL (permission changes)
- **Use Case**: Data manipulation without schema/permission changes

**DBA Mode** (`--dba_mode`):
- **Allowed**: All operations (DQL, DML, DDL, DCL, FILE)
- **Blocked**: Nothing (full access)
- **Use Case**: Complete database administration

#### Confirmation Overlays

Add confirmation requirements to allowed operations:

```bash
# Readwrite mode but require confirmation for DML
.mcp start --readwrite_mode --confirm-queries dml

# DBA mode but require confirmation for DDL and DCL
.mcp start --dba_mode --confirm-queries ddl,dcl
```

#### Fine-Grained Mode

Skip confirmation on specific categories:

```bash
# Skip confirmation on DQL and DML (require for DDL, DCL)
.mcp start --skip-confirmation dql,dml

# Skip ALL (allow everything without confirmation)
.mcp start --skip-confirmation ALL
```

### Operation Categories

- **DQL** (14 ops): SELECT, DESCRIBE, LIST ROLES/USERS/PERMISSIONS, SHOW
- **SESSION** (8 ops): CONSISTENCY, PAGING, TRACING, EXPAND, OUTPUT, CAPTURE, SAVE, AUTOFETCH
- **DML** (8 ops): INSERT, UPDATE, DELETE, BATCH variants
- **DDL** (28 ops): CREATE/ALTER/DROP for TABLE, INDEX, TYPE, FUNCTION, AGGREGATE, TRIGGER, etc.
- **DCL** (12 ops): CREATE/ALTER/DROP ROLE/USER, GRANT, REVOKE, ADD/DROP IDENTITY
- **FILE** (3 ops): COPY TO/FROM, SOURCE

---

## MCP Tools (20 Total)

### Query & Schema Tools (3)

#### 1. `list_keyspaces`
Lists all available keyspaces in the Cassandra cluster.

**Parameters**: None

**Returns**: Array of keyspace names

**Example**:
```json
["system", "system_schema", "test_mcp", "production"]
```

#### 2. `get_schema`
Gets complete schema for a specific table including columns and types.

**Parameters**:
- `keyspace` (string, required): Keyspace name
- `table` (string, required): Table name

**Returns**: Table schema with column definitions

**Example**:
```json
{
  "keyspace": "test_mcp",
  "table": "users",
  "columns": [
    {"name": "id", "type": "uuid", "kind": "partition_key"},
    {"name": "email", "type": "text", "kind": "regular"},
    {"name": "name", "type": "text", "kind": "regular"}
  ]
}
```

#### 3. `submit_query_plan`
Submits a structured query plan for execution. Supports all 37 CQL operations.

**Parameters**:
- `operation` (string, required): Operation type (SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, TRUNCATE, USE, GRANT, REVOKE, LIST, SHOW, DESCRIBE)
- `keyspace` (string, optional): Keyspace name
- `table` (string, optional): Table name
- `columns` (array, optional): Column names for SELECT/INSERT
- `values` (object, optional): Values for INSERT/UPDATE (required for these operations)
- `where` (array, optional): WHERE conditions (required for UPDATE/DELETE)
- `schema` (object, optional): Column definitions for CREATE TABLE
- `options` (object, optional): Operation-specific parameters

**Examples**:

SELECT:
```json
{
  "operation": "SELECT",
  "keyspace": "test_mcp",
  "table": "users"
}
```

INSERT:
```json
{
  "operation": "INSERT",
  "keyspace": "test_mcp",
  "table": "users",
  "values": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

UPDATE:
```json
{
  "operation": "UPDATE",
  "keyspace": "test_mcp",
  "table": "users",
  "values": {"name": "Alice Smith"},
  "where": [
    {"column": "id", "operator": "=", "value": "550e8400-e29b-41d4-a716-446655440000"}
  ]
}
```

DELETE:
```json
{
  "operation": "DELETE",
  "keyspace": "test_mcp",
  "table": "users",
  "where": [
    {"column": "id", "operator": "=", "value": "550e8400-e29b-41d4-a716-446655440000"}
  ]
}
```

CREATE TABLE:
```json
{
  "operation": "CREATE",
  "keyspace": "test_mcp",
  "table": "logs",
  "schema": {
    "id": "uuid PRIMARY KEY",
    "timestamp": "timestamp",
    "message": "text"
  },
  "options": {
    "if_not_exists": true
  }
}
```

ALTER TABLE:
```json
{
  "operation": "ALTER",
  "keyspace": "test_mcp",
  "table": "users",
  "options": {
    "object_type": "TABLE",
    "action": "ADD",
    "column_name": "age",
    "column_type": "int"
  }
}
```

CREATE INDEX:
```json
{
  "operation": "CREATE",
  "keyspace": "test_mcp",
  "table": "users",
  "options": {
    "object_type": "INDEX",
    "index_name": "users_email_idx",
    "column": "email",
    "if_not_exists": true
  }
}
```

GRANT Permission:
```json
{
  "operation": "GRANT",
  "keyspace": "test_mcp",
  "options": {
    "permission": "SELECT",
    "role": "app_readonly",
    "resource_scope": "KEYSPACE"
  }
}
```

GRANT on TABLE (fine-grained):
```json
{
  "operation": "GRANT",
  "keyspace": "production",
  "table": "users",
  "options": {
    "permission": "SELECT",
    "role": "analyst_role",
    "resource_scope": "TABLE"
  }
}
```

LIST ROLES:
```json
{
  "operation": "LIST",
  "options": {
    "object_type": "ROLES"
  }
}
```

**Permissions**: SELECT, MODIFY, ALTER, DROP, CREATE, AUTHORIZE, DESCRIBE, EXECUTE, UNMASK, SELECT_MASKED, ALL

**Resource Scopes**: ALL KEYSPACES, KEYSPACE, TABLE, ALL ROLES, ROLE, ALL FUNCTIONS, FUNCTION, ALL MBEANS, MBEAN

**Returns**:
- Success: `{"status": "executed", "query": "...", "execution_time_ms": 45, "trace_id": "abc123..."}`
- Confirmation needed: Error with request_id
- Permission denied: Error with configuration hints

---

### Status & Configuration Tools (2)

#### 4. `get_mcp_status`
Returns current MCP server status, configuration, connection info, and metrics.

**Parameters**: None

**Returns**:
```json
{
  "state": "RUNNING",
  "config": {
    "mode": "preset",
    "preset_mode": "readwrite",
    "confirm_queries": ["dml"],
    "allow_mcp_request_approval": false,
    "history_file": "~/.cqlai/cqlai_mcp_history",
    "history_max_size_mb": 10,
    "history_max_rotations": 5,
    "history_rotation_interval_seconds": 60
  },
  "connection": {
    "contact_point": "127.0.0.1:9042",
    "username": "cassandra",
    "cluster_name": "Test Cluster"
  },
  "metrics": {
    "total_requests": 150,
    "success_rate": 98.7
  }
}
```

#### 5. `update_mcp_permissions`
Updates MCP server permissions at runtime (if not locked down).

**Parameters**:
- `action` (string, required): "update_preset_mode", "update_confirm_queries", or "update_skip_confirmation"
- `preset_mode` (string, optional): "readonly", "readwrite", or "dba"
- `confirm_queries` (array, optional): Categories requiring confirmation
- `skip_confirmation` (array, optional): Categories to skip
- `user_confirmed` (boolean, required): Must be true

**Security**: Blocked if `disable_runtime_permission_changes: true`

**Example**:
```json
{
  "action": "update_preset_mode",
  "preset_mode": "readwrite",
  "user_confirmed": true
}
```

---

### Confirmation Lifecycle Tools (8)

#### 6. `get_pending_confirmations`
Lists confirmation requests awaiting approval.

**Returns**: Array of pending requests with ID, query, severity, timestamp

#### 7. `get_approved_confirmations`
Lists approved confirmation requests.

**Returns**: Array of approved requests with execution metadata

#### 8. `get_denied_confirmations`
Lists denied confirmation requests.

**Returns**: Array of denied requests with denial reason

#### 9. `get_cancelled_confirmations`
Lists cancelled confirmation requests.

**Returns**: Array of cancelled requests

#### 10. `get_confirmation_state`
Gets detailed state of a specific confirmation request.

**Parameters**:
- `request_id` (string, required): Request ID (e.g., "req_001")

**Returns**:
```json
{
  "id": "req_001",
  "status": "PENDING",
  "query": "DELETE FROM users WHERE id = ...",
  "operation": "DELETE",
  "dangerous": true,
  "severity": "HIGH",
  "timeout": "2025-12-30T20:25:00Z",
  "user_confirmed": false
}
```

#### 11. `confirm_request` ⚠️ Security-Gated
Approves a dangerous query for execution.

**Security Requirements**:
1. `allow_mcp_request_approval: true` in config (disabled by default)
2. `user_confirmed: true` parameter (must ask user)

**Parameters**:
- `request_id` (string, required): Request ID to approve
- `user_confirmed` (boolean, required): Must be true

**Returns**: Execution results including trace_id, duration, rows_affected

**Example**:
```json
{
  "request_id": "req_001",
  "user_confirmed": true
}
```

**Error if disabled**:
```
confirm_request tool is disabled. Set allow_mcp_request_approval=true in MCP config to enable (security setting).
```

#### 12. `deny_request`
Denies a dangerous query.

**Parameters**:
- `request_id` (string, required): Request ID to deny
- `reason` (string, optional): Reason for denial
- `user_confirmed` (boolean, required): Must be true

**Example**:
```json
{
  "request_id": "req_001",
  "reason": "User declined",
  "user_confirmed": true
}
```

#### 13. `cancel_confirmation`
Cancels a confirmation request in any state.

**Parameters**:
- `request_id` (string, required): Request ID to cancel
- `reason` (string, optional): Reason for cancellation

**Example**:
```json
{
  "request_id": "req_001",
  "reason": "Mistake in query"
}
```

---

### Trace Analysis Tools (1)

#### 14. `get_trace_data`
Retrieves Cassandra trace data for performance analysis.

**Parameters**:
- `trace_id` (string, required): Hex trace ID from query execution

**Returns**:
```json
{
  "trace_id": "550e8400-e29b-41d4-a716-446655440000",
  "coordinator": "127.0.0.1",
  "duration_us": 15000,
  "events": [
    {
      "event_id": "abc12345",
      "activity": "Parsing statement",
      "source": "127.0.0.1",
      "source_elapsed": 150,
      "thread": "Native-Transport-Requests-1"
    },
    ...
  ]
}
```

**Usage**:
1. Execute query with `submit_query_plan`
2. Get `trace_id` from response
3. Call `get_trace_data` with the trace_id
4. Analyze coordinator, duration, and event timeline

---

## CLI Commands (In CQLAI)

### `.mcp start [options]`

Starts the MCP server.

**Options**:

**Permission Modes** (mutually exclusive):
- `--readonly_mode` - Read-only access
- `--readwrite_mode` - Read-write access (no schema changes)
- `--dba_mode` - Full database administration

**Confirmation Overlay**:
- `--confirm-queries <categories>` - Require confirmation for specific categories (comma-separated: dql,dml,ddl,dcl)

**Fine-Grained Mode**:
- `--skip-confirmation <categories>` - Skip confirmation on specific categories

**Configuration**:
- `--config-file <path>` - Load configuration from JSON file
- `--http-host <host>` - HTTP server host (default: 127.0.0.1)
- `--http-port <port>` - HTTP server port (default: 8888)
- `--api-key <ksuid>` - KSUID API key (default: auto-generated)
- `--log-level <level>` - Logging level (debug, info, warn, error)
- `--log-file <path>` - Log file path (default: ~/.cqlai/cqlai_mcp.log)

**Security**:
- `--disable-runtime-permission-changes` - Lock down runtime permission updates
- `--allow-runtime-permission-changes` - Allow runtime updates (default)
- `--allow-mcp-request-approval` - Allow MCP tools to approve requests (default: false, security)

**Examples**:
```bash
# DBA mode with DDL confirmations required
.mcp start --dba_mode --confirm-queries ddl

# Readwrite mode, locked down (no runtime changes)
.mcp start --readwrite_mode --disable-runtime-permission-changes

# DBA mode allowing MCP approval (use with caution!)
.mcp start --dba_mode --allow-mcp-request-approval

# Load from config file
.mcp start --config-file ~/.cqlai/production_mcp.json
```

### `.mcp stop`

Stops the MCP server gracefully.

### `.mcp status`

Shows MCP server status including:
- Current state (RUNNING/STOPPED)
- Permission configuration
- Cassandra connection details
- Server configuration (HTTP endpoint, logs, history)
- Metrics (total requests, success rate, tool usage)

### `.mcp pending`

Lists pending confirmation requests.

### `.mcp confirm <request_id>`

Approves a pending confirmation request (REPL command, not MCP tool).

**Example**:
```bash
.mcp confirm req_001
```

### `.mcp deny <request_id> [reason]`

Denies a pending confirmation request (REPL command, not MCP tool).

**Example**:
```bash
.mcp deny req_001 Too dangerous
```

### `.mcp permissions-config [action]`

Shows or updates permission configuration (if not locked down).

**Examples**:
```bash
# Show current config
.mcp permissions-config

# Update mode
.mcp permissions-config update-mode readwrite

# Add confirmation requirements
.mcp permissions-config add-confirm-queries ddl
```

---

## History File

### Purpose

The MCP history file provides a complete audit trail of:
- All executed queries
- Confirmation lifecycle events (requested, approved, denied, cancelled, timeout)

### Location

- **Default**: `~/.cqlai/cqlai_mcp_history`
- **Configurable**: Via `history_file` in JSON config or parsed from `.mcp start`

### Format

```
[2025-12-30 20:15:30] QUERY: SELECT * FROM test_mcp.users
[2025-12-30 20:16:00] CONFIRM_REQUESTED: request_id=req_abc query="DELETE FROM users WHERE id = ..." details=operation=DELETE category=DML dangerous=true
[2025-12-30 20:16:30] CONFIRM_APPROVED: request_id=req_abc query="DELETE FROM users WHERE id = ..." details=confirmed_by=mcp
[2025-12-30 20:16:31] QUERY: DELETE FROM users WHERE id = ...
[2025-12-30 20:18:00] CONFIRM_REQUESTED: request_id=req_xyz query="TRUNCATE users" details=operation=TRUNCATE category=DDL dangerous=true
[2025-12-30 20:18:45] CONFIRM_DENIED: request_id=req_xyz query="TRUNCATE users" details=denied_by=cassandra reason="Too risky"
```

### Lifecycle Events

1. **CONFIRM_REQUESTED**: Query needs approval
   - Includes: request_id, query, operation, category, dangerous flag
2. **CONFIRM_APPROVED**: User/MCP approved request
   - Includes: request_id, query, confirmed_by (\"mcp\" or username)
3. **CONFIRM_DENIED**: User/MCP denied request
   - Includes: request_id, query, denied_by, reason
4. **CONFIRM_CANCELLED**: Request cancelled
   - Includes: request_id, query, cancelled_by, reason
5. **CONFIRM_TIMEOUT**: Request expired without decision
   - Includes: request_id, query, timeout_after duration
6. **QUERY**: Query executed successfully
   - Includes: complete CQL query text

### File Rotation

**Automatic rotation** via background worker:
- **Trigger**: File exceeds `history_max_size_mb` (default: 10MB)
- **Check Interval**: Every `history_rotation_interval_seconds` (default: 60s)
- **Numbered Backups**: `cqlai_mcp_history.1.gz`, `.2.gz`, ..., `.N.gz`
- **Compression**: gzip compression of rotated files
- **Retention**: Keep `history_max_rotations` files (default: 5)
- **Thread-Safe**: Uses RWMutex (writers blocked briefly during rotation)

**Example sequence**:
```
cqlai_mcp_history           (current, 9MB)
cqlai_mcp_history.1.gz      (previous rotation, compressed)
cqlai_mcp_history.2.gz      (older)
cqlai_mcp_history.3.gz
cqlai_mcp_history.4.gz
cqlai_mcp_history.5.gz      (oldest, will be deleted on next rotation)
```

### Configuration

```json
{
  "history_file": "~/.cqlai/cqlai_mcp_history",
  "history_max_size_mb": 10,
  "history_max_rotations": 5,
  "history_rotation_interval_seconds": 60
}
```

**To disable rotation**: Set `history_max_rotations: 0` (file will be deleted when size exceeded)

---

## Security Model

### Defense in Depth

CQLAI MCP implements multiple security layers:

1. **Permission Modes**: Restrict operation categories (readonly/readwrite/dba)
2. **Confirmation Requirements**: Dangerous queries require approval
3. **User Confirmation Flag**: MCP tools require user_confirmed=true
4. **Request Approval Gate**: allow_mcp_request_approval must be enabled
5. **Runtime Lockdown**: Disable runtime permission changes
6. **Audit Trail**: Complete history of all operations

### Typical Security Configurations

**Production (Locked Down)**:
```json
{
  "mode": "readonly",
  "disable_runtime_permission_changes": true,
  "allow_mcp_request_approval": false
}
```
- Read-only access only
- No runtime changes
- MCP cannot approve any operations

**Development (Moderate)**:
```json
{
  "mode": "readwrite",
  "confirm_queries": ["dml"],
  "allow_mcp_request_approval": false
}
```
- Data modifications allowed
- DML requires confirmation
- Approval via REPL only (not MCP)

**Trusted Automation (Permissive)**:
```json
{
  "mode": "dba",
  "confirm_queries": ["ddl", "dcl"],
  "allow_mcp_request_approval": true
}
```
- Full access
- Schema/permission changes require confirmation
- MCP can approve (if user confirms)

### Best Practices

1. **Start Restrictive**: Use readonly mode initially
2. **Never Auto-Approve**: Keep allow_mcp_request_approval=false unless absolutely needed
3. **Review History**: Regularly audit `~/.cqlai/cqlai_mcp_history`
4. **Lock Production**: Use disable_runtime_permission_changes=true in production
5. **Monitor Metrics**: Check success rates and tool usage patterns
6. **Rotate Credentials**: Change Cassandra passwords regularly
7. **Confirm Categories**: Require confirmation for DDL/DCL operations

---

## Confirmation Workflow

### Overview

CQLAI uses **HTTP streaming** for confirmation workflows. When a dangerous query requires user approval, the HTTP connection **stays open** and streams multiple messages until the user responds.

### Three Response Patterns

#### 1. Allowed, No Confirmation (SELECT, DESCRIBE, LIST)

**Behavior:** Single response, no streaming
```
Request: submit_query_plan (SELECT * FROM users)
Response (immediate):
{
  "id": 123,
  "result": {
    "status": "executed",
    "query": "SELECT * FROM users;",
    "rows_returned": 42
  }
}
Connection closes (<1 second)
```

#### 2. NOT ALLOWED (Policy Block)

**Behavior:** Single error response, no streaming
```
Request: submit_query_plan (INSERT in readonly mode)
Response (immediate):
{
  "id": 123,
  "error": {
    "message": "Operation not allowed in readonly mode\n
                 Suggestion: You might be able to use update_mcp_permissions
                 tool with mode='readwrite' (confirm with the user first),
                 or restart the MCP server with different permissions"
  }
}
Connection closes (<1 second)
```

**Note:** Message adapts based on `disable_runtime_permission_changes`:
- **If runtime changes allowed:** Suggests `update_mcp_permissions` tool (with user confirmation) OR restart
- **If runtime changes disabled:** Only suggests restart

#### 3. NEEDS CONFIRMATION (Allowed But Requires Approval)

**Behavior:** HTTP streaming, connection stays open, multiple messages

**Example:** INSERT in readwrite mode with `--confirm-queries dml`

```
Request: submit_query_plan (INSERT INTO users...)
Connection: STAYS OPEN (streaming)

Message 1 (Immediate - before blocking):
{
  "method": "confirmation/requested",
  "params": {
    "request_id": "req_001",
    "query": "INSERT INTO users (id, name) VALUES (...);",
    "operation_info": {
      "type": "DML",
      "operation": "INSERT",
      "description": "Insert data",
      "risk_level": "MEDIUM"
    },
    "timeout_seconds": 300,
    "timeout_message": "Request will timeout in 5 minutes",
    "approval_workflow": {
      "step_1": "Ask the user: 'This is a MEDIUM risk DML operation (Insert data). Do you want to execute this query: INSERT INTO users...?'",
      "step_2": "If user says YES: Call confirm_request tool with user_confirmed=true",
      "step_3": "If user says NO: Call deny_request tool with reason",
      "important": "NEVER approve dangerous operations without explicit user consent"
    },
    "tools": {
      "approve": {
        "tool": "confirm_request",
        "request_id": "req_001",
        "user_confirmed": true,
        "must_ask_user": "You MUST ask the user first..."
      },
      "deny": {
        "tool": "deny_request",
        "request_id": "req_001"
      }
    }
  }
}

[Optional: Heartbeats every 30 seconds if wait is long]
Heartbeat message:
{
  "method": "confirmation/waiting",
  "params": {
    "request_id": "req_001",
    "status": "STILL_PENDING",
    "elapsed_seconds": 30,
    "remaining_seconds": 270,
    "heartbeat": true,
    "message": "Still waiting for user response (270 seconds remaining)"
  }
}

[User approves via separate request: confirm_request tool]

Message 2 (After approval):
{
  "method": "confirmation/statusChanged",
  "params": {
    "request_id": "req_001",
    "status": "CONFIRMED",
    "actor": "mcp",
    "timestamp": "2026-01-01T12:00:00Z"
  }
}

Message 3 (Final result):
{
  "id": 123,
  "result": {
    "status": "executed",
    "query": "INSERT INTO users...",
    "execution_time_ms": 2
  }
}

Connection closes
```

**Key Points:**
- ✅ **ONE HTTP connection** for entire workflow
- ✅ **Connection stays open** (blocks) until user responds or timeout
- ✅ **3 messages minimum** (requested → statusChanged → result)
- ✅ **Heartbeats** every 30 seconds keep proxy connections alive
- ✅ **Parallel safe** - each waiting query = lightweight goroutine (2-8 KB)

### Confirmation Outcomes

#### Outcome 1: User Confirms (Success)

```
Message 1: confirmation/requested (initial)
Message 2: confirmation/statusChanged (status=CONFIRMED)
Message 3: Result (query executed successfully)
```

#### Outcome 2: User Denies

```
Message 1: confirmation/requested (initial)
[User denies via deny_request tool]
Message 2: confirmation/statusChanged (status=DENIED)
Message 3: Error {"message": "Query denied by mcp: user rejected operation"}
```

#### Outcome 3: User Cancels

```
Message 1: confirmation/requested (initial)
[User cancels via cancel_confirmation tool]
Message 2: confirmation/statusChanged (status=CANCELLED)
Message 3: Error {"message": "Query cancelled by mcp"}
```

#### Outcome 4: Timeout (No Response)

```
Message 1: confirmation/requested (initial)
[Wait 30s] Heartbeat: "270 seconds remaining"
[Wait 60s] Heartbeat: "240 seconds remaining"
...
[Wait 300s] Timeout!
Message 2: confirmation/statusChanged (status=TIMEOUT)
Message 3: Error {"message": "Confirmation timed out after 5m - query not executed"}
```

### How Streaming Works (Technical)

**Concurrent Execution:**
```
Thread 1 (Claude's request):
  1. Calls submit_query_plan (INSERT)
  2. Receives initial "confirmation/requested" notification
  3. HTTP connection BLOCKS (waits)
  4. (User is prompted)
  5. Receives "confirmation/statusChanged" notification
  6. Receives final result
  7. Connection closes

Thread 2 (Approval request):
  1. User says "Yes"
  2. Claude calls confirm_request with user_confirmed=true
  3. Signals Thread 1 to continue
  4. Returns success
```

**Benefits:**
- ✅ **No polling** - Claude gets instant notification when approved
- ✅ **Better UX** - Single continuous conversation
- ✅ **Scalable** - 1,000 concurrent waits = ~25 MB RAM, 0% CPU
- ✅ **Proxy-safe** - Heartbeats keep connections alive

### Example: User Approves DELETE

```
User: "Delete all test data"
  ↓
Claude: Calls submit_query_plan (DELETE FROM users WHERE name LIKE 'test%')
  ↓
Connection stays OPEN, receives Message 1:
  - Request ID: req_001
  - Query: "DELETE FROM users WHERE name LIKE 'test%'"
  - Risk: HIGH
  - Type: DML (Delete data)
  ↓
Claude: Asks user "This is a HIGH risk DML operation (Delete data). Execute: DELETE FROM users WHERE name LIKE 'test%'?"
  ↓
User: "Yes, delete test data"
  ↓
Claude: Calls confirm_request (separate request) with user_confirmed=true
  ↓
Original connection receives Message 2:
  - Status: CONFIRMED
  ↓
Original connection receives Message 3:
  - Query executed
  - Rows affected: 15
  ↓
Connection closes
```

**Total time:** ~1-2 seconds (instant after user approves)

---

## Troubleshooting

### Common Issues

**"confirm_request tool is disabled"**
- **Cause**: allow_mcp_request_approval=false (default)
- **Fix**: Add `--allow-mcp-request-approval` to `.mcp start` or set in JSON
- **Security Note**: Only enable if you trust the AI to request approval appropriately

**"Permission denied" errors**
- **Cause**: Operation not allowed in current mode
- **Fix**: Check error message for hints (e.g., "upgrade to readwrite mode")
- **Options**:
  - Use `update_mcp_permissions` to upgrade mode
  - Restart with higher permission mode
  - Use fine-grained skip_confirmation

**"WHERE clause required for DELETE/UPDATE"**
- **Cause**: Missing where parameter in submit_query_plan
- **Fix**: Add where array with column, operator, value

**"no values to insert"**
- **Cause**: INSERT without values parameter
- **Fix**: Add values object to submit_query_plan

**"Failed to open history file"**
- **Cause**: Permission denied on history directory
- **Fix**: Ensure ~/.cqlai directory exists and is writable
- **Check**: `ls -la ~/.cqlai/`

---

## Configuration Examples

### Development Setup

```json
{
  "http_host": "127.0.0.1",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30,
  "mode": "readwrite",
  "confirm_queries": ["dml"],
  "log_level": "debug",
  "history_file": "~/.cqlai/dev_mcp_history",
  "allow_mcp_request_approval": true
}
```

### Production Setup

```json
{
  "http_host": "0.0.0.0",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 7,
  "ip_allowlist": ["${OFFICE_SUBNET}"],
  "allowed_origins": ["${ALLOWED_ORIGIN}"],
  "mode": "readonly",
  "log_level": "info",
  "log_file": "/var/log/cqlai/mcp.log",
  "history_file": "/var/log/cqlai/mcp_history",
  "history_max_size_mb": 50,
  "history_max_rotations": 10,
  "disable_runtime_permission_changes": true,
  "allow_mcp_request_approval": false
}
```

### Testing Setup

```json
{
  "http_host": "127.0.0.1",
  "http_port": 8889,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 1,
  "mode": "dba",
  "log_level": "debug",
  "history_file": "/tmp/test_mcp_history",
  "history_max_rotations": 0,
  "confirmation_timeout_seconds": 30,
  "allow_mcp_request_approval": true
}
```

---

## Advanced Features

### Supported CQL Operations (37 Total)

**DQL (7)**:
- SELECT, DESCRIBE, LIST ROLES, LIST USERS, LIST PERMISSIONS, SHOW VERSION/HOST/SESSION

**DML (4)**:
- INSERT, UPDATE, DELETE, TRUNCATE

**DDL (16)**:
- CREATE/ALTER/DROP TABLE/KEYSPACE
- CREATE/DROP INDEX
- CREATE/ALTER/DROP TYPE (UDTs)
- CREATE/DROP FUNCTION (UDFs)
- CREATE/DROP AGGREGATE (UDAs)
- CREATE/DROP TRIGGER
- CREATE/DROP MATERIALIZED VIEW
- USE (switch keyspace)

**DCL (10)**:
- CREATE/ALTER/DROP ROLE
- CREATE/ALTER/DROP USER (legacy)
- GRANT/REVOKE permissions (11 types, 9 resource scopes)
- GRANT/REVOKE roles
- ADD/DROP IDENTITY

**SESSION (8)**:
- SHOW, CONSISTENCY, PAGING, TRACING, EXPAND, OUTPUT, CAPTURE, SAVE, AUTOFETCH

**FILE (3)**:
- COPY TO/FROM, SOURCE

**Not Yet Supported**:
- BATCH operations (planned for future release)

### Permission Types

All 11 Cassandra permissions validated:
- CREATE, ALTER, DROP, SELECT, MODIFY, AUTHORIZE, DESCRIBE, EXECUTE, UNMASK, SELECT_MASKED, ALL

Invalid permissions (e.g., READ, WRITE, ADMIN) rejected with error.

### Resource Scopes

All 9 Cassandra resource types supported:
- ALL KEYSPACES
- KEYSPACE <name>
- TABLE <keyspace>.<table>
- ALL ROLES
- ROLE <name>
- ALL FUNCTIONS [IN KEYSPACE <name>]
- FUNCTION <keyspace>.<name>
- ALL MBEANS
- MBEAN <name>

---

## FAQ

**Q: Can AI approve dangerous queries without asking me?**
A: No. Two safeguards: (1) allow_mcp_request_approval defaults to false, (2) tools require user_confirmed=true

**Q: What happens if a confirmation request times out?**
A: Status changes to TIMEOUT, logged to history, request cannot be executed

**Q: Can I change permissions while MCP is running?**
A: Yes, via update_mcp_permissions tool, unless disable_runtime_permission_changes=true

**Q: How do I see what queries were executed?**
A: Check ~/.cqlai/cqlai_mcp_history (complete audit trail)

**Q: Can I use MCP with multiple Cassandra clusters?**
A: Start separate CQLAI instances with different HTTP ports, one per cluster (e.g., :8888, :8889)

**Q: What's the difference between confirm_request (MCP) and .mcp confirm (REPL)?**
A: Both approve requests. MCP tool requires allow_mcp_request_approval=true. REPL command always works.

**Q: Why is allow_mcp_request_approval disabled by default?**
A: Security: Prevents AI from auto-approving dangerous operations. Requires explicit opt-in.

**Q: How often does history rotate?**
A: Checked every history_rotation_interval_seconds (default: 60s), rotates when file exceeds history_max_size_mb

---

## Support

- **Issues**: https://github.com/axonops/cqlai/issues
- **Documentation**: This file (MCP.md)
- **Testing**: See claude-notes/MANUAL_TESTING_MATRIX.md for comprehensive test cases

---

**Version**: 1.0 (feature/mcp branch)
**Last Updated**: 2025-12-30
