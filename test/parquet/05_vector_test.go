package parquet_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorTypes(t *testing.T) {
	tmpFile := "/tmp/test_vector_types.parquet"
	defer os.Remove(tmpFile)

	t.Run("Vector Data Types", func(t *testing.T) {
		// Cassandra 5.0 supports vector types for ML/AI workloads
		// Vectors are typically stored as arrays of floats
		columns := []string{
			"id", "name", "embedding", "features", "metadata",
		}
		types := []string{
			"int", "text", "list<float>", "list<float>", "text",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		testData := []map[string]interface{}{
			{
				"id":   1,
				"name": "document_1",
				// 128-dimensional embedding vector
				"embedding": generateVector(128, 1.0),
				// Feature vector
				"features": []float32{0.1, 0.2, 0.3, 0.4, 0.5},
				"metadata": mustJSON(map[string]interface{}{
					"source": "training_set",
					"label":  "category_a",
				}),
			},
			{
				"id":   2,
				"name": "document_2",
				// Different embedding
				"embedding": generateVector(128, 2.0),
				"features":  []float32{0.6, 0.7, 0.8, 0.9, 1.0},
				"metadata": mustJSON(map[string]interface{}{
					"source": "training_set",
					"label":  "category_b",
				}),
			},
			{
				"id":        3,
				"name":      "document_3",
				"embedding": nil, // Test null vector
				"features":  []float32{},
				"metadata":  nil,
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

		// Verify vector data
		firstRow := rows[0]
		if embedding, ok := firstRow["embedding"].([]interface{}); ok {
			assert.Len(t, embedding, 128)
			// Check first few values
			assert.InDelta(t, float32(0.0), embedding[0], 0.01)
		}

		if features, ok := firstRow["features"].([]interface{}); ok {
			assert.Len(t, features, 5)
			assert.InDelta(t, float32(0.1), features[0], 0.01)
		}

		// Verify null vector
		assert.Nil(t, rows[2]["embedding"])
	})
}

func TestHighDimensionalVectors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-dimensional vector test in short mode")
	}

	tmpFile := "/tmp/test_high_dim_vectors.parquet"
	defer os.Remove(tmpFile)

	t.Run("High Dimensional Vectors for ML", func(t *testing.T) {
		columns := []string{
			"id", "model_name", "vector_1536", "vector_768", "vector_384",
		}
		types := []string{
			"int", "text", "list<float>", "list<float>", "list<float>",
		}

		options := parquet.DefaultWriterOptions()
		options.ChunkSize = 100 // Smaller chunks for large vectors

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, options)
		require.NoError(t, err)

		// Common embedding dimensions for popular models
		// 1536: OpenAI ada-002
		// 768: BERT base
		// 384: Sentence transformers mini

		for i := 0; i < 100; i++ {
			row := map[string]interface{}{
				"id":          i,
				"model_name":  "test_model",
				"vector_1536": generateVector(1536, float64(i)*0.1),
				"vector_768":  generateVector(768, float64(i)*0.2),
				"vector_384":  generateVector(384, float64(i)*0.3),
			}

			err = writer.WriteRow(row)
			require.NoError(t, err)
		}

		err = writer.Close()
		require.NoError(t, err)

		// Check file size
		fileInfo, err := os.Stat(tmpFile)
		require.NoError(t, err)
		t.Logf("File size for high-dimensional vectors: %.2f MB", float64(fileInfo.Size())/(1024*1024))

		// Read back sample
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		batch, err := reader.ReadBatch(5)
		require.NoError(t, err)
		assert.Len(t, batch, 5)

		// Verify dimensions
		for _, row := range batch {
			if vec1536, ok := row["vector_1536"].([]interface{}); ok {
				assert.Len(t, vec1536, 1536)
			}

			if vec768, ok := row["vector_768"].([]interface{}); ok {
				assert.Len(t, vec768, 768)
			}

			if vec384, ok := row["vector_384"].([]interface{}); ok {
				assert.Len(t, vec384, 384)
			}
		}
	})
}

func TestVectorSimilarityData(t *testing.T) {
	tmpFile := "/tmp/test_vector_similarity.parquet"
	defer os.Remove(tmpFile)

	t.Run("Vector Similarity Search Data", func(t *testing.T) {
		// Schema for vector similarity search use case
		columns := []string{
			"id", "content_id", "content_type", "title",
			"embedding", "norm", "metadata", "created_at",
		}
		types := []string{
			"int", "uuid", "text", "text",
			"list<float>", "float", "text", "timestamp",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Generate test data for similarity search
		contentTypes := []string{"article", "product", "image", "video"}

		for i := 0; i < 50; i++ {
			embedding := generateNormalizedVector(256)

			row := map[string]interface{}{
				"id":           i,
				"content_id":   generateUUID(i),
				"content_type": contentTypes[i%len(contentTypes)],
				"title":        generateTitle(i),
				"embedding":    embedding,
				"norm":         calculateNorm(embedding),
				"metadata": mustJSON(map[string]interface{}{
					"tags":        generateTags(i),
					"category":    generateCategory(i),
					"score":       float64(i%100) / 100.0,
					"indexed":     true,
					"version":     "1.0",
				}),
				"created_at": generateTimestamp(i),
			}

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
		assert.Len(t, rows, 50)

		// Verify vector data structure
		for i, row := range rows[:5] {
			assert.Equal(t, int32(i), row["id"])

			if embedding, ok := row["embedding"].([]interface{}); ok {
				assert.Len(t, embedding, 256)
			}

			// Verify normalized vectors
			if norm, ok := row["norm"].(float32); ok {
				assert.Greater(t, norm, float32(0))
				assert.LessOrEqual(t, norm, float32(1.1)) // Allow small tolerance
			}

			// Verify metadata
			if metadataStr, ok := row["metadata"].(string); ok {
				var metadata map[string]interface{}
				err = json.Unmarshal([]byte(metadataStr), &metadata)
				require.NoError(t, err)
				assert.NotEmpty(t, metadata["tags"])
				assert.NotEmpty(t, metadata["category"])
			}
		}
	})
}

// Helper functions for vector tests

func generateVector(dim int, seed float64) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		// Generate deterministic values based on seed
		vec[i] = float32(math.Sin(float64(i)*0.1+seed) * 0.5)
	}
	return vec
}

func generateNormalizedVector(dim int) []float32 {
	vec := generateVector(dim, 0)
	norm := calculateNorm(vec)

	// Normalize the vector
	for i := range vec {
		vec[i] /= norm
	}
	return vec
}

func calculateNorm(vec []float32) float32 {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	return float32(math.Sqrt(sum))
}

func generateUUID(seed int) string {
	return fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", seed)
}

func generateTitle(seed int) string {
	titles := []string{
		"Introduction to Machine Learning",
		"Advanced Data Structures",
		"Cloud Computing Fundamentals",
		"Deep Learning Applications",
		"System Design Principles",
	}
	return titles[seed%len(titles)] + fmt.Sprintf(" - Part %d", seed)
}

func generateTags(seed int) []string {
	allTags := []string{"ml", "ai", "data", "cloud", "engineering", "tutorial", "advanced", "beginner"}
	numTags := (seed % 3) + 2 // 2-4 tags
	tags := make([]string, numTags)
	for i := 0; i < numTags; i++ {
		tags[i] = allTags[(seed+i)%len(allTags)]
	}
	return tags
}

func generateCategory(seed int) string {
	categories := []string{"technology", "science", "education", "business", "research"}
	return categories[seed%len(categories)]
}

func generateTimestamp(seed int) time.Time {
	return time.Now().Add(-time.Duration(seed) * time.Hour)
}