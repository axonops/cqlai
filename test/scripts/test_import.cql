USE test_udt_parquet;
CREATE TABLE IF NOT EXISTS users_imported (
    id int PRIMARY KEY,
    name text,
    home address
);
TRUNCATE users_imported;
COPY users_imported FROM '/tmp/test_direct.parquet';
SELECT * FROM users_imported;
