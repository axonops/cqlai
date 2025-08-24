package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// VisitCreateUser handles CREATE USER statements
func (v *CqlCommandVisitorImpl) VisitCreateUser(ctx *grammar.CreateUserContext) interface{} {
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("CREATE USER failed: %v", err)
	}

	// Extract username for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("CREATE ROLE failed: %v", err)
	}

	// Extract role name for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("ALTER USER failed: %v", err)
	}

	// Extract username for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("ALTER ROLE failed: %v", err)
	}

	// Extract role name for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("DROP USER failed: %v", err)
	}

	// Extract username for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("DROP ROLE failed: %v", err)
	}

	// Extract role name for success message
	parts := strings.Fields(ctx.GetText())
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
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("GRANT failed: %v", err)
	}

	return "GRANT successful"
}

// VisitRevoke handles REVOKE statements
func (v *CqlCommandVisitorImpl) VisitRevoke(ctx *grammar.RevokeContext) interface{} {
	if err := v.session.ExecuteRoleCommand(ctx.GetText()); err != nil {
		return fmt.Errorf("REVOKE failed: %v", err)
	}

	return "REVOKE successful"
}

// VisitListRoles handles LIST ROLES statements
func (v *CqlCommandVisitorImpl) VisitListRoles(ctx *grammar.ListRolesContext) interface{} {
	roles, err := v.session.ListRoles()
	if err != nil {
		return fmt.Errorf("LIST ROLES failed: %v", err)
	}

	if len(roles) == 0 {
		return "No roles found"
	}

	// Prepare headers
	headers := []string{"Role", "Login", "Superuser", "Member Of"}

	// Format results
	results := [][]string{headers}

	for _, role := range roles {
		memberOfStr := "None"
		if len(role.MemberOf) > 0 {
			memberOfStr = strings.Join(role.MemberOf, ", ")
		}

		row := []string{
			role.Role,
			fmt.Sprintf("%v", role.CanLogin),
			fmt.Sprintf("%v", role.IsSuperuser),
			memberOfStr,
		}
		results = append(results, row)
	}

	return results
}

// VisitListPermissions handles LIST PERMISSIONS statements
func (v *CqlCommandVisitorImpl) VisitListPermissions(ctx *grammar.ListPermissionsContext) interface{} {
	// Parse the command to see if it's for a specific role/user
	parts := strings.Fields(strings.ToUpper(ctx.GetText()))

	var specificRole string

	// Check if listing permissions for a specific role
	if len(parts) >= 3 {
		// LIST PERMISSIONS OF <role>
		for i, part := range parts {
			if part == "OF" && i+1 < len(parts) {
				specificRole = strings.Trim(parts[i+1], "'\"")
				break
			}
		}
	}

	// Get permissions from db package
	var perms []db.PermissionInfo
	var err error

	if specificRole != "" {
		perms, err = v.session.ListPermissionsForRole(specificRole)
	} else {
		perms, err = v.session.ListPermissions()
	}

	if err != nil {
		return fmt.Errorf("LIST PERMISSIONS failed: %v", err)
	}

	if len(perms) == 0 {
		return "No permissions found"
	}

	// Prepare headers
	headers := []string{"Role", "Resource", "Permissions"}

	// Format results
	results := [][]string{headers}

	for _, perm := range perms {
		permStr := "None"
		if len(perm.Permissions) > 0 {
			permStr = strings.Join(perm.Permissions, ", ")
		}

		row := []string{
			perm.Role,
			perm.Resource,
			permStr,
		}
		results = append(results, row)
	}

	if len(results) == 1 {
		return "No permissions found"
	}

	return results
}
