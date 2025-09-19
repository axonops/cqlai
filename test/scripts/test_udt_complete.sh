#!/bin/bash
echo "=== Complete UDT Test: Export and Import ==="

# Step 1: Create test data with UDTs
echo "Step 1: Creating test data with UDTs..."
./cqlai << 'EOF'
DROP KEYSPACE IF EXISTS udt_complete_test;
CREATE KEYSPACE udt_complete_test WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
USE udt_complete_test;

-- Create UDT
CREATE TYPE contact_info (
    email text,
    phone text,
    address frozen<map<text, text>>
);

-- Create table with UDT
CREATE TABLE employees (
    id int PRIMARY KEY,
    name text,
    department text,
    contact frozen<contact_info>,
    skills set<text>,
    projects list<text>
);

-- Insert test data
INSERT INTO employees (id, name, department, contact, skills, projects)
VALUES (1, 'Alice Smith', 'Engineering',
    {email: 'alice@example.com', phone: '555-0001', address: {'street': '123 Tech Blvd', 'city': 'San Francisco'}},
    {'Python', 'Go', 'Cassandra'},
    ['Project A', 'Project B']);

INSERT INTO employees (id, name, department, contact, skills, projects)
VALUES (2, 'Bob Johnson', 'Marketing',
    {email: 'bob@example.com', phone: '555-0002', address: {'street': '456 Market St', 'city': 'New York'}},
    {'SEO', 'Analytics', 'Content'},
    ['Campaign X', 'Campaign Y', 'Campaign Z']);

SELECT * FROM employees;
EOF

echo -e "\nStep 2: Exporting to Parquet..."
./cqlai << 'EOF'
USE udt_complete_test;
COPY employees TO '/tmp/employees_export.parquet' WITH FORMAT='PARQUET';
EOF

echo -e "\nStep 3: Checking Parquet file..."
ls -lh /tmp/employees_export.parquet

echo -e "\nStep 4: Creating new table for import..."
./cqlai << 'EOF'
USE udt_complete_test;
CREATE TABLE employees_imported (
    id int PRIMARY KEY,
    name text,
    department text,
    contact frozen<contact_info>,
    skills set<text>,
    projects list<text>
);
EOF

echo -e "\nStep 5: Importing from Parquet..."
./cqlai << 'EOF'
USE udt_complete_test;
COPY employees_imported FROM '/tmp/employees_export.parquet' WITH FORMAT='PARQUET';
EOF

echo -e "\nStep 6: Verifying imported data..."
./cqlai << 'EOF'
USE udt_complete_test;
SELECT * FROM employees_imported;

-- Compare counts
SELECT COUNT(*) as original_count FROM employees;
SELECT COUNT(*) as imported_count FROM employees_imported;
EOF

echo -e "\n=== Test Complete ===""