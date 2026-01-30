package models

import (
	"time"

	"gorm.io/gorm"
)

// Permission represents a single permission
type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:varchar(255)" json:"description"`
	Resource    string    `gorm:"type:varchar(100);not null;index" json:"resource"` // e.g., "news", "ping", "users"
	Action      string    `gorm:"type:varchar(50);not null;index" json:"action"`    // e.g., "read", "create", "update", "delete"
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Role represents a role that can have multiple permissions
type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	IsSystem    bool         `gorm:"default:false" json:"is_system"` // System roles cannot be deleted
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Role) TableName() string {
	return "roles"
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRole) TableName() string {
	return "user_roles"
}

// SeedDefaultRolesAndPermissions creates default roles and permissions
func SeedDefaultRolesAndPermissions(db *gorm.DB) error {
	// Define default permissions
	permissions := []Permission{
		// Ping permissions
		{Name: "ping:read", Description: "Can ping the API", Resource: "ping", Action: "read"},

		// News permissions (example)
		{Name: "news:read", Description: "Can read news articles", Resource: "news", Action: "read"},
		{Name: "news:create", Description: "Can create news articles", Resource: "news", Action: "create"},
		{Name: "news:update", Description: "Can update news articles", Resource: "news", Action: "update"},
		{Name: "news:delete", Description: "Can delete news articles", Resource: "news", Action: "delete"},

		// User management permissions
		{Name: "users:read", Description: "Can view users", Resource: "users", Action: "read"},
		{Name: "users:update", Description: "Can update users", Resource: "users", Action: "update"},
		{Name: "users:delete", Description: "Can delete users", Resource: "users", Action: "delete"},
		{Name: "users:manage_roles", Description: "Can manage user roles", Resource: "users", Action: "manage_roles"},

		// Role management permissions
		{Name: "roles:read", Description: "Can view roles", Resource: "roles", Action: "read"},
		{Name: "roles:create", Description: "Can create roles", Resource: "roles", Action: "create"},
		{Name: "roles:update", Description: "Can update roles", Resource: "roles", Action: "update"},
		{Name: "roles:delete", Description: "Can delete roles", Resource: "roles", Action: "delete"},
	}

	// Create permissions if they don't exist
	for _, perm := range permissions {
		if err := db.Where("name = ?", perm.Name).FirstOrCreate(&perm).Error; err != nil {
			return err
		}
	}

	// Get all permissions for admin role
	var allPerms []Permission
	if err := db.Find(&allPerms).Error; err != nil {
		return err
	}

	// Get basic permissions for user role
	var userPerms []Permission
	if err := db.Where("name IN ?", []string{"ping:read", "news:read"}).Find(&userPerms).Error; err != nil {
		return err
	}

	// Define default roles
	roles := []struct {
		Role        Role
		Permissions []Permission
	}{
		{
			Role: Role{
				Name:        "admin",
				Description: "Administrator with full access",
				IsSystem:    true,
			},
			Permissions: allPerms,
		},
		{
			Role: Role{
				Name:        "user",
				Description: "Regular user with basic access",
				IsSystem:    true,
			},
			Permissions: userPerms,
		},
		{
			Role: Role{
				Name:        "editor",
				Description: "Editor with news management access",
				IsSystem:    false,
			},
			Permissions: func() []Permission {
				var editorPerms []Permission
				db.Where("name IN ?", []string{"ping:read", "news:read", "news:create", "news:update"}).Find(&editorPerms)
				return editorPerms
			}(),
		},
	}

	// Create roles if they don't exist
	for _, r := range roles {
		var existingRole Role
		result := db.Where("name = ?", r.Role.Name).First(&existingRole)
		if result.Error == gorm.ErrRecordNotFound {
			r.Role.Permissions = r.Permissions
			if err := db.Create(&r.Role).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// GetUserPermissions returns all permissions for a user across all their roles
func GetUserPermissions(db *gorm.DB, userID uint) ([]string, error) {
	var permissions []string

	err := db.Raw(`
		SELECT DISTINCT p.name 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
	`, userID).Scan(&permissions).Error

	return permissions, err
}

// GetUserRoles returns all roles for a user
func GetUserRoles(db *gorm.DB, userID uint) ([]Role, error) {
	var roles []Role

	err := db.Raw(`
		SELECT r.* 
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`, userID).Scan(&roles).Error

	return roles, err
}

// AssignRoleToUser assigns a role to a user
func AssignRoleToUser(db *gorm.DB, userID uint, roleName string) error {
	var role Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	userRole := UserRole{
		UserID: userID,
		RoleID: role.ID,
	}

	return db.Where("user_id = ? AND role_id = ?", userID, role.ID).FirstOrCreate(&userRole).Error
}

// RemoveRoleFromUser removes a role from a user
func RemoveRoleFromUser(db *gorm.DB, userID uint, roleName string) error {
	var role Role
	if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
		return err
	}

	return db.Where("user_id = ? AND role_id = ?", userID, role.ID).Delete(&UserRole{}).Error
}

// HasPermission checks if a user has a specific permission
func HasPermission(db *gorm.DB, userID uint, permissionName string) (bool, error) {
	var count int64

	err := db.Raw(`
		SELECT COUNT(*) 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND p.name = ?
	`, userID, permissionName).Scan(&count).Error

	return count > 0, err
}
