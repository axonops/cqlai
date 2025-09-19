package parquet_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUDTBasic(t *testing.T) {
	tmpFile := "/tmp/test_udt_basic.parquet"
	defer os.Remove(tmpFile)

	t.Run("Basic UDT as JSON", func(t *testing.T) {
		// Since UDTs are stored as JSON strings in Parquet
		columns := []string{"id", "name", "address"}
		types := []string{"int", "text", "text"} // UDT stored as text (JSON)

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Create test data with UDT as JSON
		testData := []map[string]interface{}{
			{
				"id":   1,
				"name": "Alice",
				"address": mustJSON(map[string]interface{}{
					"street": "123 Main St",
					"city":   "New York",
					"zip":    10001,
				}),
			},
			{
				"id":   2,
				"name": "Bob",
				"address": mustJSON(map[string]interface{}{
					"street": "456 Oak Ave",
					"city":   "Boston",
					"zip":    02101,
				}),
			},
			{
				"id":      3,
				"name":    "Charlie",
				"address": nil, // Test null UDT
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

		// Verify UDT data
		firstRow := rows[0]
		assert.Equal(t, int32(1), firstRow["id"])
		assert.Equal(t, "Alice", firstRow["name"])

		// Parse JSON UDT
		var address map[string]interface{}
		err = json.Unmarshal([]byte(firstRow["address"].(string)), &address)
		require.NoError(t, err)
		assert.Equal(t, "123 Main St", address["street"])
		assert.Equal(t, "New York", address["city"])
		assert.Equal(t, float64(10001), address["zip"])

		// Verify null UDT
		assert.Nil(t, rows[2]["address"])
	})
}

func TestComplexUDT(t *testing.T) {
	tmpFile := "/tmp/test_complex_udt.parquet"
	defer os.Remove(tmpFile)

	t.Run("Nested UDT with Collections", func(t *testing.T) {
		columns := []string{"id", "user_profile"}
		types := []string{"int", "text"} // Complex UDT as JSON

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Complex UDT with nested structures
		testData := []map[string]interface{}{
			{
				"id": 1,
				"user_profile": mustJSON(map[string]interface{}{
					"personal_info": map[string]interface{}{
						"first_name": "John",
						"last_name":  "Doe",
						"age":        30,
					},
					"contact": map[string]interface{}{
						"email":  "john@example.com",
						"phone":  "555-0100",
						"social": []string{"twitter", "linkedin"},
					},
					"preferences": map[string]interface{}{
						"notifications": true,
						"theme":         "dark",
						"languages":     []string{"en", "es", "fr"},
					},
					"addresses": []map[string]interface{}{
						{
							"type":   "home",
							"street": "123 Home St",
							"city":   "HomeCity",
							"zip":    12345,
						},
						{
							"type":   "work",
							"street": "456 Work Ave",
							"city":   "WorkCity",
							"zip":    67890,
						},
					},
				}),
			},
			{
				"id": 2,
				"user_profile": mustJSON(map[string]interface{}{
					"personal_info": map[string]interface{}{
						"first_name": "Jane",
						"last_name":  "Smith",
						"age":        28,
					},
					"contact": map[string]interface{}{
						"email": "jane@example.com",
						"phone": "555-0200",
					},
					"preferences": map[string]interface{}{
						"notifications": false,
						"theme":         "light",
					},
					"addresses": []map[string]interface{}{},
				}),
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

		// Verify complex UDT structure
		firstRow := rows[0]
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(firstRow["user_profile"].(string)), &profile)
		require.NoError(t, err)

		// Check nested structures
		personalInfo := profile["personal_info"].(map[string]interface{})
		assert.Equal(t, "John", personalInfo["first_name"])
		assert.Equal(t, float64(30), personalInfo["age"])

		contact := profile["contact"].(map[string]interface{})
		assert.Equal(t, "john@example.com", contact["email"])
		social := contact["social"].([]interface{})
		assert.Len(t, social, 2)
		assert.Contains(t, social, "twitter")

		addresses := profile["addresses"].([]interface{})
		assert.Len(t, addresses, 2)
		homeAddr := addresses[0].(map[string]interface{})
		assert.Equal(t, "home", homeAddr["type"])
		assert.Equal(t, "123 Home St", homeAddr["street"])
	})
}

func TestTupleTypes(t *testing.T) {
	tmpFile := "/tmp/test_tuple_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Tuple Types as JSON", func(t *testing.T) {
		columns := []string{"id", "coordinates", "rgb_color", "metadata"}
		types := []string{"int", "text", "text", "text"} // Tuples as JSON

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":          1,
				"coordinates": mustJSON([]interface{}{40.7128, -74.0060}), // lat, lon
				"rgb_color":   mustJSON([]interface{}{255, 0, 0}),         // red
				"metadata":    mustJSON([]interface{}{"version", 1, true}),
			},
			{
				"id":          2,
				"coordinates": mustJSON([]interface{}{51.5074, -0.1278}), // London
				"rgb_color":   mustJSON([]interface{}{0, 255, 0}),        // green
				"metadata":    mustJSON([]interface{}{"version", 2, false}),
			},
			{
				"id":          3,
				"coordinates": nil, // null tuple
				"rgb_color":   mustJSON([]interface{}{0, 0, 255}), // blue
				"metadata":    nil,
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

		// Verify tuple data
		firstRow := rows[0]
		var coordinates []interface{}
		err = json.Unmarshal([]byte(firstRow["coordinates"].(string)), &coordinates)
		require.NoError(t, err)
		assert.Len(t, coordinates, 2)
		assert.Equal(t, float64(40.7128), coordinates[0])
		assert.Equal(t, float64(-74.0060), coordinates[1])

		var rgbColor []interface{}
		err = json.Unmarshal([]byte(firstRow["rgb_color"].(string)), &rgbColor)
		require.NoError(t, err)
		assert.Len(t, rgbColor, 3)
		assert.Equal(t, float64(255), rgbColor[0])

		// Verify null tuples
		assert.Nil(t, rows[2]["coordinates"])
		assert.Nil(t, rows[2]["metadata"])
	})
}

func TestMixedComplexTypes(t *testing.T) {
	tmpFile := "/tmp/test_mixed_complex.parquet"
	defer os.Remove(tmpFile)

	t.Run("UDT with Collections and Nested Types", func(t *testing.T) {
		columns := []string{
			"id", "name", "tags", "scores", "profile", "settings",
		}
		types := []string{
			"int", "text", "list<text>", "map<text,int>", "text", "text",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":     1,
				"name":   "Test User",
				"tags":   []string{"admin", "developer", "tester"},
				"scores": map[string]interface{}{"math": 95, "science": 88, "english": 92},
				"profile": mustJSON(map[string]interface{}{
					"bio":      "Software developer",
					"location": "San Francisco",
					"skills":   []string{"Go", "Python", "Cassandra"},
					"contact": map[string]interface{}{
						"email": "test@example.com",
						"phone": "555-1234",
					},
				}),
				"settings": mustJSON(map[string]interface{}{
					"notifications": map[string]interface{}{
						"email": true,
						"sms":   false,
						"push":  true,
					},
					"privacy": map[string]interface{}{
						"profile_visible": true,
						"show_email":      false,
					},
				}),
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
		assert.Len(t, rows, 1)

		row := rows[0]

		// Verify collections
		tags := row["tags"].([]interface{})
		assert.Len(t, tags, 3)
		assert.Contains(t, tags, "admin")

		scores := row["scores"].(map[interface{}]interface{})
		assert.Equal(t, int32(95), scores["math"])

		// Verify UDT data
		var profile map[string]interface{}
		err = json.Unmarshal([]byte(row["profile"].(string)), &profile)
		require.NoError(t, err)
		assert.Equal(t, "Software developer", profile["bio"])

		skills := profile["skills"].([]interface{})
		assert.Len(t, skills, 3)
		assert.Contains(t, skills, "Go")

		contact := profile["contact"].(map[string]interface{})
		assert.Equal(t, "test@example.com", contact["email"])

		// Verify nested settings
		var settings map[string]interface{}
		err = json.Unmarshal([]byte(row["settings"].(string)), &settings)
		require.NoError(t, err)

		notifications := settings["notifications"].(map[string]interface{})
		assert.Equal(t, true, notifications["email"])
		assert.Equal(t, false, notifications["sms"])
	})
}

// Helper function to convert data to JSON string
func mustJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return string(data)
}