# MCP Server Security Guide

**CQLAI MCP Server - Comprehensive Security Documentation**

This document explains all security features implemented in the CQLAI MCP server, including what threats they protect against, how to configure them, and deployment examples.

---

## Table of Contents

1. [Security Architecture Overview](#security-architecture-overview)
2. [Layer 1: API Key Authentication (KSUID)](#layer-1-api-key-authentication-ksuid)
3. [Layer 2: Origin Header Validation](#layer-2-origin-header-validation)
4. [Layer 3: IP Allowlisting](#layer-3-ip-allowlisting)
5. [Layer 4: Required Header Validation](#layer-4-required-header-validation)
6. [Header Auditing](#header-auditing)
7. [Configuration Reference](#configuration-reference)
8. [Deployment Scenarios](#deployment-scenarios)
9. [Threat Models](#threat-models)
10. [Best Practices](#best-practices)
11. [Troubleshooting](#troubleshooting)

---

## Security Architecture Overview

The CQLAI MCP server implements **defense-in-depth** with four independent security layers. Every HTTP request must pass ALL enabled layers:

```
┌─────────────────────────────────────────────┐
│ Incoming HTTP Request                       │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│ Layer 1: API Key Authentication              │
│ - Validates X-API-Key header                 │
│ - KSUID format enforcement                   │
│ - Age-based expiration (default: 30 days)    │
│ - Constant-time comparison                   │
└──────────────┬───────────────────────────────┘
               │ PASS
               ▼
┌──────────────────────────────────────────────┐
│ Layer 2: Origin Header Validation            │
│ - Prevents DNS rebinding attacks             │
│ - Dynamic based on bind address              │
│ - Subdomain attack prevention                │
└──────────────┬───────────────────────────────┘
               │ PASS
               ▼
┌──────────────────────────────────────────────┐
│ Layer 3: IP Allowlisting                     │
│ - Default: 127.0.0.1 (localhost only)        │
│ - Supports CIDR notation (10.0.0.0/24)       │
│ - IPv4 and IPv6 support                      │
│ - Can be disabled (triggers warnings)        │
└──────────────┬───────────────────────────────┘
               │ PASS
               ▼
┌──────────────────────────────────────────────┐
│ Layer 4: Required Header Validation          │
│ - Verifies proxy added required headers      │
│ - Exact match or regex pattern support       │
│ - Detects bypassed proxy                     │
└──────────────┬───────────────────────────────┘
               │ PASS
               ▼
┌──────────────────────────────────────────────┐
│ Header Auditing (Logging)                    │
│ - Logs X-Forwarded-For, User-Agent           │
│ - Custom headers configurable                │
│ - Creates audit trail                        │
└──────────────┬───────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────┐
│ MCP Request Processing                       │
│ - Tool execution                             │
│ - Query execution with permissions           │
└──────────────────────────────────────────────┘
```

**If ANY layer fails → Request rejected with 403 Forbidden**

---

## Layer 1: API Key Authentication (KSUID)

### What It Is

API keys are **KSUID** (K-Sortable Unique IDentifiers) - cryptographically secure identifiers with embedded timestamps.

**Example KSUID:**
```
2ABCDEFGHIJKLMNOPQRSTUVWXYZa
```

### Why KSUID (Not UUID or Random Bytes)

| Feature | KSUID | TimeUUID (UUIDv1) | UUIDv4 | Random Bytes |
|---------|-------|-------------------|---------|--------------|
| Crypto-random | ✅ 128 bits | ❌ 14 bits | ✅ 122 bits | ✅ Configurable |
| Timestamp | ✅ Sortable | ✅ Sortable | ❌ No | ❌ No |
| Privacy | ✅ No MAC | ❌ MAC leak | ✅ No MAC | ✅ No MAC |
| Enumerable | ❌ No | ✅ Yes | ❌ No | ❌ No |
| URL-safe | ✅ Base62 | ❌ Hex+dashes | ❌ Hex+dashes | Depends |
| Expiration | ✅ Yes | ✅ Yes | ❌ No | ❌ No |

**Security improvements over TimeUUID:**
- **128 bits of randomness** (vs 14-bit clock sequence)
- **No MAC address leakage** (privacy protection)
- **Not predictable** (TimeUUID clock sequence can be guessed)
- **Still has timestamp** for expiration support

### Configuration

**JSON config (.mcp.json):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30
}
```

**CLI (.mcp start inside CQLAI):**
```bash
.mcp start --api-key=2ABCDEFGHIJKLMNOPQRSTUVWXYZa --api-key-max-age-days=30
```

**Auto-generation:**
```bash
# If no key provided, one is auto-generated and displayed
.mcp start

# Output:
=== MCP Server API Key ===
API Key: 2ABCDEFGHIJKLMNOPQRSTUVWXYZa
(Save this key - it won't be shown again)
==========================
```

### API Key Expiration

**Default:** Keys expire after 30 days

**Why:** Limits compromise window. Stolen keys become useless after expiration.

**Configuration:**
```json
{
  "api_key_max_age_days": 7    // Expire after 7 days
}
```

**Disable expiration (NOT RECOMMENDED):**
```json
{
  "api_key_max_age_days": 0    // Never expire
}
```

**Startup warning when disabled:**
```
⚠️  WARNING: API KEY AGE VALIDATION DISABLED ⚠️
API keys will NEVER expire. This is a security risk.
Recommendation: Set 'api_key_max_age_days' in config (default: 30 days)
```

### Timestamp Validation

**Future timestamp rejection:**
- Prevents expiration bypass attacks
- Attackers can't create keys dated far in the future
- Allows 1-minute clock skew tolerance (NTP drift)

**Example attack prevented:**
```
# Attacker creates KSUID dated 10 years in the future
# Even if stolen, key would never expire
# This is REJECTED at validation time
```

### Constant-Time Comparison

API keys are compared using `crypto/subtle.ConstantTimeCompare()`:

**Why:** Prevents timing attacks where attacker measures response time to guess key bytes.

**Without constant-time:**
```go
// VULNERABLE - fails fast on first wrong byte
if provided == expected { ... }
```

**With constant-time:**
```go
// SECURE - always compares all bytes regardless of mismatches
subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
```

### What This Layer Protects Against

| Threat | Protected | How |
|--------|-----------|-----|
| Unauthorized access | ✅ | Requires valid KSUID API key |
| Brute force attacks | ✅ | 128-bit randomness = 2^128 combinations |
| Timing attacks | ✅ | Constant-time comparison |
| Key enumeration | ✅ | Crypto-random, not predictable |
| Long-term key compromise | ✅ | Keys expire (default 30 days) |
| Expiration bypass | ✅ | Future timestamps rejected |
| MAC address tracking | ✅ | KSUID has no MAC (unlike TimeUUID) |

### Environment Variable Support

**ALL configuration fields** support environment variable expansion in **both** JSON config and CLI flags.

**Syntax:**
- `${VAR}` - Expand environment variable (required - fails if not set)
- `${VAR:-default}` - Use default value if VAR not set
- Explicit values: `"http_host": "192.168.1.100"` (no expansion)

**Environment variables are OPTIONAL** - mix and match with explicit values as needed.

**JSON config file (all fields support env vars):**
```json
{
  "http_host": "${MCP_HOST:-127.0.0.1}",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30,
  "allowed_origins": ["${ALLOWED_ORIGIN}"],
  "ip_allowlist": ["${OFFICE_SUBNET}", "${VPN_GATEWAY}"],
  "audit_http_headers": ["${CUSTOM_HEADER}"],
  "log_level": "${LOG_LEVEL:-info}",
  "log_file": "${MCP_LOG_FILE:-~/.cqlai/cqlai_mcp.log}",
  "required_headers": {
    "${PROXY_HEADER}": "${PROXY_VALUE}",
    "X-Request-ID": "${REQ_ID_PATTERN}"
  }
}
```

**CLI flags (inside CQLAI console - all support env vars):**
```bash
# Set environment variables
export MCP_HOST="192.168.1.100"
export MCP_API_KEY="2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
export OFFICE_SUBNET="10.0.1.0/24"
export PROXY_HEADER="X-Proxy-Verified"

# Use in CLI flags (single quotes to prevent shell expansion!)
.mcp start \
  --http-host='${MCP_HOST}' \
  --api-key='${MCP_API_KEY}' \
  --ip-allowlist='${OFFICE_SUBNET}' \
  --require-headers='${PROXY_HEADER}:true' \
  --log-level='${LOG_LEVEL:-debug}'
```

**Important:** Use single quotes (`'${VAR}'`) in CLI flags to prevent shell expansion.
CQLAI will expand the variable internally.

**Supported fields:**
- `http_host`, `http_port` (port is numeric, but can use var for string representation)
- `api_key`
- `allowed_origins` (array - each element expanded)
- `ip_allowlist` (array - each element expanded)
- `audit_http_headers` (array - each element expanded)
- `required_headers` (map - both keys and values expanded)
- `log_level`, `log_file`, `history_file`
- `mode`, `preset_mode`

### Example Configuration

**High security (7-day expiration):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 7
}
```

**Balanced (30-day expiration, default):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30
}
```

**Long-lived (90-day expiration):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 90
}
```

**No expiration (NOT RECOMMENDED):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 0
}
```

---

## Layer 2: Origin Header Validation

### What It Is

Validates the `Origin` HTTP header sent by web browsers to prevent DNS rebinding attacks.

### What It Protects Against

**DNS Rebinding Attack:**

1. Attacker creates malicious website: `evil.com`
2. User visits `evil.com` in browser
3. JavaScript on `evil.com` tries to call `http://localhost:8888/mcp`
4. Browser sends `Origin: https://evil.com` header
5. **cqlai rejects** because origin not in allowlist

**Without Origin validation:**
- Malicious websites could call your local MCP server
- Exfiltrate Cassandra data
- Execute queries on your behalf

**With Origin validation:**
- Only allowed origins accepted
- Localhost-only by default
- Remote deployments require explicit allowlist

### How It Works

**Localhost binding (default):**
```go
HttpHost: "127.0.0.1"
→ Only allows Origin: http://localhost or http://127.0.0.1
→ Rejects all other origins
```

**Remote binding:**
```go
HttpHost: "0.0.0.0"
AllowedOrigins: ["https://app.company.com"]
→ Only allows Origin: https://app.company.com
→ Rejects all other origins
```

**No Origin header:**
- Direct API calls (curl, HTTP clients) don't send Origin
- These are ALLOWED (not browser-based)

### Configuration

**JSON config:**
```json
{
  "http_host": "0.0.0.0",
  "allowed_origins": [
    "https://app.company.com",
    "http://localhost"
  ]
}
```

**CLI:**
```bash
.mcp start --http-host=0.0.0.0 --allowed-origins "https://app.company.com,http://localhost"
```

### Subdomain Attack Prevention

**Vulnerable approach (using `strings.HasPrefix`):**
```
Allowed: https://app.company.com
Attacker: https://app.company.com.evil.com
→ WOULD PASS (prefix match) ← SECURITY BUG
```

**Secure approach (matchOrigin function):**
```
Allowed: https://app.company.com
Matches:
  ✅ https://app.company.com (exact match)
  ✅ https://app.company.com:443 (with port)
  ✅ https://app.company.com/path (with path)
  ❌ https://app.company.com.evil.com (subdomain attack)
  ❌ https://evil-app.company.com (different subdomain)
```

### Example Origins

**Valid:**
```
http://localhost
http://localhost:3000
https://localhost:3000
http://127.0.0.1
http://127.0.0.1:8080
https://app.company.com
https://app.company.com:443
https://app.company.com/mcp
```

**Invalid (Attack Attempts):**
```
https://app.company.com.evil.com     (subdomain attack)
https://evil-app.company.com         (different subdomain)
https://company.com                  (different domain)
http://192.168.1.100                 (different IP)
```

---

## Layer 3: IP Allowlisting

### What It Is

IP allowlisting restricts which client IP addresses can connect to the MCP server.

**Default:** `127.0.0.1` (localhost only) - Works out-of-the-box for local development.

### Why IP Allowlisting

**Defense in depth:**
- API key could be stolen/leaked
- Origin header can be spoofed by non-browser clients
- IP allowlist provides network-layer protection

**Use cases:**
- Restrict to specific developer machines
- Allow only office subnet
- Allow only trusted proxy servers
- Prevent unauthorized network access

### Configuration

**Default (localhost only):**
```json
{
  "ip_allowlist": ["127.0.0.1"]
}
```

**Single IP:**
```json
{
  "ip_allowlist": ["203.0.113.10"]
}
```

**Multiple IPs:**
```json
{
  "ip_allowlist": [
    "203.0.113.10",
    "203.0.113.11",
    "192.168.1.100"
  ]
}
```

**CIDR Subnet:**
```json
{
  "ip_allowlist": ["10.0.1.0/24"]
}
```
This allows all IPs from `10.0.1.0` to `10.0.1.255` (256 addresses).

**Mixed (IPs + Subnets):**
```json
{
  "ip_allowlist": [
    "127.0.0.1",
    "10.0.1.0/24",
    "203.0.113.10"
  ]
}
```

**IPv6 Support:**
```json
{
  "ip_allowlist": [
    "::1",
    "fe80::/10",
    "2001:db8::1"
  ]
}
```

**Disable IP Allowlist (SECURITY RISK):**
```json
{
  "ip_allowlist_disabled": true
}
```

**Startup warning when disabled:**
```
⚠️  WARNING: IP ALLOWLIST DISABLED ⚠️
All client IPs will be accepted. This is a security risk.
Recommendation: Use default IP allowlist (127.0.0.1) or configure specific IPs
Only disable in fully trusted networks.
```

### How It Works

**IP extraction:**
```
HTTP Request → RemoteAddr: "192.168.1.100:54321"
              → Extract IP: "192.168.1.100"
```

**Validation:**
```
1. Extract client IP from r.RemoteAddr
2. If IP allowlist disabled → ALLOW (with warning)
3. Parse client IP
4. For each entry in allowlist:
   - If CIDR (contains "/"): Check if IP in subnet
   - If direct IP: Check exact match
5. If match found → ALLOW
6. If no match → REJECT (403 Forbidden)
```

**Important:** Only validates **direct connection IP**, not `X-Forwarded-For` header.

### Why X-Forwarded-For Is NOT Validated

**Scenario: Nginx proxy → CQLAI**

```
Client (192.168.1.100) → Nginx (127.0.0.1) → CQLAI (127.0.0.1)
                          │
                          └─ Adds: X-Forwarded-For: 192.168.1.100
```

**CQLAI sees:**
- `RemoteAddr`: `127.0.0.1` (Nginx's IP) ← **This is validated**
- `X-Forwarded-For`: `192.168.1.100` (Original client) ← **Not validated, only logged**

**Why this is secure:**
- Nginx is in allowlist (127.0.0.1) → trusted
- Nginx is responsible for adding correct X-Forwarded-For
- Nginx should validate/filter client IPs at its layer
- CQLAI trusts the proxy to do its job
- CQLAI logs X-Forwarded-For for audit trail only

### CIDR Notation Examples

| CIDR | Range | Total IPs | Use Case |
|------|-------|-----------|----------|
| `10.0.1.0/24` | 10.0.1.0 - 10.0.1.255 | 256 | Office subnet |
| `192.168.0.0/16` | 192.168.0.0 - 192.168.255.255 | 65,536 | Home network |
| `10.0.0.0/8` | 10.0.0.0 - 10.255.255.255 | 16,777,216 | Entire private network |
| `203.0.113.0/25` | 203.0.113.0 - 203.0.113.127 | 128 | Small subnet |

**Tool to calculate CIDR ranges:** https://www.ipaddressguide.com/cidr

### What This Layer Protects Against

| Threat | Protected | How |
|--------|-----------|-----|
| Network-based attacks | ✅ | Only allowlisted IPs can connect |
| Stolen API key from remote attacker | ✅ | Even with valid key, wrong IP rejected |
| Port scanning | ✅ | Non-allowlisted IPs get 403 |
| DDoS from internet | ✅ | Only specific IPs can reach server |
| Lateral movement after breach | ✅ | Attacker on wrong subnet can't connect |

### Deployment Examples

**Local development (default, no config needed):**
```json
{
  "ip_allowlist": ["127.0.0.1"]
}
```
Claude Code on same machine → Works automatically.

**Remote development (specific developer IPs):**
```json
{
  "ip_allowlist": [
    "203.0.113.10",
    "203.0.113.11"
  ]
}
```

**Office network (subnet):**
```json
{
  "ip_allowlist": ["10.0.1.0/24"]
}
```
All developers on office network can connect.

**Nginx proxy (local):**
```json
{
  "ip_allowlist": ["127.0.0.1"]
}
```
Nginx on same machine forwards to localhost → Works.

**Fully trusted network (disable allowlist):**
```json
{
  "ip_allowlist_disabled": true
}
```
⚠️  Triggers security warning at startup.

---

## Layer 4: Required Header Validation

### What It Is

Requires specific HTTP headers with exact values or regex patterns. Used to:
- Verify proxy added required headers
- Detect if request bypassed proxy
- Add custom security markers

### Configuration

**JSON config:**
```json
{
  "required_headers": {
    "X-Proxy-Verified": "true",
    "X-Request-ID": "^req_[0-9a-f]{16}$"
  }
}
```

**CLI:**
```bash
.mcp start --require-headers "X-Proxy-Verified:true,X-Request-ID:^req_.*"
```

### How It Works

**Exact match:**
```
Required: X-Proxy-Verified: "true"
Request header: X-Proxy-Verified: true  → PASS
Request header: X-Proxy-Verified: false → REJECT (403)
Request header: (missing)               → REJECT (403)
```

**Regex pattern:**
```
Required: X-Request-ID: "^req_[0-9a-f]{16}$"
Request header: X-Request-ID: req_a1b2c3d4e5f60708 → PASS
Request header: X-Request-ID: invalid-format       → REJECT (403)
Request header: (missing)                          → REJECT (403)
```

**Pattern detection:**
- If value contains: `^`, `$`, `*`, `+`, `?`, `[`, `]`, `(`, `)`, `|`, `.`
- → Treated as regex pattern
- Otherwise → Exact match

### What This Layer Protects Against

| Threat | Protected | How |
|--------|-----------|-----|
| Bypassed proxy | ✅ | Required header only proxy adds |
| Direct access to internal port | ✅ | No required header → rejected |
| Misconfigured proxy | ✅ | Missing headers detected |
| Unauthorized clients | ✅ | Custom marker validation |

### Deployment Examples

**Nginx proxy adds verification marker:**

**nginx.conf:**
```nginx
location /mcp {
    proxy_pass http://localhost:9999/mcp;
    proxy_set_header X-Proxy-Verified true;
}
```

**cqlai config:**
```json
{
  "required_headers": {
    "X-Proxy-Verified": "true"
  }
}
```

**Result:**
- Requests through Nginx → `X-Proxy-Verified: true` added → PASS
- Direct requests to :9999 → No header → REJECT (403)
- Detects bypassed proxy

**Apache proxy with request ID:**

**apache.conf:**
```apache
RequestHeader set X-Proxy-Verified "true"
RequestHeader set X-Request-ID "req_%{UNIQUE_ID}e"
```

**cqlai config:**
```json
{
  "required_headers": {
    "X-Proxy-Verified": "true",
    "X-Request-ID": "^req_.*"
  }
}
```

**Custom application marker:**
```json
{
  "required_headers": {
    "X-Application-ID": "cqlai-dashboard-v1"
  }
}
```
Only requests with specific application ID accepted.

---

## Header Auditing

### What It Is

Logs specific HTTP headers (if present) for security audit trail and debugging.

**Default:** Logs `X-Forwarded-For` and `User-Agent`

### Configuration

**JSON config:**
```json
{
  "audit_http_headers": ["X-Forwarded-For", "User-Agent", "X-Request-ID"]
}
```

**CLI:**
```bash
.mcp start --audit-http-headers "X-Forwarded-For,User-Agent,X-Request-ID"
```

**Log all headers (debugging):**
```bash
.mcp start --audit-http-headers "ALL"
```

### Logging Format

**Default (X-Forwarded-For + User-Agent):**
```
Request audit: method=POST path=/mcp client_ip=127.0.0.1
  X-Forwarded-For: 192.168.1.100
  User-Agent: Claude Code/1.0
```

**With custom headers:**
```
Request audit: method=POST path=/mcp client_ip=127.0.0.1
  X-Forwarded-For: 192.168.1.100
  User-Agent: Claude Code/1.0
  X-Request-ID: req_a1b2c3d4e5f60708
  X-Proxy-Verified: true
```

**With ALL headers:**
```
Request audit: method=POST path=/mcp client_ip=127.0.0.1
  Accept: application/json
  Content-Type: application/json
  X-Api-Key: 2ABC...XYZa (masked in logs)
  X-Forwarded-For: 192.168.1.100
  User-Agent: Claude Code/1.0
  ... (all request headers)
```

### Use Cases

**Debugging proxy configuration:**
```bash
# Check if proxy is adding headers correctly
.mcp start --audit-http-headers "ALL"
```

**Security investigation:**
```
# Track which client IPs called which tools
# Correlate requests with X-Forwarded-For
# Identify suspicious patterns
```

**Compliance/audit trail:**
```
# Log request IDs for tracking
# Log user agents for access patterns
# Meet audit requirements
```

### What This Provides

- **Visibility:** See all request metadata
- **Debugging:** Troubleshoot proxy issues
- **Audit trail:** Track who accessed what
- **Security investigation:** Identify attack patterns
- **No failures:** Headers logged if present, doesn't reject requests

---

## Configuration Reference

### All Security Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `http_host` | string | `"127.0.0.1"` | HTTP server bind address |
| `http_port` | int | `8888` | HTTP server port |
| `api_key` | string | (auto-generated) | KSUID API key |
| `api_key_max_age_days` | int | `30` | Key expiration (0=disabled) |
| `allowed_origins` | []string | `nil` | Allowed Origin headers |
| `ip_allowlist` | []string | `["127.0.0.1"]` | Allowed client IPs/CIDRs |
| `ip_allowlist_disabled` | bool | `false` | Disable IP checking |
| `audit_http_headers` | []string | `["X-Forwarded-For", "User-Agent"]` | Headers to log |
| `required_headers` | map | `{}` | Required headers with values/patterns |

### JSON Config Example (Full)

```json
{
  "http_host": "0.0.0.0",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 30,
  "allowed_origins": [
    "https://app.company.com",
    "http://localhost"
  ],
  "ip_allowlist": [
    "127.0.0.1",
    "10.0.1.0/24",
    "203.0.113.10"
  ],
  "audit_http_headers": [
    "X-Forwarded-For",
    "User-Agent",
    "X-Request-ID"
  ],
  "required_headers": {
    "X-Proxy-Verified": "true",
    "X-Request-ID": "^req_[0-9a-f]{16}$"
  }
}
```

### CLI Flags (.mcp start)

```bash
.mcp start \
  --http-host=0.0.0.0 \
  --http-port=8888 \
  --api-key=2ABCDEFGHIJKLMNOPQRSTUVWXYZa \
  --api-key-max-age-days=30 \
  --allowed-origins "https://app.company.com,http://localhost" \
  --ip-allowlist "127.0.0.1,10.0.1.0/24" \
  --audit-http-headers "X-Forwarded-For,User-Agent,X-Request-ID" \
  --require-headers "X-Proxy-Verified:true,X-Request-ID:^req_.*"
```

---

## Deployment Scenarios

### Scenario 1: Local Development (Default)

**Setup:**
- Claude Code and CQLAI on same machine
- Direct connection (no proxy)

**Configuration:**
```bash
# No configuration needed - uses secure defaults
.mcp start
```

**What happens:**
1. API key auto-generated and displayed
2. Server binds to 127.0.0.1:8888
3. IP allowlist: 127.0.0.1 (default)
4. Origin validation: localhost only
5. No required headers
6. Logs X-Forwarded-For and User-Agent if present

**Security:**
- ✅ Only localhost can connect
- ✅ API key required
- ✅ Keys expire in 30 days
- ✅ Works out-of-the-box

---

### Scenario 2: Remote CQLAI, Multiple Developers

**Setup:**
- CQLAI on server `cqlai.internal` (10.0.2.50)
- 3 developers on different IPs

**Configuration:**
```json
{
  "http_host": "0.0.0.0",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "api_key_max_age_days": 7,
  "allowed_origins": ["https://developer-portal.company.com"],
  "ip_allowlist": [
    "203.0.113.10",
    "203.0.113.11",
    "203.0.113.12"
  ]
}
```

**What happens:**
1. Server accepts connections on all interfaces (0.0.0.0)
2. Only 3 developer IPs can connect
3. All other IPs get 403 Forbidden
4. Keys expire after 7 days (weekly rotation)

**Security:**
- ✅ Only specific developer IPs allowed
- ✅ Stolen key from other IP useless
- ✅ Short expiration (7 days)

---

### Scenario 3: Office Subnet

**Setup:**
- CQLAI on server in office
- All developers on 10.0.1.0/24 network

**Configuration:**
```json
{
  "http_host": "0.0.0.0",
  "http_port": 8888,
  "api_key": "${MCP_API_KEY}",
  "ip_allowlist": ["10.0.1.0/24"]
}
```

**What happens:**
1. Any IP from 10.0.1.0 to 10.0.1.255 allowed
2. External IPs rejected
3. Easy management (don't list every developer IP)

**Security:**
- ✅ Entire subnet protected
- ✅ No individual IP management
- ✅ External attackers blocked

---

### Scenario 4: Nginx Proxy (Local, Same Machine)

**Setup:**
- Nginx and CQLAI both on same server
- Nginx listens on :8888 (public)
- CQLAI listens on :9999 (localhost only)

**nginx.conf:**
```nginx
server {
    listen 8888;
    server_name cqlai.company.com;

    location /mcp {
        proxy_pass http://localhost:9999/mcp;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Proxy-Verified true;
    }
}
```

**cqlai config (.mcp.json):**
```json
{
  "http_host": "127.0.0.1",
  "http_port": 9999,
  "api_key": "${MCP_API_KEY}",
  "ip_allowlist": ["127.0.0.1"],
  "required_headers": {
    "X-Proxy-Verified": "true"
  },
  "audit_http_headers": ["X-Forwarded-For", "User-Agent"]
}
```

**What happens:**
1. Claude connects to `https://cqlai.company.com:8888/mcp`
2. Nginx adds `X-Proxy-Verified: true` header
3. Nginx forwards to `http://localhost:9999/mcp`
4. CQLAI sees:
   - RemoteAddr: 127.0.0.1 (Nginx) → IP allowlist PASS
   - X-Proxy-Verified: true → Required header PASS
   - X-Forwarded-For: (original client IP) → Logged for audit
5. Request processed

**Security:**
- ✅ Direct access to :9999 blocked (no X-Proxy-Verified header)
- ✅ Only requests through Nginx accepted
- ✅ Audit trail shows original client IP
- ✅ Detects bypassed proxy

---

### Scenario 5: Apache Proxy (Local, Same Machine)

**Setup:**
- Apache and CQLAI both on same server
- Apache listens on :8888 (public)
- CQLAI listens on :9999 (localhost only)

**apache.conf:**
```apache
<VirtualHost *:8888>
    ServerName cqlai.company.com

    ProxyPreserveHost On
    ProxyPass /mcp http://localhost:9999/mcp
    ProxyPassReverse /mcp http://localhost:9999/mcp

    # Add required headers
    RequestHeader set X-Forwarded-For "%{REMOTE_ADDR}s"
    RequestHeader set X-Proxy-Verified "true"
</VirtualHost>
```

**cqlai config (.mcp.json):**
```json
{
  "http_host": "127.0.0.1",
  "http_port": 9999,
  "api_key": "${MCP_API_KEY}",
  "ip_allowlist": ["127.0.0.1"],
  "required_headers": {
    "X-Proxy-Verified": "true"
  }
}
```

**What happens:**
- Same as Nginx scenario
- Apache adds required headers
- Direct access to :9999 rejected (no X-Proxy-Verified)

---

### Scenario 6: Remote Nginx with Request ID Tracking

**Setup:**
- Nginx on proxy.internal (10.0.0.50)
- CQLAI on cqlai.internal (10.0.2.100)
- Different servers

**nginx.conf:**
```nginx
upstream cqlai_backend {
    server cqlai.internal:9999;
}

server {
    listen 8888;

    location /mcp {
        # Generate request ID for tracking
        set $request_id $http_x_request_id;
        if ($request_id = "") {
            set $request_id "req_${msec}_${connection_number}";
        }

        proxy_pass http://cqlai_backend;
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Request-ID $request_id;
        proxy_set_header X-Proxy-Verified true;
    }
}
```

**cqlai config (.mcp.json):**
```json
{
  "http_host": "0.0.0.0",
  "http_port": 9999,
  "api_key": "${MCP_API_KEY}",
  "ip_allowlist": ["10.0.0.50"],
  "required_headers": {
    "X-Proxy-Verified": "true",
    "X-Request-ID": "^req_.*"
  },
  "audit_http_headers": [
    "X-Forwarded-For",
    "User-Agent",
    "X-Request-ID"
  ]
}
```

**What happens:**
1. Developer connects to proxy.internal:8888
2. Nginx generates request ID (if not present)
3. Nginx adds all required headers
4. Nginx connects to cqlai.internal:9999
5. CQLAI sees:
   - RemoteAddr: 10.0.0.50 (Nginx) → IP allowlist PASS
   - X-Proxy-Verified: true → Required header PASS
   - X-Request-ID: req_1234567890_123 → Pattern PASS
6. CQLAI logs:
   - X-Forwarded-For: (original client IP)
   - User-Agent: (client's user agent)
   - X-Request-ID: (for request tracking)
7. Request processed

**Security:**
- ✅ Only Nginx can connect (IP allowlist)
- ✅ Direct access rejected (no required headers)
- ✅ Complete audit trail with request IDs
- ✅ Can track requests across systems

---

## Threat Models

### Threat 1: Stolen API Key

**Attack:**
- Attacker steals API key from developer's config file
- Tries to connect from attacker's machine

**Defense layers:**
- ❌ API key: Valid (stolen)
- ✅ **IP allowlist: Attacker's IP not in allowlist → REJECTED**

**Result:** Attack failed at Layer 3

---

### Threat 2: DNS Rebinding

**Attack:**
1. Attacker creates `evil.com`
2. User visits `evil.com` in browser
3. JavaScript tries: `fetch('http://localhost:8888/mcp')`
4. Browser sends `Origin: https://evil.com`

**Defense layers:**
- ❌ API key: Attacker doesn't have it → REJECTED at Layer 1
- Even if attacker had key:
  - ✅ **Origin: evil.com not in allowlist → REJECTED at Layer 2**

**Result:** Attack failed at Layer 1 (or Layer 2 if key stolen)

---

### Threat 3: Bypassed Proxy

**Attack:**
- Proxy is at localhost:8888
- CQLAI is at localhost:9999
- Attacker discovers port 9999 is open
- Tries direct connection to :9999, bypassing proxy

**Defense layers:**
- ❌ API key: Attacker doesn't have it → REJECTED at Layer 1
- Even if attacker had key:
  - ✅ IP allowlist: 127.0.0.1 allowed → PASS
  - ✅ **Required header (X-Proxy-Verified): Missing → REJECTED at Layer 4**

**Result:** Attack failed (at Layer 1, or Layer 4 if key stolen)

**Why this matters:**
- Proxy may add additional authentication
- Proxy may rate-limit requests
- Proxy may add audit logging
- Direct access bypasses all proxy protections
- Required header ensures requests came through proxy

---

### Threat 4: Expiration Bypass

**Attack:**
- Attacker crafts KSUID with timestamp 10 years in the future
- Even when you add expiration, this key never expires

**Defense:**
- ✅ **Future timestamp validation: Rejected at config load/validation**
- Allows 1-minute clock skew (NTP drift)
- 10 years in future → REJECTED

**Result:** Attack prevented

---

### Threat 5: Timing Attack (Key Guessing)

**Attack:**
- Attacker tries many API keys
- Measures response time to guess correct bytes
- Non-constant-time comparison leaks information via timing

**Defense:**
- ✅ **Constant-time comparison: All bytes compared regardless**
- Uses `crypto/subtle.ConstantTimeCompare()`
- No timing information leaked

**Result:** Attack mitigated

---

### Threat 6: Origin Subdomain Attack

**Attack:**
- Allowed origin: `https://app.company.com`
- Attacker registers: `https://app.company.com.evil.com`
- If using `strings.HasPrefix()`, this would match!

**Defense:**
- ✅ **matchOrigin() function: Only matches exact, :port, or /path**
- `app.company.com.evil.com` does NOT match `app.company.com`

**Result:** Attack prevented

---

## Best Practices

### 1. Never Disable Security Features Without Good Reason

**Bad:**
```json
{
  "api_key_max_age_days": 0,
  "ip_allowlist_disabled": true
}
```

**Good:**
```json
{
  "api_key_max_age_days": 30,
  "ip_allowlist": ["10.0.1.0/24"]
}
```

### 2. Use Environment Variables for Configuration

**ALL config fields support environment variables** - not just API keys!

**Bad (hardcoded sensitive/environment-specific values):**
```json
{
  "api_key": "2ABCDEFGHIJKLMNOPQRSTUVWXYZa",
  "http_host": "192.168.1.100",
  "ip_allowlist": ["10.0.1.0/24", "10.0.2.0/24"]
}
```

**Good (environment variables for everything):**
```json
{
  "api_key": "${MCP_API_KEY}",
  "http_host": "${MCP_HOST:-127.0.0.1}",
  "http_port": 8888,
  "allowed_origins": ["${ALLOWED_ORIGIN}"],
  "ip_allowlist": ["${OFFICE_SUBNET}", "${VPN_SUBNET}"],
  "audit_http_headers": ["X-Forwarded-For", "${CUSTOM_HEADER}"],
  "required_headers": {
    "${PROXY_MARKER}": "${PROXY_VALUE}"
  }
}
```

Set environment variables (names can be anything):
```bash
export MCP_API_KEY="2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
export MCP_HOST="192.168.1.100"
export OFFICE_SUBNET="10.0.1.0/24"
export VPN_SUBNET="10.0.2.0/24"
export ALLOWED_ORIGIN="https://app.company.com"
export PROXY_MARKER="X-Proxy-Verified"
export PROXY_VALUE="true"
export CUSTOM_HEADER="X-Correlation-ID"
```

**Benefits:**
- ✅ No sensitive data in config files
- ✅ Environment-specific configuration (dev/staging/prod)
- ✅ CI/CD friendly (inject secrets at runtime)
- ✅ Secret management integration (Vault, AWS Secrets, etc.)
- ✅ Team collaboration (share config file, not secrets)

**CLI flags also support env vars (inside CQLAI console):**
```bash
export MCP_API_KEY="2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
export OFFICE_SUBNET="10.0.1.0/24"

# Inside CQLAI console (use single quotes!):
.mcp start \
  --api-key='${MCP_API_KEY}' \
  --ip-allowlist='${OFFICE_SUBNET}'
```

**Important:** Use single quotes (`'${VAR}'`) in CLI to prevent shell expansion.
CQLAI expands the variable internally.

### 3. Use Narrow IP Allowlists

**Bad (entire private network):**
```json
{
  "ip_allowlist": ["10.0.0.0/8"]
}
```
16 million IPs allowed!

**Good (specific subnet):**
```json
{
  "ip_allowlist": ["10.0.1.0/24"]
}
```
Only 256 IPs allowed.

### 4. Rotate API Keys Regularly

**Set appropriate expiration:**
```json
{
  "api_key_max_age_days": 30
}
```

**Generate new key (before expiration):**
```bash
# Method 1: Command-line (recommended for automation)
cqlai --generate-mcp-api-key

# Method 2: Inside CQLAI console
.mcp generate-api-key
```

**Update config with new key:**
```bash
# Update config file with new key
# OR update environment variable:
export MCP_API_KEY="new-key-here"

# Then restart MCP server (if running):
.mcp stop
.mcp start
```

**Automated rotation script:**
```bash
#!/bin/bash
# rotate-mcp-key.sh - Run monthly via cron

# Generate new key
NEW_KEY=$(cqlai --generate-mcp-api-key | grep "API Key:" | awk '{print $3}')

# Store in secret manager (example: AWS Secrets Manager)
aws secretsmanager update-secret \
  --secret-id cqlai/mcp-api-key \
  --secret-string "$NEW_KEY"

# Or update env file
echo "MCP_API_KEY=$NEW_KEY" > ~/.cqlai/.env
```

### 5. Use Required Headers with Proxies

**Nginx/Apache proxy:**
```nginx
# Add security marker
proxy_set_header X-Proxy-Verified true;
```

**CQLAI config:**
```json
{
  "required_headers": {
    "X-Proxy-Verified": "true"
  }
}
```

**Why:** Detects if someone discovers the internal port and tries direct access.

### 6. Enable Comprehensive Audit Logging

**Production deployments:**
```json
{
  "audit_http_headers": [
    "X-Forwarded-For",
    "User-Agent",
    "X-Request-ID",
    "X-Correlation-ID"
  ]
}
```

**Benefits:**
- Track which clients accessed server
- Correlate requests across systems
- Security investigation
- Debugging

### 7. Review Logs Regularly

**Check for:**
- Failed authentication attempts (invalid API keys)
- IP rejections (unauthorized IPs)
- Missing required headers (bypassed proxy)
- Unusual patterns (same IP, high frequency)

**Log locations:**
```
~/.cqlai/cqlai_mcp.log
```

---

## Troubleshooting

### Problem: "IP not in allowlist" Error

**Symptom:**
```
Client IP 192.168.1.100 rejected (not in allowlist: [127.0.0.1])
```

**Solution:**
Add your IP to allowlist:
```bash
.mcp start --ip-allowlist "127.0.0.1,192.168.1.100"
```

Or use CIDR if you're on a subnet:
```bash
.mcp start --ip-allowlist "192.168.1.0/24"
```

---

### Problem: "Required header missing" Error

**Symptom:**
```
Required header validation failed: required header 'X-Proxy-Verified' is missing
```

**Cause:** Request didn't come through proxy, or proxy not configured correctly.

**Solution:**

1. **Check if proxy is adding header:**
```nginx
# In nginx.conf
proxy_set_header X-Proxy-Verified true;
```

2. **Temporarily disable required headers for debugging:**
```bash
.mcp start  # Don't specify --require-headers
```

3. **Check proxy is forwarding to correct port:**
```nginx
proxy_pass http://localhost:9999/mcp;  # Correct port?
```

---

### Problem: Claude Code Can't Connect

**Symptom:**
Claude Code shows connection error.

**Debug steps:**

1. **Check MCP server is running:**
```bash
.mcp status
```

2. **Check IP allowlist includes Claude's IP:**
```bash
# If Claude Code on same machine, should be 127.0.0.1 (default)
# If remote, add Claude's IP
```

3. **Check API key is correct in Claude config:**
```json
{
  "mcpServers": {
    "cqlai": {
      "url": "http://127.0.0.1:8888/mcp",
      "headers": {
        "X-API-Key": "2ABCDEFGHIJKLMNOPQRSTUVWXYZa"
      }
    }
  }
}
```

4. **Check logs:**
```bash
tail -f ~/.cqlai/cqlai_mcp.log
```

Look for rejection reasons.

---

### Problem: API Key Expired

**Symptom:**
```
API key expired: created at 2025-11-01T10:00:00Z (age: 744h, max allowed: 720h)
```

**Solution:**

1. **Generate new API key:**
```bash
.mcp generate-api-key

# Output:
Generated API key: 2NEWKEYGHIJKLMNOPQRSTUVWXYZa
```

2. **Update config with new key**

3. **Restart MCP server:**
```bash
.mcp stop
.mcp start --api-key=2NEWKEYGHIJKLMNOPQRSTUVWXYZa
```

---

### Problem: "Origin not allowed" Error

**Symptom:**
```
Origin validation failed: https://app.example.com
```

**Cause:** Origin header from browser not in allowed list.

**Solution:**

Add origin to allowlist:
```bash
.mcp start --allowed-origins "https://app.example.com,http://localhost"
```

Or in JSON:
```json
{
  "allowed_origins": ["https://app.example.com"]
}
```

---

## Security Checklist

### Local Development
- [ ] Use default IP allowlist (127.0.0.1)
- [ ] Use auto-generated API key
- [ ] 30-day expiration enabled
- [ ] No additional configuration needed

### Remote Development
- [ ] Configure IP allowlist (specific IPs or CIDR)
- [ ] Use strong API key (auto-generated)
- [ ] Enable 7-30 day expiration
- [ ] Configure allowed origins if browser access needed
- [ ] Enable audit logging
- [ ] Review logs regularly

### Production Deployment with Proxy
- [ ] Nginx/Apache on public interface
- [ ] CQLAI bound to localhost only
- [ ] IP allowlist includes only proxy IP (127.0.0.1)
- [ ] Required headers configured (X-Proxy-Verified)
- [ ] Audit logging enabled (X-Forwarded-For, X-Request-ID)
- [ ] Short API key expiration (7-14 days)
- [ ] Regular key rotation process
- [ ] Monitor logs for suspicious activity

---

## Quick Reference

### Secure Defaults (No Config Needed)

```
HTTP Host:       127.0.0.1 (localhost only)
HTTP Port:       8888
API Key:         Auto-generated KSUID
Key Expiration:  30 days
IP Allowlist:    127.0.0.1 (localhost only)
Origin Allow:    localhost only (when bound to 127.0.0.1)
Audit Headers:   X-Forwarded-For, User-Agent
Required Headers: None
```

**These defaults work for local development with zero configuration.**

### Generate New API Key

**Method 1: Command-line (Before Starting CQLAI):**
```bash
# Generate key without running CQLAI
cqlai --generate-mcp-api-key

# Output:
═══════════════════════════════════════════════════════════
  MCP API Key Generated
═══════════════════════════════════════════════════════════

API Key: 2ABCDEFGHIJKLMNOPQRSTUVWXYZa

Key Details:
  Format:     KSUID (K-Sortable Unique ID)
  Length:     27 characters (base62 encoding)
  Generated:  2026-01-01 10:00:00 UTC
  Entropy:    128 bits of cryptographically secure random data

[... usage instructions ...]
```

**Use cases:**
- Initial setup (before first run)
- CI/CD pipelines
- Secret management systems (Vault, AWS Secrets Manager)
- Team distribution (generate and share securely)

**Method 2: Inside CQLAI Console:**
```bash
# When CQLAI is already running
.mcp generate-api-key
```

**Use cases:**
- Key rotation during session
- Quick regeneration while working

### Check Current Security Configuration

```bash
# Inside CQLAI console
.mcp status

# Output includes:
#   HTTP endpoint: http://127.0.0.1:8888/mcp
#   API key: 2ABCDEFG...XYZa (generated: 2026-01-01 10:00:00, age: 5h, expires in 25 days)
#   IP allowlist: [127.0.0.1]
#   Allowed origins: localhost only (secure default)
```

### MCP Tool: get_mcp_status

Claude can query security configuration via MCP tool:
```json
{
  "config": {
    "http_endpoint": "http://127.0.0.1:8888/mcp",
    "api_key_masked": "2ABCDEFG...XYZa",
    "api_key_timestamp": "2026-01-01T10:00:00Z",
    "api_key_age": "5h0m0s",
    "api_key_max_age_days": 30,
    "api_key_expired": false,
    "ip_allowlist": ["127.0.0.1"],
    "ip_allowlist_disabled": false,
    "allowed_origins": null
  }
}
```

---

## Additional Resources

- **KSUID Specification:** https://github.com/segmentio/ksuid
- **CIDR Calculator:** https://www.ipaddressguide.com/cidr
- **MCP Protocol Specification:** https://modelcontextprotocol.io/
- **CQLAI MCP User Guide:** [MCP.md](MCP.md)

---

**Last Updated:** 2026-01-01
**CQLAI Version:** 1.0+
**Security Review:** Complete
