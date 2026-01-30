package admin

import (
	"net/http"
	"strconv"

	"github.com/chattycathy/api/db/models"
	"github.com/chattycathy/api/pkg/logger"
	"github.com/chattycathy/api/pkg/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles admin endpoints for role and permission management
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new admin handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers admin routes (requires admin role)
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.RequireRole("admin"))
	{
		// Permissions
		admin.GET("/permissions", h.ListPermissions)

		// Roles
		admin.GET("/roles", h.ListRoles)
		admin.GET("/roles/:id", h.GetRole)
		admin.POST("/roles", h.CreateRole)
		admin.PUT("/roles/:id", h.UpdateRole)
		admin.DELETE("/roles/:id", h.DeleteRole)

		// Role permissions
		admin.PUT("/roles/:id/permissions", h.SetRolePermissions)
	}
}

// PermissionResponse represents a permission in the API response
type PermissionResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// RoleResponse represents a role in the API response
type RoleResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions"`
}

// ListPermissions returns all available permissions
func (h *Handler) ListPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := h.db.Order("resource, action").Find(&permissions).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to list permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch permissions"})
		return
	}

	response := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		response[i] = PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ListRoles returns all roles with their permissions
func (h *Handler) ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Preload("Permissions").Order("name").Find(&roles).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to list roles")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch roles"})
		return
	}

	response := make([]RoleResponse, len(roles))
	for i, r := range roles {
		permissions := make([]PermissionResponse, len(r.Permissions))
		for j, p := range r.Permissions {
			permissions[j] = PermissionResponse{
				ID:          p.ID,
				Name:        p.Name,
				Description: p.Description,
				Resource:    p.Resource,
				Action:      p.Action,
			}
		}
		response[i] = RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			IsSystem:    r.IsSystem,
			Permissions: permissions,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetRole returns a single role with its permissions
func (h *Handler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to get role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role"})
		return
	}

	permissions := make([]PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
		}
	}

	c.JSON(http.StatusOK, RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: permissions,
	})
}

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
}

// CreateRole creates a new role
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Check if role already exists
	var existing models.Role
	if err := h.db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role already exists"})
		return
	}

	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false, // User-created roles are never system roles
	}

	if err := h.db.Create(&role).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to create role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: []PermissionResponse{},
	})
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
}

// UpdateRole updates an existing role
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to get role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role"})
		return
	}

	// Don't allow renaming system roles
	if role.IsSystem && role.Name != req.Name {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot rename system roles"})
		return
	}

	// Check for name conflict
	var existing models.Role
	if err := h.db.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "role name already exists"})
		return
	}

	role.Name = req.Name
	role.Description = req.Description

	if err := h.db.Save(&role).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to update role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	// Reload with permissions
	h.db.Preload("Permissions").First(&role, id)

	permissions := make([]PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permissions[i] = PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
		}
	}

	c.JSON(http.StatusOK, RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: permissions,
	})
}

// DeleteRole deletes a role
func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to get role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role"})
		return
	}

	// Don't allow deleting system roles
	if role.IsSystem {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete system roles"})
		return
	}

	// Delete role (GORM will handle the role_permissions junction table)
	if err := h.db.Select("Permissions").Delete(&role).Error; err != nil {
		logger.Error().Err(err).Msg("Failed to delete role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}

	// Also delete user_roles associations
	h.db.Where("role_id = ?", id).Delete(&models.UserRole{})

	c.JSON(http.StatusOK, gin.H{"message": "role deleted"})
}

// SetRolePermissionsRequest represents a request to set role permissions
type SetRolePermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// SetRolePermissions sets the permissions for a role
func (h *Handler) SetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role ID"})
		return
	}

	var req SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
			return
		}
		logger.Error().Err(err).Msg("Failed to get role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch role"})
		return
	}

	// Fetch the requested permissions
	var permissions []models.Permission
	if len(req.PermissionIDs) > 0 {
		if err := h.db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			logger.Error().Err(err).Msg("Failed to fetch permissions")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch permissions"})
			return
		}
	}

	// Replace the role's permissions
	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		logger.Error().Err(err).Msg("Failed to update role permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role permissions"})
		return
	}

	// Return updated role
	h.db.Preload("Permissions").First(&role, id)

	permResponse := make([]PermissionResponse, len(role.Permissions))
	for i, p := range role.Permissions {
		permResponse[i] = PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Resource:    p.Resource,
			Action:      p.Action,
		}
	}

	logger.Info().
		Uint("role_id", uint(id)).
		Str("role_name", role.Name).
		Int("permission_count", len(permissions)).
		Msg("Role permissions updated")

	c.JSON(http.StatusOK, RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: permResponse,
	})
}
