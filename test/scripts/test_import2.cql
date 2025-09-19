USE test_udt_parquet;
COPY users_imported (id, name, home) FROM '/tmp/test_direct.parquet';
SELECT * FROM users_imported;
