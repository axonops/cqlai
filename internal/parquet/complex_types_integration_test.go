package parquet

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplexTypesEndToEnd(t *testing.T) {
	// This test verifies that complex types work end-to-end
	// by writing and reading a Parquet file with complex types

	tmpFile := "/tmp/test_complex_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Write and Read List Types", func(t *testing.T) {
		// Define columns with list types
		columns := []string{"id", "tags", "scores"}
		types := []string{"int", "list<text>", "list<float>"}

		// Write test
		writer, err := NewParquetCaptureWriter(tmpFile, columns, types, DefaultWriterOptions())
		require.NoError(t, err)

		// Write data rows
		rows := []map[string]interface{}{
			{"id": 1, "tags": []string{"tag1", "tag2", "tag3"}, "scores": []float64{1.1, 2.2, 3.3}},
			{"id": 2, "tags": []string{"tag4"}, "scores": []float64{4.4}},
			{"id": 3, "tags": nil, "scores": []float64{5.5, 6.6}},
			{"id": 4, "tags": []string{"tag5", "tag6"}, "scores": nil},
		}

		for _, row := range rows {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read test
		reader, err := NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		// Verify schema
		readColumns, _ := reader.GetSchema()
		assert.Equal(t, columns, readColumns)

		// Read all data
		readRows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, len(rows), len(readRows))

		// Verify first row
		assert.Equal(t, int32(1), readRows[0]["id"])
		if tags, ok := readRows[0]["tags"].([]interface{}); ok {
			assert.Equal(t, 3, len(tags))
			assert.Equal(t, "tag1", tags[0])
			assert.Equal(t, "tag2", tags[1])
			assert.Equal(t, "tag3", tags[2])
		}

		// Verify third row with nil tags
		assert.Equal(t, int32(3), readRows[2]["id"])
		assert.Nil(t, readRows[2]["tags"])
		if scores, ok := readRows[2]["scores"].([]interface{}); ok {
			assert.Equal(t, 2, len(scores))
		}
	})

	t.Run("Write and Read Map Types", func(t *testing.T) {
		tmpMapFile := "/tmp/test_map_types.parquet"
		defer os.Remove(tmpMapFile)

		// Define columns with map types
		columns := []string{"id", "attributes", "counts"}
		types := []string{"int", "map<text,text>", "map<text,int>"}

		writer, err := NewParquetCaptureWriter(tmpMapFile, columns, types, DefaultWriterOptions())
		require.NoError(t, err)

		// Write data rows
		rows := []map[string]interface{}{
			{"id": 1, "attributes": map[string]interface{}{"color": "red", "size": "large"}, "counts": map[string]interface{}{"apples": 5, "oranges": 3}},
			{"id": 2, "attributes": map[string]interface{}{"color": "blue"}, "counts": nil},
			{"id": 3, "attributes": nil, "counts": map[string]interface{}{"bananas": 10}},
		}

		for _, row := range rows {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read test
		reader, err := NewParquetReader(tmpMapFile)
		require.NoError(t, err)
		defer reader.Close()

		readRows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, len(rows), len(readRows))

		// Verify first row
		attrs := readRows[0]["attributes"].(map[interface{}]interface{})
		assert.Equal(t, "red", attrs["color"])
		assert.Equal(t, "large", attrs["size"])
		counts := readRows[0]["counts"].(map[interface{}]interface{})
		assert.Equal(t, int32(5), counts["apples"])
	})

	t.Run("Write and Read Nested Complex Types", func(t *testing.T) {
		tmpNestedFile := "/tmp/test_nested_types.parquet"
		defer os.Remove(tmpNestedFile)

		// Define columns with nested complex types
		columns := []string{"id", "list_of_lists", "map_of_lists"}
		types := []string{"int", "list<list<int>>", "map<text,list<text>>"}

		writer, err := NewParquetCaptureWriter(tmpNestedFile, columns, types, DefaultWriterOptions())
		require.NoError(t, err)

		// Write data with nested structures
		rows := []map[string]interface{}{
			{
				"id": 1,
				"list_of_lists": []interface{}{
					[]int{1, 2, 3},
					[]int{4, 5},
					[]int{6, 7, 8, 9},
				},
				"map_of_lists": map[string]interface{}{
					"fruits":  []string{"apple", "banana"},
					"veggies": []string{"carrot", "lettuce", "tomato"},
				},
			},
		}

		for _, row := range rows {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read and verify
		reader, err := NewParquetReader(tmpNestedFile)
		require.NoError(t, err)
		defer reader.Close()

		readRows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, 1, len(readRows))

		// Verify nested list
		listOfLists := readRows[0]["list_of_lists"].([]interface{})
		assert.Equal(t, 3, len(listOfLists))
		firstList := listOfLists[0].([]interface{})
		assert.Equal(t, 3, len(firstList))
		assert.Equal(t, int32(1), firstList[0])

		// Verify map of lists
		mapOfLists := readRows[0]["map_of_lists"].(map[interface{}]interface{})
		fruits := mapOfLists["fruits"].([]interface{})
		assert.Equal(t, 2, len(fruits))
		assert.Equal(t, "apple", fruits[0])
		veggies := mapOfLists["veggies"].([]interface{})
		assert.Equal(t, 3, len(veggies))
	})

	t.Run("Write and Read Set Types", func(t *testing.T) {
		tmpSetFile := "/tmp/test_set_types.parquet"
		defer os.Remove(tmpSetFile)

		// Define columns with set types (sets are represented as lists)
		columns := []string{"id", "unique_tags", "unique_numbers"}
		types := []string{"int", "set<text>", "set<int>"}

		writer, err := NewParquetCaptureWriter(tmpSetFile, columns, types, DefaultWriterOptions())
		require.NoError(t, err)

		// Write data rows
		rows := []map[string]interface{}{
			{"id": 1, "unique_tags": []string{"unique1", "unique2", "unique3"}, "unique_numbers": []int{10, 20, 30}},
			{"id": 2, "unique_tags": []string{"tag1"}, "unique_numbers": nil},
			{"id": 3, "unique_tags": nil, "unique_numbers": []int{40, 50}},
		}

		for _, row := range rows {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read test
		reader, err := NewParquetReader(tmpSetFile)
		require.NoError(t, err)
		defer reader.Close()

		readRows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, len(rows), len(readRows))

		// Verify set data (stored as lists in Arrow)
		tags := readRows[0]["unique_tags"].([]interface{})
		assert.Equal(t, 3, len(tags))
		assert.Contains(t, tags, "unique1")
		assert.Contains(t, tags, "unique2")
		assert.Contains(t, tags, "unique3")
	})

	t.Run("Write and Read Frozen Collections", func(t *testing.T) {
		tmpFrozenFile := "/tmp/test_frozen_types.parquet"
		defer os.Remove(tmpFrozenFile)

		// Frozen collections are handled by unwrapping the frozen<> wrapper
		columns := []string{"id", "frozen_list", "frozen_map"}
		types := []string{"int", "frozen<list<text>>", "frozen<map<text,int>>"}

		writer, err := NewParquetCaptureWriter(tmpFrozenFile, columns, types, DefaultWriterOptions())
		require.NoError(t, err)

		// Write data rows
		rows := []map[string]interface{}{
			{"id": 1, "frozen_list": []string{"frozen1", "frozen2"}, "frozen_map": map[string]interface{}{"key1": 100, "key2": 200}},
			{"id": 2, "frozen_list": nil, "frozen_map": map[string]interface{}{"key3": 300}},
		}

		for _, row := range rows {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read and verify
		reader, err := NewParquetReader(tmpFrozenFile)
		require.NoError(t, err)
		defer reader.Close()

		readRows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, len(rows), len(readRows))

		// Frozen collections should be read the same as regular collections
		frozenList := readRows[0]["frozen_list"].([]interface{})
		assert.Equal(t, 2, len(frozenList))
		assert.Equal(t, "frozen1", frozenList[0])

		frozenMap := readRows[0]["frozen_map"].(map[interface{}]interface{})
		assert.Equal(t, int32(100), frozenMap["key1"])
		assert.Equal(t, int32(200), frozenMap["key2"])
	})
}