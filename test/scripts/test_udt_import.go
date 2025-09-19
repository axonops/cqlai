package main

import (
	"fmt"
	"log"

	"github.com/axonops/cqlai/internal/parquet"
)

func main() {
	// Create a Parquet file with UDT data
	filename := "/tmp/udt_test_data.parquet"

	columns := []string{"id", "name", "home_address"}
	types := []string{"int", "text", "text"} // UDT stored as text (JSON)

	writer, err := parquet.NewParquetCaptureWriter(filename, columns, types, parquet.DefaultWriterOptions())
	if err != nil {
		log.Fatal(err)
	}

	// Write rows with UDT data as JSON
	rows := []map[string]interface{}{
		{
			"id": 10,
			"name": "Test User 1",
			"home_address": `{"street": "999 Test St", "city": "TestCity", "zip": 99999}`,
		},
		{
			"id": 20,
			"name": "Test User 2",
			"home_address": `{"street": "888 Demo Ave", "city": "DemoTown", "zip": 88888}`,
		},
	}

	for _, row := range rows {
		if err := writer.WriteRow(row); err != nil {
			log.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created Parquet file: %s\n", filename)

	// Now read it back to verify
	reader, err := parquet.NewParquetReader(filename)
	if err != nil {
		log.Fatal(err)
	}

	data, err := reader.ReadAll()
	if err != nil {
		_ = reader.Close()
		log.Fatal(err)
	}

	// Close reader at the end
	defer reader.Close()

	fmt.Println("\nParquet file contents:")
	for i, row := range data {
		fmt.Printf("Row %d: id=%v, name=%v, home_address=%v\n",
			i+1, row["id"], row["name"], row["home_address"])
	}
}