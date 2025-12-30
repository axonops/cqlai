package ai

import (
	"fmt"
	"strings"
)

// CheckOperationPermission determines if an operation is allowed and if it needs confirmation
// This method is thread-safe and works with both preset and fine-grained modes
func (s *MCPServer) CheckOperationPermission(opInfo OperationInfo) (allowed bool, needsConfirmation bool, reason string) {
	// Get thread-safe snapshot of config
	snapshot := s.config.GetConfigSnapshot()

	// SESSION operations always allowed, never confirmed
	if opInfo.Category == CategorySESSION {
		return true, false, ""
	}

	// Delegate to mode-specific logic
	if snapshot.Mode == ConfigModePreset {
		return checkPresetModePermission(snapshot, opInfo)
	} else if snapshot.Mode == ConfigModeFineGrained {
		return checkFineGrainedPermission(snapshot, opInfo)
	}

	// Unknown mode - deny by default for safety
	return false, false, fmt.Sprintf("unknown configuration mode: %s", snapshot.Mode)
}

// checkPresetModePermission handles permission checking for preset modes
func checkPresetModePermission(snapshot MCPConfigSnapshot, opInfo OperationInfo) (allowed bool, needsConfirmation bool, reason string) {
	// Determine what's allowed in this preset mode
	allowedCategories := getAllowedCategoriesForPresetMode(snapshot.PresetMode)

	// Special handling for FILE category in readonly mode
	if snapshot.PresetMode == "readonly" && opInfo.Category == CategoryFILE {
		// Only COPY TO (export) allowed in readonly
		if opInfo.Operation == "COPY TO" {
			needsConf := shouldConfirmInPresetMode(snapshot, opInfo.Category)
			return true, needsConf, ""
		}
		// COPY FROM and SOURCE not allowed in readonly
		return false, false, fmt.Sprintf("operation %s not allowed in readonly mode (only COPY TO allowed)", opInfo.Operation)
	}

	// Check if operation's category is allowed
	if !containsCategory(allowedCategories, string(opInfo.Category)) {
		return false, false, fmt.Sprintf("operation %s not allowed in %s mode", opInfo.Operation, snapshot.PresetMode)
	}

	// Operation is allowed - check if it needs confirmation
	needsConf := shouldConfirmInPresetMode(snapshot, opInfo.Category)

	return true, needsConf, ""
}

// checkFineGrainedPermission handles permission checking for fine-grained mode
func checkFineGrainedPermission(snapshot MCPConfigSnapshot, opInfo OperationInfo) (allowed bool, needsConfirmation bool, reason string) {
	// In fine-grained mode, all operations are allowed (no blocking)
	// Only question is whether confirmation is needed

	// Check if "ALL" is in skip list
	if containsCategory(snapshot.SkipConfirmation, "ALL") {
		return true, false, "" // Skip confirmation on everything
	}

	// Check if this operation's category is in the skip list
	if containsCategory(snapshot.SkipConfirmation, string(opInfo.Category)) {
		return true, false, "" // Skip confirmation for this category
	}

	// Not in skip list - requires confirmation
	return true, true, ""
}

// getAllowedCategoriesForPresetMode returns which categories are allowed in a preset mode
func getAllowedCategoriesForPresetMode(presetMode string) []string {
	switch presetMode {
	case "readonly":
		// DQL: Queries
		// SESSION: Always allowed (handled separately)
		// FILE Export: COPY TO only
		return []string{"dql", "session", "file_export"}

	case "readwrite":
		// DQL: Queries
		// SESSION: Always allowed
		// DML: Data modifications
		// FILE: All file operations (COPY TO/FROM, SOURCE)
		return []string{"dql", "session", "dml", "file"}

	case "dba":
		// Everything allowed
		return []string{"dql", "session", "dml", "ddl", "dcl", "file"}

	default:
		// Unknown mode - allow nothing (safe default)
		return []string{"session"} // Only session always allowed
	}
}

// shouldConfirmInPresetMode determines if confirmation is needed based on confirm-queries overlay
func shouldConfirmInPresetMode(snapshot MCPConfigSnapshot, category OperationCategory) bool {
	// No confirm-queries overlay - no confirmation needed
	if len(snapshot.ConfirmQueries) == 0 {
		return false
	}

	// Check for special values
	if containsCategory(snapshot.ConfirmQueries, "ALL") {
		return true // Confirm everything
	}

	if containsCategory(snapshot.ConfirmQueries, "none") ||
		containsCategory(snapshot.ConfirmQueries, "disable") {
		return false // No confirmations
	}

	// Check if this category is in the confirm list
	return containsCategory(snapshot.ConfirmQueries, string(category))
}

// IsFileExportOperation checks if an operation is a file export (allowed in readonly)
func IsFileExportOperation(command string) bool {
	cmd := strings.ToUpper(strings.TrimSpace(command))
	// COPY TO is export (safe in readonly)
	return strings.Contains(cmd, "COPY") && strings.Contains(cmd, " TO ")
}

// IsFileImportOperation checks if an operation is a file import (not allowed in readonly)
func IsFileImportOperation(command string) bool {
	cmd := strings.ToUpper(strings.TrimSpace(command))
	// COPY FROM or SOURCE are imports (not safe in readonly)
	if strings.Contains(cmd, "COPY") && strings.Contains(cmd, " FROM ") {
		return true
	}
	if strings.HasPrefix(cmd, "SOURCE") {
		return true
	}
	return false
}

// ValidatePresetMode validates a preset mode name
func ValidatePresetMode(mode string) error {
	validModes := map[string]bool{
		"readonly":  true,
		"readwrite": true,
		"dba":       true,
	}

	if !validModes[mode] {
		return fmt.Errorf("invalid preset mode: %s (valid: readonly, readwrite, dba)", mode)
	}

	return nil
}

// ValidateCategory validates a category name
func ValidateCategory(category string) error {
	validCategories := map[string]bool{
		"dql": true, "session": true, "dml": true,
		"ddl": true, "dcl": true, "file": true,
		"all": true, "none": true, "disable": true,
	}

	if !validCategories[strings.ToLower(category)] {
		return fmt.Errorf("invalid category: %s (valid: dql, session, dml, ddl, dcl, file, ALL, none, disable)", category)
	}

	return nil
}

// ValidateCategoryList validates a list of categories
func ValidateCategoryList(categories []string) error {
	for _, cat := range categories {
		if err := ValidateCategory(cat); err != nil {
			return err
		}
	}
	return nil
}

// FormatConfigForDisplay formats the current configuration for status display
func (c *MCPServerConfig) FormatConfigForDisplay() string {
	snapshot := c.GetConfigSnapshot()

	var sb strings.Builder

	if snapshot.Mode == ConfigModePreset {
		sb.WriteString(fmt.Sprintf("  Mode: %s (preset)\n", snapshot.PresetMode))

		// Always show runtime permission changes setting
		if snapshot.DisableRuntimePermissionChanges {
			sb.WriteString("  Runtime permission changes: DISABLED (locked at startup)\n")
		} else {
			sb.WriteString("  Runtime permission changes: ENABLED\n")
		}

		// Show confirm-queries if any
		if len(snapshot.ConfirmQueries) > 0 {
			if containsCategory(snapshot.ConfirmQueries, "ALL") {
				sb.WriteString("  Confirm-queries: ALL (confirm everything)\n")
			} else if containsCategory(snapshot.ConfirmQueries, "none") || containsCategory(snapshot.ConfirmQueries, "disable") {
				sb.WriteString("  Confirm-queries: none (disabled)\n")
			} else {
				sb.WriteString(fmt.Sprintf("  Confirm-queries: %s\n", strings.Join(snapshot.ConfirmQueries, ", ")))
			}
		}

		// Show what's allowed
		sb.WriteString("\n")
		sb.WriteString(formatPresetModePermissions(snapshot))

	} else if snapshot.Mode == ConfigModeFineGrained {
		sb.WriteString("  Mode: fine-grained\n")

		// Always show runtime permission changes setting
		if snapshot.DisableRuntimePermissionChanges {
			sb.WriteString("  Runtime permission changes: DISABLED (locked at startup)\n")
		} else {
			sb.WriteString("  Runtime permission changes: ENABLED\n")
		}

		if containsCategory(snapshot.SkipConfirmation, "ALL") {
			sb.WriteString("  Skip confirmation: ALL (no confirmations)\n")
		} else {
			sb.WriteString(fmt.Sprintf("  Skip confirmation: %s\n", strings.Join(snapshot.SkipConfirmation, ", ")))
		}

		sb.WriteString("\n")
		sb.WriteString(formatFineGrainedPermissions(snapshot))
	}

	return sb.String()
}

// formatPresetModePermissions formats permissions for preset modes
func formatPresetModePermissions(snapshot MCPConfigSnapshot) string {
	var sb strings.Builder

	// Determine what categories are allowed and if they need confirmation
	allowedCategories := getAllowedCategoriesForPresetMode(snapshot.PresetMode)

	// Build allowed operations list
	sb.WriteString("Allowed Operations:\n")

	if containsCategory(allowedCategories, "dql") {
		needsConf := shouldConfirmInPresetMode(snapshot, CategoryDQL)
		if needsConf {
			sb.WriteString("  DQL (14 ops): Queries - REQUIRES CONFIRMATION\n")
		} else {
			sb.WriteString("  DQL (14 ops): Queries - no confirmation\n")
		}
	}

	sb.WriteString("  SESSION (8 ops): Settings - no confirmation (always)\n")

	if containsCategory(allowedCategories, "dml") {
		needsConf := shouldConfirmInPresetMode(snapshot, CategoryDML)
		if needsConf {
			sb.WriteString("  DML (8 ops): Data modifications - REQUIRES CONFIRMATION\n")
		} else {
			sb.WriteString("  DML (8 ops): Data modifications - no confirmation\n")
		}
	}

	if containsCategory(allowedCategories, "ddl") {
		needsConf := shouldConfirmInPresetMode(snapshot, CategoryDDL)
		if needsConf {
			sb.WriteString("  DDL (28 ops): Schema changes - REQUIRES CONFIRMATION\n")
		} else {
			sb.WriteString("  DDL (28 ops): Schema changes - no confirmation\n")
		}
	}

	if containsCategory(allowedCategories, "dcl") {
		needsConf := shouldConfirmInPresetMode(snapshot, CategoryDCL)
		if needsConf {
			sb.WriteString("  DCL (13 ops): Security - REQUIRES CONFIRMATION\n")
		} else {
			sb.WriteString("  DCL (13 ops): Security - no confirmation\n")
		}
	}

	if containsCategory(allowedCategories, "file") || containsCategory(allowedCategories, "file_export") {
		needsConf := shouldConfirmInPresetMode(snapshot, CategoryFILE)
		if containsCategory(allowedCategories, "file_export") {
			sb.WriteString("  FILE (1 op): Export only (COPY TO) - no confirmation\n")
		} else if needsConf {
			sb.WriteString("  FILE (3 ops): COPY TO/FROM, SOURCE - REQUIRES CONFIRMATION\n")
		} else {
			sb.WriteString("  FILE (3 ops): COPY TO/FROM, SOURCE - no confirmation\n")
		}
	}

	// Show what's NOT allowed
	sb.WriteString("\nNot Allowed:\n")
	notAllowed := []string{}

	if !containsCategory(allowedCategories, "dml") {
		notAllowed = append(notAllowed, "DML (8 ops): Data modifications")
	}
	if !containsCategory(allowedCategories, "ddl") {
		notAllowed = append(notAllowed, "DDL (28 ops): Schema changes")
	}
	if !containsCategory(allowedCategories, "dcl") {
		notAllowed = append(notAllowed, "DCL (13 ops): Security")
	}
	if !containsCategory(allowedCategories, "file") && !containsCategory(allowedCategories, "file_export") {
		notAllowed = append(notAllowed, "FILE (3 ops): Import/export")
	} else if containsCategory(allowedCategories, "file_export") {
		notAllowed = append(notAllowed, "FILE Import (2 ops): COPY FROM, SOURCE")
	}

	if len(notAllowed) == 0 {
		sb.WriteString("  (none - all operations allowed)\n")
	} else {
		for _, item := range notAllowed {
			sb.WriteString("  " + item + "\n")
		}
	}

	return sb.String()
}

// formatFineGrainedPermissions formats permissions for fine-grained mode
func formatFineGrainedPermissions(snapshot MCPConfigSnapshot) string {
	var sb strings.Builder

	if containsCategory(snapshot.SkipConfirmation, "ALL") {
		sb.WriteString("All Operations:\n")
		sb.WriteString("  DQL, SESSION, DML, DDL, DCL, FILE - no confirmation on anything\n")
		return sb.String()
	}

	sb.WriteString("Skip Confirmation:\n")
	for _, cat := range snapshot.SkipConfirmation {
		catType := OperationCategory(cat)
		count := GetCategoryOperationCount(catType)
		sb.WriteString(fmt.Sprintf("  %s (%d ops): %s\n",
			strings.ToUpper(cat), count, GetCategoryDescription(catType)))
	}

	sb.WriteString("\nRequire Confirmation:\n")
	allCategories := []string{"dql", "session", "dml", "ddl", "dcl", "file"}
	for _, cat := range allCategories {
		if !containsCategory(snapshot.SkipConfirmation, cat) {
			catType := OperationCategory(cat)
			count := GetCategoryOperationCount(catType)
			sb.WriteString(fmt.Sprintf("  %s (%d ops): %s\n",
				strings.ToUpper(cat), count, GetCategoryDescription(catType)))
		}
	}

	return sb.String()
}
