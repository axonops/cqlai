package db

import (
	"fmt"
)

// RoleInfo holds role information from system_auth.roles
type RoleInfo struct {
	Role        string
	CanLogin    bool
	IsSuperuser bool
	MemberOf    []string
	SaltedHash  string
}

// PermissionInfo holds permission information from system_auth.role_permissions
type PermissionInfo struct {
	Role        string
	Resource    string
	Permissions []string
}

// ExecuteRoleCommand executes a role-related command (CREATE/ALTER/DROP USER/ROLE, GRANT/REVOKE)
func (s *Session) ExecuteRoleCommand(query string) error {
	return s.Query(query).Exec()
}

// ListRoles queries system_auth.roles for role information
func (s *Session) ListRoles() ([]RoleInfo, error) {
	query := "SELECT role, can_login, is_superuser, member_of, salted_hash FROM system_auth.roles"
	
	iter := s.Query(query).Iter()
	defer iter.Close()
	
	var roles []RoleInfo
	var role string
	var canLogin, isSuperuser bool
	var memberOf []string
	var saltedHash string
	
	for iter.Scan(&role, &canLogin, &isSuperuser, &memberOf, &saltedHash) {
		roles = append(roles, RoleInfo{
			Role:        role,
			CanLogin:    canLogin,
			IsSuperuser: isSuperuser,
			MemberOf:    memberOf,
			SaltedHash:  saltedHash,
		})
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to list roles: %v", err)
	}
	
	return roles, nil
}

// ListPermissions queries system_auth.role_permissions for all permissions
func (s *Session) ListPermissions() ([]PermissionInfo, error) {
	query := "SELECT role, resource, permissions FROM system_auth.role_permissions"
	
	iter := s.Query(query).Iter()
	defer iter.Close()
	
	var permissions []PermissionInfo
	var role, resource string
	var perms []string
	
	for iter.Scan(&role, &resource, &perms) {
		permissions = append(permissions, PermissionInfo{
			Role:        role,
			Resource:    resource,
			Permissions: perms,
		})
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to list permissions: %v", err)
	}
	
	return permissions, nil
}

// ListPermissionsForRole queries system_auth.role_permissions for a specific role
func (s *Session) ListPermissionsForRole(roleName string) ([]PermissionInfo, error) {
	query := fmt.Sprintf("SELECT role, resource, permissions FROM system_auth.role_permissions WHERE role = '%s'", roleName)
	
	iter := s.Query(query).Iter()
	defer iter.Close()
	
	var permissions []PermissionInfo
	var role, resource string
	var perms []string
	
	for iter.Scan(&role, &resource, &perms) {
		permissions = append(permissions, PermissionInfo{
			Role:        role,
			Resource:    resource,
			Permissions: perms,
		})
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to list permissions for role %s: %v", roleName, err)
	}
	
	return permissions, nil
}