//go:build integration
// +build integration

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNullFirstRowDoesNotPanic covers the crash in #96.
//
// Rows used to be scanned into *interface{}, which cannot hold a NULL. On the
// first row the interface is empty, so gocql had no type to take the zero value
// from and panicked with "reflect: call of reflect.Value.Type on zero Value".
// A table whose first row contains a NULL was therefore impossible to export.
func TestNullFirstRowDoesNotPanic(t *testing.T) {
	dbSession, handler, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS null_first (
		id int PRIMARY KEY,
		name text,
		uid uuid,
		ts timestamp,
		amount double,
		ok boolean,
		tags list<text>
	)`).Exec())

	// id 1 sorts first and has nothing but its key, so the first row the
	// exporter sees is all NULL.
	require.NoError(t, dbSession.Query(`INSERT INTO null_first (id) VALUES (1)`).Exec())
	require.NoError(t, dbSession.Query(
		`INSERT INTO null_first (id, name, ts, amount, ok) VALUES (2, '', 0, 0.0, false)`).Exec())

	parquetFile := filepath.Join(os.TempDir(), "null_first.parquet")
	defer os.Remove(parquetFile)

	// This is the call that used to panic and take the process with it.
	result := handler.HandleMetaCommand(
		fmt.Sprintf("COPY null_first TO '%s' WITH FORMAT='PARQUET';", parquetFile))
	t.Logf("COPY TO result: %v", result)

	require.FileExists(t, parquetFile)
}

// TestNullSurvivesParquetRoundTrip is the other half of #96: a NULL must stay a
// NULL rather than turning into the zero value of whatever the previous row
// happened to hold, and a genuine zero must stay a zero.
func TestNullSurvivesParquetRoundTrip(t *testing.T) {
	dbSession, handler, cleanup := getTestSession(t)
	defer cleanup()

	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS null_rt (
		id int PRIMARY KEY,
		name text,
		amount double,
		ok boolean
	)`).Exec())

	// Row 1 is entirely NULL. Row 2 holds real zeros that must not be mistaken
	// for NULLs: an empty string, 0.0 and false are all legitimate values.
	require.NoError(t, dbSession.Query(`INSERT INTO null_rt (id) VALUES (1)`).Exec())
	require.NoError(t, dbSession.Query(
		`INSERT INTO null_rt (id, name, amount, ok) VALUES (2, '', 0.0, false)`).Exec())

	parquetFile := filepath.Join(os.TempDir(), "null_rt.parquet")
	defer os.Remove(parquetFile)

	result := handler.HandleMetaCommand(
		fmt.Sprintf("COPY null_rt TO '%s' WITH FORMAT='PARQUET';", parquetFile))
	t.Logf("COPY TO result: %v", result)

	// Read the file directly: this is where a lost NULL shows up as "" or 0.
	reader, err := parquet.NewParquetReader(parquetFile)
	require.NoError(t, err)
	// Close reports "file already closed" even on success, so ignore it here
	// the way every other caller does.
	defer reader.Close()

	rows, err := reader.ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 2)

	byID := map[int32]map[string]interface{}{}
	for _, row := range rows {
		id, ok := row["id"].(int32)
		require.True(t, ok, "unexpected id type %T", row["id"])
		byID[id] = row
	}

	nullRow := byID[1]
	require.NotNil(t, nullRow)
	for _, col := range []string{"name", "amount", "ok"} {
		assert.Nil(t, nullRow[col],
			"column %q was NULL in Cassandra and must be null in the Parquet file, got %#v",
			col, nullRow[col])
	}

	zeroRow := byID[2]
	require.NotNil(t, zeroRow)
	assert.NotNil(t, zeroRow["name"], "an empty string is a value, not a NULL")
	assert.Equal(t, "", zeroRow["name"])
	assert.Equal(t, float64(0), zeroRow["amount"], "0.0 is a value, not a NULL")
	assert.Equal(t, false, zeroRow["ok"], "false is a value, not a NULL")

	// And back into Cassandra: both rows must return exactly as they went out.
	require.NoError(t, dbSession.Query(`CREATE TABLE IF NOT EXISTS null_rt_dest (
		id int PRIMARY KEY,
		name text,
		amount double,
		ok boolean
	)`).Exec())

	result = handler.HandleMetaCommand(
		fmt.Sprintf("COPY null_rt_dest FROM '%s' WITH FORMAT='PARQUET';", parquetFile))
	t.Logf("COPY FROM result: %v", result)

	var gotName *string
	var gotAmount *float64
	var gotOK *bool

	require.NoError(t, dbSession.Query(
		`SELECT name, amount, ok FROM null_rt_dest WHERE id = 1`).Scan(&gotName, &gotAmount, &gotOK))
	assert.Nil(t, gotName, "a NULL must not come back as an empty string")
	assert.Nil(t, gotAmount, "a NULL must not come back as 0")
	assert.Nil(t, gotOK, "a NULL must not come back as false")

	require.NoError(t, dbSession.Query(
		`SELECT name, amount, ok FROM null_rt_dest WHERE id = 2`).Scan(&gotName, &gotAmount, &gotOK))
	require.NotNil(t, gotName)
	assert.Equal(t, "", *gotName)
	require.NotNil(t, gotAmount)
	assert.Equal(t, float64(0), *gotAmount)
	require.NotNil(t, gotOK)
	assert.Equal(t, false, *gotOK)
}
