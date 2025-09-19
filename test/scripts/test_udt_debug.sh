#!/bin/bash
# Debug UDT values

echo "=== Debug UDT Values ==="

./cqlai << 'EOF'
USE udt_test;
-- Query to see what we get
SELECT * FROM users;
EXIT;
EOF

echo "=== Check what's in the Parquet file with hexdump ==="
hexdump -C /tmp/udt_export.parquet | head -50