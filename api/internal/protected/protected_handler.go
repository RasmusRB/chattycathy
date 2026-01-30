package protected

import (
	"net/http"

	"github.com/chattycathy/api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

// Handler handles protected endpoints
type Handler struct{}

// NewHandler creates a new protected handler
func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers protected routes (requires JWT)
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// All routes in this group require valid JWT
	protected := router.Group("/protected")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/secret", h.Secret)
		protected.GET("/profile", h.Profile)

		// Admin-only routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/dashboard", h.AdminDashboard)
		}
	}
}

// Secret is a protected endpoint that requires authentication
func (h *Handler) Secret(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "This is a secret message!",
		"hint":    "You can only see this because you're authenticated",
		"user":    claims.Username,
	})
}

// Profile returns the authenticated user's profile
func (h *Handler) Profile(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"permissions": []string{
			"read:own-data",
			"write:own-data",
		},
	})
}

// AdminDashboard is an admin-only endpoint
func (h *Handler) AdminDashboard(c *gin.Context) {
	claims, _ := middleware.GetClaims(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the admin dashboard!",
		"admin":   claims.Username,
		"stats": gin.H{
			"total_users":     42,
			"active_sessions": 7,
			"api_calls_today": 1337,
		},
	})
}
