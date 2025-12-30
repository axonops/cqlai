package ai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCreatePermissionDeniedError tests permission denied error generation
func TestCreatePermissionDeniedError(t *testing.T) {
	tests := []struct {
		name              string
		opInfo            OperationInfo
		config            MCPConfigSnapshot
		expectContains    []string
		expectSuggestions bool
	}{
		{
			name: "INSERT blocked in readonly - suggests readwrite",
			opInfo: OperationInfo{
				Category:    CategoryDML,
				Operation:   "INSERT",
				Description: "Insert data",
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readonly",
			},
			expectContains:    []string{"not allowed", "readonly", "INSERT", "readwrite", "dba"},
			expectSuggestions: true,
		},
		{
			name: "CREATE TABLE blocked in readwrite - suggests dba",
			opInfo: OperationInfo{
				Category:    CategoryDDL,
				Operation:   "CREATE TABLE",
				Description: "Create table",
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readwrite",
			},
			expectContains:    []string{"not allowed", "readwrite", "dba"},
			expectSuggestions: true,
		},
		{
			name: "DROP blocked in readwrite - suggests dba",
			opInfo: OperationInfo{
				Category:  CategoryDDL,
				Operation: "DROP TABLE",
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readwrite",
			},
			expectContains:    []string{"dba"},
			expectSuggestions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := CreatePermissionDeniedError(tt.opInfo, tt.config)

			// Check error contains expected strings
			for _, expectedStr := range tt.expectContains {
				assert.Contains(t, errorMsg, expectedStr,
					"Error should contain '%s'", expectedStr)
			}

			// Check JSON structure is present
			assert.Contains(t, errorMsg, "Structured error",
				"Error should include structured JSON")

			// Verify JSON is parseable
			if strings.Contains(errorMsg, "{") {
				jsonStart := strings.Index(errorMsg, "{")
				jsonPart := errorMsg[jsonStart:]
				var response MCPErrorResponse
				err := json.Unmarshal([]byte(jsonPart), &response)
				if assert.NoError(t, err, "JSON should be parseable") {
					assert.Equal(t, "permission_denied", response.ErrorType)
					assert.Equal(t, string(tt.opInfo.Category), response.OperationCategory)

					if tt.expectSuggestions {
						assert.NotNil(t, response.ConfigurationHints,
							"Should include configuration hints")
						assert.NotEmpty(t, response.ConfigurationHints.Message)
					}
				}
			}
		})
	}
}

// TestCreateConfirmationRequiredError tests confirmation required error generation
func TestCreateConfirmationRequiredError(t *testing.T) {
	tests := []struct {
		name           string
		opInfo         OperationInfo
		config         MCPConfigSnapshot
		requestID      string
		expectContains []string
	}{
		{
			name: "DELETE requires confirmation in dangerous_only mode",
			opInfo: OperationInfo{
				Category:  CategoryDML,
				Operation: "DELETE",
			},
			config: MCPConfigSnapshot{
				Mode:           ConfigModePreset,
				PresetMode:     "dba",
				ConfirmQueries: []string{"dml"},
			},
			requestID:      "req_001",
			expectContains: []string{"requires user confirmation", "req_001", "dml", "disable"},
		},
		{
			name: "GRANT requires confirmation with dcl in confirm-queries",
			opInfo: OperationInfo{
				Category:  CategoryDCL,
				Operation: "GRANT",
			},
			config: MCPConfigSnapshot{
				Mode:           ConfigModePreset,
				PresetMode:     "dba",
				ConfirmQueries: []string{"dcl"},
			},
			requestID:      "req_002",
			expectContains: []string{"requires user confirmation", "req_002", "disable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorMsg := CreateConfirmationRequiredError(tt.opInfo, tt.config, tt.requestID)

			// Check error contains expected strings
			for _, expectedStr := range tt.expectContains {
				assert.Contains(t, errorMsg, expectedStr,
					"Error should contain '%s'", expectedStr)
			}

			// Verify JSON structure
			if strings.Contains(errorMsg, "{") {
				jsonStart := strings.Index(errorMsg, "{")
				jsonPart := errorMsg[jsonStart:]
				var response MCPErrorResponse
				err := json.Unmarshal([]byte(jsonPart), &response)
				if assert.NoError(t, err) {
					assert.Equal(t, "confirmation_required", response.ErrorType)
					assert.NotNil(t, response.ConfigurationHints)
				}
			}
		})
	}
}

// TestGeneratePermissionDeniedHints tests hint generation for permission denials
func TestGeneratePermissionDeniedHints(t *testing.T) {
	tests := []struct {
		name               string
		opInfo             OperationInfo
		config             MCPConfigSnapshot
		expectSuggestModes bool
		expectedModes      []string
	}{
		{
			name: "DML in readonly suggests readwrite or dba",
			opInfo: OperationInfo{
				Category: CategoryDML,
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readonly",
			},
			expectSuggestModes: true,
			expectedModes:      []string{"readwrite", "dba"},
		},
		{
			name: "DDL in readwrite suggests dba",
			opInfo: OperationInfo{
				Category: CategoryDDL,
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readwrite",
			},
			expectSuggestModes: true,
			expectedModes:      []string{"dba"},
		},
		{
			name: "DCL in readwrite suggests dba",
			opInfo: OperationInfo{
				Category: CategoryDCL,
			},
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readwrite",
			},
			expectSuggestModes: true,
			expectedModes:      []string{"dba"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := generatePermissionDeniedHints(tt.opInfo, tt.config)

			assert.NotNil(t, hints)
			assert.True(t, hints.CanUpdateRuntime)
			assert.NotEmpty(t, hints.Message)

			if tt.expectSuggestModes {
				assert.Equal(t, tt.expectedModes, hints.SuggestedModes)
				assert.NotEmpty(t, hints.UpdateCommandExample)
			}
		})
	}
}

// TestGenerateConfirmationHints tests hint generation for confirmation required
func TestGenerateConfirmationHints(t *testing.T) {
	tests := []struct {
		name              string
		opInfo            OperationInfo
		config            MCPConfigSnapshot
		requestID         string
		expectContains    []string
		expectSuggestions bool
	}{
		{
			name: "DML confirmation with hint to disable",
			opInfo: OperationInfo{
				Category: CategoryDML,
			},
			config: MCPConfigSnapshot{
				Mode:           ConfigModePreset,
				PresetMode:     "dba",
				ConfirmQueries: []string{"dml"},
			},
			requestID:         "req_001",
			expectContains:    []string{"req_001", "disable", "dml"},
			expectSuggestions: true,
		},
		{
			name: "DCL confirmation with disable hint",
			opInfo: OperationInfo{
				Category: CategoryDCL,
			},
			config: MCPConfigSnapshot{
				Mode:           ConfigModePreset,
				PresetMode:     "dba",
				ConfirmQueries: []string{"dcl", "ddl"},
			},
			requestID:         "req_002",
			expectContains:    []string{"req_002", "disable", "dcl"},
			expectSuggestions: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := generateConfirmationHints(tt.opInfo, tt.config, tt.requestID)

			assert.NotNil(t, hints)
			assert.True(t, hints.CanUpdateRuntime)
			assert.NotEmpty(t, hints.Message)

			for _, expectedStr := range tt.expectContains {
				assert.Contains(t, hints.Message, expectedStr)
			}

			if tt.expectSuggestions {
				assert.NotEmpty(t, hints.UpdateCommandExample)
			}
		})
	}
}

// TestFormatModeName tests mode name formatting
func TestFormatModeName(t *testing.T) {
	tests := []struct {
		name     string
		config   MCPConfigSnapshot
		expected string
	}{
		{
			name: "readonly preset",
			config: MCPConfigSnapshot{
				Mode:       ConfigModePreset,
				PresetMode: "readonly",
			},
			expected: "readonly",
		},
		{
			name: "readwrite with confirm overlay",
			config: MCPConfigSnapshot{
				Mode:           ConfigModePreset,
				PresetMode:     "readwrite",
				ConfirmQueries: []string{"dml"},
			},
			expected: "readwrite (confirm: dml)",
		},
		{
			name: "fine-grained with skip list",
			config: MCPConfigSnapshot{
				Mode:             ConfigModeFineGrained,
				SkipConfirmation: []string{"dql", "dml"},
			},
			expected: "fine-grained (skip: dql,dml)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatModeName(tt.config)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatErrorResponse tests error response formatting
func TestFormatErrorResponse(t *testing.T) {
	response := MCPErrorResponse{
		Error:             "Test error",
		ErrorType:         "permission_denied",
		CurrentMode:       "readonly",
		Operation:         "INSERT",
		OperationCategory: "dml",
		ConfigurationHints: &ConfigurationHints{
			Message:              "This operation requires readwrite mode",
			SuggestedModes:       []string{"readwrite", "dba"},
			CanUpdateRuntime:     true,
			UpdateCommandExample: "Use update_mcp_config tool with mode='readwrite'",
		},
	}

	result := formatErrorResponse(response)

	// Check all expected sections present
	assert.Contains(t, result, "Test error")
	assert.Contains(t, result, "This operation requires readwrite mode")
	assert.Contains(t, result, "update_mcp_config")
	assert.Contains(t, result, "Structured error")

	// Verify JSON is parseable
	jsonStart := strings.Index(result, "{")
	if jsonStart >= 0 {
		jsonPart := result[jsonStart:]
		var parsed MCPErrorResponse
		err := json.Unmarshal([]byte(jsonPart), &parsed)
		assert.NoError(t, err, "JSON should be valid")
		assert.Equal(t, response.Error, parsed.Error)
		assert.Equal(t, response.ErrorType, parsed.ErrorType)
	}
}

// TestErrorMessageStructure tests that all error messages are valid JSON
func TestErrorMessageStructure(t *testing.T) {
	scenarios := []struct {
		name   string
		opInfo OperationInfo
		config MCPConfigSnapshot
	}{
		{
			name:   "INSERT in readonly",
			opInfo: OperationInfo{Category: CategoryDML, Operation: "INSERT"},
			config: MCPConfigSnapshot{Mode: ConfigModePreset, PresetMode: "readonly"},
		},
		{
			name:   "CREATE in readwrite",
			opInfo: OperationInfo{Category: CategoryDDL, Operation: "CREATE TABLE"},
			config: MCPConfigSnapshot{Mode: ConfigModePreset, PresetMode: "readwrite"},
		},
		{
			name:   "GRANT in readwrite",
			opInfo: OperationInfo{Category: CategoryDCL, Operation: "GRANT"},
			config: MCPConfigSnapshot{Mode: ConfigModePreset, PresetMode: "readwrite"},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			errorMsg := CreatePermissionDeniedError(scenario.opInfo, scenario.config)

			// Must contain structured error
			assert.Contains(t, errorMsg, "Structured error")

			// Extract and parse JSON
			jsonStart := strings.Index(errorMsg, "{")
			assert.GreaterOrEqual(t, jsonStart, 0, "Must contain JSON")

			jsonPart := errorMsg[jsonStart:]
			var response MCPErrorResponse
			err := json.Unmarshal([]byte(jsonPart), &response)
			assert.NoError(t, err, "JSON must be valid")

			// Verify required fields
			assert.NotEmpty(t, response.Error)
			assert.Equal(t, "permission_denied", response.ErrorType)
			assert.NotEmpty(t, response.CurrentMode)
			assert.Equal(t, scenario.opInfo.Operation, response.Operation)
			assert.Equal(t, string(scenario.opInfo.Category), response.OperationCategory)
		})
	}
}

// TestConfirmationErrorStructure tests confirmation required errors
func TestConfirmationErrorStructure(t *testing.T) {
	opInfo := OperationInfo{
		Category:  CategoryDML,
		Operation: "DELETE",
	}

	config := MCPConfigSnapshot{
		Mode:           ConfigModePreset,
		PresetMode:     "dba",
		ConfirmQueries: []string{"dml"},
	}

	errorMsg := CreateConfirmationRequiredError(opInfo, config, "req_001")

	// Verify structure
	assert.Contains(t, errorMsg, "requires user confirmation")
	assert.Contains(t, errorMsg, "req_001")
	assert.Contains(t, errorMsg, "Structured error")

	// Parse JSON
	jsonStart := strings.Index(errorMsg, "{")
	jsonPart := errorMsg[jsonStart:]
	var response MCPErrorResponse
	err := json.Unmarshal([]byte(jsonPart), &response)

	assert.NoError(t, err)
	assert.Equal(t, "confirmation_required", response.ErrorType)
	assert.NotNil(t, response.ConfigurationHints)
	assert.Contains(t, response.ConfigurationHints.Message, "req_001")
}
