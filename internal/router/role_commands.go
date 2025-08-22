package router

import (
	"fmt"
	"strings"
	
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// VisitCreateUser handles CREATE USER statements
func (v *CqlCommandVisitorImpl) VisitCreateUser(ctx *grammar.CreateUserContext) interface{} {
	// Get the full text of the CREATE USER statement
	query := ctx.GetText()
	
	// Execute the CREATE USER query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("CREATE USER failed: %v", err)
	}
	
	// Extract username for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	username := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "USER" && i+1 < len(parts) {
			username = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if username != "" {
		return fmt.Sprintf("User '%s' created successfully", username)
	}
	return "CREATE USER successful"
}

// VisitCreateRole handles CREATE ROLE statements
func (v *CqlCommandVisitorImpl) VisitCreateRole(ctx *grammar.CreateRoleContext) interface{} {
	// Get the full text of the CREATE ROLE statement
	query := ctx.GetText()
	
	// Execute the CREATE ROLE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("CREATE ROLE failed: %v", err)
	}
	
	// Extract role name for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	roleName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "ROLE" && i+1 < len(parts) {
			roleName = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if roleName != "" {
		return fmt.Sprintf("Role '%s' created successfully", roleName)
	}
	return "CREATE ROLE successful"
}

// VisitAlterUser handles ALTER USER statements
func (v *CqlCommandVisitorImpl) VisitAlterUser(ctx *grammar.AlterUserContext) interface{} {
	// Get the full text of the ALTER USER statement
	query := ctx.GetText()
	
	// Execute the ALTER USER query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("ALTER USER failed: %v", err)
	}
	
	// Extract username for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	username := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "USER" && i+1 < len(parts) {
			username = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if username != "" {
		return fmt.Sprintf("User '%s' altered successfully", username)
	}
	return "ALTER USER successful"
}

// VisitAlterRole handles ALTER ROLE statements
func (v *CqlCommandVisitorImpl) VisitAlterRole(ctx *grammar.AlterRoleContext) interface{} {
	// Get the full text of the ALTER ROLE statement
	query := ctx.GetText()
	
	// Execute the ALTER ROLE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("ALTER ROLE failed: %v", err)
	}
	
	// Extract role name for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	roleName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "ROLE" && i+1 < len(parts) {
			roleName = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if roleName != "" {
		return fmt.Sprintf("Role '%s' altered successfully", roleName)
	}
	return "ALTER ROLE successful"
}

// VisitDropUser handles DROP USER statements
func (v *CqlCommandVisitorImpl) VisitDropUser(ctx *grammar.DropUserContext) interface{} {
	// Get the full text of the DROP USER statement
	query := ctx.GetText()
	
	// Execute the DROP USER query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("DROP USER failed: %v", err)
	}
	
	// Extract username for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	username := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "USER" && i+1 < len(parts) {
			username = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if username != "" {
		return fmt.Sprintf("User '%s' dropped successfully", username)
	}
	return "DROP USER successful"
}

// VisitDropRole handles DROP ROLE statements
func (v *CqlCommandVisitorImpl) VisitDropRole(ctx *grammar.DropRoleContext) interface{} {
	// Get the full text of the DROP ROLE statement
	query := ctx.GetText()
	
	// Execute the DROP ROLE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("DROP ROLE failed: %v", err)
	}
	
	// Extract role name for success message
	text := ctx.GetText()
	parts := strings.Fields(text)
	roleName := ""
	for i, part := range parts {
		if strings.ToUpper(part) == "ROLE" && i+1 < len(parts) {
			roleName = strings.Trim(parts[i+1], "'\"")
			break
		}
	}
	
	if roleName != "" {
		return fmt.Sprintf("Role '%s' dropped successfully", roleName)
	}
	return "DROP ROLE successful"
}

// VisitGrant handles GRANT statements
func (v *CqlCommandVisitorImpl) VisitGrant(ctx *grammar.GrantContext) interface{} {
	// Get the full text of the GRANT statement
	query := ctx.GetText()
	
	// Execute the GRANT query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("GRANT failed: %v", err)
	}
	
	return "GRANT successful"
}

// VisitRevoke handles REVOKE statements
func (v *CqlCommandVisitorImpl) VisitRevoke(ctx *grammar.RevokeContext) interface{} {
	// Get the full text of the REVOKE statement
	query := ctx.GetText()
	
	// Execute the REVOKE query
	if err := v.session.Query(query).Exec(); err != nil {
		return fmt.Errorf("REVOKE failed: %v", err)
	}
	
	return "REVOKE successful"
}

// VisitListRoles handles LIST ROLES statements
func (v *CqlCommandVisitorImpl) VisitListRoles(ctx *grammar.ListRolesContext) interface{} {
	// Query system_auth.roles for role information
	query := "SELECT role, can_login, is_superuser, member_of, salted_hash FROM system_auth.roles"
	
	iter := v.session.Query(query).Iter()
	
	// Get column info
	columns := iter.Columns()
	if len(columns) == 0 {
		if err := iter.Close(); err != nil {
			return fmt.Errorf("LIST ROLES failed: %v", err)
		}
		return "No roles found"
	}
	
	// Prepare headers
	headers := []string{"Role", "Login", "Superuser", "Member Of"}
	
	// Collect results
	results := [][]string{headers}
	
	var role string
	var canLogin, isSuperuser bool
	var memberOf []string
	var saltedHash string
	
	for iter.Scan(&role, &canLogin, &isSuperuser, &memberOf, &saltedHash) {
		memberOfStr := "None"
		if len(memberOf) > 0 {
			memberOfStr = strings.Join(memberOf, ", ")
		}
		
		row := []string{
			role,
			fmt.Sprintf("%v", canLogin),
			fmt.Sprintf("%v", isSuperuser),
			memberOfStr,
		}
		results = append(results, row)
	}
	
	if err := iter.Close(); err != nil {
		return fmt.Errorf("LIST ROLES failed: %v", err)
	}
	
	if len(results) == 1 {
		return "No roles found"
	}
	
	return results
}

// VisitListPermissions handles LIST PERMISSIONS statements
func (v *CqlCommandVisitorImpl) VisitListPermissions(ctx *grammar.ListPermissionsContext) interface{} {
	// Parse the command to see if it's for a specific role/user
	text := ctx.GetText()
	parts := strings.Fields(strings.ToUpper(text))
	
	var query string
	if len(parts) >= 3 {
		// LIST PERMISSIONS OF <role>
		for i, part := range parts {
			if part == "OF" && i+1 < len(parts) {
				role := strings.Trim(parts[i+1], "'\"")
				query = fmt.Sprintf("SELECT role, resource, permissions FROM system_auth.role_permissions WHERE role = '%s'", role)
				break
			}
		}
	}
	
	if query == "" {
		// List all permissions
		query = "SELECT role, resource, permissions FROM system_auth.role_permissions"
	}
	
	iter := v.session.Query(query).Iter()
	
	// Get column info
	columns := iter.Columns()
	if len(columns) == 0 {
		if err := iter.Close(); err != nil {
			return fmt.Errorf("LIST PERMISSIONS failed: %v", err)
		}
		return "No permissions found"
	}
	
	// Prepare headers
	headers := []string{"Role", "Resource", "Permissions"}
	
	// Collect results
	results := [][]string{headers}
	
	var role, resource string
	var permissions []string
	
	for iter.Scan(&role, &resource, &permissions) {
		permStr := "None"
		if len(permissions) > 0 {
			permStr = strings.Join(permissions, ", ")
		}
		
		row := []string{
			role,
			resource,
			permStr,
		}
		results = append(results, row)
	}
	
	if err := iter.Close(); err != nil {
		return fmt.Errorf("LIST PERMISSIONS failed: %v", err)
	}
	
	if len(results) == 1 {
		return "No permissions found"
	}
	
	return results
}