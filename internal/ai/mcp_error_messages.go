package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MCPErrorResponse represents a structured error response for Claude
type MCPErrorResponse struct {
	Error              string              `json:"error"`
	ErrorType          string              `json:"error_type"` // "permission_denied", "confirmation_required"
	CurrentMode        string              `json:"current_mode"`
	Operation          string              `json:"operation"`
	OperationCategory  string              `json:"operation_category"`
	ConfigurationHints *ConfigurationHints `json:"configuration_hints,omitempty"`
}

// ConfigurationHints provides suggestions for changing configuration
type ConfigurationHints struct {
	Message            string   `json:"message"`
	SuggestedModes     []string `json:"suggested_modes,omitempty"`
	SuggestedSkipList  []string `json:"suggested_skip_list,omitempty"`
	CanUpdateRuntime   bool     `json:"can_update_runtime"`
	UpdateCommandExample string `json:"update_command_example,omitempty"`
}

// CreatePermissionDeniedError creates a detailed error for operations not allowed
func CreatePermissionDeniedError(opInfo OperationInfo, config MCPConfigSnapshot) string {
	response := MCPErrorResponse{
		Error:             fmt.Sprintf("Operation not allowed in %s mode", formatModeName(config)),
		ErrorType:         "permission_denied",
		CurrentMode:       formatModeName(config),
		Operation:         opInfo.Operation,
		OperationCategory: string(opInfo.Category),
	}

	// Generate configuration hints
	hints := generatePermissionDeniedHints(opInfo, config)
	if hints != nil {
		response.ConfigurationHints = hints
	}

	// Format as JSON for Claude to parse, but also human-readable
	return formatErrorResponse(response)
}

// CreateConfirmationRequiredError creates a detailed error for operations needing confirmation
func CreateConfirmationRequiredError(opInfo OperationInfo, config MCPConfigSnapshot, requestID string) string {
	response := MCPErrorResponse{
		Error:             "Operation requires user confirmation",
		ErrorType:         "confirmation_required",
		CurrentMode:       formatModeName(config),
		Operation:         opInfo.Operation,
		OperationCategory: string(opInfo.Category),
	}

	// Generate configuration hints
	hints := generateConfirmationHints(opInfo, config, requestID)
	if hints != nil {
		response.ConfigurationHints = hints
	}

	return formatErrorResponse(response)
}

// generatePermissionDeniedHints suggests how to enable blocked operations
func generatePermissionDeniedHints(opInfo OperationInfo, config MCPConfigSnapshot) *ConfigurationHints {
	hints := &ConfigurationHints{
		CanUpdateRuntime: true,
	}

	if config.Mode == ConfigModePreset {
		// Suggest mode upgrades
		switch config.PresetMode {
		case "readonly":
			if opInfo.Category == CategoryDML || opInfo.Category == CategoryFILE {
				hints.Message = "This operation requires readwrite or dba mode"
				hints.SuggestedModes = []string{"readwrite", "dba"}
				hints.UpdateCommandExample = "Use update_mcp_permissions tool with mode='readwrite'"
			} else if opInfo.Category == CategoryDDL || opInfo.Category == CategoryDCL {
				hints.Message = "This operation requires dba mode"
				hints.SuggestedModes = []string{"dba"}
				hints.UpdateCommandExample = "Use update_mcp_permissions tool with mode='dba'"
			}

		case "readwrite":
			if opInfo.Category == CategoryDDL || opInfo.Category == CategoryDCL {
				hints.Message = "This operation requires dba mode"
				hints.SuggestedModes = []string{"dba"}
				hints.UpdateCommandExample = "Use update_mcp_permissions tool with mode='dba'"
			}
		}
	} else {
		// Fine-grained mode - suggest adding category to skip list
		hints.Message = fmt.Sprintf("Add %s to skip-confirmation list to allow this operation", opInfo.Category)
		hints.SuggestedSkipList = append(config.SkipConfirmation, string(opInfo.Category))
		hints.UpdateCommandExample = fmt.Sprintf("Use update_mcp_permissions tool with skip_confirmation='%s'",
			strings.Join(hints.SuggestedSkipList, ","))
	}

	return hints
}

// generateConfirmationHints suggests how to disable confirmations
func generateConfirmationHints(opInfo OperationInfo, config MCPConfigSnapshot, requestID string) *ConfigurationHints {
	hints := &ConfigurationHints{
		Message:          fmt.Sprintf("User must confirm this %s operation via REPL: .mcp confirm %s", opInfo.Category, requestID),
		CanUpdateRuntime: true,
	}

	if config.Mode == ConfigModePreset {
		// Suggest disabling confirmations for this category
		if len(config.ConfirmQueries) > 0 {
			hints.Message += "\n\nTo disable confirmations for " + string(opInfo.Category) + " operations:"
			hints.UpdateCommandExample = "Use update_mcp_permissions tool with confirm_queries='disable'"
		}
	} else {
		// Fine-grained mode - suggest adding to skip list
		hints.Message += "\n\nTo skip confirmations for " + string(opInfo.Category) + ":"
		newSkipList := append(config.SkipConfirmation, string(opInfo.Category))
		hints.SuggestedSkipList = newSkipList
		hints.UpdateCommandExample = fmt.Sprintf("Use update_mcp_permissions tool with skip_confirmation='%s'",
			strings.Join(newSkipList, ","))
	}

	return hints
}

// formatModeName returns a user-friendly mode name
func formatModeName(config MCPConfigSnapshot) string {
	if config.Mode == ConfigModePreset {
		overlay := ""
		if len(config.ConfirmQueries) > 0 {
			overlay = fmt.Sprintf(" (confirm: %s)", strings.Join(config.ConfirmQueries, ","))
		}
		return config.PresetMode + overlay
	}

	if len(config.SkipConfirmation) > 0 {
		return fmt.Sprintf("fine-grained (skip: %s)", strings.Join(config.SkipConfirmation, ","))
	}

	return "fine-grained"
}

// formatErrorResponse formats the error response as both JSON and human-readable
func formatErrorResponse(response MCPErrorResponse) string {
	// Try to marshal as JSON
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		// Fallback to simple string
		return response.Error
	}

	// Return JSON wrapped in a way that's both human and machine readable
	var sb strings.Builder
	sb.WriteString(response.Error)
	sb.WriteString("\n\n")

	if response.ConfigurationHints != nil {
		sb.WriteString(response.ConfigurationHints.Message)
		sb.WriteString("\n\n")

		if response.ConfigurationHints.UpdateCommandExample != "" {
			sb.WriteString("Suggestion: ")
			sb.WriteString(response.ConfigurationHints.UpdateCommandExample)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("---\nStructured error (for programmatic access):\n")
	sb.WriteString(string(jsonBytes))

	return sb.String()
}
