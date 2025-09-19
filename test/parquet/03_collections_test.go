package parquet_test

import (
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTypes(t *testing.T) {
	tmpFile := "/tmp/test_list_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("List Collections", func(t *testing.T) {
		columns := []string{
			"id", "tags", "scores", "nested_lists", "mixed_list",
		}
		types := []string{
			"int", "list<text>", "list<int>", "list<list<int>>", "list<text>",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":   1,
				"tags": []string{"tag1", "tag2", "tag3"},
				"scores": []int{100, 200, 300},
				"nested_lists": []interface{}{
					[]int{1, 2, 3},
					[]int{4, 5, 6},
					[]int{7, 8, 9},
				},
				"mixed_list": []string{"text1", "text2"},
			},
			{
				"id":           2,
				"tags":         []string{"single_tag"},
				"scores":       []int{999},
				"nested_lists": []interface{}{[]int{10}},
				"mixed_list":   nil, // Test null list
			},
			{
				"id":           3,
				"tags":         []string{}, // Empty list
				"scores":       []int{},
				"nested_lists": []interface{}{},
				"mixed_list":   []string{},
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 3)

		// Verify first row lists
		firstRow := rows[0]
		tags := firstRow["tags"].([]interface{})
		assert.Len(t, tags, 3)
		assert.Equal(t, "tag1", tags[0])

		scores := firstRow["scores"].([]interface{})
		assert.Len(t, scores, 3)
		assert.Equal(t, int32(100), scores[0])

		// Verify nested lists
		nested := firstRow["nested_lists"].([]interface{})
		assert.Len(t, nested, 3)
		firstNested := nested[0].([]interface{})
		assert.Len(t, firstNested, 3)
		assert.Equal(t, int32(1), firstNested[0])

		// Verify null handling
		assert.Nil(t, rows[1]["mixed_list"])

		// Verify empty lists
		emptyTags := rows[2]["tags"].([]interface{})
		assert.Len(t, emptyTags, 0)
	})
}

func TestSetTypes(t *testing.T) {
	tmpFile := "/tmp/test_set_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Set Collections", func(t *testing.T) {
		columns := []string{
			"id", "unique_tags", "unique_numbers", "categories",
		}
		types := []string{
			"int", "set<text>", "set<int>", "set<text>",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":             1,
				"unique_tags":    []string{"unique1", "unique2", "unique3"},
				"unique_numbers": []int{10, 20, 30, 40, 50},
				"categories":     []string{"category_a", "category_b"},
			},
			{
				"id":             2,
				"unique_tags":    []string{"single"},
				"unique_numbers": []int{100},
				"categories":     nil, // Null set
			},
			{
				"id":             3,
				"unique_tags":    []string{}, // Empty set
				"unique_numbers": []int{},
				"categories":     []string{},
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 3)

		// Sets are stored as lists in Parquet
		firstRow := rows[0]
		uniqueTags := firstRow["unique_tags"].([]interface{})
		assert.Len(t, uniqueTags, 3)
		assert.Contains(t, uniqueTags, "unique1")

		uniqueNumbers := firstRow["unique_numbers"].([]interface{})
		assert.Len(t, uniqueNumbers, 5)
		assert.Contains(t, uniqueNumbers, int32(10))
	})
}

func TestMapTypes(t *testing.T) {
	tmpFile := "/tmp/test_map_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Map Collections", func(t *testing.T) {
		columns := []string{
			"id", "attributes", "counts", "nested_map", "config",
		}
		types := []string{
			"int", "map<text,text>", "map<text,int>", "map<text,list<text>>", "map<text,text>",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id": 1,
				"attributes": map[string]interface{}{
					"color": "red",
					"size":  "large",
					"type":  "primary",
				},
				"counts": map[string]interface{}{
					"apples":  5,
					"oranges": 10,
					"bananas": 3,
				},
				"nested_map": map[string]interface{}{
					"fruits":     []string{"apple", "orange"},
					"vegetables": []string{"carrot", "lettuce"},
				},
				"config": map[string]interface{}{
					"setting1": "value1",
					"setting2": "value2",
				},
			},
			{
				"id": 2,
				"attributes": map[string]interface{}{
					"single_key": "single_value",
				},
				"counts": map[string]interface{}{
					"total": 100,
				},
				"nested_map": nil, // Null map
				"config":     map[string]interface{}{},
			},
			{
				"id":         3,
				"attributes": map[string]interface{}{}, // Empty map
				"counts":     map[string]interface{}{},
				"nested_map": map[string]interface{}{},
				"config":     nil,
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 3)

		// Verify first row maps
		firstRow := rows[0]
		attributes := firstRow["attributes"].(map[interface{}]interface{})
		assert.Len(t, attributes, 3)
		assert.Equal(t, "red", attributes["color"])
		assert.Equal(t, "large", attributes["size"])

		counts := firstRow["counts"].(map[interface{}]interface{})
		assert.Len(t, counts, 3)
		assert.Equal(t, int32(5), counts["apples"])

		// Verify nested map
		nestedMap := firstRow["nested_map"].(map[interface{}]interface{})
		assert.Len(t, nestedMap, 2)
		fruits := nestedMap["fruits"].([]interface{})
		assert.Len(t, fruits, 2)
		assert.Contains(t, fruits, "apple")

		// Verify null handling
		assert.Nil(t, rows[1]["nested_map"])
		assert.Nil(t, rows[2]["config"])

		// Verify empty maps
		emptyAttrs := rows[2]["attributes"].(map[interface{}]interface{})
		assert.Len(t, emptyAttrs, 0)
	})
}

func TestFrozenCollections(t *testing.T) {
	tmpFile := "/tmp/test_frozen_collections.parquet"
	defer os.Remove(tmpFile)

	t.Run("Frozen Collections", func(t *testing.T) {
		columns := []string{
			"id", "frozen_list", "frozen_set", "frozen_map",
		}
		types := []string{
			"int", "frozen<list<text>>", "frozen<set<int>>", "frozen<map<text,int>>",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":          1,
				"frozen_list": []string{"item1", "item2", "item3"},
				"frozen_set":  []int{100, 200, 300},
				"frozen_map": map[string]interface{}{
					"key1": 10,
					"key2": 20,
					"key3": 30,
				},
			},
			{
				"id":          2,
				"frozen_list": nil,
				"frozen_set":  nil,
				"frozen_map":  nil,
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 2)

		// Frozen collections should behave the same as regular collections
		firstRow := rows[0]
		frozenList := firstRow["frozen_list"].([]interface{})
		assert.Len(t, frozenList, 3)
		assert.Equal(t, "item1", frozenList[0])

		frozenSet := firstRow["frozen_set"].([]interface{})
		assert.Len(t, frozenSet, 3)
		assert.Contains(t, frozenSet, int32(100))

		frozenMap := firstRow["frozen_map"].(map[interface{}]interface{})
		assert.Len(t, frozenMap, 3)
		assert.Equal(t, int32(10), frozenMap["key1"])

		// Verify null handling
		assert.Nil(t, rows[1]["frozen_list"])
		assert.Nil(t, rows[1]["frozen_set"])
		assert.Nil(t, rows[1]["frozen_map"])
	})
}

func TestComplexNestedCollections(t *testing.T) {
	tmpFile := "/tmp/test_complex_nested.parquet"
	defer os.Remove(tmpFile)

	t.Run("Complex Nested Collections", func(t *testing.T) {
		columns := []string{
			"id", "list_of_maps", "map_of_lists", "list_of_lists_of_maps",
		}
		types := []string{
			"int", "list<map<text,int>>", "map<text,list<text>>", "list<list<map<text,text>>>",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id": 1,
				"list_of_maps": []interface{}{
					map[string]interface{}{"a": 1, "b": 2},
					map[string]interface{}{"c": 3, "d": 4},
					map[string]interface{}{"e": 5},
				},
				"map_of_lists": map[string]interface{}{
					"fruits":  []string{"apple", "banana", "orange"},
					"colors":  []string{"red", "green", "blue"},
					"numbers": []string{"one", "two", "three", "four"},
				},
				"list_of_lists_of_maps": []interface{}{
					[]interface{}{
						map[string]interface{}{"k1": "v1", "k2": "v2"},
						map[string]interface{}{"k3": "v3"},
					},
					[]interface{}{
						map[string]interface{}{"k4": "v4", "k5": "v5", "k6": "v6"},
					},
				},
			},
		}

		for _, row := range testData {
			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Read back and verify complex nested structures
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		rows, err := reader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, rows, 1)

		row := rows[0]

		// Verify list of maps
		listOfMaps := row["list_of_maps"].([]interface{})
		assert.Len(t, listOfMaps, 3)
		firstMap := listOfMaps[0].(map[interface{}]interface{})
		assert.Equal(t, int32(1), firstMap["a"])
		assert.Equal(t, int32(2), firstMap["b"])

		// Verify map of lists
		mapOfLists := row["map_of_lists"].(map[interface{}]interface{})
		assert.Len(t, mapOfLists, 3)
		fruits := mapOfLists["fruits"].([]interface{})
		assert.Len(t, fruits, 3)
		assert.Contains(t, fruits, "apple")

		// Verify deeply nested structure
		listOfListsOfMaps := row["list_of_lists_of_maps"].([]interface{})
		assert.Len(t, listOfListsOfMaps, 2)
		firstListOfMaps := listOfListsOfMaps[0].([]interface{})
		assert.Len(t, firstListOfMaps, 2)
		deepMap := firstListOfMaps[0].(map[interface{}]interface{})
		assert.Equal(t, "v1", deepMap["k1"])
	})
}