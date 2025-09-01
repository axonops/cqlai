-- Test file for comment handling
-- This entire line should be ignored

-- Single line comment before a query
SELECT * FROM system.local;

SELECT * FROM system.local; -- Inline comment after query

/* Block comment on single line */ SELECT * FROM system.local;

/* 
   Multi-line 
   block 
   comment 
*/
SELECT * FROM system.local;

-- Mixed comments
/* block */ SELECT * FROM system.local; -- line comment

-- Empty statement (should be skipped)
--;

/* Just a block comment - should be skipped */

-- Test with multiple statements
SELECT release_version FROM system.local; -- Get version
SELECT cluster_name FROM system.local;    -- Get cluster name

-- End of test