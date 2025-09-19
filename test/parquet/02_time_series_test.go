package parquet_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/axonops/cqlai/internal/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeSeriesData(t *testing.T) {
	tmpFile := "/tmp/test_time_series.parquet"
	defer os.Remove(tmpFile)

	t.Run("Wide Partition Time Series", func(t *testing.T) {
		// Typical time series schema with partition key (sensor_id)
		// and clustering columns (timestamp, measurement_type)
		columns := []string{
			"sensor_id", "timestamp", "measurement_type",
			"value", "unit", "quality", "metadata",
		}
		types := []string{
			"text", "timestamp", "text",
			"double", "text", "int", "text",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		// Simulate wide partition: multiple sensors with many measurements over time
		sensors := []string{"sensor_001", "sensor_002", "sensor_003"}
		measurementTypes := []string{"temperature", "humidity", "pressure"}

		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		rowsWritten := 0

		// Generate time series data
		for _, sensorID := range sensors {
			// Each sensor has 1000 measurements (wide partition)
			for i := 0; i < 1000; i++ {
				timestamp := baseTime.Add(time.Duration(i) * time.Minute)

				for _, measurementType := range measurementTypes {
					var value float64
					var unit string

					switch measurementType {
					case "temperature":
						value = 20.0 + float64(i%10) - 5.0
						unit = "celsius"
					case "humidity":
						value = 50.0 + float64(i%20) - 10.0
						unit = "percent"
					case "pressure":
						value = 1013.25 + float64(i%5) - 2.5
						unit = "hPa"
					}

					row := map[string]interface{}{
						"sensor_id":        sensorID,
						"timestamp":        timestamp,
						"measurement_type": measurementType,
						"value":            value,
						"unit":             unit,
						"quality":          100 - (i % 5), // Quality score
						"metadata":         fmt.Sprintf(`{"batch": %d, "source": "automated"}`, i/100),
					}

					err = writer.WriteRow(row)
					require.NoError(t, err)
					rowsWritten++
				}
			}
		}

		err = writer.Close()
		require.NoError(t, err)

		t.Logf("Wrote %d time series rows", rowsWritten)

		// Read back and verify
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, int64(rowsWritten), reader.GetRowCount())

		// Read a batch and verify structure
		batch, err := reader.ReadBatch(100)
		require.NoError(t, err)
		assert.Len(t, batch, 100)

		// Verify first row
		firstRow := batch[0]
		assert.Equal(t, "sensor_001", firstRow["sensor_id"])
		assert.NotNil(t, firstRow["timestamp"])
		assert.Contains(t, []string{"temperature", "humidity", "pressure"}, firstRow["measurement_type"])
		assert.NotNil(t, firstRow["value"])
	})
}

func TestIOTDeviceMetrics(t *testing.T) {
	tmpFile := "/tmp/test_iot_metrics.parquet"
	defer os.Remove(tmpFile)

	t.Run("IOT Device Metrics with Multiple Clustering Columns", func(t *testing.T) {
		// Complex time series with multiple clustering columns
		// Partition: (device_id, date)
		// Clustering: (hour, minute, metric_name)
		columns := []string{
			"device_id", "date", "hour", "minute", "metric_name",
			"metric_value", "tags", "alert_level",
		}
		types := []string{
			"text", "date", "int", "int", "text",
			"double", "text", "int",
		}

		options := parquet.DefaultWriterOptions()
		options.ChunkSize = 5000 // Optimize for time series

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, options)
		require.NoError(t, err)

		// Generate IOT metrics
		devices := []string{"iot_device_001", "iot_device_002"}
		metrics := []string{"cpu_usage", "memory_usage", "network_throughput", "disk_io", "battery_level"}
		baseDate := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)

		rowsWritten := 0

		for _, deviceID := range devices {
			// Simulate 24 hours of data
			for hour := 0; hour < 24; hour++ {
				// Data every minute
				for minute := 0; minute < 60; minute++ {
					for _, metricName := range metrics {
						var value float64
						var alertLevel int

						// Simulate realistic metric values
						switch metricName {
						case "cpu_usage":
							value = 30.0 + float64(hour*2) + float64(minute%10)
							switch {
							case value > 90:
								alertLevel = 3 // Critical
							case value > 80:
								alertLevel = 2 // Warning
							default:
								alertLevel = 1 // Normal
							}
						case "memory_usage":
							value = 40.0 + float64(hour) + float64(minute%5)*2
							if value > 85 {
								alertLevel = 2
							} else {
								alertLevel = 1
							}
						case "network_throughput":
							value = float64(minute*100 + hour*50)
							alertLevel = 1
						case "disk_io":
							value = float64(minute*10 + hour*5)
							alertLevel = 1
						case "battery_level":
							value = 100.0 - float64(hour*2) - float64(minute)/30.0
							switch {
							case value < 20:
								alertLevel = 3
							case value < 40:
								alertLevel = 2
							default:
								alertLevel = 1
							}
						}

						row := map[string]interface{}{
							"device_id":    deviceID,
							"date":         baseDate,
							"hour":         hour,
							"minute":       minute,
							"metric_name":  metricName,
							"metric_value": value,
							"tags":         fmt.Sprintf("location:zone_%d,env:production", hour%3),
							"alert_level":  alertLevel,
						}

						err = writer.WriteRow(row)
						require.NoError(t, err)
						rowsWritten++
					}
				}
			}
		}

		err = writer.Close()
		require.NoError(t, err)

		t.Logf("Wrote %d IOT metric rows", rowsWritten)

		// Verify the data
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, int64(rowsWritten), reader.GetRowCount())

		// Verify we can read back the metrics
		batch, err := reader.ReadBatch(500)
		require.NoError(t, err)
		assert.Len(t, batch, 500)

		// Check data integrity
		for _, row := range batch[:10] {
			assert.Contains(t, devices, row["device_id"])
			assert.Contains(t, metrics, row["metric_name"])
			assert.NotNil(t, row["metric_value"])
			assert.NotNil(t, row["alert_level"])
		}
	})
}

func TestEventLogData(t *testing.T) {
	tmpFile := "/tmp/test_event_log.parquet"
	defer os.Remove(tmpFile)

	t.Run("Event Log with Wide Partitions", func(t *testing.T) {
		// Event log schema common in Cassandra
		// Partition: (application, date_bucket)
		// Clustering: (timestamp, event_id)
		columns := []string{
			"application", "date_bucket", "timestamp", "event_id",
			"event_type", "severity", "message", "user_id",
			"session_id", "metadata",
		}
		types := []string{
			"text", "text", "timestamp", "uuid",
			"text", "text", "text", "text",
			"uuid", "text",
		}

		writer, err := parquet.NewParquetCaptureWriter(tmpFile, columns, types, parquet.DefaultWriterOptions())
		require.NoError(t, err)

		applications := []string{"web_app", "mobile_app", "api_service"}
		eventTypes := []string{"LOGIN", "LOGOUT", "ERROR", "WARNING", "INFO", "DEBUG"}
		severities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}

		baseTime := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		rowsWritten := 0

		// Generate event log data for each application
		for _, app := range applications {
			// Create events for 7 days
			for day := 0; day < 7; day++ {
				dateBucket := baseTime.AddDate(0, 0, day).Format("2006-01-02")

				// Generate 1000 events per day per application (wide partition)
				for i := 0; i < 1000; i++ {
					timestamp := baseTime.AddDate(0, 0, day).Add(time.Duration(i) * time.Minute)

					row := map[string]interface{}{
						"application":  app,
						"date_bucket":  dateBucket,
						"timestamp":    timestamp,
						"event_id":     fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", rowsWritten),
						"event_type":   eventTypes[i%len(eventTypes)],
						"severity":     severities[i%len(severities)],
						"message":      fmt.Sprintf("Event message for %s at index %d", app, i),
						"user_id":      fmt.Sprintf("user_%04d", i%100),
						"session_id":   fmt.Sprintf("6ba7b810-9dad-11d1-80b4-%012d", i),
						"metadata":     fmt.Sprintf(`{"ip": "192.168.1.%d", "browser": "Chrome", "version": "%d.0"}`, i%255, 90+i%10),
					}

					err = writer.WriteRow(row)
					require.NoError(t, err)
					rowsWritten++
				}
			}
		}

		err = writer.Close()
		require.NoError(t, err)

		t.Logf("Wrote %d event log rows", rowsWritten)

		// Verify the data
		reader, err := parquet.NewParquetReader(tmpFile)
		require.NoError(t, err)
		defer reader.Close()

		assert.Equal(t, int64(rowsWritten), reader.GetRowCount())

		// Read sample data
		batch, err := reader.ReadBatch(100)
		require.NoError(t, err)
		assert.Len(t, batch, 100)

		// Verify event structure
		for _, row := range batch[:5] {
			assert.Contains(t, applications, row["application"])
			assert.NotEmpty(t, row["date_bucket"])
			assert.NotNil(t, row["timestamp"])
			assert.Contains(t, eventTypes, row["event_type"])
			assert.Contains(t, severities, row["severity"])
			assert.NotEmpty(t, row["message"])
		}
	})
}