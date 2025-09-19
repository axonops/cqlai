#!/bin/bash

echo "=== Checking UDT column types ==="

./cqlai << 'EOF'
USE udt_test;
DESCRIBE users;
EOF