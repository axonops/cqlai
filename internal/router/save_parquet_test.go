package router

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/file"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// savedResult is a result as the UI holds it: the header row, then the rows.
func savedResult() ([][]string, []string) {
	data := [][]string{
		{"id (PK)", "name", "age"},
		{"1", "alice", "30"},
		{"2", "bob", "25"},
	}
	return data, []string{"int", "text", "int"}
}

// readParquet returns the column names and the values of a written file, one
// string per cell.
//
// It reads the values rather than only counting rows because that is the
// failure this is guarding against: a file with the wrong schema opens cleanly
// and reports the right number of rows, and is empty in every column.
func readParquet(t *testing.T, path string) ([]string, [][]string) {
	t.Helper()

	f, err := os.Open(path) //nolint:gosec // a path this test just wrote
	require.NoError(t, err)
	defer f.Close()

	reader, err := file.NewParquetReader(f)
	require.NoError(t, err)
	defer reader.Close()

	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	require.NoError(t, err)

	table, err := arrowReader.ReadTable(context.Background())
	require.NoError(t, err)
	defer table.Release()

	names := make([]string, 0, table.NumCols())
	for i := 0; i < int(table.NumCols()); i++ {
		names = append(names, table.Schema().Field(i).Name)
	}

	rows := make([][]string, table.NumRows())
	for r := range rows {
		rows[r] = make([]string, table.NumCols())
	}
	for c := 0; c < int(table.NumCols()); c++ {
		col := table.Column(c)
		r := 0
		for _, chunk := range col.Data().Chunks() {
			for i := 0; i < chunk.Len(); i++ {
				if !chunk.IsNull(i) {
					rows[r][c] = chunk.ValueStr(i)
				}
				r++
			}
		}
	}
	return names, rows
}

// TestSaveWritesParquet is the reported gap: CAPTURE writes Parquet and SAVE
// did not, so picking it for a result already on screen meant running the query
// again under CAPTURE, which is the thing SAVE exists to avoid.
func TestSaveWritesParquet(t *testing.T) {
	data, types := savedResult()
	path := filepath.Join(t.TempDir(), "out.parquet")

	cmd, err := ParseSaveCommand("SAVE TO '" + path + "' AS PARQUET")
	require.NoError(t, err)
	require.Equal(t, "PARQUET", cmd.Format)

	require.NoError(t, HandleSaveCommand(*cmd, data, types))

	names, rows := readParquet(t, path)
	assert.Equal(t, []string{"id", "name", "age"}, names,
		"the key marker is not part of the column name")
	assert.Equal(t, [][]string{
		{"1", "alice", "30"},
		{"2", "bob", "25"},
	}, rows, "the values arrive, and the header is not one of them")
}

// TestTheExtensionIsEnoughToPickParquet, the same as .csv and .json.
func TestTheExtensionIsEnoughToPickParquet(t *testing.T) {
	data, types := savedResult()
	path := filepath.Join(t.TempDir(), "out.parquet")

	cmd, err := ParseSaveCommand("SAVE TO '" + path + "'")
	require.NoError(t, err)
	assert.Equal(t, "PARQUET", cmd.Format)

	require.NoError(t, HandleSaveCommand(*cmd, data, types))
	_, rows := readParquet(t, path)
	assert.Len(t, rows, 2)
}

// TestParquetRefusesWithoutTypes.
//
// The writer builds an Arrow schema from the CQL types, and
// AppendValueToBuilder swallows a conversion it cannot do - there is a TODO in
// the writer saying so - so a wrong schema writes a file that opens cleanly and
// holds nothing. Refusing is the better answer, and it must leave no file
// behind for someone to find later and trust.
func TestParquetRefusesWithoutTypes(t *testing.T) {
	data, types := savedResult()

	for name, given := range map[string][]string{
		"none":      nil,
		"too few":   {"int"},
		"too many":  append(slices.Clone(types), "text"),
		"empty set": {},
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out.parquet")
			cmd := SaveCommand{Filename: path, Format: "PARQUET"}

			err := HandleSaveCommand(cmd, data, given)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "PARQUET")
			assert.NoFileExists(t, path, "nothing half-written left behind")
		})
	}
}

// TestEverySaveFormatIsAccepted ties the list offered to the list that works.
//
// Five separate bugs this week had this shape: a list written in more than one
// place with only one copy deciding anything. A format the window offers and
// the command refuses is that bug again.
func TestEverySaveFormatIsAccepted(t *testing.T) {
	data, types := savedResult()

	for _, format := range SaveFormats() {
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out")

			cmd, err := ParseSaveCommand("SAVE TO '" + path + "' AS " + format)
			require.NoError(t, err, "the parser accepts every offered format")
			require.Equal(t, format, cmd.Format)

			require.NoError(t, HandleSaveCommand(*cmd, data, types),
				"and so does the writer")
			assert.FileExists(t, path)
		})
	}
}

// TestParquetTypesUsable is the one check the window and the command share, so
// neither can decide differently about whether Parquet is available.
func TestParquetTypesUsable(t *testing.T) {
	data, types := savedResult()

	assert.True(t, ParquetTypesUsable(data, types))
	assert.False(t, ParquetTypesUsable(data, nil))
	assert.False(t, ParquetTypesUsable(data, types[:2]))
	assert.False(t, ParquetTypesUsable(nil, types), "nothing to save")
}

// TestParquetWritesEveryColumnType is what the feature is actually for.
//
// A result reaches SAVE as [][]string - the values as they are displayed - so
// everything has to be parsed back into its CQL type to build the file. Before
// this, WriteStringRows put the raw string into a builder expecting a number,
// the conversion failed, and AppendValueToBuilder swallowed it and appended
// null. The file opened cleanly, reported the right number of rows, and every
// column that was not text was empty.
//
// That path is CAPTURE's fallback too, for when it has no typed rows to hand.
func TestParquetWritesEveryColumnType(t *testing.T) {
	data := [][]string{
		{"id", "big", "small", "tiny", "f", "d", "ok", "when", "who", "note"},
		{
			"42", "9000000000", "7", "3",
			"1.5", "2.25",
			"true",
			"2026-09-10T11:22:33Z",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"hello",
		},
	}
	types := []string{
		"int", "bigint", "smallint", "tinyint",
		"float", "double",
		"boolean",
		"timestamp",
		"uuid",
		"text",
	}

	path := filepath.Join(t.TempDir(), "types.parquet")
	require.NoError(t, HandleSaveCommand(
		SaveCommand{Filename: path, Format: "PARQUET"}, data, types))

	names, rows := readParquet(t, path)
	assert.Equal(t, data[0], names)
	require.Len(t, rows, 1)

	for i, name := range names {
		assert.NotEmpty(t, rows[0][i], "column %s (%s) must not be null", name, types[i])
	}

	assert.Equal(t, "42", rows[0][0])
	assert.Equal(t, "9000000000", rows[0][1])
	assert.Equal(t, "7", rows[0][2])
	assert.Equal(t, "3", rows[0][3])
	assert.Equal(t, "true", rows[0][6])
	assert.Equal(t, "hello", rows[0][9])
}

// TestAValueThatCannotBeParsedBecomesNull rather than failing the whole save.
// A column that is genuinely empty in Cassandra is displayed as "null", and one
// bad cell should not cost you the other ten thousand rows.
func TestAValueThatCannotBeParsedBecomesNull(t *testing.T) {
	data := [][]string{
		{"id", "name"},
		{"1", "alice"},
		{"null", "bob"},
	}
	path := filepath.Join(t.TempDir(), "nulls.parquet")

	require.NoError(t, HandleSaveCommand(
		SaveCommand{Filename: path, Format: "PARQUET"}, data, []string{"int", "text"}))

	_, rows := readParquet(t, path)
	require.Len(t, rows, 2)
	assert.Equal(t, "1", rows[0][0])
	assert.Empty(t, rows[1][0], "an unparseable id is null")
	assert.Equal(t, "bob", rows[1][1], "and the rest of the row survives")
}

// TestSaveNoLongerWritesASCII.
//
// It wrote the table with its box drawing, padded to the widths the terminal
// happened to be using: a picture of a result rather than the result, which
// nothing could read back. The spellings that mapped onto it go with it.
func TestSaveNoLongerWritesASCII(t *testing.T) {
	assert.NotContains(t, SaveFormats(), "ASCII")

	for _, format := range []string{"ASCII", "TXT", "TEXT"} {
		_, err := ParseSaveCommand("SAVE TO '/tmp/out' AS " + format)
		assert.Error(t, err, "AS %s is refused", format)
	}
}

// TestATextExtensionFallsToTheDefault.
//
// .txt used to pick ASCII. With ASCII gone it takes the same route as .dat,
// .out or no extension at all, which is CSV - the existing rule for anything
// unrecognised rather than a new one, and AS CSV says so outright when it
// matters.
func TestATextExtensionFallsToTheDefault(t *testing.T) {
	for _, name := range []string{"report.txt", "report.text", "report.dat", "report"} {
		cmd, err := ParseSaveCommand("SAVE TO '" + name + "'")
		require.NoError(t, err, name)
		assert.Equal(t, "CSV", cmd.Format, name)
	}
}
