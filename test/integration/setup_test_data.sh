#!/bin/bash

# Setup test data for MCP integration testing
# Creates keyspaces, tables, roles, and sample data in Cassandra

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "========================================="
echo "Setting Up Cassandra Test Data for MCP"
echo "========================================="

# Configuration
CASSANDRA_HOST=${CASSANDRA_HOST:-127.0.0.1}
CASSANDRA_PORT=${CASSANDRA_PORT:-9042}
CASSANDRA_USER=${CASSANDRA_USER:-cassandra}
CASSANDRA_PASS=${CASSANDRA_PASS:-cassandra}

# Check if Cassandra is running
echo -e "${YELLOW}Checking Cassandra connection...${NC}"
if command -v cqlsh &> /dev/null; then
    if cqlsh -u "$CASSANDRA_USER" -p "$CASSANDRA_PASS" "$CASSANDRA_HOST" "$CASSANDRA_PORT" -e "SELECT release_version FROM system.local" &> /dev/null; then
        echo -e "${GREEN}✓ Connected to Cassandra${NC}"
    else
        echo -e "${RED}✗ Cannot connect to Cassandra${NC}"
        echo "Start Cassandra with: podman run -d -p 9042:9042 cassandra:latest"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ cqlsh not found, attempting to seed via podman exec...${NC}"
fi

echo ""
echo -e "${YELLOW}Creating test keyspace and schema...${NC}"

# Create CQL script
CQLSH_CMD="cqlsh -u $CASSANDRA_USER -p $CASSANDRA_PASS $CASSANDRA_HOST $CASSANDRA_PORT"

# Use podman exec if running in container, otherwise use cqlsh
if podman ps --format "{{.Names}}" | grep -q "cassandra"; then
    CONTAINER_NAME=$(podman ps --format "{{.Names}}" | grep cassandra | head -1)
    echo "Using container: $CONTAINER_NAME"
    EXEC_CMD="podman exec -i $CONTAINER_NAME cqlsh -u $CASSANDRA_USER -p $CASSANDRA_PASS"
else
    EXEC_CMD="$CQLSH_CMD"
fi

# Execute CQL script
$EXEC_CMD <<'EOF'
-- Create test keyspace
CREATE KEYSPACE IF NOT EXISTS test_mcp
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

-- Create users table
CREATE TABLE IF NOT EXISTS test_mcp.users (
  id uuid PRIMARY KEY,
  email text,
  name text,
  created_at timestamp,
  role text,
  is_active boolean
);

-- Create events table
CREATE TABLE IF NOT EXISTS test_mcp.events (
  id uuid,
  timestamp timestamp,
  event_type text,
  user_id uuid,
  data text,
  PRIMARY KEY (id, timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC);

-- Create orders table
CREATE TABLE IF NOT EXISTS test_mcp.orders (
  order_id uuid PRIMARY KEY,
  user_id uuid,
  total decimal,
  status text,
  created_at timestamp
);

-- Insert sample users (20 users)
INSERT INTO test_mcp.users (id, email, name, created_at, role, is_active)
  VALUES (11111111-1111-1111-1111-111111111111, 'alice@example.com', 'Alice Admin', toTimestamp(now()), 'admin', true);
INSERT INTO test_mcp.users (id, email, name, created_at, role, is_active)
  VALUES (22222222-2222-2222-2222-222222222222, 'bob@example.com', 'Bob User', toTimestamp(now()), 'customer', true);
INSERT INTO test_mcp.users (id, email, name, created_at, role, is_active)
  VALUES (33333333-3333-3333-3333-333333333333, 'charlie@example.com', 'Charlie Dev', toTimestamp(now()), 'developer', true);
INSERT INTO test_mcp.users (id, email, name, created_at, role, is_active)
  VALUES (44444444-4444-4444-4444-444444444444, 'diana@example.com', 'Diana Analyst', toTimestamp(now()), 'analyst', true);
INSERT INTO test_mcp.users (id, email, name, created_at, role, is_active)
  VALUES (55555555-5555-5555-5555-555555555555, 'eve@example.com', 'Eve Support', toTimestamp(now()), 'support', false);

-- Insert sample events (10 events)
INSERT INTO test_mcp.events (id, timestamp, event_type, user_id, data)
  VALUES (uuid(), toTimestamp(now()), 'login', 11111111-1111-1111-1111-111111111111, 'User logged in');
INSERT INTO test_mcp.events (id, timestamp, event_type, user_id, data)
  VALUES (uuid(), toTimestamp(now()), 'purchase', 22222222-2222-2222-2222-222222222222, 'Order placed');
INSERT INTO test_mcp.events (id, timestamp, event_type, user_id, data)
  VALUES (uuid(), toTimestamp(now()), 'login', 33333333-3333-3333-3333-333333333333, 'Developer login');

-- Insert sample orders (5 orders)
INSERT INTO test_mcp.orders (order_id, user_id, total, status, created_at)
  VALUES (uuid(), 22222222-2222-2222-2222-222222222222, 99.99, 'completed', toTimestamp(now()));
INSERT INTO test_mcp.orders (order_id, user_id, total, status, created_at)
  VALUES (uuid(), 44444444-4444-4444-4444-444444444444, 149.50, 'pending', toTimestamp(now()));

-- Create indexes
CREATE INDEX IF NOT EXISTS users_email_idx ON test_mcp.users (email);
CREATE INDEX IF NOT EXISTS users_role_idx ON test_mcp.users (role);

EOF

echo -e "${GREEN}✓ Test schema and data created${NC}"

echo ""
echo -e "${YELLOW}Creating custom roles for RBAC testing...${NC}"

$EXEC_CMD <<'EOF'
-- Create custom roles (if they don't exist, ignore errors)
CREATE ROLE IF NOT EXISTS app_readonly WITH PASSWORD = 'readonly123' AND LOGIN = true;
CREATE ROLE IF NOT EXISTS app_readwrite WITH PASSWORD = 'readwrite123' AND LOGIN = true;
CREATE ROLE IF NOT EXISTS app_admin WITH PASSWORD = 'admin123' AND LOGIN = true AND SUPERUSER = false;
CREATE ROLE IF NOT EXISTS developer WITH PASSWORD = 'dev123' AND LOGIN = true;

-- Grant permissions
GRANT SELECT ON KEYSPACE test_mcp TO app_readonly;
GRANT SELECT, MODIFY ON KEYSPACE test_mcp TO app_readwrite;
GRANT ALL PERMISSIONS ON KEYSPACE test_mcp TO app_admin;
GRANT SELECT, MODIFY ON KEYSPACE test_mcp TO developer;

-- Grant some system access for testing
GRANT SELECT ON KEYSPACE system TO app_readonly;
GRANT SELECT ON KEYSPACE system_schema TO app_readonly;

EOF

echo -e "${GREEN}✓ Custom roles created and permissions granted${NC}"

echo ""
echo -e "${YELLOW}Verifying setup...${NC}"

$EXEC_CMD <<'EOF'
-- List roles
SELECT role FROM system_auth.roles;

-- Count tables
SELECT keyspace_name, table_name FROM system_schema.tables WHERE keyspace_name = 'test_mcp';

-- Count users
SELECT COUNT(*) FROM test_mcp.users;

EOF

echo ""
echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}✓ Test Data Setup Complete${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""
echo "Test Keyspace: test_mcp"
echo "Tables:"
echo "  - users (5+ rows, indexed by email and role)"
echo "  - events (3+ rows, time-series)"
echo "  - orders (2+ rows)"
echo ""
echo "Custom Roles:"
echo "  - app_readonly (password: readonly123) - SELECT only"
echo "  - app_readwrite (password: readwrite123) - SELECT + MODIFY"
echo "  - app_admin (password: admin123) - ALL PERMISSIONS (not superuser)"
echo "  - developer (password: dev123) - SELECT + MODIFY"
echo ""
echo "Usage:"
echo "  cqlai -h $CASSANDRA_HOST -p $CASSANDRA_PORT -u app_readonly -k test_mcp"
echo "  cqlai -h $CASSANDRA_HOST -p $CASSANDRA_PORT -u app_readwrite -k test_mcp"
echo ""
