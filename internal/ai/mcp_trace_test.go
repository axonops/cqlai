package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseHexTraceID tests hex trace ID parsing
func TestParseHexTraceID(t *testing.T) {
	tests := []struct {
		name    string
		hexStr  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid UUID hex (32 chars)",
			hexStr:  "550e8400e29b41d4a716446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID hex with dashes (gets parsed after removing dashes elsewhere)",
			hexStr:  "550e8400e29b41d4a716446655440000",
			wantErr: false,
		},
		{
			name:    "invalid hex (non-hex characters)",
			hexStr:  "zzze8400e29b41d4a716446655440000",
			wantErr: true,
			errMsg:  "invalid hex",
		},
		{
			name:    "wrong length (too short)",
			hexStr:  "550e8400e29b41d4a716",
			wantErr: true,
			errMsg:  "must be 16 bytes",
		},
		{
			name:    "wrong length (too long)",
			hexStr:  "550e8400e29b41d4a716446655440000550e8400",
			wantErr: true,
			errMsg:  "must be 16 bytes",
		},
		{
			name:    "empty string",
			hexStr:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHexTraceID(tt.hexStr)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, 16, len(got), "UUID should be 16 bytes")
			}
		})
	}
}

// TestParseHexTraceID_ValidUUIDs tests various valid UUID formats
func TestParseHexTraceID_ValidUUIDs(t *testing.T) {
	validUUIDs := []string{
		"00000000000000000000000000000000", // All zeros
		"ffffffffffffffffffffffffffffffff", // All Fs
		"550e8400e29b41d4a716446655440000", // Random UUID
	}

	for _, hexStr := range validUUIDs {
		t.Run("valid_"+hexStr[:8], func(t *testing.T) {
			got, err := parseHexTraceID(hexStr)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, 16, len(got))
		})
	}
}

// Note: getTraceDataByID requires a live Cassandra connection with trace data,
// so it's tested via integration tests, not unit tests here.
// The integration tests verify:
// 1. Trace ID is captured from query execution
// 2. get_trace_data tool accepts the trace ID
// 3. Trace events are retrieved from system_traces
// 4. Response contains coordinator, duration, and event list
