package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/axonops/cqlai/internal/parquet"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <parquet_file>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	reader, err := parquet.NewParquetReader(filename)
	if err != nil {
		log.Fatalf("Failed to open parquet file: %v", err)
	}

	// Get schema
	columns, types := reader.GetSchema()
	fmt.Println("Schema:")
	for i, col := range columns {
		fmt.Printf("  %s: %s\n", col, types[i])
	}

	fmt.Printf("\nTotal rows: %d\n", reader.GetRowCount())

	// Read all data
	rows, err := reader.ReadAll()
	if err != nil {
		_ = reader.Close()
		log.Fatalf("Failed to read data: %v", err)
	}

	// Close reader at the end
	defer reader.Close()

	fmt.Println("\nData:")
	for i, row := range rows {
		jsonData, _ := json.MarshalIndent(row, "", "  ")
		fmt.Printf("Row %d: %s\n", i+1, string(jsonData))
	}
}